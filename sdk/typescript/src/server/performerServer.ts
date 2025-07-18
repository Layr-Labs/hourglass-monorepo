// Core PerformerServer implementation
// Based on the Go PonosPerformer in ponos/pkg/performer/server/server.go

import * as grpc from '@grpc/grpc-js';
import * as protoLoader from '@grpc/proto-loader';
import path from 'path';
import {
  PerformerServerConfig,
  PerformerService,
  TaskRequest,
  TaskResponse,
  HealthCheckRequest,
  HealthCheckResponse,
  StartSyncRequest,
  StartSyncResponse,
  PerformerStatus,
  TaskValidationError,
  TaskExecutionError,
} from '../types/performer';
import { IWorker } from '../worker/iWorker';
import { createLogger, ILogger } from '../utils/logger';
import { TaskProcessor, TaskProcessorConfig } from './taskProcessor';
import { HealthManager, HealthCheckConfig } from './healthManager';
import { MetricsCollector, MetricsConfig } from '../utils/metrics';
import { DiagnosticTool } from '../utils/diagnostics';
import { resolveProtoPath } from '../utils/protoResolver';

/**
 * PerformerServer class that implements the gRPC PerformerService
 * This mirrors the Go PonosPerformer structure
 */
export class PerformerServer {
  private server: grpc.Server;
  private config: Required<PerformerServerConfig>;
  private worker: IWorker;
  private logger: ILogger;
  private taskProcessor: TaskProcessor;
  private healthManager: HealthManager;
  private metricsCollector: MetricsCollector;
  private diagnosticTool: DiagnosticTool;
  private isRunning: boolean = false;

  constructor(
    worker: IWorker, 
    config: PerformerServerConfig = {}, 
    taskProcessorConfig?: TaskProcessorConfig,
    healthConfig?: HealthCheckConfig,
    metricsConfig?: MetricsConfig
  ) {
    this.worker = worker;
    this.config = {
      port: config.port ?? 8080,
      timeout: config.timeout ?? 5000,
      debug: config.debug ?? false,
    };
    this.logger = createLogger({ debug: this.config.debug });
    
    // Initialize health manager
    this.healthManager = new HealthManager(this.logger, {
      checkInterval: 30000,
      maxFailures: 3,
      ...healthConfig,
    });

    // Initialize metrics collector
    this.metricsCollector = new MetricsCollector(this.logger, {
      enabled: true,
      defaultLabels: { service: 'hourglass-performer' },
      ...metricsConfig,
    });

    // Initialize diagnostic tool
    this.diagnosticTool = new DiagnosticTool(this.logger, {
      enabled: this.config.debug,
      enablePerformanceMonitoring: true,
      enableMemoryLeakDetection: true,
    });
    this.diagnosticTool.setMetricsCollector(this.metricsCollector);
    
    // Initialize task processor with enhanced capabilities
    this.taskProcessor = new TaskProcessor(worker, this.logger, {
      timeout: this.config.timeout,
      enableMetrics: true,
      ...taskProcessorConfig,
    });
    
    this.server = new grpc.Server();
    this.setupGrpcServer();
  }

  /**
   * Set up the gRPC server with the PerformerService
   */
  private setupGrpcServer(): void {
    // Load protobuf definition using the robust path resolver
    const protoPath = resolveProtoPath('performer.proto');
    
    const packageDefinition = protoLoader.loadSync(protoPath, {
      keepCase: true,
      longs: String,
      enums: String,
      defaults: true,
      oneofs: true,
    });

    const protoDescriptor = grpc.loadPackageDefinition(packageDefinition) as any;
    const performerService = protoDescriptor.eigenlayer.hourglass.v1.performer.PerformerService;

    // Register service implementation
    this.server.addService(performerService.service, {
      ExecuteTask: this.executeTask.bind(this),
      HealthCheck: this.healthCheck.bind(this),
      StartSync: this.startSync.bind(this),
    });
  }

  /**
   * Start the gRPC server
   */
  async start(): Promise<void> {
    if (this.isRunning) {
      this.logger.warn('Server is already running');
      return;
    }

    // Start monitoring services
    this.healthManager.start();
    this.metricsCollector.start();
    this.diagnosticTool.start();

    return new Promise((resolve, reject) => {
      const bindAddress = `0.0.0.0:${this.config.port}`;
      
      this.server.bindAsync(
        bindAddress,
        grpc.ServerCredentials.createInsecure(),
        (error, port) => {
          if (error) {
            this.logger.error('Failed to bind server', { error: error.message });
            this.healthManager.setStatus(PerformerStatus.ERROR);
            reject(error);
            return;
          }

          this.server.start();
          this.isRunning = true;
          this.healthManager.setStatus(PerformerStatus.READY_FOR_TASK);
          this.logger.info(`Performer server started on port ${port}`);
          this.metricsCollector.counter('server_starts_total');
          resolve();
        }
      );
    });
  }

  /**
   * Stop the gRPC server gracefully
   */
  async stop(): Promise<void> {
    if (!this.isRunning) {
      this.logger.warn('Server is not running');
      return;
    }

    this.healthManager.setStatus(PerformerStatus.STOPPING);

    return new Promise((resolve) => {
      this.logger.info('Shutting down server...');
      
      this.server.tryShutdown((error) => {
        if (error) {
          this.logger.error('Error during shutdown', { error: error.message });
          // Force shutdown
          this.server.forceShutdown();
        } else {
          this.logger.info('Server shutdown complete');
        }
        
        // Stop monitoring services
        this.healthManager.stop();
        this.metricsCollector.stop();
        this.diagnosticTool.stop();
        
        this.isRunning = false;
        resolve();
      });
    });
  }

  /**
   * Execute a task (gRPC handler)
   */
  async executeTask(
    call: grpc.ServerUnaryCall<TaskRequest, TaskResponse>,
    callback: grpc.sendUnaryData<TaskResponse>
  ): Promise<void> {
    const request = call.request;
    const timer = this.metricsCollector.timer('task_execution');
    
    this.logger.info('Received task', { taskId: request.taskId });
    this.metricsCollector.counter('tasks_received_total');
    this.healthManager.recordTask();

    // Start diagnostic trace if enabled
    this.diagnosticTool.startTrace(request.taskId, 'execute_task', {
      payloadSize: request.payload.length,
    });

    try {
      this.healthManager.setStatus(PerformerStatus.BUSY);
      
      // Use the enhanced task processor
      const response = await this.taskProcessor.processTask(request);
      
      timer(); // Record execution time
      this.metricsCollector.counter('tasks_completed_total');
      this.metricsCollector.histogram('task_payload_size_bytes', request.payload.length);
      this.metricsCollector.histogram('task_result_size_bytes', response.result.length);
      
      this.diagnosticTool.endTrace(request.taskId, { success: true });
      this.healthManager.setStatus(PerformerStatus.READY_FOR_TASK);
      
      callback(null, response);
    } catch (error) {
      timer(); // Record execution time even on failure
      this.metricsCollector.counter('tasks_failed_total');
      
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      this.logger.error('Task processing failed', { 
        taskId: request.taskId,
        error: errorMessage
      });

      this.healthManager.recordError(`Task ${request.taskId}: ${errorMessage}`);
      this.diagnosticTool.recordError(errorMessage);
      this.diagnosticTool.endTrace(request.taskId, { success: false, error: errorMessage });
      
      // Convert to gRPC error
      const grpcError = this.createGrpcError(error, request.taskId);
      callback(grpcError);
    }
  }

  /**
   * Get task execution metrics
   */
  getTaskMetrics() {
    return this.taskProcessor.getMetrics();
  }

  /**
   * Clear accumulated task metrics
   */
  clearTaskMetrics() {
    this.taskProcessor.clearMetrics();
  }

  /**
   * Add custom serialization strategy
   */
  addSerializationStrategy(strategy: any) {
    this.taskProcessor.addSerializationStrategy(strategy);
  }

  /**
   * Health check (gRPC handler)
   */
  async healthCheck(
    call: grpc.ServerUnaryCall<HealthCheckRequest, HealthCheckResponse>,
    callback: grpc.sendUnaryData<HealthCheckResponse>
  ): Promise<void> {
    try {
      // Perform comprehensive health check
      const healthResult = await this.healthManager.performHealthCheck();
      const currentStatus = this.healthManager.getStatus();
      
      this.metricsCollector.counter('health_checks_total');
      this.metricsCollector.gauge('health_status', currentStatus === PerformerStatus.READY_FOR_TASK ? 1 : 0);
      
      const response: HealthCheckResponse = {
        status: currentStatus,
      };

      callback(null, response);
    } catch (error) {
      this.logger.error('Health check failed', { 
        error: error instanceof Error ? error.message : 'Unknown error'
      });
      
      const response: HealthCheckResponse = {
        status: PerformerStatus.ERROR,
      };

      callback(null, response);
    }
  }

  /**
   * Start sync (gRPC handler)
   */
  async startSync(
    call: grpc.ServerUnaryCall<StartSyncRequest, StartSyncResponse>,
    callback: grpc.sendUnaryData<StartSyncResponse>
  ): Promise<void> {
    const response: StartSyncResponse = {};
    callback(null, response);
  }

  /**
   * Convert application errors to gRPC errors
   */
  private createGrpcError(error: any, taskId?: string): grpc.ServiceError {
    let code = grpc.status.INTERNAL;
    let message = 'Internal server error';

    if (error instanceof TaskValidationError) {
      code = grpc.status.INVALID_ARGUMENT;
      message = `Task validation failed: ${error.message}`;
    } else if (error instanceof TaskExecutionError) {
      code = grpc.status.INTERNAL;
      message = `Task execution failed: ${error.message}`;
    } else if (error instanceof Error) {
      message = error.message;
    }

    const grpcError: grpc.ServiceError = {
      name: 'ServiceError',
      message,
      code,
      details: taskId ? `Task ID: ${taskId}` : '',
      metadata: new grpc.Metadata(),
    };

    return grpcError;
  }

  /**
   * PerformerService interface implementation
   */
  async ExecuteTask(request: TaskRequest): Promise<TaskResponse> {
    return new Promise((resolve, reject) => {
      const call = {} as grpc.ServerUnaryCall<TaskRequest, TaskResponse>;
      call.request = request;
      
      this.executeTask(call, (error, response) => {
        if (error) {
          reject(error);
        } else {
          resolve(response!);
        }
      });
    });
  }

  async HealthCheck(request: HealthCheckRequest): Promise<HealthCheckResponse> {
    return new Promise((resolve, reject) => {
      const call = {} as grpc.ServerUnaryCall<HealthCheckRequest, HealthCheckResponse>;
      call.request = request;
      
      this.healthCheck(call, (error, response) => {
        if (error) {
          reject(error);
        } else {
          resolve(response!);
        }
      });
    });
  }

  async StartSync(request: StartSyncRequest): Promise<StartSyncResponse> {
    return new Promise((resolve, reject) => {
      const call = {} as grpc.ServerUnaryCall<StartSyncRequest, StartSyncResponse>;
      call.request = request;
      
      this.startSync(call, (error, response) => {
        if (error) {
          reject(error);
        } else {
          resolve(response!);
        }
      });
    });
  }

  /**
   * Set up graceful shutdown handlers
   */
  setupGracefulShutdown(): void {
    const shutdown = async (signal: string) => {
      this.logger.info(`Received ${signal}, shutting down gracefully...`);
      await this.stop();
      process.exit(0);
    };

    process.on('SIGTERM', () => shutdown('SIGTERM'));
    process.on('SIGINT', () => shutdown('SIGINT'));
  }

  /**
   * Get server status
   */
  getStatus(): { isRunning: boolean; config: Required<PerformerServerConfig> } {
    return {
      isRunning: this.isRunning,
      config: this.config,
    };
  }

  /**
   * Get comprehensive health information
   */
  async getHealthInfo() {
    return await this.healthManager.performHealthCheck();
  }

  /**
   * Get health summary
   */
  getHealthSummary() {
    return this.healthManager.getHealthSummary();
  }

  /**
   * Get metrics summary
   */
  getMetricsSummary() {
    return this.metricsCollector.getSummary();
  }

  /**
   * Get diagnostic report
   */
  getDiagnosticReport(): string {
    return this.diagnosticTool.generateReport(this.healthManager.getStatus(), this.config);
  }

  /**
   * Add custom health check provider
   */
  addHealthProvider(provider: any) {
    this.healthManager.addProvider(provider);
  }

  /**
   * Add custom metrics exporter
   */
  addMetricsExporter(exporter: any) {
    this.metricsCollector.addExporter(exporter);
  }

  /**
   * Record custom metric
   */
  recordMetric(name: string, value: number, labels?: Record<string, string>) {
    this.metricsCollector.gauge(name, value, labels);
  }

  /**
   * Get active diagnostic traces
   */
  getActiveTraces() {
    return this.diagnosticTool.getActiveTraces();
  }
}
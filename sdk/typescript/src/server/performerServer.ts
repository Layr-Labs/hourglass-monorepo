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

/**
 * PerformerServer class that implements the gRPC PerformerService
 * This mirrors the Go PonosPerformer structure
 */
export class PerformerServer implements PerformerService {
  private server: grpc.Server;
  private config: Required<PerformerServerConfig>;
  private worker: IWorker;
  private logger: ILogger;
  private taskProcessor: TaskProcessor;
  private isRunning: boolean = false;

  constructor(worker: IWorker, config: PerformerServerConfig = {}, taskProcessorConfig?: TaskProcessorConfig) {
    this.worker = worker;
    this.config = {
      port: config.port ?? 8080,
      timeout: config.timeout ?? 5000,
      debug: config.debug ?? false,
    };
    this.logger = createLogger({ debug: this.config.debug });
    
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
    // Load protobuf definition
    const protoPath = path.join(__dirname, '../../proto/performer.proto');
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

    return new Promise((resolve, reject) => {
      const bindAddress = `0.0.0.0:${this.config.port}`;
      
      this.server.bindAsync(
        bindAddress,
        grpc.ServerCredentials.createInsecure(),
        (error, port) => {
          if (error) {
            this.logger.error('Failed to bind server', { error: error.message });
            reject(error);
            return;
          }

          this.server.start();
          this.isRunning = true;
          this.logger.info(`Performer server started on port ${port}`);
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
    
    this.logger.info('Received task', { taskId: request.taskId });

    try {
      // Use the enhanced task processor
      const response = await this.taskProcessor.processTask(request);
      callback(null, response);
    } catch (error) {
      this.logger.error('Task processing failed', { 
        taskId: request.taskId,
        error: error instanceof Error ? error.message : 'Unknown error'
      });

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
    const response: HealthCheckResponse = {
      status: PerformerStatus.READY_FOR_TASK,
    };

    callback(null, response);
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
      details: taskId ? `Task ID: ${taskId}` : undefined,
    };

    return grpcError;
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
}
// Enhanced task processing framework with serialization and validation

import { TaskRequest, TaskResponse, TaskValidationError, TaskExecutionError } from '../types/performer';
import { IWorker } from '../worker/iWorker';
import { TaskContext, TaskContextBuilder, TaskPipeline, ValidationStage, TaskMetrics } from './taskContext';
import { ILogger } from '../utils/logger';

/**
 * Serialization strategy interface
 */
export interface SerializationStrategy {
  /** Strategy name */
  name: string;
  /** Serialize data to bytes */
  serialize(data: any): Uint8Array;
  /** Deserialize bytes to data */
  deserialize<T = any>(bytes: Uint8Array): T;
  /** Validate if bytes can be deserialized */
  canDeserialize(bytes: Uint8Array): boolean;
}

/**
 * JSON serialization strategy
 */
export class JsonSerializationStrategy implements SerializationStrategy {
  name = 'json';

  serialize(data: any): Uint8Array {
    const json = JSON.stringify(data);
    return new TextEncoder().encode(json);
  }

  deserialize<T = any>(bytes: Uint8Array): T {
    const json = new TextDecoder().decode(bytes);
    return JSON.parse(json);
  }

  canDeserialize(bytes: Uint8Array): boolean {
    try {
      const json = new TextDecoder().decode(bytes);
      JSON.parse(json);
      return true;
    } catch {
      return false;
    }
  }
}

/**
 * Raw bytes serialization strategy
 */
export class RawSerializationStrategy implements SerializationStrategy {
  name = 'raw';

  serialize(data: any): Uint8Array {
    if (data instanceof Uint8Array) {
      return data;
    }
    if (typeof data === 'string') {
      return new TextEncoder().encode(data);
    }
    throw new Error('Raw serialization only supports Uint8Array or string');
  }

  deserialize<T = any>(bytes: Uint8Array): T {
    return bytes as T;
  }

  canDeserialize(bytes: Uint8Array): boolean {
    return true;
  }
}

/**
 * Number serialization strategy (64-bit big-endian)
 */
export class NumberSerializationStrategy implements SerializationStrategy {
  name = 'number';

  serialize(data: any): Uint8Array {
    const num = Number(data);
    if (isNaN(num)) {
      throw new Error('Invalid number for serialization');
    }
    
    const bytes = new Uint8Array(8);
    const view = new DataView(bytes.buffer);
    view.setBigUint64(0, BigInt(num), false); // big-endian
    return bytes;
  }

  deserialize<T = any>(bytes: Uint8Array): T {
    if (bytes.length !== 8) {
      throw new Error('Number deserialization requires exactly 8 bytes');
    }
    
    const view = new DataView(bytes.buffer);
    return Number(view.getBigUint64(0, false)) as T; // big-endian
  }

  canDeserialize(bytes: Uint8Array): boolean {
    return bytes.length === 8;
  }
}

/**
 * Task processing configuration
 */
export interface TaskProcessorConfig {
  /** Default timeout in milliseconds */
  timeout?: number;
  /** Enable metrics collection */
  enableMetrics?: boolean;
  /** Custom validation stages */
  validationStages?: ValidationStage[];
  /** Default serialization strategy */
  defaultSerialization?: SerializationStrategy;
  /** Available serialization strategies */
  serializationStrategies?: Map<string, SerializationStrategy>;
}

/**
 * Enhanced task processor with pipeline support
 */
export class TaskProcessor {
  private config: Required<TaskProcessorConfig>;
  private pipeline: TaskPipeline;
  private metrics: TaskMetrics[] = [];

  constructor(
    private worker: IWorker,
    private logger: ILogger,
    config: TaskProcessorConfig = {}
  ) {
    // Set up default configuration
    this.config = {
      timeout: config.timeout ?? 5000,
      enableMetrics: config.enableMetrics ?? true,
      validationStages: config.validationStages ?? [],
      defaultSerialization: config.defaultSerialization ?? new JsonSerializationStrategy(),
      serializationStrategies: config.serializationStrategies ?? new Map([
        ['json', new JsonSerializationStrategy()],
        ['raw', new RawSerializationStrategy()],
        ['number', new NumberSerializationStrategy()],
      ]),
    };

    this.setupPipeline();
  }

  /**
   * Process a task with full pipeline support
   */
  async processTask(request: TaskRequest): Promise<TaskResponse> {
    const contextBuilder = new TaskContextBuilder(request, this.logger, this.config.timeout);
    const context = contextBuilder.build();

    try {
      // Execute the processing pipeline
      await this.pipeline.execute(context);

      // Execute the worker task
      const response = await this.executeWorkerTask(context);

      // Record success metrics
      if (this.config.enableMetrics) {
        this.recordTaskCompletion(context, response);
      }

      return response;
    } catch (error) {
      // Record failure metrics
      if (this.config.enableMetrics) {
        this.recordTaskFailure(context, error);
      }

      // Re-throw the error for gRPC handling
      throw error;
    }
  }

  /**
   * Execute the worker task with enhanced error handling
   */
  private async executeWorkerTask(context: TaskContext): Promise<TaskResponse> {
    const { request } = context;

    try {
      // Validate task using worker
      await this.worker.validateTask(request);
      context.logger.debug('Task validation passed');

      // Handle task using worker
      const response = await this.worker.handleTask(request);
      context.logger.debug('Task execution completed', { 
        resultSize: response.result.length 
      });

      // Validate response
      this.validateResponse(response, request);

      return response;
    } catch (error) {
      if (error instanceof TaskValidationError || error instanceof TaskExecutionError) {
        throw error;
      }
      
      // Wrap unknown errors
      throw new TaskExecutionError(
        error instanceof Error ? error.message : 'Unknown task execution error',
        request.taskId
      );
    }
  }

  /**
   * Validate task response
   */
  private validateResponse(response: TaskResponse, request: TaskRequest): void {
    if (response.taskId !== request.taskId) {
      throw new TaskExecutionError(
        `Response task ID (${response.taskId}) does not match request task ID (${request.taskId})`,
        request.taskId
      );
    }

    if (!response.result) {
      throw new TaskExecutionError(
        'Task response must include a result',
        request.taskId
      );
    }
  }

  /**
   * Set up the processing pipeline
   */
  private setupPipeline(): void {
    this.pipeline = new TaskPipeline();

    // Add built-in validation stages
    this.pipeline.addStage(new ValidationStage(async (context) => {
      this.validateTaskRequest(context.request);
    }));

    // Add custom validation stages
    for (const stage of this.config.validationStages) {
      this.pipeline.addStage(stage);
    }

    // Add metrics collection stage if enabled
    if (this.config.enableMetrics) {
      this.pipeline.addStage({
        name: 'metrics-start',
        execute: async (context) => {
          const metrics: TaskMetrics = {
            taskId: context.request.taskId,
            executionId: context.executionId,
            startTime: context.startTime,
            payloadSize: context.request.payload.length,
          };
          context.metadata.set('metrics', metrics);
        },
      });
    }
  }

  /**
   * Validate task request
   */
  private validateTaskRequest(request: TaskRequest): void {
    if (!request.taskId) {
      throw new TaskValidationError('Task ID is required', '');
    }

    if (!request.payload) {
      throw new TaskValidationError('Task payload is required', request.taskId);
    }

    if (request.payload.length === 0) {
      throw new TaskValidationError('Task payload cannot be empty', request.taskId);
    }
  }

  /**
   * Record successful task completion metrics
   */
  private recordTaskCompletion(context: TaskContext, response: TaskResponse): void {
    const metrics = context.metadata.get('metrics') as TaskMetrics;
    if (metrics) {
      metrics.endTime = Date.now();
      metrics.duration = metrics.endTime - metrics.startTime;
      metrics.success = true;
      metrics.resultSize = response.result.length;
      
      this.metrics.push(metrics);
      context.logger.info('Task completed successfully', {
        duration: `${metrics.duration}ms`,
        payloadSize: metrics.payloadSize,
        resultSize: metrics.resultSize,
      });
    }
  }

  /**
   * Record task failure metrics
   */
  private recordTaskFailure(context: TaskContext, error: any): void {
    const metrics = context.metadata.get('metrics') as TaskMetrics;
    if (metrics) {
      metrics.endTime = Date.now();
      metrics.duration = metrics.endTime - metrics.startTime;
      metrics.success = false;
      metrics.error = error instanceof Error ? error.message : 'Unknown error';
      
      this.metrics.push(metrics);
      context.logger.error('Task execution failed', {
        duration: `${metrics.duration}ms`,
        error: metrics.error,
      });
    }
  }

  /**
   * Get serialization strategy by name
   */
  getSerializationStrategy(name: string): SerializationStrategy | undefined {
    return this.config.serializationStrategies.get(name);
  }

  /**
   * Add custom serialization strategy
   */
  addSerializationStrategy(strategy: SerializationStrategy): void {
    this.config.serializationStrategies.set(strategy.name, strategy);
  }

  /**
   * Get task execution metrics
   */
  getMetrics(): TaskMetrics[] {
    return [...this.metrics];
  }

  /**
   * Clear accumulated metrics
   */
  clearMetrics(): void {
    this.metrics = [];
  }
}
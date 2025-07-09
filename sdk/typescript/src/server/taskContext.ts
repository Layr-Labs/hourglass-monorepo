// Task context and metadata handling for the performer framework

import { TaskRequest } from '../types/performer';
import { ILogger } from '../utils/logger';

/**
 * Task execution context containing metadata and utilities
 */
export interface TaskContext {
  /** The original task request */
  readonly request: TaskRequest;
  /** Task start timestamp */
  readonly startTime: number;
  /** Unique execution ID for this task run */
  readonly executionId: string;
  /** Logger instance for this task */
  readonly logger: ILogger;
  /** Cancellation signal */
  readonly signal: AbortSignal;
  /** Custom metadata storage */
  readonly metadata: Map<string, any>;
}

/**
 * Task execution result with metadata
 */
export interface TaskExecutionResult {
  /** The task response */
  result: Uint8Array;
  /** Execution duration in milliseconds */
  duration: number;
  /** Success status */
  success: boolean;
  /** Error message if failed */
  error?: string;
  /** Custom metadata */
  metadata?: Record<string, any>;
}

/**
 * Task processing pipeline stage
 */
export interface TaskProcessingStage {
  /** Stage name for logging */
  name: string;
  /** Execute the stage */
  execute(context: TaskContext): Promise<void>;
}

/**
 * Task execution metrics
 */
export interface TaskMetrics {
  /** Task ID */
  taskId: string;
  /** Execution ID */
  executionId: string;
  /** Start timestamp */
  startTime: number;
  /** End timestamp */
  endTime?: number;
  /** Duration in milliseconds */
  duration?: number;
  /** Success status */
  success?: boolean;
  /** Error message */
  error?: string;
  /** Payload size in bytes */
  payloadSize: number;
  /** Result size in bytes */
  resultSize?: number;
  /** Custom metrics */
  customMetrics?: Record<string, number | string>;
}

/**
 * Task context builder for creating task execution contexts
 */
export class TaskContextBuilder {
  private request: TaskRequest;
  private logger: ILogger;
  private timeout: number;
  private abortController: AbortController;

  constructor(request: TaskRequest, logger: ILogger, timeout: number = 5000) {
    this.request = request;
    this.logger = logger;
    this.timeout = timeout;
    this.abortController = new AbortController();
  }

  /**
   * Build the task context
   */
  build(): TaskContext {
    const executionId = this.generateExecutionId();
    const startTime = Date.now();

    // Set up timeout
    setTimeout(() => {
      this.abortController.abort();
    }, this.timeout);

    return {
      request: this.request,
      startTime,
      executionId,
      logger: this.createTaskLogger(executionId),
      signal: this.abortController.signal,
      metadata: new Map<string, any>(),
    };
  }

  /**
   * Cancel the task execution
   */
  cancel(): void {
    this.abortController.abort();
  }

  /**
   * Generate a unique execution ID
   */
  private generateExecutionId(): string {
    const timestamp = Date.now().toString(36);
    const random = Math.random().toString(36).substr(2, 5);
    return `${this.request.taskId}-${timestamp}-${random}`;
  }

  /**
   * Create a task-specific logger
   */
  private createTaskLogger(executionId: string): ILogger {
    const taskMeta = {
      taskId: this.request.taskId,
      executionId,
    };

    return {
      error: (message: string, meta?: any) => 
        this.logger.error(message, { ...taskMeta, ...meta }),
      warn: (message: string, meta?: any) => 
        this.logger.warn(message, { ...taskMeta, ...meta }),
      info: (message: string, meta?: any) => 
        this.logger.info(message, { ...taskMeta, ...meta }),
      debug: (message: string, meta?: any) => 
        this.logger.debug(message, { ...taskMeta, ...meta }),
    };
  }
}

/**
 * Task execution pipeline for processing tasks through multiple stages
 */
export class TaskPipeline {
  private stages: TaskProcessingStage[] = [];

  /**
   * Add a processing stage to the pipeline
   */
  addStage(stage: TaskProcessingStage): this {
    this.stages.push(stage);
    return this;
  }

  /**
   * Execute all stages in the pipeline
   */
  async execute(context: TaskContext): Promise<void> {
    for (const stage of this.stages) {
      if (context.signal.aborted) {
        throw new Error('Task execution was cancelled');
      }

      context.logger.debug(`Executing stage: ${stage.name}`);
      const stageStart = Date.now();
      
      try {
        await stage.execute(context);
        const stageDuration = Date.now() - stageStart;
        context.logger.debug(`Stage completed: ${stage.name}`, { duration: `${stageDuration}ms` });
      } catch (error) {
        const stageDuration = Date.now() - stageStart;
        context.logger.error(`Stage failed: ${stage.name}`, { 
          duration: `${stageDuration}ms`,
          error: error instanceof Error ? error.message : 'Unknown error'
        });
        throw error;
      }
    }
  }
}

/**
 * Built-in processing stages
 */
export class ValidationStage implements TaskProcessingStage {
  name = 'validation';
  
  constructor(private validator: (context: TaskContext) => Promise<void>) {}

  async execute(context: TaskContext): Promise<void> {
    await this.validator(context);
  }
}

export class MetricsStage implements TaskProcessingStage {
  name = 'metrics';
  
  constructor(private metricsCollector: (metrics: TaskMetrics) => void) {}

  async execute(context: TaskContext): Promise<void> {
    const metrics: TaskMetrics = {
      taskId: context.request.taskId,
      executionId: context.executionId,
      startTime: context.startTime,
      payloadSize: context.request.payload.length,
    };

    context.metadata.set('metrics', metrics);
    this.metricsCollector(metrics);
  }
}
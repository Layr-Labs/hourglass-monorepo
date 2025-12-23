// TypeScript types for Hourglass Performer gRPC interface
// Based on github.com/Layr-Labs/protocol-apis/gen/protos/eigenlayer/hourglass/v1/performer

/**
 * Task request from the executor to the performer
 */
export interface TaskRequest {
  /** Unique identifier for the task */
  taskId: string;
  /** Task payload as bytes */
  payload: Uint8Array;
}

/**
 * Task response from the performer back to the executor
 */
export interface TaskResponse {
  /** Unique identifier for the task (must match request) */
  taskId: string;
  /** Task result as bytes */
  result: Uint8Array;
}

/**
 * Health check request
 */
export interface HealthCheckRequest {
  // Empty request
}

/**
 * Performer status enumeration
 */
export enum PerformerStatus {
  READY_FOR_TASK = 'READY_FOR_TASK',
  BUSY = 'BUSY',
  ERROR = 'ERROR',
  STARTING = 'STARTING',
  STOPPING = 'STOPPING',
}

/**
 * Health check response
 */
export interface HealthCheckResponse {
  /** Current performer status */
  status: PerformerStatus;
}

/**
 * Start sync request
 */
export interface StartSyncRequest {
  // Empty request
}

/**
 * Start sync response
 */
export interface StartSyncResponse {
  // Empty response
}

/**
 * PerformerService interface that must be implemented by performers
 */
export interface PerformerService {
  /**
   * Execute a task
   * @param request Task request with payload
   * @returns Task response with result
   */
  ExecuteTask(request: TaskRequest): Promise<TaskResponse>;

  /**
   * Check performer health
   * @param request Health check request
   * @returns Health check response with status
   */
  HealthCheck(request: HealthCheckRequest): Promise<HealthCheckResponse>;

  /**
   * Start synchronization
   * @param request Start sync request
   * @returns Start sync response
   */
  StartSync(request: StartSyncRequest): Promise<StartSyncResponse>;
}

/**
 * Configuration for the performer server
 */
export interface PerformerServerConfig {
  /** Port to listen on (default: 8080) */
  port?: number;
  /** Timeout for task execution in milliseconds (default: 5000) */
  timeout?: number;
  /** Enable debug logging (default: false) */
  debug?: boolean;
}

/**
 * Task execution result for internal use
 */
export interface TaskResult {
  /** Task ID */
  taskId: string;
  /** Result payload */
  result: Uint8Array;
}

/**
 * Task validation error
 */
export class TaskValidationError extends Error {
  constructor(message: string, public readonly taskId: string) {
    super(message);
    this.name = 'TaskValidationError';
  }
}

/**
 * Task execution error
 */
export class TaskExecutionError extends Error {
  constructor(message: string, public readonly taskId: string) {
    super(message);
    this.name = 'TaskExecutionError';
  }
}
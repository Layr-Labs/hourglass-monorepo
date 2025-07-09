// Worker interface for user implementation
// Based on the Go IWorker interface in ponos/pkg/performer/worker/worker.go

import { TaskRequest, TaskResponse } from '../types/performer';

/**
 * IWorker interface that users must implement to handle tasks
 * This mirrors the Go IWorker interface pattern
 */
export interface IWorker {
  /**
   * Validate a task before execution
   * @param task The task request to validate
   * @throws TaskValidationError if task is invalid
   */
  validateTask(task: TaskRequest): Promise<void>;

  /**
   * Handle and execute a task
   * @param task The task request to handle
   * @returns The task response with result
   * @throws TaskExecutionError if task execution fails
   */
  handleTask(task: TaskRequest): Promise<TaskResponse>;
}

/**
 * Abstract base class for Workers that provides common functionality
 * Users can extend this class for convenience
 */
export abstract class BaseWorker implements IWorker {
  /**
   * Default validation - can be overridden by subclasses
   * @param task The task request to validate
   */
  async validateTask(task: TaskRequest): Promise<void> {
    if (!task.taskId) {
      throw new Error('Task ID is required');
    }
    if (!task.payload) {
      throw new Error('Task payload is required');
    }
  }

  /**
   * Abstract method that must be implemented by subclasses
   * @param task The task request to handle
   * @returns The task response with result
   */
  abstract handleTask(task: TaskRequest): Promise<TaskResponse>;

  /**
   * Helper method to create a task response
   * @param taskId The task ID
   * @param result The result as bytes
   * @returns TaskResponse object
   */
  protected createResponse(taskId: string, result: Uint8Array): TaskResponse {
    return {
      taskId,
      result,
    };
  }
}

/**
 * Simple worker implementation for testing and examples
 */
export class EchoWorker extends BaseWorker {
  /**
   * Echo worker that returns the input payload as result
   * @param task The task request
   * @returns The task response with the same payload
   */
  async handleTask(task: TaskRequest): Promise<TaskResponse> {
    return this.createResponse(task.taskId, task.payload);
  }
}
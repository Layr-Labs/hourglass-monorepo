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
   * Handle task - can be overridden in two ways:
   * 1. Override this method for full TaskRequest/TaskResponse handling
   * 2. Override handleTask(input) for simplified payload handling
   */
  async handleTask(task: TaskRequest): Promise<TaskResponse> {
    // Parse payload and call simplified handler
    const input = this.parsePayload(task.payload);
    const result = await this.handleSimpleTask(input);
    const encoded = this.encodePayload(result);
    return this.createResponse(task.taskId, encoded);
  }

  /**
   * Simplified task handler - override this for easy development
   * @param input Parsed input from task payload
   * @returns Result to be encoded and returned
   */
  async handleSimpleTask(input: any): Promise<any> {
    throw new Error('Either handleTask or handleSimpleTask must be implemented');
  }

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

  /**
   * Parse payload from bytes to JavaScript object
   * @param payload The raw payload bytes
   * @returns Parsed object
   */
  protected parsePayload(payload: Uint8Array): any {
    try {
      const text = new TextDecoder().decode(payload);
      return JSON.parse(text);
    } catch {
      // If not JSON, return as string
      return new TextDecoder().decode(payload);
    }
  }

  /**
   * Encode result to bytes
   * @param result The result to encode
   * @returns Encoded bytes
   */
  protected encodePayload(result: any): Uint8Array {
    if (result instanceof Uint8Array) {
      return result;
    }
    
    const json = typeof result === 'object' ? JSON.stringify(result) : String(result);
    return new TextEncoder().encode(json);
  }

  /**
   * Simple start method for one-line usage
   */
  async start(port: number = 8080): Promise<void> {
    const { PerformerServer } = await import('../server/performerServer');
    
    const server = new PerformerServer(this, {
      port,
      timeout: 10000,
      debug: true,
    });

    // Set up graceful shutdown
    server.setupGracefulShutdown();

    try {
      await server.start();
      console.log(`🚀 Performer is running on port ${port}!`);
    } catch (error) {
      console.error('❌ Failed to start server:', error);
      process.exit(1);
    }
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
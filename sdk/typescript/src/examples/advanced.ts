// Advanced performer implementation examples

import { PerformerServer } from '../server/performerServer';
import { BaseWorker } from '../worker/iWorker';
import { TaskRequest, TaskResponse } from '../types/performer';
import { stringToBytes, bytesToString, jsonToBytes, bytesToJson } from '../types/protobuf';

/**
 * String processing worker example
 */
class StringProcessor extends BaseWorker {
  async handleTask(task: TaskRequest): Promise<TaskResponse> {
    // Convert bytes to string
    const input = bytesToString(task.payload);
    
    // Process the string (uppercase)
    const result = input.toUpperCase();
    
    // Convert back to bytes
    const resultBytes = stringToBytes(result);
    
    return this.createResponse(task.taskId, resultBytes);
  }
}

/**
 * JSON processing worker example
 */
class JsonProcessor extends BaseWorker {
  async handleTask(task: TaskRequest): Promise<TaskResponse> {
    // Parse JSON from bytes
    const input = bytesToJson(task.payload);
    
    // Process the JSON (add timestamp)
    const result = {
      ...input,
      processed_at: new Date().toISOString(),
      processed_by: 'typescript-performer',
    };
    
    // Convert back to bytes
    const resultBytes = jsonToBytes(result);
    
    return this.createResponse(task.taskId, resultBytes);
  }

  async validateTask(task: TaskRequest): Promise<void> {
    await super.validateTask(task);
    
    // Validate JSON format
    try {
      bytesToJson(task.payload);
    } catch (error) {
      throw new Error('Invalid JSON payload');
    }
  }
}

/**
 * Multi-step processing worker example
 */
class MultiStepProcessor extends BaseWorker {
  async handleTask(task: TaskRequest): Promise<TaskResponse> {
    // Step 1: Parse input
    const input = bytesToJson(task.payload);
    
    // Step 2: Validate input structure
    if (!input.operation || !input.data) {
      throw new Error('Invalid input structure. Expected: {operation: string, data: any}');
    }
    
    // Step 3: Process based on operation
    let result: any;
    switch (input.operation) {
      case 'count':
        result = { count: Array.isArray(input.data) ? input.data.length : 0 };
        break;
      case 'sum':
        result = { sum: Array.isArray(input.data) ? input.data.reduce((a, b) => a + b, 0) : 0 };
        break;
      case 'reverse':
        result = { reversed: Array.isArray(input.data) ? input.data.reverse() : input.data };
        break;
      default:
        throw new Error(`Unknown operation: ${input.operation}`);
    }
    
    // Step 4: Add metadata
    const finalResult = {
      ...result,
      metadata: {
        taskId: task.taskId,
        operation: input.operation,
        processedAt: new Date().toISOString(),
      },
    };
    
    return this.createResponse(task.taskId, jsonToBytes(finalResult));
  }
}

/**
 * Example usage of different workers
 */
async function runExamples() {
  // Example 1: String processor
  console.log('Starting String Processor on port 8081...');
  const stringWorker = new StringProcessor();
  const stringServer = new PerformerServer(stringWorker, { port: 8081, debug: true });
  await stringServer.start();
  
  // Example 2: JSON processor
  console.log('Starting JSON Processor on port 8082...');
  const jsonWorker = new JsonProcessor();
  const jsonServer = new PerformerServer(jsonWorker, { port: 8082, debug: true });
  await jsonServer.start();
  
  // Example 3: Multi-step processor
  console.log('Starting Multi-step Processor on port 8083...');
  const multiWorker = new MultiStepProcessor();
  const multiServer = new PerformerServer(multiWorker, { port: 8083, debug: true });
  await multiServer.start();
  
  console.log('All servers started successfully!');
  console.log('- String Processor: port 8081');
  console.log('- JSON Processor: port 8082');
  console.log('- Multi-step Processor: port 8083');
  
  // Set up graceful shutdown for all servers
  const shutdown = async () => {
    console.log('Shutting down all servers...');
    await Promise.all([
      stringServer.stop(),
      jsonServer.stop(),
      multiServer.stop(),
    ]);
    process.exit(0);
  };
  
  process.on('SIGTERM', shutdown);
  process.on('SIGINT', shutdown);
}

// Run if this file is executed directly
if (require.main === module) {
  runExamples().catch(console.error);
}
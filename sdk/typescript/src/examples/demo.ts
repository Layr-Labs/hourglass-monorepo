// Demo performer implementation
// Similar to the Go demo in demo/main.go

import { PerformerServer } from '../server/performerServer';
import { BaseWorker } from '../worker/iWorker';
import { TaskRequest, TaskResponse } from '../types/performer';
import { bytesToNumber, numberToBytes } from '../types/protobuf';

/**
 * Simple demo worker that squares numbers
 * This mirrors the Go demo performer functionality
 */
class SquareWorker extends BaseWorker {
  async handleTask(task: TaskRequest): Promise<TaskResponse> {
    // Convert bytes to number (assuming the payload is a number)
    const input = bytesToNumber(task.payload);
    
    // Square the number
    const result = input * input;
    
    // Convert back to bytes
    const resultBytes = numberToBytes(result);
    
    return this.createResponse(task.taskId, resultBytes);
  }

  async validateTask(task: TaskRequest): Promise<void> {
    // Call parent validation
    await super.validateTask(task);
    
    // Additional validation for numeric input
    if (task.payload.length !== 8) {
      throw new Error('Payload must be 8 bytes (64-bit number)');
    }
  }
}

/**
 * Main function to start the demo performer
 */
async function main() {
  const worker = new SquareWorker();
  
  const server = new PerformerServer(worker, {
    port: 8080,
    timeout: 5000,
    debug: true,
  });

  // Set up graceful shutdown
  server.setupGracefulShutdown();

  try {
    await server.start();
    console.log('Demo performer is running! Send tasks to port 8080');
    
    // Keep the process running
    await new Promise(() => {});
  } catch (error) {
    console.error('Failed to start server:', error);
    process.exit(1);
  }
}

// Run if this file is executed directly
if (require.main === module) {
  main().catch(console.error);
}
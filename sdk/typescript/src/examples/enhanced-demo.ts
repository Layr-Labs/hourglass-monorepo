// Enhanced demo showing advanced task processing features

import { PerformerServer } from '../server/performerServer';
import { BaseWorker } from '../worker/iWorker';
import { TaskRequest, TaskResponse } from '../types/performer';
import { ValidationStage } from '../server/taskContext';
import { JsonSerializationStrategy, NumberSerializationStrategy } from '../server/taskProcessor';

/**
 * Enhanced demo worker with custom validation and serialization
 */
class EnhancedCalculatorWorker extends BaseWorker {
  async handleTask(task: TaskRequest): Promise<TaskResponse> {
    // Parse JSON input
    const input = JSON.parse(new TextDecoder().decode(task.payload));
    
    let result: any;
    
    switch (input.operation) {
      case 'add':
        result = { result: input.a + input.b, operation: 'add' };
        break;
      case 'multiply':
        result = { result: input.a * input.b, operation: 'multiply' };
        break;
      case 'power':
        result = { result: Math.pow(input.a, input.b), operation: 'power' };
        break;
      case 'fibonacci':
        result = { result: this.fibonacci(input.n), operation: 'fibonacci' };
        break;
      default:
        throw new Error(`Unknown operation: ${input.operation}`);
    }
    
    // Add metadata
    result.metadata = {
      processedAt: new Date().toISOString(),
      taskId: task.taskId,
    };
    
    // Convert result to bytes
    const resultBytes = new TextEncoder().encode(JSON.stringify(result));
    
    return this.createResponse(task.taskId, resultBytes);
  }

  async validateTask(task: TaskRequest): Promise<void> {
    await super.validateTask(task);
    
    // Parse and validate JSON structure
    let input: any;
    try {
      input = JSON.parse(new TextDecoder().decode(task.payload));
    } catch {
      throw new Error('Task payload must be valid JSON');
    }
    
    if (!input.operation) {
      throw new Error('Task must specify an operation');
    }
    
    const validOperations = ['add', 'multiply', 'power', 'fibonacci'];
    if (!validOperations.includes(input.operation)) {
      throw new Error(`Invalid operation. Must be one of: ${validOperations.join(', ')}`);
    }
    
    // Validate operation-specific parameters
    switch (input.operation) {
      case 'add':
      case 'multiply':
      case 'power':
        if (typeof input.a !== 'number' || typeof input.b !== 'number') {
          throw new Error(`Operation ${input.operation} requires numeric parameters 'a' and 'b'`);
        }
        break;
      case 'fibonacci':
        if (typeof input.n !== 'number' || input.n < 0 || input.n > 40) {
          throw new Error('Fibonacci operation requires parameter "n" (0-40)');
        }
        break;
    }
  }

  private fibonacci(n: number): number {
    if (n <= 1) return n;
    return this.fibonacci(n - 1) + this.fibonacci(n - 2);
  }
}

/**
 * Demo with custom validation stages
 */
async function main() {
  const worker = new EnhancedCalculatorWorker();
  
  // Create custom validation stages
  const securityValidation = new ValidationStage(async (context) => {
    const input = JSON.parse(new TextDecoder().decode(context.request.payload));
    
    // Example: Block dangerous operations
    if (input.operation === 'fibonacci' && input.n > 35) {
      throw new Error('Fibonacci operations limited to n <= 35 for performance');
    }
    
    context.logger.debug('Security validation passed');
  });
  
  const rateLimit = new ValidationStage(async (context) => {
    // Example: Simple rate limiting based on task ID patterns
    if (context.request.taskId.includes('bulk')) {
      throw new Error('Bulk operations not allowed');
    }
    
    context.logger.debug('Rate limit validation passed');
  });
  
  // Create server with enhanced task processing
  const server = new PerformerServer(
    worker,
    {
      port: 8080,
      timeout: 10000, // 10 second timeout
      debug: true,
    },
    {
      enableMetrics: true,
      validationStages: [securityValidation, rateLimit],
      defaultSerialization: new JsonSerializationStrategy(),
    }
  );

  // Add additional serialization strategies
  server.addSerializationStrategy(new NumberSerializationStrategy());

  // Set up graceful shutdown
  server.setupGracefulShutdown();

  try {
    await server.start();
    console.log('Enhanced Calculator Performer is running on port 8080!');
    console.log('\nSupported operations:');
    console.log('- add: {"operation": "add", "a": 5, "b": 3}');
    console.log('- multiply: {"operation": "multiply", "a": 4, "b": 7}');
    console.log('- power: {"operation": "power", "a": 2, "b": 8}');
    console.log('- fibonacci: {"operation": "fibonacci", "n": 10}');
    
    // Periodically log metrics
    setInterval(() => {
      const metrics = server.getTaskMetrics();
      if (metrics.length > 0) {
        console.log('\n📊 Task Metrics:');
        metrics.forEach(metric => {
          console.log(`  ${metric.taskId}: ${metric.success ? '✅' : '❌'} ${metric.duration}ms`);
        });
        server.clearTaskMetrics(); // Clear after logging
      }
    }, 30000); // Every 30 seconds
    
    // Keep the process running
    await new Promise(() => {});
  } catch (error) {
    console.error('Failed to start enhanced performer:', error);
    process.exit(1);
  }
}

// Run if this file is executed directly
if (require.main === module) {
  main().catch(console.error);
}
// Advanced performer template with full monitoring and TypeChain integration

import { PerformerServer } from '../server/performerServer';
import { SolidityWorker } from '../worker/solidityWorker';
import { MemoryHealthProvider } from '../server/healthManager';
import { FileMetricsExporter } from '../utils/metrics';

/**
 * Advanced worker with comprehensive error handling and monitoring
 */
class AdvancedSolidityWorker extends SolidityWorker<any, 'processAdvancedTask'> {
  private processedCount = 0;
  private errorCount = 0;
  private lastProcessedTime = 0;

  constructor() {
    super({
      // TODO: Replace with your contract ABI
      abi: [
        {
          inputs: [
            { name: 'taskId', type: 'bytes32' },
            { name: 'payload', type: 'bytes' },
            { name: 'metadata', type: 'tuple', components: [
              { name: 'timestamp', type: 'uint256' },
              { name: 'priority', type: 'uint8' },
              { name: 'user', type: 'address' },
            ]},
          ],
          name: 'processAdvancedTask',
          outputs: [
            { name: 'result', type: 'bytes' },
            { name: 'success', type: 'bool' },
            { name: 'gasUsed', type: 'uint256' },
          ],
          stateMutability: 'nonpayable',
          type: 'function',
        },
      ],
      functionName: 'processAdvancedTask',
      autoDetectPayload: true,
      strictMode: false,
    });
  }

  async handleSolidityTask(
    params: {
      taskId: string;
      payload: Uint8Array;
      metadata: { timestamp: bigint; priority: number; user: string };
    }
  ): Promise<{ result: Uint8Array; success: boolean; gasUsed: bigint }> {
    const startTime = Date.now();
    
    try {
      const { taskId, payload, metadata } = params;
      
      console.log(`Processing advanced task ${taskId}`, {
        timestamp: metadata.timestamp,
        priority: metadata.priority,
        user: metadata.user,
      });

      // Simulate processing with priority handling
      const processingTime = metadata.priority === 1 ? 100 : 1000; // High priority = faster
      await new Promise(resolve => setTimeout(resolve, processingTime));

      // TODO: Implement your advanced task processing logic here
      const processedData = new TextEncoder().encode(
        JSON.stringify({
          taskId,
          originalPayload: new TextDecoder().decode(payload),
          processedAt: new Date().toISOString(),
          priority: metadata.priority,
          user: metadata.user,
          processingTime,
        })
      );

      // Update metrics
      this.processedCount++;
      this.lastProcessedTime = Date.now();

      // Simulate gas usage calculation
      const gasUsed = BigInt(Math.floor(processingTime * 21)); // Simulate gas cost

      return {
        result: processedData,
        success: true,
        gasUsed,
      };
    } catch (error) {
      console.error('Advanced task processing error:', error);
      this.errorCount++;
      
      const errorData = new TextEncoder().encode(
        JSON.stringify({
          error: error instanceof Error ? error.message : 'Unknown error',
          taskId: params.taskId,
          timestamp: new Date().toISOString(),
        })
      );

      return {
        result: errorData,
        success: false,
        gasUsed: BigInt(0),
      };
    }
  }

  // Custom health check for this worker
  async checkHealth(): Promise<{ healthy: boolean; details?: any; error?: string }> {
    const timeSinceLastTask = Date.now() - this.lastProcessedTime;
    const errorRate = this.errorCount / Math.max(this.processedCount, 1);
    
    return {
      healthy: errorRate < 0.1 && timeSinceLastTask < 600000, // Healthy if error rate < 10% and processed task in last 10 minutes
      details: {
        processedCount: this.processedCount,
        errorCount: this.errorCount,
        errorRate: errorRate,
        timeSinceLastTask: timeSinceLastTask,
      },
      error: errorRate >= 0.1 ? 'Error rate too high' : 
             timeSinceLastTask >= 600000 ? 'No recent task processing' : undefined,
    };
  }
}

/**
 * Custom health provider for the worker
 */
class AdvancedWorkerHealthProvider {
  name = 'advanced-worker';
  
  constructor(private worker: AdvancedSolidityWorker) {}
  
  async check(): Promise<{ healthy: boolean; details?: any; error?: string }> {
    return await this.worker.checkHealth();
  }
}

/**
 * Main performer server with advanced configuration
 */
async function main() {
  const worker = new AdvancedSolidityWorker();
  
  // Create server with comprehensive monitoring
  const server = new PerformerServer(
    worker,
    {
      port: 8080,
      timeout: 30000, // 30 second timeout for complex tasks
      debug: true,
    },
    // Task processor config
    {
      enableMetrics: true,
      timeout: 30000,
    },
    // Health config
    {
      checkInterval: 10000, // Check every 10 seconds
      maxFailures: 3,
      autoRecover: true,
    },
    // Metrics config
    {
      enabled: true,
      exportInterval: 60000, // Export every minute
      defaultLabels: {
        service: 'advanced-performer',
        version: '1.0.0',
      },
    }
  );

  // Add custom health provider
  server.addHealthProvider(new AdvancedWorkerHealthProvider(worker));

  // Add file metrics exporter
  server.addMetricsExporter(new FileMetricsExporter('./metrics.json'));

  // Add memory health provider with custom threshold
  server.addHealthProvider(new MemoryHealthProvider(512)); // 512MB threshold

  // Set up graceful shutdown
  server.setupGracefulShutdown();

  try {
    await server.start();
    console.log('🚀 Advanced Performer is running on port 8080!');
    console.log('📊 Monitoring Features:');
    console.log('  • Health checks every 10 seconds');
    console.log('  • Metrics export every minute');
    console.log('  • Memory leak detection');
    console.log('  • Advanced error handling');
    console.log('  • TypeChain integration');
    console.log('💡 Make sure to run "npm run typechain" to generate contract types');

    // Periodic status reporting
    setInterval(() => {
      const healthSummary = server.getHealthSummary();
      const metricsSummary = server.getMetricsSummary();
      const memoryMB = Math.round(process.memoryUsage().heapUsed / 1024 / 1024);
      
      console.log(`📈 Status: ${healthSummary.status} | Tasks: ${healthSummary.taskCount} | Metrics: ${metricsSummary.totalMetrics} | Memory: ${memoryMB}MB`);
    }, 60000); // Every minute

  } catch (error) {
    console.error('❌ Failed to start server:', error);
    process.exit(1);
  }
}

main().catch(console.error);
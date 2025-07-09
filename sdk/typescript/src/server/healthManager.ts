// Health and status management for the performer server

import { PerformerStatus } from '../types/performer';
import { ILogger } from '../utils/logger';

/**
 * Health check configuration
 */
export interface HealthCheckConfig {
  /** Interval for automatic health checks in milliseconds */
  checkInterval?: number;
  /** Maximum number of consecutive failures before marking unhealthy */
  maxFailures?: number;
  /** Timeout for health check operations in milliseconds */
  timeout?: number;
  /** Enable automatic recovery attempts */
  autoRecover?: boolean;
}

/**
 * Health check result
 */
export interface HealthCheckResult {
  /** Overall health status */
  status: PerformerStatus;
  /** Health check timestamp */
  timestamp: number;
  /** Detailed health information */
  details: {
    uptime: number;
    memoryUsage: NodeJS.MemoryUsage;
    taskCount: number;
    lastTaskTime?: number;
    errors: string[];
  };
  /** Custom health indicators */
  custom?: Record<string, any>;
}

/**
 * Health check provider interface
 */
export interface HealthCheckProvider {
  /** Provider name */
  name: string;
  /** Check health and return status */
  check(): Promise<{ healthy: boolean; details?: any; error?: string }>;
}

/**
 * Built-in health check providers
 */
export class MemoryHealthProvider implements HealthCheckProvider {
  name = 'memory';
  
  constructor(private maxMemoryMB: number = 512) {}

  async check(): Promise<{ healthy: boolean; details?: any; error?: string }> {
    const memUsage = process.memoryUsage();
    const usedMB = memUsage.heapUsed / 1024 / 1024;
    
    return {
      healthy: usedMB < this.maxMemoryMB,
      details: {
        usedMB: Math.round(usedMB),
        maxMB: this.maxMemoryMB,
        heapUsed: memUsage.heapUsed,
        heapTotal: memUsage.heapTotal,
      },
      error: usedMB >= this.maxMemoryMB ? `Memory usage ${Math.round(usedMB)}MB exceeds limit ${this.maxMemoryMB}MB` : undefined,
    };
  }
}

export class UptimeHealthProvider implements HealthCheckProvider {
  name = 'uptime';
  
  constructor(private startTime: number = Date.now()) {}

  async check(): Promise<{ healthy: boolean; details?: any }> {
    const uptime = Date.now() - this.startTime;
    
    return {
      healthy: true,
      details: {
        uptimeMs: uptime,
        uptimeSeconds: Math.round(uptime / 1000),
        startTime: this.startTime,
      },
    };
  }
}

/**
 * Health manager for tracking performer health and status
 */
export class HealthManager {
  private currentStatus: PerformerStatus = PerformerStatus.STARTING;
  private config: Required<HealthCheckConfig>;
  private providers: Map<string, HealthCheckProvider> = new Map();
  private failures: number = 0;
  private lastHealthCheck?: HealthCheckResult;
  private healthCheckInterval?: NodeJS.Timeout;
  private startTime: number = Date.now();
  private taskCount: number = 0;
  private lastTaskTime?: number;
  private errors: string[] = [];
  private customMetrics: Map<string, any> = new Map();

  constructor(
    private logger: ILogger,
    config: HealthCheckConfig = {}
  ) {
    this.config = {
      checkInterval: config.checkInterval ?? 30000, // 30 seconds
      maxFailures: config.maxFailures ?? 3,
      timeout: config.timeout ?? 5000, // 5 seconds
      autoRecover: config.autoRecover ?? true,
    };

    // Add built-in health check providers
    this.addProvider(new MemoryHealthProvider());
    this.addProvider(new UptimeHealthProvider(this.startTime));
  }

  /**
   * Start the health manager
   */
  start(): void {
    this.currentStatus = PerformerStatus.READY_FOR_TASK;
    this.logger.info('Health manager started');

    // Start periodic health checks
    if (this.config.checkInterval > 0) {
      this.healthCheckInterval = setInterval(() => {
        this.performHealthCheck().catch(error => {
          this.logger.error('Health check failed', { error: error.message });
        });
      }, this.config.checkInterval);
    }
  }

  /**
   * Stop the health manager
   */
  stop(): void {
    this.currentStatus = PerformerStatus.STOPPING;
    
    if (this.healthCheckInterval) {
      clearInterval(this.healthCheckInterval);
      this.healthCheckInterval = undefined;
    }

    this.logger.info('Health manager stopped');
  }

  /**
   * Add a custom health check provider
   */
  addProvider(provider: HealthCheckProvider): void {
    this.providers.set(provider.name, provider);
    this.logger.debug(`Added health check provider: ${provider.name}`);
  }

  /**
   * Remove a health check provider
   */
  removeProvider(name: string): void {
    this.providers.delete(name);
    this.logger.debug(`Removed health check provider: ${name}`);
  }

  /**
   * Get current status
   */
  getStatus(): PerformerStatus {
    return this.currentStatus;
  }

  /**
   * Set status manually
   */
  setStatus(status: PerformerStatus): void {
    const oldStatus = this.currentStatus;
    this.currentStatus = status;
    
    if (oldStatus !== status) {
      this.logger.info(`Status changed: ${oldStatus} -> ${status}`);
    }
  }

  /**
   * Record a task execution
   */
  recordTask(): void {
    this.taskCount++;
    this.lastTaskTime = Date.now();
  }

  /**
   * Record an error
   */
  recordError(error: string): void {
    this.errors.push(`${new Date().toISOString()}: ${error}`);
    
    // Keep only last 10 errors
    if (this.errors.length > 10) {
      this.errors = this.errors.slice(-10);
    }

    // Update status if too many recent errors
    const recentErrors = this.errors.filter(e => {
      const errorTime = new Date(e.split(':')[0]).getTime();
      return Date.now() - errorTime < 60000; // Last minute
    });

    if (recentErrors.length >= 3) {
      this.setStatus(PerformerStatus.ERROR);
    }
  }

  /**
   * Set custom metric
   */
  setCustomMetric(key: string, value: any): void {
    this.customMetrics.set(key, value);
  }

  /**
   * Get custom metric
   */
  getCustomMetric(key: string): any {
    return this.customMetrics.get(key);
  }

  /**
   * Perform comprehensive health check
   */
  async performHealthCheck(): Promise<HealthCheckResult> {
    const timestamp = Date.now();
    let overallHealthy = true;
    const providerResults: Record<string, any> = {};

    // Run all health check providers
    for (const [name, provider] of this.providers) {
      try {
        const result = await Promise.race([
          provider.check(),
          new Promise((_, reject) => 
            setTimeout(() => reject(new Error('Health check timeout')), this.config.timeout)
          )
        ]) as { healthy: boolean; details?: any; error?: string };

        providerResults[name] = result;
        
        if (!result.healthy) {
          overallHealthy = false;
          if (result.error) {
            this.recordError(`Health check failed [${name}]: ${result.error}`);
          }
        }
      } catch (error) {
        overallHealthy = false;
        const errorMsg = error instanceof Error ? error.message : 'Unknown error';
        providerResults[name] = { healthy: false, error: errorMsg };
        this.recordError(`Health check error [${name}]: ${errorMsg}`);
      }
    }

    // Update failure count and status
    if (overallHealthy) {
      this.failures = 0;
      if (this.currentStatus === PerformerStatus.ERROR && this.config.autoRecover) {
        this.setStatus(PerformerStatus.READY_FOR_TASK);
      }
    } else {
      this.failures++;
      if (this.failures >= this.config.maxFailures) {
        this.setStatus(PerformerStatus.ERROR);
      }
    }

    // Create health check result
    const result: HealthCheckResult = {
      status: this.currentStatus,
      timestamp,
      details: {
        uptime: timestamp - this.startTime,
        memoryUsage: process.memoryUsage(),
        taskCount: this.taskCount,
        lastTaskTime: this.lastTaskTime,
        errors: [...this.errors],
      },
      custom: {
        providers: providerResults,
        failures: this.failures,
        maxFailures: this.config.maxFailures,
        ...Object.fromEntries(this.customMetrics),
      },
    };

    this.lastHealthCheck = result;
    return result;
  }

  /**
   * Get last health check result
   */
  getLastHealthCheck(): HealthCheckResult | undefined {
    return this.lastHealthCheck;
  }

  /**
   * Get health summary for monitoring
   */
  getHealthSummary(): {
    status: PerformerStatus;
    uptime: number;
    taskCount: number;
    errorCount: number;
    memoryUsageMB: number;
  } {
    const memUsage = process.memoryUsage();
    
    return {
      status: this.currentStatus,
      uptime: Date.now() - this.startTime,
      taskCount: this.taskCount,
      errorCount: this.errors.length,
      memoryUsageMB: Math.round(memUsage.heapUsed / 1024 / 1024),
    };
  }
}
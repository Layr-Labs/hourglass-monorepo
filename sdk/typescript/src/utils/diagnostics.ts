// Debugging and troubleshooting utilities for the performer

import { PerformerStatus } from '../types/performer';
import { MetricsCollector } from './metrics';
import { ILogger } from './logger';

/**
 * System information
 */
export interface SystemInfo {
  /** Node.js version */
  nodeVersion: string;
  /** Platform information */
  platform: string;
  /** Architecture */
  arch: string;
  /** CPU information */
  cpus: number;
  /** Total memory in bytes */
  totalMemory: number;
  /** Free memory in bytes */
  freeMemory: number;
  /** Uptime in seconds */
  uptime: number;
  /** Load average */
  loadAverage: number[];
}

/**
 * Performance snapshot
 */
export interface PerformanceSnapshot {
  /** Timestamp when snapshot was taken */
  timestamp: number;
  /** Memory usage */
  memory: NodeJS.MemoryUsage;
  /** CPU usage (if available) */
  cpu?: {
    user: number;
    system: number;
  };
  /** Event loop lag in milliseconds */
  eventLoopLag: number;
  /** Active handles and requests */
  handles: number;
  requests: number;
}

/**
 * Diagnostic information
 */
export interface DiagnosticInfo {
  /** System information */
  system: SystemInfo;
  /** Current performance snapshot */
  performance: PerformanceSnapshot;
  /** Performer status */
  status: PerformerStatus;
  /** Configuration */
  config: any;
  /** Environment variables (filtered) */
  environment: Record<string, string>;
  /** Recent errors */
  recentErrors: string[];
  /** Metrics summary */
  metricsSummary?: any;
}

/**
 * Debug mode configuration
 */
export interface DebugConfig {
  /** Enable debug mode */
  enabled: boolean;
  /** Enable performance monitoring */
  enablePerformanceMonitoring: boolean;
  /** Enable memory leak detection */
  enableMemoryLeakDetection: boolean;
  /** Enable request tracing */
  enableRequestTracing: boolean;
  /** Performance monitoring interval in milliseconds */
  performanceInterval: number;
  /** Memory threshold for warnings (in MB) */
  memoryThreshold: number;
}

/**
 * Memory leak detector
 */
export class MemoryLeakDetector {
  private snapshots: PerformanceSnapshot[] = [];
  private config: DebugConfig;
  private logger: ILogger;
  private checkInterval?: NodeJS.Timeout;

  constructor(logger: ILogger, config: DebugConfig) {
    this.logger = logger;
    this.config = config;
  }

  start(): void {
    if (!this.config.enableMemoryLeakDetection) return;

    this.checkInterval = setInterval(() => {
      this.checkMemoryUsage();
    }, this.config.performanceInterval);

    this.logger.debug('Memory leak detector started');
  }

  stop(): void {
    if (this.checkInterval) {
      clearInterval(this.checkInterval);
      this.checkInterval = undefined;
    }
    this.logger.debug('Memory leak detector stopped');
  }

  private checkMemoryUsage(): void {
    const snapshot = this.createPerformanceSnapshot();
    this.snapshots.push(snapshot);

    // Keep only last 10 snapshots
    if (this.snapshots.length > 10) {
      this.snapshots = this.snapshots.slice(-10);
    }

    const memoryMB = snapshot.memory.heapUsed / 1024 / 1024;

    // Check for memory threshold
    if (memoryMB > this.config.memoryThreshold) {
      this.logger.warn('Memory usage exceeds threshold', {
        currentMB: Math.round(memoryMB),
        thresholdMB: this.config.memoryThreshold,
      });
    }

    // Check for memory growth trend
    if (this.snapshots.length >= 5) {
      const trend = this.calculateMemoryTrend();
      if (trend > 10) { // Growing by more than 10MB
        this.logger.warn('Potential memory leak detected', {
          trendMB: Math.round(trend),
          currentMB: Math.round(memoryMB),
        });
      }
    }
  }

  private calculateMemoryTrend(): number {
    const recent = this.snapshots.slice(-5);
    const first = recent[0].memory.heapUsed / 1024 / 1024;
    const last = recent[recent.length - 1].memory.heapUsed / 1024 / 1024;
    return last - first;
  }

  private createPerformanceSnapshot(): PerformanceSnapshot {
    const hrTime = process.hrtime();
    const start = hrTime[0] * 1000 + hrTime[1] / 1000000;
    
    // Measure event loop lag
    setImmediate(() => {
      const end = process.hrtime()[0] * 1000 + process.hrtime()[1] / 1000000;
      const lag = end - start;
      
      return {
        timestamp: Date.now(),
        memory: process.memoryUsage(),
        eventLoopLag: lag,
        handles: (process as any)._getActiveHandles().length,
        requests: (process as any)._getActiveRequests().length,
      };
    });

    return {
      timestamp: Date.now(),
      memory: process.memoryUsage(),
      eventLoopLag: 0, // Will be updated asynchronously
      handles: (process as any)._getActiveHandles?.().length || 0,
      requests: (process as any)._getActiveRequests?.().length || 0,
    };
  }
}

/**
 * Request tracer for debugging
 */
export class RequestTracer {
  private traces: Map<string, any> = new Map();
  private config: DebugConfig;
  private logger: ILogger;

  constructor(logger: ILogger, config: DebugConfig) {
    this.logger = logger;
    this.config = config;
  }

  startTrace(id: string, operation: string, metadata?: any): void {
    if (!this.config.enableRequestTracing) return;

    this.traces.set(id, {
      id,
      operation,
      startTime: Date.now(),
      metadata: metadata || {},
      steps: [],
    });

    this.logger.debug(`Started trace: ${operation}`, { traceId: id });
  }

  addStep(id: string, step: string, metadata?: any): void {
    if (!this.config.enableRequestTracing) return;

    const trace = this.traces.get(id);
    if (trace) {
      trace.steps.push({
        step,
        timestamp: Date.now(),
        metadata: metadata || {},
      });
    }
  }

  endTrace(id: string, result?: any): void {
    if (!this.config.enableRequestTracing) return;

    const trace = this.traces.get(id);
    if (trace) {
      trace.endTime = Date.now();
      trace.duration = trace.endTime - trace.startTime;
      trace.result = result;

      this.logger.debug(`Completed trace: ${trace.operation}`, {
        traceId: id,
        duration: trace.duration,
        steps: trace.steps.length,
      });

      // Clean up old traces
      this.traces.delete(id);
    }
  }

  getActiveTraces(): any[] {
    return Array.from(this.traces.values());
  }
}

/**
 * Comprehensive diagnostics tool
 */
export class DiagnosticTool {
  private config: DebugConfig;
  private logger: ILogger;
  private memoryLeakDetector: MemoryLeakDetector;
  private requestTracer: RequestTracer;
  private metricsCollector?: MetricsCollector;
  private recentErrors: string[] = [];

  constructor(logger: ILogger, config: Partial<DebugConfig> = {}) {
    this.config = {
      enabled: config.enabled ?? false,
      enablePerformanceMonitoring: config.enablePerformanceMonitoring ?? true,
      enableMemoryLeakDetection: config.enableMemoryLeakDetection ?? true,
      enableRequestTracing: config.enableRequestTracing ?? false,
      performanceInterval: config.performanceInterval ?? 30000,
      memoryThreshold: config.memoryThreshold ?? 512,
    };

    this.logger = logger;
    this.memoryLeakDetector = new MemoryLeakDetector(logger, this.config);
    this.requestTracer = new RequestTracer(logger, this.config);
  }

  /**
   * Start diagnostic monitoring
   */
  start(): void {
    if (!this.config.enabled) {
      this.logger.debug('Diagnostics disabled');
      return;
    }

    this.memoryLeakDetector.start();
    this.logger.info('Diagnostic monitoring started');
  }

  /**
   * Stop diagnostic monitoring
   */
  stop(): void {
    this.memoryLeakDetector.stop();
    this.logger.info('Diagnostic monitoring stopped');
  }

  /**
   * Set metrics collector
   */
  setMetricsCollector(collector: MetricsCollector): void {
    this.metricsCollector = collector;
  }

  /**
   * Record error for diagnostics
   */
  recordError(error: string): void {
    this.recentErrors.push(`${new Date().toISOString()}: ${error}`);
    
    // Keep only last 20 errors
    if (this.recentErrors.length > 20) {
      this.recentErrors = this.recentErrors.slice(-20);
    }
  }

  /**
   * Get system information
   */
  getSystemInfo(): SystemInfo {
    const os = require('os');
    
    return {
      nodeVersion: process.version,
      platform: os.platform(),
      arch: os.arch(),
      cpus: os.cpus().length,
      totalMemory: os.totalmem(),
      freeMemory: os.freemem(),
      uptime: os.uptime(),
      loadAverage: os.loadavg(),
    };
  }

  /**
   * Get current performance snapshot
   */
  getPerformanceSnapshot(): PerformanceSnapshot {
    return {
      timestamp: Date.now(),
      memory: process.memoryUsage(),
      eventLoopLag: 0, // Would need async measurement
      handles: (process as any)._getActiveHandles?.().length || 0,
      requests: (process as any)._getActiveRequests?.().length || 0,
    };
  }

  /**
   * Get comprehensive diagnostic information
   */
  getDiagnosticInfo(status: PerformerStatus, config: any): DiagnosticInfo {
    return {
      system: this.getSystemInfo(),
      performance: this.getPerformanceSnapshot(),
      status,
      config: this.sanitizeConfig(config),
      environment: this.getFilteredEnvironment(),
      recentErrors: [...this.recentErrors],
      metricsSummary: this.metricsCollector?.getSummary(),
    };
  }

  /**
   * Start request trace
   */
  startTrace(id: string, operation: string, metadata?: any): void {
    this.requestTracer.startTrace(id, operation, metadata);
  }

  /**
   * Add step to trace
   */
  addTraceStep(id: string, step: string, metadata?: any): void {
    this.requestTracer.addStep(id, step, metadata);
  }

  /**
   * End request trace
   */
  endTrace(id: string, result?: any): void {
    this.requestTracer.endTrace(id, result);
  }

  /**
   * Get active traces
   */
  getActiveTraces(): any[] {
    return this.requestTracer.getActiveTraces();
  }

  /**
   * Perform health check on the system
   */
  performHealthCheck(): {
    healthy: boolean;
    issues: string[];
    recommendations: string[];
  } {
    const issues: string[] = [];
    const recommendations: string[] = [];

    const perf = this.getPerformanceSnapshot();
    const system = this.getSystemInfo();

    // Check memory usage
    const memoryUsageMB = perf.memory.heapUsed / 1024 / 1024;
    if (memoryUsageMB > this.config.memoryThreshold) {
      issues.push(`High memory usage: ${Math.round(memoryUsageMB)}MB`);
      recommendations.push('Consider optimizing memory usage or increasing memory limits');
    }

    // Check system memory
    const systemMemoryUsage = (system.totalMemory - system.freeMemory) / system.totalMemory;
    if (systemMemoryUsage > 0.9) {
      issues.push('System memory usage is very high');
      recommendations.push('Consider scaling or optimizing system resources');
    }

    // Check load average
    const loadAvg = system.loadAverage[0];
    if (loadAvg > system.cpus * 2) {
      issues.push(`High system load: ${loadAvg.toFixed(2)}`);
      recommendations.push('System may be overloaded, consider scaling');
    }

    // Check active handles
    if (perf.handles > 1000) {
      issues.push(`High number of active handles: ${perf.handles}`);
      recommendations.push('Check for resource leaks or connection pooling issues');
    }

    return {
      healthy: issues.length === 0,
      issues,
      recommendations,
    };
  }

  /**
   * Generate diagnostic report
   */
  generateReport(status: PerformerStatus, config: any): string {
    const info = this.getDiagnosticInfo(status, config);
    const healthCheck = this.performHealthCheck();

    let report = '# Performer Diagnostic Report\n\n';
    report += `Generated: ${new Date().toISOString()}\n\n`;

    report += '## System Information\n';
    report += `- Node.js: ${info.system.nodeVersion}\n`;
    report += `- Platform: ${info.system.platform} ${info.system.arch}\n`;
    report += `- CPUs: ${info.system.cpus}\n`;
    report += `- Memory: ${Math.round(info.system.totalMemory / 1024 / 1024 / 1024)}GB total, ${Math.round(info.system.freeMemory / 1024 / 1024 / 1024)}GB free\n`;
    report += `- Uptime: ${Math.round(info.system.uptime / 3600)}h\n\n`;

    report += '## Performance\n';
    report += `- Status: ${info.status}\n`;
    report += `- Memory Usage: ${Math.round(info.performance.memory.heapUsed / 1024 / 1024)}MB\n`;
    report += `- Active Handles: ${info.performance.handles}\n`;
    report += `- Active Requests: ${info.performance.requests}\n\n`;

    if (info.metricsSummary) {
      report += '## Metrics Summary\n';
      report += `- Total Metrics: ${info.metricsSummary.totalMetrics}\n`;
      report += `- Counters: ${info.metricsSummary.counters}\n`;
      report += `- Gauges: ${info.metricsSummary.gauges}\n`;
      report += `- Histograms: ${info.metricsSummary.histograms}\n\n`;
    }

    report += '## Health Check\n';
    report += `- Overall Health: ${healthCheck.healthy ? '✅ Healthy' : '❌ Issues Detected'}\n`;
    
    if (healthCheck.issues.length > 0) {
      report += '- Issues:\n';
      healthCheck.issues.forEach(issue => {
        report += `  - ${issue}\n`;
      });
    }

    if (healthCheck.recommendations.length > 0) {
      report += '- Recommendations:\n';
      healthCheck.recommendations.forEach(rec => {
        report += `  - ${rec}\n`;
      });
    }

    if (info.recentErrors.length > 0) {
      report += '\n## Recent Errors\n';
      info.recentErrors.slice(-5).forEach(error => {
        report += `- ${error}\n`;
      });
    }

    return report;
  }

  /**
   * Sanitize configuration for diagnostics
   */
  private sanitizeConfig(config: any): any {
    const sensitive = ['password', 'secret', 'key', 'token'];
    const sanitized = { ...config };

    const sanitizeObject = (obj: any): any => {
      if (typeof obj !== 'object' || obj === null) return obj;
      
      const result: any = {};
      for (const [key, value] of Object.entries(obj)) {
        if (sensitive.some(s => key.toLowerCase().includes(s))) {
          result[key] = '[REDACTED]';
        } else if (typeof value === 'object') {
          result[key] = sanitizeObject(value);
        } else {
          result[key] = value;
        }
      }
      return result;
    };

    return sanitizeObject(sanitized);
  }

  /**
   * Get filtered environment variables
   */
  private getFilteredEnvironment(): Record<string, string> {
    const sensitive = ['password', 'secret', 'key', 'token'];
    const filtered: Record<string, string> = {};

    for (const [key, value] of Object.entries(process.env)) {
      if (value === undefined) continue;
      
      if (sensitive.some(s => key.toLowerCase().includes(s))) {
        filtered[key] = '[REDACTED]';
      } else {
        filtered[key] = value;
      }
    }

    return filtered;
  }
}
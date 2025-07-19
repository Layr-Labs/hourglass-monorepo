// Comprehensive metrics collection and monitoring system

import { ILogger } from './logger';

/**
 * Metric types
 */
export enum MetricType {
  COUNTER = 'counter',
  GAUGE = 'gauge',
  HISTOGRAM = 'histogram',
  SUMMARY = 'summary',
}

/**
 * Metric data point
 */
export interface MetricPoint {
  /** Metric name */
  name: string;
  /** Metric type */
  type: MetricType;
  /** Metric value */
  value: number;
  /** Timestamp when metric was recorded */
  timestamp: number;
  /** Optional labels/tags */
  labels?: Record<string, string>;
  /** Optional help text */
  help?: string;
}

/**
 * Metric aggregation result
 */
export interface MetricAggregation {
  /** Metric name */
  name: string;
  /** Number of data points */
  count: number;
  /** Sum of all values */
  sum: number;
  /** Average value */
  avg: number;
  /** Minimum value */
  min: number;
  /** Maximum value */
  max: number;
  /** Standard deviation */
  stdDev: number;
  /** Percentiles (50th, 90th, 95th, 99th) */
  percentiles: {
    p50: number;
    p90: number;
    p95: number;
    p99: number;
  };
}

/**
 * Metrics exporter interface
 */
export interface MetricsExporter {
  /** Exporter name */
  name: string;
  /** Export metrics */
  export(metrics: MetricPoint[]): Promise<void>;
}

/**
 * Console metrics exporter
 */
export class ConsoleMetricsExporter implements MetricsExporter {
  name = 'console';

  async export(metrics: MetricPoint[]): Promise<void> {
    console.log('\n📊 Metrics Report:');
    console.log('═'.repeat(50));
    
    const grouped = this.groupMetricsByName(metrics);
    
    for (const [name, points] of grouped) {
      const latest = points[points.length - 1];
      const agg = this.calculateAggregation(name, points);
      
      console.log(`\n${name} (${latest.type})`);
      console.log(`  Current: ${latest.value}`);
      if (points.length > 1) {
        console.log(`  Count: ${agg.count}, Avg: ${agg.avg.toFixed(2)}, Min: ${agg.min}, Max: ${agg.max}`);
      }
      if (latest.labels && Object.keys(latest.labels).length > 0) {
        console.log(`  Labels: ${JSON.stringify(latest.labels)}`);
      }
    }
    console.log('═'.repeat(50));
  }

  private groupMetricsByName(metrics: MetricPoint[]): Map<string, MetricPoint[]> {
    const grouped = new Map<string, MetricPoint[]>();
    
    for (const metric of metrics) {
      const existing = grouped.get(metric.name) || [];
      existing.push(metric);
      grouped.set(metric.name, existing);
    }
    
    return grouped;
  }

  private calculateAggregation(name: string, points: MetricPoint[]): MetricAggregation {
    const values = points.map(p => p.value).sort((a, b) => a - b);
    const count = values.length;
    const sum = values.reduce((a, b) => a + b, 0);
    const avg = sum / count;
    const min = values[0];
    const max = values[count - 1];
    
    // Calculate standard deviation
    const variance = values.reduce((acc, val) => acc + Math.pow(val - avg, 2), 0) / count;
    const stdDev = Math.sqrt(variance);
    
    // Calculate percentiles
    const getPercentile = (p: number) => {
      const index = Math.ceil((p / 100) * count) - 1;
      return values[Math.max(0, index)];
    };
    
    return {
      name,
      count,
      sum,
      avg,
      min,
      max,
      stdDev,
      percentiles: {
        p50: getPercentile(50),
        p90: getPercentile(90),
        p95: getPercentile(95),
        p99: getPercentile(99),
      },
    };
  }
}

/**
 * JSON file metrics exporter
 */
export class FileMetricsExporter implements MetricsExporter {
  name = 'file';

  constructor(private filePath: string) {}

  async export(metrics: MetricPoint[]): Promise<void> {
    const fs = await import('fs/promises');
    const data = {
      timestamp: new Date().toISOString(),
      metrics,
    };
    
    await fs.writeFile(this.filePath, JSON.stringify(data, null, 2));
  }
}

/**
 * Metrics collector configuration
 */
export interface MetricsConfig {
  /** Enable metrics collection */
  enabled?: boolean;
  /** Maximum number of metric points to keep in memory */
  maxPoints?: number;
  /** Export interval in milliseconds */
  exportInterval?: number;
  /** Metrics retention period in milliseconds */
  retentionPeriod?: number;
  /** Default labels to add to all metrics */
  defaultLabels?: Record<string, string>;
}

/**
 * Comprehensive metrics collector
 */
export class MetricsCollector {
  private config: Required<MetricsConfig>;
  private metrics: MetricPoint[] = [];
  private exporters: MetricsExporter[] = [];
  private exportInterval?: NodeJS.Timeout;
  private counters: Map<string, number> = new Map();
  private gauges: Map<string, number> = new Map();
  private histograms: Map<string, number[]> = new Map();

  constructor(
    private logger: ILogger,
    config: MetricsConfig = {}
  ) {
    this.config = {
      enabled: config.enabled ?? true,
      maxPoints: config.maxPoints ?? 10000,
      exportInterval: config.exportInterval ?? 60000, // 1 minute
      retentionPeriod: config.retentionPeriod ?? 3600000, // 1 hour
      defaultLabels: config.defaultLabels ?? {},
    };

    // Add default console exporter
    this.addExporter(new ConsoleMetricsExporter());
  }

  /**
   * Start the metrics collector
   */
  start(): void {
    if (!this.config.enabled) {
      this.logger.debug('Metrics collection is disabled');
      return;
    }

    this.logger.info('Metrics collector started');

    // Start periodic export
    if (this.config.exportInterval > 0) {
      this.exportInterval = setInterval(() => {
        this.exportMetrics().catch(error => {
          this.logger.error('Failed to export metrics', { error: error.message });
        });
      }, this.config.exportInterval);
    }

    // Start periodic cleanup
    setInterval(() => {
      this.cleanup();
    }, this.config.retentionPeriod / 4); // Cleanup every quarter of retention period
  }

  /**
   * Stop the metrics collector
   */
  stop(): void {
    if (this.exportInterval) {
      clearInterval(this.exportInterval);
      this.exportInterval = undefined;
    }

    this.logger.info('Metrics collector stopped');
  }

  /**
   * Add metrics exporter
   */
  addExporter(exporter: MetricsExporter): void {
    this.exporters.push(exporter);
    this.logger.debug(`Added metrics exporter: ${exporter.name}`);
  }

  /**
   * Remove metrics exporter
   */
  removeExporter(name: string): void {
    this.exporters = this.exporters.filter(e => e.name !== name);
    this.logger.debug(`Removed metrics exporter: ${name}`);
  }

  /**
   * Record a counter metric
   */
  counter(name: string, value: number = 1, labels?: Record<string, string>): void {
    if (!this.config.enabled) return;

    const current = this.counters.get(name) || 0;
    const newValue = current + value;
    this.counters.set(name, newValue);

    this.recordMetric({
      name,
      type: MetricType.COUNTER,
      value: newValue,
      timestamp: Date.now(),
      labels: { ...this.config.defaultLabels, ...labels },
    });
  }

  /**
   * Record a gauge metric
   */
  gauge(name: string, value: number, labels?: Record<string, string>): void {
    if (!this.config.enabled) return;

    this.gauges.set(name, value);

    this.recordMetric({
      name,
      type: MetricType.GAUGE,
      value,
      timestamp: Date.now(),
      labels: { ...this.config.defaultLabels, ...labels },
    });
  }

  /**
   * Record a histogram metric
   */
  histogram(name: string, value: number, labels?: Record<string, string>): void {
    if (!this.config.enabled) return;

    const existing = this.histograms.get(name) || [];
    existing.push(value);
    this.histograms.set(name, existing);

    this.recordMetric({
      name,
      type: MetricType.HISTOGRAM,
      value,
      timestamp: Date.now(),
      labels: { ...this.config.defaultLabels, ...labels },
    });
  }

  /**
   * Record timing metric (convenience method for histogram)
   */
  timing(name: string, duration: number, labels?: Record<string, string>): void {
    this.histogram(`${name}_duration_ms`, duration, labels);
  }

  /**
   * Create a timer function
   */
  timer(name: string, labels?: Record<string, string>): () => void {
    const startTime = Date.now();
    return () => {
      const duration = Date.now() - startTime;
      this.timing(name, duration, labels);
    };
  }

  /**
   * Record custom metric
   */
  recordMetric(metric: Omit<MetricPoint, 'timestamp'> & { timestamp?: number }): void {
    if (!this.config.enabled) return;

    const point: MetricPoint = {
      ...metric,
      timestamp: metric.timestamp ?? Date.now(),
      labels: { ...this.config.defaultLabels, ...metric.labels },
    };

    this.metrics.push(point);

    // Enforce max points limit
    if (this.metrics.length > this.config.maxPoints) {
      this.metrics = this.metrics.slice(-this.config.maxPoints);
    }
  }

  /**
   * Get all recorded metrics
   */
  getMetrics(): MetricPoint[] {
    return [...this.metrics];
  }

  /**
   * Get metrics by name
   */
  getMetricsByName(name: string): MetricPoint[] {
    return this.metrics.filter(m => m.name === name);
  }

  /**
   * Get metric aggregation
   */
  getAggregation(name: string): MetricAggregation | null {
    const points = this.getMetricsByName(name);
    if (points.length === 0) return null;

    const values = points.map(p => p.value).sort((a, b) => a - b);
    const count = values.length;
    const sum = values.reduce((a, b) => a + b, 0);
    const avg = sum / count;
    const min = values[0];
    const max = values[count - 1];
    
    const variance = values.reduce((acc, val) => acc + Math.pow(val - avg, 2), 0) / count;
    const stdDev = Math.sqrt(variance);
    
    const getPercentile = (p: number) => {
      const index = Math.ceil((p / 100) * count) - 1;
      return values[Math.max(0, index)];
    };
    
    return {
      name,
      count,
      sum,
      avg,
      min,
      max,
      stdDev,
      percentiles: {
        p50: getPercentile(50),
        p90: getPercentile(90),
        p95: getPercentile(95),
        p99: getPercentile(99),
      },
    };
  }

  /**
   * Export metrics using all configured exporters
   */
  async exportMetrics(): Promise<void> {
    if (this.metrics.length === 0) return;

    const promises = this.exporters.map(async exporter => {
      try {
        await exporter.export([...this.metrics]);
      } catch (error) {
        this.logger.error(`Failed to export metrics with ${exporter.name}`, {
          error: error instanceof Error ? error.message : 'Unknown error'
        });
      }
    });

    await Promise.allSettled(promises);
  }

  /**
   * Clean up old metrics
   */
  private cleanup(): void {
    const cutoff = Date.now() - this.config.retentionPeriod;
    const before = this.metrics.length;
    
    this.metrics = this.metrics.filter(m => m.timestamp >= cutoff);
    
    const removed = before - this.metrics.length;
    if (removed > 0) {
      this.logger.debug(`Cleaned up ${removed} old metrics`);
    }
  }

  /**
   * Get current metrics summary
   */
  getSummary(): {
    totalMetrics: number;
    counters: number;
    gauges: number;
    histograms: number;
    oldestTimestamp?: number;
    newestTimestamp?: number;
  } {
    const byType = new Map<MetricType, number>();
    let oldest = Number.MAX_SAFE_INTEGER;
    let newest = 0;

    for (const metric of this.metrics) {
      byType.set(metric.type, (byType.get(metric.type) || 0) + 1);
      oldest = Math.min(oldest, metric.timestamp);
      newest = Math.max(newest, metric.timestamp);
    }

    return {
      totalMetrics: this.metrics.length,
      counters: byType.get(MetricType.COUNTER) || 0,
      gauges: byType.get(MetricType.GAUGE) || 0,
      histograms: byType.get(MetricType.HISTOGRAM) || 0,
      oldestTimestamp: oldest === Number.MAX_SAFE_INTEGER ? undefined : oldest,
      newestTimestamp: newest === 0 ? undefined : newest,
    };
  }
}
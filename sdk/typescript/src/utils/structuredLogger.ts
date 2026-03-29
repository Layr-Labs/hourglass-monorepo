// Enhanced structured logging with correlation IDs and context

import winston from 'winston';
import { ILogger } from './logger';

/**
 * Log levels
 */
export enum LogLevel {
  ERROR = 'error',
  WARN = 'warn',
  INFO = 'info',
  DEBUG = 'debug',
  TRACE = 'trace',
}

/**
 * Log context interface
 */
export interface LogContext {
  /** Correlation ID for tracing across services */
  correlationId?: string;
  /** Request ID for tracking individual requests */
  requestId?: string;
  /** Task ID for task-specific logging */
  taskId?: string;
  /** User ID or performer ID */
  performerId?: string;
  /** Component or module name */
  component?: string;
  /** Operation being performed */
  operation?: string;
  /** Additional custom fields */
  [key: string]: any;
}

/**
 * Log entry structure
 */
export interface LogEntry {
  /** Log level */
  level: LogLevel;
  /** Log message */
  message: string;
  /** Timestamp */
  timestamp: string;
  /** Log context */
  context: LogContext;
  /** Stack trace for errors */
  stack?: string;
  /** Additional metadata */
  metadata?: any;
}

/**
 * Log appender interface
 */
export interface LogAppender {
  /** Appender name */
  name: string;
  /** Write log entry */
  write(entry: LogEntry): Promise<void>;
}

/**
 * Console log appender with structured output
 */
export class ConsoleLogAppender implements LogAppender {
  name = 'console';

  async write(entry: LogEntry): Promise<void> {
    const colorMap = {
      [LogLevel.ERROR]: '\x1b[31m', // Red
      [LogLevel.WARN]: '\x1b[33m',  // Yellow
      [LogLevel.INFO]: '\x1b[36m',  // Cyan
      [LogLevel.DEBUG]: '\x1b[37m', // White
      [LogLevel.TRACE]: '\x1b[90m', // Gray
    };

    const reset = '\x1b[0m';
    const color = colorMap[entry.level] || reset;

    // Format the main log line
    const time = new Date(entry.timestamp).toISOString();
    const level = entry.level.toUpperCase().padEnd(5);
    const component = entry.context.component ? `[${entry.context.component}]` : '';
    const correlation = entry.context.correlationId ? `{${entry.context.correlationId.slice(0, 8)}}` : '';
    
    console.log(`${color}${time} ${level}${reset} ${component}${correlation} ${entry.message}`);

    // Add context details if present
    const contextKeys = Object.keys(entry.context).filter(k => 
      !['component', 'correlationId'].includes(k) && entry.context[k] !== undefined
    );
    
    if (contextKeys.length > 0) {
      const contextStr = contextKeys.map(k => `${k}=${entry.context[k]}`).join(' ');
      console.log(`  └─ ${contextStr}`);
    }

    // Add metadata if present
    if (entry.metadata && Object.keys(entry.metadata).length > 0) {
      console.log(`  └─ ${JSON.stringify(entry.metadata)}`);
    }

    // Add stack trace for errors
    if (entry.stack) {
      console.log(`  └─ Stack:\n${entry.stack.split('\n').map(line => `     ${line}`).join('\n')}`);
    }
  }
}

/**
 * File log appender with JSON format
 */
export class FileLogAppender implements LogAppender {
  name = 'file';

  constructor(private filePath: string) {}

  async write(entry: LogEntry): Promise<void> {
    const fs = await import('fs/promises');
    const logLine = JSON.stringify(entry) + '\n';
    
    try {
      await fs.appendFile(this.filePath, logLine);
    } catch (error) {
      // Fallback to console if file write fails
      console.error('Failed to write to log file:', error);
      console.log(logLine);
    }
  }
}

/**
 * Structured logger configuration
 */
export interface StructuredLoggerConfig {
  /** Log level */
  level?: LogLevel;
  /** Default context to include in all logs */
  defaultContext?: LogContext;
  /** Enable/disable logging */
  enabled?: boolean;
  /** Log appenders */
  appenders?: LogAppender[];
  /** Include stack traces for warnings and errors */
  includeStack?: boolean;
}

/**
 * Enhanced structured logger
 */
export class StructuredLogger implements ILogger {
  private config: Required<StructuredLoggerConfig>;
  private appenders: LogAppender[] = [];
  private correlationIdGenerator: () => string;

  constructor(config: StructuredLoggerConfig = {}) {
    this.config = {
      level: config.level ?? LogLevel.INFO,
      defaultContext: config.defaultContext ?? {},
      enabled: config.enabled ?? true,
      appenders: config.appenders ?? [new ConsoleLogAppender()],
      includeStack: config.includeStack ?? true,
    };

    this.appenders = [...this.config.appenders];
    this.correlationIdGenerator = this.createCorrelationIdGenerator();
  }

  /**
   * Create a child logger with additional context
   */
  child(context: LogContext): StructuredLogger {
    return new StructuredLogger({
      ...this.config,
      defaultContext: { ...this.config.defaultContext, ...context },
    });
  }

  /**
   * Generate a new correlation ID
   */
  generateCorrelationId(): string {
    return this.correlationIdGenerator();
  }

  /**
   * Add log appender
   */
  addAppender(appender: LogAppender): void {
    this.appenders.push(appender);
  }

  /**
   * Remove log appender
   */
  removeAppender(name: string): void {
    this.appenders = this.appenders.filter(a => a.name !== name);
  }

  /**
   * Log an error
   */
  error(message: string, meta?: any): void {
    this.log(LogLevel.ERROR, message, meta);
  }

  /**
   * Log a warning
   */
  warn(message: string, meta?: any): void {
    this.log(LogLevel.WARN, message, meta);
  }

  /**
   * Log info
   */
  info(message: string, meta?: any): void {
    this.log(LogLevel.INFO, message, meta);
  }

  /**
   * Log debug information
   */
  debug(message: string, meta?: any): void {
    this.log(LogLevel.DEBUG, message, meta);
  }

  /**
   * Log trace information
   */
  trace(message: string, meta?: any): void {
    this.log(LogLevel.TRACE, message, meta);
  }

  /**
   * Log with specific level and context
   */
  logWithContext(level: LogLevel, message: string, context: LogContext, meta?: any): void {
    this.log(level, message, meta, context);
  }

  /**
   * Core logging method
   */
  private log(level: LogLevel, message: string, meta?: any, additionalContext?: LogContext): void {
    if (!this.config.enabled || !this.shouldLog(level)) {
      return;
    }

    // Extract error stack if meta is an Error
    let stack: string | undefined;
    let metadata = meta;
    
    if (meta instanceof Error) {
      stack = meta.stack;
      metadata = {
        errorName: meta.name,
        errorMessage: meta.message,
        ...meta,
      };
    } else if (this.config.includeStack && (level === LogLevel.ERROR || level === LogLevel.WARN)) {
      // Capture stack trace for errors and warnings
      const error = new Error();
      stack = error.stack?.split('\n').slice(3).join('\n'); // Remove the first 3 lines (Error creation)
    }

    const entry: LogEntry = {
      level,
      message,
      timestamp: new Date().toISOString(),
      context: {
        ...this.config.defaultContext,
        ...additionalContext,
        // Add correlation ID if not present
        correlationId: additionalContext?.correlationId || 
                      this.config.defaultContext.correlationId || 
                      this.generateCorrelationId(),
      },
      stack,
      metadata,
    };

    // Write to all appenders
    this.writeToAppenders(entry);
  }

  /**
   * Check if we should log at this level
   */
  private shouldLog(level: LogLevel): boolean {
    const levels = [LogLevel.ERROR, LogLevel.WARN, LogLevel.INFO, LogLevel.DEBUG, LogLevel.TRACE];
    const currentIndex = levels.indexOf(this.config.level);
    const messageIndex = levels.indexOf(level);
    
    return messageIndex <= currentIndex;
  }

  /**
   * Write log entry to all appenders
   */
  private async writeToAppenders(entry: LogEntry): Promise<void> {
    const promises = this.appenders.map(async appender => {
      try {
        await appender.write(entry);
      } catch (error) {
        // Fallback: log to console if appender fails
        console.error(`Log appender ${appender.name} failed:`, error);
        console.log(JSON.stringify(entry));
      }
    });

    // Don't await to avoid blocking the calling code
    Promise.allSettled(promises);
  }

  /**
   * Create correlation ID generator
   */
  private createCorrelationIdGenerator(): () => string {
    let counter = 0;
    const instanceId = Math.random().toString(36).substr(2, 4);
    
    return () => {
      counter = (counter + 1) % 10000;
      const timestamp = Date.now().toString(36);
      return `${instanceId}-${timestamp}-${counter.toString().padStart(4, '0')}`;
    };
  }

  /**
   * Create request-scoped logger
   */
  static forRequest(requestId: string, additionalContext?: LogContext): StructuredLogger {
    return new StructuredLogger({
      defaultContext: {
        requestId,
        ...additionalContext,
      },
    });
  }

  /**
   * Create task-scoped logger
   */
  static forTask(taskId: string, additionalContext?: LogContext): StructuredLogger {
    return new StructuredLogger({
      defaultContext: {
        taskId,
        component: 'task-processor',
        ...additionalContext,
      },
    });
  }

  /**
   * Create component-scoped logger
   */
  static forComponent(component: string, additionalContext?: LogContext): StructuredLogger {
    return new StructuredLogger({
      defaultContext: {
        component,
        ...additionalContext,
      },
    });
  }
}
// Logger utility for the Hourglass TypeScript SDK
import winston from 'winston';

/**
 * Logger configuration options
 */
export interface LoggerConfig {
  /** Enable debug logging */
  debug?: boolean;
  /** Log level (default: 'info') */
  level?: string;
  /** Enable console logging (default: true) */
  console?: boolean;
  /** Log file path (optional) */
  file?: string;
}

/**
 * Create a configured winston logger
 */
export function createLogger(config: LoggerConfig = {}): winston.Logger {
  const { debug = false, level = debug ? 'debug' : 'info', console = true, file } = config;

  const transports: winston.transport[] = [];

  // Console transport
  if (console) {
    transports.push(
      new winston.transports.Console({
        format: winston.format.combine(
          winston.format.colorize(),
          winston.format.timestamp(),
          winston.format.printf(({ timestamp, level, message, ...meta }) => {
            const metaStr = Object.keys(meta).length > 0 ? ` ${JSON.stringify(meta)}` : '';
            return `${timestamp} [${level}]: ${message}${metaStr}`;
          })
        ),
      })
    );
  }

  // File transport
  if (file) {
    transports.push(
      new winston.transports.File({
        filename: file,
        format: winston.format.combine(
          winston.format.timestamp(),
          winston.format.json()
        ),
      })
    );
  }

  return winston.createLogger({
    level,
    transports,
    exitOnError: false,
  });
}

/**
 * Default logger instance
 */
export const logger = createLogger();

/**
 * Logger interface for dependency injection
 */
export interface ILogger {
  error(message: string, meta?: any): void;
  warn(message: string, meta?: any): void;
  info(message: string, meta?: any): void;
  debug(message: string, meta?: any): void;
}
import winston from 'winston';
import chalk from 'chalk';

export interface Logger {
  info(message: string, ...args: any[]): void;
  warn(message: string, ...args: any[]): void;
  error(message: string, ...args: any[]): void;
  debug(message: string, ...args: any[]): void;
  title(message: string, ...args: any[]): void;
}

class WinstonLogger implements Logger {
  private winston: winston.Logger;

  constructor(verbose: boolean = false) {
    const format = winston.format.combine(
      winston.format.timestamp({ format: 'HH:mm:ss' }),
      winston.format.printf(({ level, message, timestamp }) => {
        const color = {
          info: chalk.blue,
          warn: chalk.yellow,
          error: chalk.red,
          debug: chalk.gray
        }[level] || chalk.white;
        
        return color(`[${timestamp}] ${level.toUpperCase()}: ${message}`);
      })
    );

    this.winston = winston.createLogger({
      level: verbose ? 'debug' : 'info',
      format,
      transports: [
        new winston.transports.Console()
      ]
    });
  }

  info(message: string, ...args: any[]): void {
    this.winston.info(this.format(message, args));
  }

  warn(message: string, ...args: any[]): void {
    this.winston.warn(this.format(message, args));
  }

  error(message: string, ...args: any[]): void {
    this.winston.error(this.format(message, args));
  }

  debug(message: string, ...args: any[]): void {
    this.winston.debug(this.format(message, args));
  }

  title(message: string, ...args: any[]): void {
    console.log(chalk.bold.cyan(`\n${this.format(message, args)}\n`));
  }

  private format(message: string, args: any[]): string {
    return args.length > 0 ? message.replace(/%s/g, () => args.shift()) : message;
  }
}

export function setupLogger(verbose: boolean = false): Logger {
  return new WinstonLogger(verbose);
}
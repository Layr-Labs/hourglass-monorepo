// Environment configuration utilities for Hourglass performers

/**
 * Environment configuration interface
 */
export interface EnvironmentConfig {
  // Server configuration
  port: number;
  host: string;
  timeout: number;
  
  // Application configuration
  nodeEnv: string;
  debug: boolean;
  logLevel: string;
  
  // Performance configuration
  maxConcurrentTasks: number;
  taskTimeout: number;
  
  // Health and monitoring
  healthCheckInterval: number;
  metricsEnabled: boolean;
  metricsPort: number;
  
  // TypeChain configuration
  typechainEnabled: boolean;
  contractsPath: string;
  
  // Security configuration
  corsEnabled: boolean;
  corsOrigins: string[];
  
  // External services
  aggregatorUrl?: string;
  executorUrl?: string;
  
  // Storage configuration
  dataDir: string;
  logsDir: string;
}

/**
 * Load and validate environment configuration
 */
export function loadEnvironmentConfig(): EnvironmentConfig {
  const config: EnvironmentConfig = {
    // Server configuration
    port: parseInt(process.env.PORT || '8080', 10),
    host: process.env.HOST || '0.0.0.0',
    timeout: parseInt(process.env.TIMEOUT || '10000', 10),
    
    // Application configuration
    nodeEnv: process.env.NODE_ENV || 'development',
    debug: process.env.DEBUG === 'true' || process.env.NODE_ENV === 'development',
    logLevel: process.env.LOG_LEVEL || 'info',
    
    // Performance configuration
    maxConcurrentTasks: parseInt(process.env.MAX_CONCURRENT_TASKS || '10', 10),
    taskTimeout: parseInt(process.env.TASK_TIMEOUT || '30000', 10),
    
    // Health and monitoring
    healthCheckInterval: parseInt(process.env.HEALTH_CHECK_INTERVAL || '30000', 10),
    metricsEnabled: process.env.METRICS_ENABLED !== 'false',
    metricsPort: parseInt(process.env.METRICS_PORT || '9090', 10),
    
    // TypeChain configuration
    typechainEnabled: process.env.TYPECHAIN_ENABLED === 'true',
    contractsPath: process.env.CONTRACTS_PATH || './contracts',
    
    // Security configuration
    corsEnabled: process.env.CORS_ENABLED === 'true',
    corsOrigins: process.env.CORS_ORIGINS?.split(',') || ['*'],
    
    // External services
    aggregatorUrl: process.env.AGGREGATOR_URL,
    executorUrl: process.env.EXECUTOR_URL,
    
    // Storage configuration
    dataDir: process.env.DATA_DIR || './data',
    logsDir: process.env.LOGS_DIR || './logs',
  };
  
  // Validate configuration
  validateConfig(config);
  
  return config;
}

/**
 * Validate environment configuration
 */
function validateConfig(config: EnvironmentConfig): void {
  const errors: string[] = [];
  
  // Validate port
  if (config.port < 1 || config.port > 65535) {
    errors.push('PORT must be between 1 and 65535');
  }
  
  // Validate timeout
  if (config.timeout < 1000) {
    errors.push('TIMEOUT must be at least 1000ms');
  }
  
  // Validate task timeout
  if (config.taskTimeout < 1000) {
    errors.push('TASK_TIMEOUT must be at least 1000ms');
  }
  
  // Validate max concurrent tasks
  if (config.maxConcurrentTasks < 1 || config.maxConcurrentTasks > 1000) {
    errors.push('MAX_CONCURRENT_TASKS must be between 1 and 1000');
  }
  
  // Validate log level
  const validLogLevels = ['error', 'warn', 'info', 'debug', 'trace'];
  if (!validLogLevels.includes(config.logLevel)) {
    errors.push(`LOG_LEVEL must be one of: ${validLogLevels.join(', ')}`);
  }
  
  // Validate node environment
  const validNodeEnvs = ['development', 'production', 'test'];
  if (!validNodeEnvs.includes(config.nodeEnv)) {
    errors.push(`NODE_ENV must be one of: ${validNodeEnvs.join(', ')}`);
  }
  
  if (errors.length > 0) {
    throw new Error(`Environment configuration validation failed:\n${errors.join('\n')}`);
  }
}

/**
 * Get configuration for specific environment
 */
export function getProductionConfig(): Partial<EnvironmentConfig> {
  return {
    debug: false,
    logLevel: 'info',
    metricsEnabled: true,
    healthCheckInterval: 30000,
    maxConcurrentTasks: 50,
    corsEnabled: true,
    corsOrigins: [], // Must be explicitly set in production
  };
}

/**
 * Get configuration for development
 */
export function getDevelopmentConfig(): Partial<EnvironmentConfig> {
  return {
    debug: true,
    logLevel: 'debug',
    metricsEnabled: true,
    healthCheckInterval: 15000,
    maxConcurrentTasks: 5,
    corsEnabled: true,
    corsOrigins: ['*'],
  };
}

/**
 * Get configuration for testing
 */
export function getTestConfig(): Partial<EnvironmentConfig> {
  return {
    debug: false,
    logLevel: 'error',
    metricsEnabled: false,
    healthCheckInterval: 60000,
    maxConcurrentTasks: 1,
    corsEnabled: false,
    timeout: 5000,
    taskTimeout: 5000,
  };
}

/**
 * Print configuration summary (safe for logging)
 */
export function printConfigSummary(config: EnvironmentConfig): void {
  console.log('🔧 Hourglass Performer Configuration:');
  console.log(`  Environment: ${config.nodeEnv}`);
  console.log(`  Port: ${config.port}`);
  console.log(`  Debug: ${config.debug}`);
  console.log(`  Log Level: ${config.logLevel}`);
  console.log(`  Max Concurrent Tasks: ${config.maxConcurrentTasks}`);
  console.log(`  Task Timeout: ${config.taskTimeout}ms`);
  console.log(`  Metrics: ${config.metricsEnabled ? 'enabled' : 'disabled'}`);
  console.log(`  TypeChain: ${config.typechainEnabled ? 'enabled' : 'disabled'}`);
  console.log(`  CORS: ${config.corsEnabled ? 'enabled' : 'disabled'}`);
  console.log(`  Data Directory: ${config.dataDir}`);
  console.log(`  Logs Directory: ${config.logsDir}`);
}
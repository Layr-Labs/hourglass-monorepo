# Hourglass TypeScript Performer SDK - API Documentation

Complete API reference for the Hourglass TypeScript Performer SDK.

## Table of Contents

- [Core Classes](#core-classes)
- [Worker Classes](#worker-classes)
- [Server Classes](#server-classes)
- [Utility Functions](#utility-functions)
- [Type Definitions](#type-definitions)
- [Configuration](#configuration)
- [Error Handling](#error-handling)

## Core Classes

### BaseWorker

The base class for implementing simple task processors.

#### Constructor

```typescript
class BaseWorker implements IWorker
```

#### Methods

##### `handleSimpleTask(input: any): Promise<any>`

**Abstract method** that must be implemented by subclasses.

```typescript
abstract handleSimpleTask(input: any): Promise<any>
```

**Parameters:**
- `input` - The parsed input from the task payload

**Returns:** `Promise<any>` - The result to be encoded and returned

**Example:**
```typescript
class NumberProcessor extends BaseWorker {
  async handleSimpleTask(input: number) {
    return input * input;
  }
}
```

##### `handleTask(task: TaskRequest): Promise<TaskResponse>`

Internal method that handles the full task lifecycle. Can be overridden for advanced use cases.

```typescript
async handleTask(task: TaskRequest): Promise<TaskResponse>
```

**Parameters:**
- `task` - The complete task request

**Returns:** `Promise<TaskResponse>` - The formatted task response

##### `validateTask(task: TaskRequest): Promise<void>`

Validates the incoming task request. Override for custom validation.

```typescript
async validateTask(task: TaskRequest): Promise<void>
```

**Parameters:**
- `task` - The task request to validate

**Throws:** `Error` if validation fails

**Example:**
```typescript
class ValidatedWorker extends BaseWorker {
  async validateTask(task: TaskRequest) {
    await super.validateTask(task);
    
    const input = this.parsePayload(task.payload);
    if (typeof input !== 'number') {
      throw new Error('Input must be a number');
    }
  }
}
```

##### `start(port?: number): Promise<void>`

Starts the performer server with automatic configuration.

```typescript
async start(port?: number): Promise<void>
```

**Parameters:**
- `port` - Optional port override (defaults to environment configuration)

**Example:**
```typescript
const worker = new MyWorker();
await worker.start(8080);
```

#### Protected Methods

##### `parsePayload(payload: Uint8Array): any`

Parses the task payload from bytes to JavaScript object.

```typescript
protected parsePayload(payload: Uint8Array): any
```

##### `encodePayload(result: any): Uint8Array`

Encodes the result to bytes for transmission.

```typescript
protected encodePayload(result: any): Uint8Array
```

##### `createResponse(taskId: string, result: Uint8Array): TaskResponse`

Creates a properly formatted task response.

```typescript
protected createResponse(taskId: string, result: Uint8Array): TaskResponse
```

---

### SolidityWorker

Extended worker class for type-safe Solidity contract interaction.

#### Constructor

```typescript
class SolidityWorker<TContract = any, TFunction extends keyof TContract = keyof TContract> extends BaseWorker
```

**Generic Parameters:**
- `TContract` - TypeChain-generated contract interface
- `TFunction` - Function name from the contract

**Example:**
```typescript
import type { MyContract } from './typechain-types';

class ContractWorker extends SolidityWorker<MyContract, 'processTask'> {
  constructor() {
    super({
      abi: contractAbi,
      functionName: 'processTask',
      autoDetectPayload: true,
      strictMode: false,
    });
  }
}
```

#### Configuration

```typescript
interface SolidityWorkerConfig {
  abi: any[];                    // Contract ABI JSON
  functionName: string;          // Function name to decode/encode
  autoDetectPayload?: boolean;   // Auto-detect payload format (default: true)
  strictMode?: boolean;          // Strict mode - throw on decoding errors (default: false)
}
```

#### Methods

##### `handleSolidityTask(params: any): Promise<any>`

**Abstract method** for handling decoded Solidity parameters.

```typescript
abstract handleSolidityTask(params: any): Promise<any>
```

**Parameters:**
- `params` - Decoded ABI parameters (typed if using TypeChain)

**Returns:** `Promise<any>` - Result to be ABI-encoded

**Example:**
```typescript
async handleSolidityTask(params: { amount: bigint; user: string }) {
  const { amount, user } = params;
  
  // Validate
  if (amount <= 0n) {
    throw new Error('Invalid amount');
  }
  
  // Process
  return {
    result: amount * 2n,
    success: true,
  };
}
```

#### Protected Methods

##### `decodeTaskPayload(payload: Uint8Array): any`

Decodes task payload using ABI or auto-detection.

```typescript
protected decodeTaskPayload(payload: Uint8Array): any
```

##### `encodeTaskResult(result: any): Uint8Array`

Encodes result using ABI.

```typescript
protected encodeTaskResult(result: any): Uint8Array
```

##### `getFunctionInfo(): AbiFunctionInfo | undefined`

Gets function information from ABI.

```typescript
protected getFunctionInfo(): AbiFunctionInfo | undefined
```

##### `getAllFunctions(): AbiFunctionInfo[]`

Gets all available functions from ABI.

```typescript
protected getAllFunctions(): AbiFunctionInfo[]
```

---

## Server Classes

### PerformerServer

The main gRPC server implementation.

#### Constructor

```typescript
class PerformerServer implements PerformerService
```

**Parameters:**
```typescript
constructor(
  worker: IWorker,
  config?: PerformerServerConfig,
  taskProcessorConfig?: TaskProcessorConfig,
  healthConfig?: HealthCheckConfig,
  metricsConfig?: MetricsConfig
)
```

#### Configuration

```typescript
interface PerformerServerConfig {
  port?: number;     // Server port (default: 8080)
  timeout?: number;  // Request timeout (default: 5000ms)
  debug?: boolean;   // Debug mode (default: false)
}
```

#### Methods

##### `start(): Promise<void>`

Starts the gRPC server.

```typescript
async start(): Promise<void>
```

##### `stop(): Promise<void>`

Stops the gRPC server gracefully.

```typescript
async stop(): Promise<void>
```

##### `setupGracefulShutdown(): void`

Sets up graceful shutdown handlers for SIGTERM/SIGINT.

```typescript
setupGracefulShutdown(): void
```

##### `addHealthProvider(provider: HealthProvider): void`

Adds a custom health check provider.

```typescript
addHealthProvider(provider: HealthProvider): void
```

##### `addMetricsExporter(exporter: MetricsExporter): void`

Adds a custom metrics exporter.

```typescript
addMetricsExporter(exporter: MetricsExporter): void
```

#### gRPC Service Methods

##### `ExecuteTask(request: TaskRequest): Promise<TaskResponse>`

Executes a task via gRPC.

```typescript
async ExecuteTask(request: TaskRequest): Promise<TaskResponse>
```

##### `HealthCheck(request: HealthCheckRequest): Promise<HealthCheckResponse>`

Performs a health check via gRPC.

```typescript
async HealthCheck(request: HealthCheckRequest): Promise<HealthCheckResponse>
```

##### `StartSync(request: StartSyncRequest): Promise<StartSyncResponse>`

Starts synchronization via gRPC.

```typescript
async StartSync(request: StartSyncRequest): Promise<StartSyncResponse>
```

---

## Utility Functions

### Environment Configuration

#### `loadEnvironmentConfig(): EnvironmentConfig`

Loads and validates environment configuration.

```typescript
function loadEnvironmentConfig(): EnvironmentConfig
```

**Returns:** `EnvironmentConfig` - Validated configuration object

**Example:**
```typescript
import { loadEnvironmentConfig } from '@layr-labs/hourglass-performer';

const config = loadEnvironmentConfig();
console.log(`Server will run on port ${config.port}`);
```

#### `printConfigSummary(config: EnvironmentConfig): void`

Prints a safe configuration summary.

```typescript
function printConfigSummary(config: EnvironmentConfig): void
```

#### `getProductionConfig(): Partial<EnvironmentConfig>`

Gets production-specific configuration defaults.

```typescript
function getProductionConfig(): Partial<EnvironmentConfig>
```

#### `getDevelopmentConfig(): Partial<EnvironmentConfig>`

Gets development-specific configuration defaults.

```typescript
function getDevelopmentConfig(): Partial<EnvironmentConfig>
```

### ABI Utilities

#### `AbiCodec`

Class for encoding/decoding Solidity function calls.

```typescript
class AbiCodec {
  constructor(abi: any[])
  
  getFunctionInfo(functionName: string): AbiFunctionInfo | undefined
  getAllFunctions(): AbiFunctionInfo[]
  decodeFunctionCall(functionName: string, data: Uint8Array): any
  encodeFunctionCall(functionName: string, params: any[]): Uint8Array
  decodeFunctionResult(functionName: string, data: Uint8Array): any
  encodeFunctionResult(functionName: string, result: any): Uint8Array
  detectFunction(data: Uint8Array): string | null
}
```

#### `SolidityTypeUtils`

Utilities for Solidity type conversion.

```typescript
class SolidityTypeUtils {
  static toSolidityType(value: any, solidityType: string): any
  static fromSolidityType(value: any, solidityType: string): any
  static validateType(value: any, solidityType: string): boolean
}
```

#### `PayloadAutoDecoder`

Auto-detects payload format and decodes accordingly.

```typescript
class PayloadAutoDecoder {
  static decode(payload: Uint8Array): {
    format: 'abi' | 'json' | 'string' | 'raw';
    data: any;
    confidence: number;
  }
}
```

### SolidityWorker Utilities

#### `SolidityWorkerUtils`

Utility functions for creating SolidityWorker instances.

```typescript
class SolidityWorkerUtils {
  static createFromAbi<T = any>(
    abi: any[],
    functionName: string,
    handler: (params: T) => Promise<any>
  ): JsonSolidityWorker
  
  static createTyped<TContract, TFunction extends keyof TContract>(
    config: SolidityWorkerConfig,
    handler: (params: ExtractFunctionParams<TContract, TFunction>[0]) => Promise<ExtractFunctionReturn<TContract, TFunction>>
  ): SolidityWorker<TContract, TFunction>
}
```

**Example:**
```typescript
// Quick worker creation
const worker = SolidityWorkerUtils.createFromAbi(
  contractAbi,
  'processData',
  async (params) => {
    return { result: params.value * 2 };
  }
);
```

---

## Type Definitions

### Core Types

#### `TaskRequest`

```typescript
interface TaskRequest {
  taskId: string;      // Unique task identifier
  payload: Uint8Array; // Task payload as bytes
}
```

#### `TaskResponse`

```typescript
interface TaskResponse {
  taskId: string;      // Task identifier (must match request)
  result: Uint8Array;  // Task result as bytes
}
```

#### `PerformerStatus`

```typescript
enum PerformerStatus {
  READY_FOR_TASK = 'READY_FOR_TASK',
  BUSY = 'BUSY',
  ERROR = 'ERROR',
  STARTING = 'STARTING',
  STOPPING = 'STOPPING'
}
```

#### `HealthCheckRequest`

```typescript
interface HealthCheckRequest {
  // Empty request
}
```

#### `HealthCheckResponse`

```typescript
interface HealthCheckResponse {
  status: PerformerStatus;
}
```

### Worker Interface

#### `IWorker`

```typescript
interface IWorker {
  validateTask(task: TaskRequest): Promise<void>;
  handleTask(task: TaskRequest): Promise<TaskResponse>;
}
```

### Environment Configuration

#### `EnvironmentConfig`

```typescript
interface EnvironmentConfig {
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
```

### ABI Types

#### `AbiFunctionInfo`

```typescript
interface AbiFunctionInfo {
  name: string;                    // Function name
  signature: string;               // Function signature
  inputs: readonly ParamType[];    // Input parameter types
  outputs: readonly ParamType[];   // Output parameter types
  selector: string;                // Function selector (first 4 bytes)
}
```

### TypeChain Integration

#### `ExtractFunctionParams<T, K>`

Type helper for extracting function parameters from TypeChain-generated types.

```typescript
type ExtractFunctionParams<T, K extends keyof T> = T[K] extends (...args: infer P) => any ? P : never;
```

#### `ExtractFunctionReturn<T, K>`

Type helper for extracting function return type from TypeChain-generated types.

```typescript
type ExtractFunctionReturn<T, K extends keyof T> = T[K] extends (...args: any[]) => Promise<infer R> ? R : never;
```

---

## Configuration

### Environment Variables

All configuration can be set via environment variables:

```bash
# Server
PORT=8080
HOST=0.0.0.0
TIMEOUT=10000

# Application
NODE_ENV=production
DEBUG=false
LOG_LEVEL=info

# Performance
MAX_CONCURRENT_TASKS=10
TASK_TIMEOUT=30000

# Health & Monitoring
HEALTH_CHECK_INTERVAL=30000
METRICS_ENABLED=true
METRICS_PORT=9090

# TypeChain
TYPECHAIN_ENABLED=true
CONTRACTS_PATH=./contracts

# Security
CORS_ENABLED=true
CORS_ORIGINS=https://yourdomain.com

# Storage
DATA_DIR=./data
LOGS_DIR=./logs
```

### Configuration Files

#### `.env.example`

Template for environment variables.

#### `.env.production`

Production environment defaults.

#### `typechain.config.json`

TypeChain configuration for contract type generation.

```json
{
  "files": ["contracts/**/*.json"],
  "target": "ethers-v6",
  "outDir": "typechain-types"
}
```

---

## Error Handling

### Error Types

#### `TaskValidationError`

Thrown when task validation fails.

```typescript
class TaskValidationError extends Error {
  constructor(message: string, taskId?: string)
}
```

#### `TaskExecutionError`

Thrown when task execution fails.

```typescript
class TaskExecutionError extends Error {
  constructor(message: string, taskId?: string)
}
```

### Error Handling Patterns

#### Basic Error Handling

```typescript
class SafeWorker extends BaseWorker {
  async handleSimpleTask(input: any) {
    try {
      return await this.processInput(input);
    } catch (error) {
      console.error('Processing failed:', error);
      throw new Error(`Processing failed: ${error.message}`);
    }
  }
}
```

#### Validation Error Handling

```typescript
class ValidatedWorker extends BaseWorker {
  async validateTask(task: TaskRequest) {
    await super.validateTask(task);
    
    const input = this.parsePayload(task.payload);
    if (typeof input !== 'number') {
      throw new TaskValidationError('Input must be a number', task.taskId);
    }
  }
}
```

#### Solidity Error Handling

```typescript
class ContractWorker extends SolidityWorker<MyContract, 'processTask'> {
  async handleSolidityTask(params: ProcessTaskParams) {
    try {
      // Validate parameters
      if (params.amount <= 0n) {
        throw new Error('Amount must be positive');
      }
      
      // Process task
      return await this.processContract(params);
    } catch (error) {
      console.error('Contract processing failed:', error);
      
      // Return error result instead of throwing
      return {
        success: false,
        error: error.message,
        result: 0n,
      };
    }
  }
}
```

---

## Best Practices

### Performance

1. **Use TypeChain** for type safety and better performance
2. **Implement proper validation** to catch errors early
3. **Use appropriate timeouts** for long-running tasks
4. **Monitor memory usage** in production

### Security

1. **Validate all inputs** before processing
2. **Use environment variables** for sensitive configuration
3. **Run containers as non-root user**
4. **Implement proper CORS** for web access

### Development

1. **Use strict TypeScript** configuration
2. **Write tests** for your performers
3. **Use structured logging** for debugging
4. **Implement health checks** for monitoring

### Deployment

1. **Use Docker** for consistent deployments
2. **Set resource limits** in production
3. **Implement proper logging** for troubleshooting
4. **Monitor performance metrics** in production

---

For more examples and advanced usage, see the [main README](../README.md) and [examples directory](../examples/).
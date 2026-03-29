// SolidityWorker base class for TypeChain integration and ABI handling

import { TaskRequest, TaskResponse } from '../types/performer';
import { BaseWorker } from './iWorker';
import { AbiCodec, PayloadAutoDecoder } from '../utils/abiUtils';
import { hexToBytes, bytesToHex } from '../types/protobuf';

/**
 * Configuration for SolidityWorker
 */
export interface SolidityWorkerConfig {
  /** Contract ABI JSON */
  abi: any[];
  /** Function name to decode/encode */
  functionName: string;
  /** Auto-detect payload format (default: true) */
  autoDetectPayload?: boolean;
  /** Strict mode - throw on decoding errors (default: false) */
  strictMode?: boolean;
}

/**
 * Type helper for extracting function parameters from TypeChain-generated types
 */
export type ExtractFunctionParams<T, K extends keyof T> = T[K] extends (...args: infer P) => any ? P : never;

/**
 * Type helper for extracting function return type from TypeChain-generated types
 */
export type ExtractFunctionReturn<T, K extends keyof T> = T[K] extends (...args: any[]) => Promise<infer R> ? R : never;

/**
 * SolidityWorker base class with TypeChain integration
 * 
 * Generic parameters:
 * - TContract: TypeChain-generated contract interface
 * - TFunction: Function name from the contract
 */
export abstract class SolidityWorker<
  TContract = any,
  TFunction extends keyof TContract = keyof TContract
> extends BaseWorker {
  protected abiCodec: AbiCodec | undefined;
  protected config: Required<SolidityWorkerConfig>;

  constructor(config?: SolidityWorkerConfig) {
    super();
    
    // Use default config if none provided (for simple usage)
    this.config = {
      autoDetectPayload: true,
      strictMode: false,
      abi: [],
      functionName: 'handleTask',
      ...config,
    };
    
    if (this.config.abi.length > 0) {
      this.abiCodec = new AbiCodec(this.config.abi);
    }
  }

  /**
   * Validate task with ABI decoding
   */
  async validateTask(task: TaskRequest): Promise<void> {
    await super.validateTask(task);

    // Validate ABI decoding
    try {
      this.decodeTaskPayload(task.payload);
    } catch (error) {
      if (this.config.strictMode) {
        throw new Error(`ABI validation failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
      }
      // In non-strict mode, just warn
      console.warn(`ABI validation warning: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  /**
   * Handle task with automatic ABI decoding/encoding
   */
  async handleTask(task: TaskRequest): Promise<TaskResponse> {
    try {
      // Decode payload using ABI (if available)
      const decodedParams = this.abiCodec ? this.decodeTaskPayload(task.payload) : this.parsePayload(task.payload);
      
      // Call user-implemented handler with decoded parameters
      const result = await this.handleSolidityTask(decodedParams);
      
      // Encode result using ABI (if available)
      const encodedResult = this.abiCodec ? this.encodeTaskResult(result) : this.encodePayload(result);
      
      return this.createResponse(task.taskId, encodedResult);
    } catch (error) {
      throw new Error(`Solidity task handling failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  /**
   * Abstract method for handling decoded Solidity parameters
   * Override this method in your implementation
   */
  abstract handleSolidityTask(params: any): Promise<any>;

  /**
   * Simple start method for one-line usage
   */
  async start(port: number = 8080): Promise<void> {
    const { PerformerServer } = await import('../server/performerServer');
    
    const server = new PerformerServer(this, {
      port,
      timeout: 10000,
      debug: true,
    });

    // Set up graceful shutdown
    server.setupGracefulShutdown();

    try {
      await server.start();
      console.log(`🚀 Performer is running on port ${port}!`);
    } catch (error) {
      console.error('❌ Failed to start server:', error);
      process.exit(1);
    }
  }

  /**
   * Decode task payload using ABI
   */
  protected decodeTaskPayload(payload: Uint8Array): any {
    if (this.config.autoDetectPayload) {
      const detection = PayloadAutoDecoder.decode(payload);
      
      if (detection.format === 'abi' && this.abiCodec) {
        return this.abiCodec.decodeFunctionCall(this.config.functionName, payload);
      } else if (detection.format === 'json') {
        // For JSON payloads, try to map to function parameters
        return this.mapJsonToAbiParams(detection.data);
      } else {
        // For other formats, pass through as raw data
        return { data: payload };
      }
    } else {
      // Direct ABI decoding
      return this.abiCodec ? this.abiCodec.decodeFunctionCall(this.config.functionName, payload) : { data: payload };
    }
  }

  /**
   * Encode task result using ABI
   */
  protected encodeTaskResult(result: any): Uint8Array {
    return this.abiCodec ? this.abiCodec.encodeFunctionResult(this.config.functionName, result) : this.encodePayload(result);
  }

  /**
   * Map JSON data to ABI function parameters
   */
  protected mapJsonToAbiParams(jsonData: any): any {
    if (!this.abiCodec) {
      return jsonData;
    }
    
    const funcInfo = this.abiCodec.getFunctionInfo(this.config.functionName);
    if (!funcInfo) {
      throw new Error(`Function ${this.config.functionName} not found in ABI`);
    }

    const mapped: any = {};
    for (const param of funcInfo.inputs) {
      const paramName = param.name || `param${funcInfo.inputs.indexOf(param)}`;
      if (jsonData.hasOwnProperty(paramName)) {
        mapped[paramName] = jsonData[paramName];
      }
    }
    
    return mapped;
  }

  /**
   * Get function information from ABI
   */
  protected getFunctionInfo() {
    return this.abiCodec ? this.abiCodec.getFunctionInfo(this.config.functionName) : undefined;
  }

  /**
   * Get all available functions from ABI
   */
  protected getAllFunctions() {
    return this.abiCodec ? this.abiCodec.getAllFunctions() : [];
  }

  /**
   * Manually decode payload (for advanced use cases)
   */
  protected manualDecode(payload: Uint8Array): any {
    return this.abiCodec ? this.abiCodec.decodeFunctionCall(this.config.functionName, payload) : { data: payload };
  }

  /**
   * Manually encode result (for advanced use cases)
   */
  protected manualEncode(result: any): Uint8Array {
    return this.abiCodec ? this.abiCodec.encodeFunctionResult(this.config.functionName, result) : this.encodePayload(result);
  }
}

/**
 * Simplified SolidityWorker for JSON-based configuration
 */
export abstract class JsonSolidityWorker extends BaseWorker {
  private abiCodec: AbiCodec;
  private functionName: string;

  constructor(abi: any[], functionName: string) {
    super();
    this.abiCodec = new AbiCodec(abi);
    this.functionName = functionName;
  }

  async handleTask(task: TaskRequest): Promise<TaskResponse> {
    try {
      // Auto-detect and decode payload
      const detection = PayloadAutoDecoder.decode(task.payload);
      let decodedParams: any;

      if (detection.format === 'abi') {
        decodedParams = this.abiCodec.decodeFunctionCall(this.functionName, task.payload);
      } else if (detection.format === 'json') {
        decodedParams = detection.data;
      } else {
        decodedParams = { data: task.payload };
      }

      // Call user handler
      const result = await this.handleDecodedTask(decodedParams);

      // Encode result
      let encodedResult: Uint8Array;
      if (typeof result === 'object' && result !== null) {
        encodedResult = this.abiCodec.encodeFunctionResult(this.functionName, result);
      } else {
        encodedResult = this.abiCodec.encodeFunctionResult(this.functionName, { result });
      }

      return this.createResponse(task.taskId, encodedResult);
    } catch (error) {
      throw new Error(`JSON Solidity task handling failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  /**
   * Abstract method for handling decoded parameters
   */
  abstract handleDecodedTask(params: any): Promise<any>;
}

/**
 * Utility functions for SolidityWorker
 */
export class SolidityWorkerUtils {
  /**
   * Create a simple worker from ABI and handler function
   */
  static createFromAbi<T = any>(
    abi: any[],
    functionName: string,
    handler: (params: T) => Promise<any>
  ): JsonSolidityWorker {
    return new (class extends JsonSolidityWorker {
      async handleDecodedTask(params: T): Promise<any> {
        return handler(params);
      }
    })(abi, functionName);
  }

  /**
   * Create a worker with automatic TypeChain integration
   */
  static createTyped<TContract, TFunction extends keyof TContract>(
    config: SolidityWorkerConfig,
    handler: (params: ExtractFunctionParams<TContract, TFunction>[0]) => Promise<ExtractFunctionReturn<TContract, TFunction>>
  ): SolidityWorker<TContract, TFunction> {
    return new (class extends SolidityWorker<TContract, TFunction> {
      async handleSolidityTask(params: ExtractFunctionParams<TContract, TFunction>[0]): Promise<ExtractFunctionReturn<TContract, TFunction>> {
        return handler(params);
      }
    })(config);
  }
}

// Decorator removed - use SolidityWorker class directly for better type safety
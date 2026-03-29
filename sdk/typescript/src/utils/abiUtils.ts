// ABI encoding/decoding utilities for Solidity integration

import { AbiCoder, Interface, Fragment, FunctionFragment, ParamType } from 'ethers';
import { hexToBytes, bytesToHex } from '../types/protobuf';

/**
 * ABI function signature information
 */
export interface AbiFunctionInfo {
  /** Function name */
  name: string;
  /** Function signature (e.g., "processTask(uint256,bytes32,address)") */
  signature: string;
  /** Input parameter types */
  inputs: readonly ParamType[];
  /** Output parameter types */
  outputs: readonly ParamType[];
  /** Function selector (first 4 bytes of keccak256 hash) */
  selector: string;
}

/**
 * ABI codec for encoding/decoding Solidity function calls
 */
export class AbiCodec {
  private abiCoder: AbiCoder;
  private contractInterface: Interface;
  private functionInfo: Map<string, AbiFunctionInfo> = new Map();

  constructor(abi: any[]) {
    this.abiCoder = AbiCoder.defaultAbiCoder();
    this.contractInterface = new Interface(abi);
    this.buildFunctionInfo();
  }

  /**
   * Build function information from ABI
   */
  private buildFunctionInfo(): void {
    for (const fragment of this.contractInterface.fragments) {
      if (fragment.type === 'function') {
        const funcFragment = fragment as FunctionFragment;
        const info: AbiFunctionInfo = {
          name: funcFragment.name,
          signature: funcFragment.format('minimal'),
          inputs: funcFragment.inputs,
          outputs: funcFragment.outputs,
          selector: funcFragment.selector,
        };
        this.functionInfo.set(funcFragment.name, info);
      }
    }
  }

  /**
   * Get function information by name
   */
  getFunctionInfo(functionName: string): AbiFunctionInfo | undefined {
    return this.functionInfo.get(functionName);
  }

  /**
   * Get all available functions
   */
  getAllFunctions(): AbiFunctionInfo[] {
    return Array.from(this.functionInfo.values());
  }

  /**
   * Decode function call data
   */
  decodeFunctionCall(functionName: string, data: Uint8Array): any {
    const funcInfo = this.functionInfo.get(functionName);
    if (!funcInfo) {
      throw new Error(`Function ${functionName} not found in ABI`);
    }

    try {
      // Convert bytes to hex string for ethers
      const hexData = bytesToHex(data);
      
      // Decode using contract interface
      const result = this.contractInterface.decodeFunctionData(functionName, hexData);
      
      // Convert result to plain object with parameter names
      const decoded: any = {};
      for (let i = 0; i < funcInfo.inputs.length; i++) {
        const param = funcInfo.inputs[i];
        decoded[param?.name || `param${i}`] = result[i];
      }
      
      return decoded;
    } catch (error) {
      throw new Error(`Failed to decode function call: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  /**
   * Encode function call data
   */
  encodeFunctionCall(functionName: string, params: any[]): Uint8Array {
    const funcInfo = this.functionInfo.get(functionName);
    if (!funcInfo) {
      throw new Error(`Function ${functionName} not found in ABI`);
    }

    try {
      const encoded = this.contractInterface.encodeFunctionData(functionName, params);
      return hexToBytes(encoded);
    } catch (error) {
      throw new Error(`Failed to encode function call: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  /**
   * Decode function result
   */
  decodeFunctionResult(functionName: string, data: Uint8Array): any {
    const funcInfo = this.functionInfo.get(functionName);
    if (!funcInfo) {
      throw new Error(`Function ${functionName} not found in ABI`);
    }

    try {
      const hexData = bytesToHex(data);
      const result = this.contractInterface.decodeFunctionResult(functionName, hexData);
      
      // Convert result to plain object with parameter names
      if (funcInfo.outputs.length === 1) {
        return result[0];
      } else {
        const decoded: any = {};
        for (let i = 0; i < funcInfo.outputs.length; i++) {
          const param = funcInfo.outputs[i];
          decoded[param?.name || `result${i}`] = result[i];
        }
        return decoded;
      }
    } catch (error) {
      throw new Error(`Failed to decode function result: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  /**
   * Encode function result
   */
  encodeFunctionResult(functionName: string, result: any): Uint8Array {
    const funcInfo = this.functionInfo.get(functionName);
    if (!funcInfo) {
      throw new Error(`Function ${functionName} not found in ABI`);
    }

    try {
      // Handle single vs multiple return values
      let values: any[];
      if (funcInfo.outputs.length === 1) {
        values = [result];
      } else {
        values = [];
        for (let i = 0; i < funcInfo.outputs.length; i++) {
          const param = funcInfo.outputs[i];
          const key = param?.name || `result${i}`;
          values.push(result[key]);
        }
      }

      const encoded = this.abiCoder.encode(
        funcInfo.outputs.map(p => p.format()),
        values
      );
      return hexToBytes(encoded);
    } catch (error) {
      throw new Error(`Failed to encode function result: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  /**
   * Detect if data looks like ABI-encoded function call
   */
  static isAbiEncoded(data: Uint8Array): boolean {
    // ABI-encoded function calls start with 4-byte function selector
    return data.length >= 4;
  }

  /**
   * Try to detect function from encoded data
   */
  detectFunction(data: Uint8Array): string | null {
    if (data.length < 4) return null;
    
    const selector = bytesToHex(data.slice(0, 4));
    
    for (const [name, info] of this.functionInfo) {
      if (info.selector === selector) {
        return name;
      }
    }
    
    return null;
  }
}

/**
 * Utility functions for common Solidity types
 */
export class SolidityTypeUtils {
  /**
   * Convert JavaScript value to Solidity type
   */
  static toSolidityType(value: any, solidityType: string): any {
    switch (solidityType) {
      case 'address':
        return value.toLowerCase();
      case 'bool':
        return Boolean(value);
      case 'string':
        return String(value);
      case 'bytes':
      case 'bytes32':
        return value instanceof Uint8Array ? bytesToHex(value) : value;
      default:
        if (solidityType.startsWith('uint') || solidityType.startsWith('int')) {
          return BigInt(value);
        }
        return value;
    }
  }

  /**
   * Convert Solidity type to JavaScript value
   */
  static fromSolidityType(value: any, solidityType: string): any {
    switch (solidityType) {
      case 'address':
        return value.toLowerCase();
      case 'bool':
        return Boolean(value);
      case 'string':
        return String(value);
      case 'bytes':
      case 'bytes32':
        return typeof value === 'string' ? hexToBytes(value) : value;
      default:
        if (solidityType.startsWith('uint') || solidityType.startsWith('int')) {
          return typeof value === 'bigint' ? value : BigInt(value);
        }
        return value;
    }
  }

  /**
   * Validate value against Solidity type
   */
  static validateType(value: any, solidityType: string): boolean {
    try {
      this.toSolidityType(value, solidityType);
      return true;
    } catch {
      return false;
    }
  }
}

/**
 * Auto-detect payload format and decode accordingly
 */
export class PayloadAutoDecoder {
  /**
   * Attempt to decode payload using multiple strategies
   */
  static decode(payload: Uint8Array): {
    format: 'abi' | 'json' | 'string' | 'raw';
    data: any;
    confidence: number;
  } {
    // Try ABI decoding first
    if (AbiCodec.isAbiEncoded(payload)) {
      return {
        format: 'abi',
        data: payload,
        confidence: 0.8,
      };
    }

    // Try JSON decoding
    try {
      const text = new TextDecoder().decode(payload);
      const json = JSON.parse(text);
      return {
        format: 'json',
        data: json,
        confidence: 0.9,
      };
    } catch {
      // Not JSON
    }

    // Try string decoding
    try {
      const text = new TextDecoder().decode(payload);
      // Check if it's valid UTF-8 and printable
      if (text.length > 0 && !/[\x00-\x08\x0B\x0C\x0E-\x1F\x7F-\x9F]/.test(text)) {
        return {
          format: 'string',
          data: text,
          confidence: 0.7,
        };
      }
    } catch {
      // Not valid UTF-8
    }

    // Default to raw bytes
    return {
      format: 'raw',
      data: payload,
      confidence: 0.5,
    };
  }
}
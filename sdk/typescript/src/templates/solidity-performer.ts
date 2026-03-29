// Solidity performer template with TypeChain integration

import { SolidityWorker } from '../worker/solidityWorker';

// TODO: Import your contract types from typechain-types
// import type { MyContract } from './typechain-types';

/**
 * Simple Solidity performer - just implement handleSolidityTask!
 */
class MyAVSPerformer extends SolidityWorker<any, 'processTask'> {
  async handleSolidityTask(params: { taskId: string; data: Uint8Array; user: string }) {
    // TODO: Implement your AVS logic here
    // params are automatically typed based on your contract ABI
    
    const { taskId, data, user } = params;
    
    // Example: Process the data and return result
    const processedData = new TextEncoder().encode(
      `Processed: ${new TextDecoder().decode(data)}`
    );
    
    return {
      result: processedData,
      success: true,
    };
  }
}

// One-line server startup
new MyAVSPerformer().start();
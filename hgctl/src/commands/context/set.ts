import { Command } from '@commander-js/extra-typings';
import { updateContext } from '../../config/context';
import { Logger } from '../../logger';

export function setupSetCommand(parent: Command) {
  parent
    .command('set [context-name]')
    .description('Set context configuration values')
    .option('--executor-address <address>', 'Executor gRPC server address')
    .option('--avs-address <address>', 'Default AVS address')
    .option('--operator-set-id <id>', 'Default operator set ID')
    .option('--network-id <id>', 'Network ID')
    .option('--rpc-url <url>', 'Ethereum RPC URL')
    .option('--release-manager-address <address>', 'ReleaseManager contract address')
    .action(async (contextName, options, command) => {
      const globalOpts = command.optsWithGlobals();
      const logger = globalOpts.logger as Logger;
      
      const name = contextName || 'default';
      
      // Build updates object from provided options
      const updates: Record<string, any> = {};
      
      if (options.executorAddress) updates.executorAddress = options.executorAddress;
      if (options.avsAddress) updates.avsAddress = options.avsAddress;
      if (options.operatorSetId) updates.operatorSetId = parseInt(options.operatorSetId);
      if (options.networkId) updates.networkId = parseInt(options.networkId);
      if (options.rpcUrl) updates.rpcUrl = options.rpcUrl;
      if (options.releaseManagerAddress) updates.releaseManagerAddress = options.releaseManagerAddress;
      
      if (Object.keys(updates).length === 0) {
        logger.error('No configuration values provided');
        console.log('\nAvailable options:');
        console.log('  --executor-address <address>');
        console.log('  --avs-address <address>');
        console.log('  --operator-set-id <id>');
        console.log('  --network-id <id>');
        console.log('  --rpc-url <url>');
        console.log('  --release-manager-address <address>');
        process.exit(1);
      }
      
      try {
        await updateContext(name, updates);
        logger.info(`Context "${name}" updated successfully`);
        
        // Show what was updated
        Object.entries(updates).forEach(([key, value]) => {
          console.log(`  ${key}: ${value}`);
        });
      } catch (error) {
        logger.error(`Failed to update context: ${(error as Error).message}`);
        process.exit(1);
      }
    });
}
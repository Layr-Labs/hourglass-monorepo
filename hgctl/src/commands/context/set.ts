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
      // Extract the injected values first
      const globalOpts = command.optsWithGlobals();
      const logger = globalOpts.logger as Logger;
      
      const name = contextName || 'default';
      
      // Get the actual command options by looking at process.argv
      const args = process.argv.slice(2);
      const updates: Record<string, any> = {};
      
      // Parse command line arguments manually
      for (let i = 0; i < args.length; i++) {
        if (args[i] === '--executor-address' && args[i + 1]) {
          updates.executorAddress = args[i + 1];
          i++;
        } else if (args[i] === '--avs-address' && args[i + 1]) {
          updates.avsAddress = args[i + 1];
          i++;
        } else if (args[i] === '--operator-set-id' && args[i + 1]) {
          updates.operatorSetId = parseInt(args[i + 1]);
          i++;
        } else if (args[i] === '--network-id' && args[i + 1]) {
          updates.networkId = parseInt(args[i + 1]);
          i++;
        } else if (args[i] === '--rpc-url' && args[i + 1]) {
          updates.rpcUrl = args[i + 1];
          i++;
        } else if (args[i] === '--release-manager-address' && args[i + 1]) {
          updates.releaseManagerAddress = args[i + 1];
          i++;
        }
      }
      
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
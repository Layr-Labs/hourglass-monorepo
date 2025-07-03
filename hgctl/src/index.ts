#!/usr/bin/env node
import { Command } from '@commander-js/extra-typings';
import { setupCommands } from './commands';
import { setupLogger } from './logger';
import { loadContext } from './config/context';
import { setupTelemetry } from './telemetry/client';

async function main() {
  const program = new Command()
    .name('hgctl')
    .description('Hourglass Control - CLI for managing AVS deployments')
    .version('1.0.0')
    .option('-v, --verbose', 'Enable verbose logging')
    .option('--executor-address <address>', 'Executor gRPC server address')
    .option('--rpc-url <url>', 'Ethereum RPC URL')
    .option('--release-manager-address <address>', 'ReleaseManager contract address')
    .option('--context <name>', 'Context to use')
    .hook('preAction', async (thisCommand, actionCommand) => {
      // Setup logger
      const logger = setupLogger(thisCommand.opts().verbose);
      actionCommand.setOptionValue('logger', logger);
      
      // Load context
      const config = await loadContext();
      const contextName = thisCommand.opts().context || config.currentContext || 'default';
      const context = config.contexts[contextName] || config.contexts.default;
      
      // Apply overrides
      if (thisCommand.opts().executorAddress) {
        context.executorAddress = thisCommand.opts().executorAddress as string;
      }
      if (thisCommand.opts().rpcUrl) {
        context.rpcUrl = thisCommand.opts().rpcUrl;
      }
      if (thisCommand.opts().releaseManagerAddress) {
        context.releaseManagerAddress = thisCommand.opts().releaseManagerAddress;
      }
      
      actionCommand.setOptionValue('context', context);
      
      // Setup telemetry
      await setupTelemetry(actionCommand as any);
    });

  setupCommands(program);
  
  await program.parseAsync(process.argv);
}

main().catch(console.error);
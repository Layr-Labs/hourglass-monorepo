import { Command } from '@commander-js/extra-typings';
import { createExecutorClient } from '../../client/executor';
import { OutputFormatter, OutputFormat } from '../../output';
import { Logger } from '../../logger';
import { Context } from '../../config/context';

export function setupPerformersCommand(parent: Command) {
  parent
    .command('performers [avs-address]')
    .description('List all performers')
    .option('--avs-address <address>', 'Filter by AVS address')
    .option('-o, --output <format>', 'Output format (json|yaml|table)', 'table')
    .action(async (avsAddressArg, options, command) => {
      const globalOpts = command.optsWithGlobals();
      const logger = globalOpts.logger as Logger;
      const context = globalOpts.context as Context;
      
      // Determine AVS address from args, options, or context
      const avsAddress = avsAddressArg || options.avsAddress || context.avsAddress;
      
      const client = createExecutorClient(context, logger);
      
      try {
        if (avsAddress) {
          logger.info(`Listing performers for AVS: ${avsAddress}`);
        } else {
          logger.info('Listing all performers');
        }
        
        const performers = await client.listPerformers(avsAddress);
        
        OutputFormatter.print(performers, options.output as OutputFormat);
      } catch (error) {
        logger.error(`Failed to list performers: ${(error as Error).message}`);
        process.exit(1);
      } finally {
        client.close();
      }
    });
}
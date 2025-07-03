import { Command } from '@commander-js/extra-typings';
import { createContractClient } from '../../client/contract';
import { OutputFormatter, OutputFormat } from '../../output';
import { Logger } from '../../logger';
import { Context } from '../../config/context';

export function setupReleasesCommand(parent: Command) {
  parent
    .command('releases <avs-address>')
    .description('List releases for an AVS')
    .option('--operator-set-id <id>', 'Operator set ID', '0')
    .option('--latest', 'Show only the latest release')
    .option('--current', 'Show only the current release')
    .option('--limit <number>', 'Maximum number of releases to show', '10')
    .option('-o, --output <format>', 'Output format (json|yaml|table)', 'table')
    .action(async (avsAddress, options, command) => {
      const globalOpts = command.optsWithGlobals();
      const logger = globalOpts.logger as Logger;
      const context = globalOpts.context as Context;
      
      // Validate AVS address
      if (!avsAddress.match(/^0x[a-fA-F0-9]{40}$/)) {
        logger.error(`Invalid AVS address: ${avsAddress}`);
        logger.info('AVS address must be a valid Ethereum address (e.g., 0x1234567890123456789012345678901234567890)');
        process.exit(1);
      }
      
      const operatorSetId = parseInt(options.operatorSetId) || context.operatorSetId || 0;
      
      if (!context.rpcUrl || !context.releaseManagerAddress) {
        logger.error('RPC URL and ReleaseManager address must be configured');
        logger.info('Use "hgctl context set" to configure these values');
        process.exit(1);
      }
      
      const client = createContractClient(context, logger);
      
      try {
        if (options.latest) {
          logger.info(`Getting latest release for AVS ${avsAddress}, operator set ${operatorSetId}`);
          const release = await client.getLatestRelease(avsAddress, operatorSetId);
          
          if (!release) {
            logger.warn('No releases found');
            return;
          }
          
          OutputFormatter.print(release, options.output as OutputFormat);
        } else if (options.current) {
          logger.info(`Getting current release for AVS ${avsAddress}, operator set ${operatorSetId}`);
          // Note: getCurrentRelease is not available in IReleaseManager, using getLatestRelease
          const release = await client.getLatestRelease(avsAddress, operatorSetId);
          
          if (!release) {
            logger.info('No current release found');
            return;
          }
          
          OutputFormatter.print(release, options.output as OutputFormat);
        } else {
          const limit = parseInt(options.limit) || 10;
          logger.info(`Listing releases for AVS ${avsAddress}, operator set ${operatorSetId} (limit: ${limit})`);
          
          const releases = await client.getReleases(avsAddress, operatorSetId, limit);
          
          if (releases.length === 0) {
            logger.info('No releases found');
            return;
          }
          
          OutputFormatter.print(releases, options.output as OutputFormat);
        }
      } catch (error) {
        logger.error(`Failed to get releases: ${(error as Error).message}`);
        process.exit(1);
      }
    });
}
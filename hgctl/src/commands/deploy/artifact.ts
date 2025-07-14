import { Command } from '@commander-js/extra-typings';
import { createExecutorClient } from '../../client/executor';
import { Logger } from '../../logger';
import { Context } from '../../config/context';
import ora from 'ora';

export function setupArtifactCommand(parent: Command) {
  parent
    .command('artifact <avs-address> <digest>')
    .description('Deploy an artifact for an AVS')
    .option('--registry-url <url>', 'Registry URL', 'docker.io')
    .action(async (avsAddress, digest, options, command) => {
      const globalOpts = command.optsWithGlobals();
      const logger = globalOpts.logger as Logger;
      const context = globalOpts.context as Context;
      
      const spinner = ora('Deploying artifact...').start();
      const client = createExecutorClient(context, logger);
      
      try {
        logger.debug(`Deploying artifact ${digest} for AVS ${avsAddress}`);
        
        const response = await client.deployArtifact({
          avsAddress,
          digest,
          registryUrl: options.registryUrl
        });
        
        if (response.success) {
          spinner.succeed('Artifact deployed successfully');
          if (response.performerId) {
            console.log(`Performer ID: ${response.performerId}`);
          }
        } else {
          spinner.fail(`Deployment failed: ${response.message || 'Unknown error'}`);
          process.exit(1);
        }
      } catch (error) {
        spinner.fail(`Failed to deploy artifact: ${(error as Error).message}`);
        logger.debug((error as Error).stack || '');
        process.exit(1);
      } finally {
        client.close();
      }
    });
}
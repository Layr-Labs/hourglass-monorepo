import { Command } from '@commander-js/extra-typings';
import { createExecutorClient } from '../../client/executor';
import { Logger } from '../../logger';
import { Context } from '../../config/context';
import prompts from 'prompts';
import ora from 'ora';

export function setupPerformerCommand(parent: Command) {
  parent
    .command('performer <performer-id>')
    .description('Remove a performer')
    .option('-f, --force', 'Skip confirmation prompt')
    .action(async (performerId, options, command) => {
      const globalOpts = command.optsWithGlobals();
      const logger = globalOpts.logger as Logger;
      const context = globalOpts.context as Context;
      
      // Confirm deletion unless --force is used
      if (!options.force) {
        const response = await prompts({
          type: 'confirm',
          name: 'confirm',
          message: `Are you sure you want to remove performer ${performerId}?`,
          initial: false
        });
        
        if (!response.confirm) {
          console.log('Removal cancelled');
          return;
        }
      }
      
      const spinner = ora(`Removing performer ${performerId}...`).start();
      const client = createExecutorClient(context, logger);
      
      try {
        await client.removePerformer(performerId);
        spinner.succeed('Performer removed successfully');
      } catch (error) {
        spinner.fail(`Failed to remove performer: ${(error as Error).message}`);
        logger.debug((error as Error).stack || '');
        process.exit(1);
      } finally {
        client.close();
      }
    });
}
import { Command } from '@commander-js/extra-typings';
import { listContexts } from '../../config/context';
import { Logger } from '../../logger';
import Table from 'cli-table3';
import chalk from 'chalk';

export function setupListCommand(parent: Command) {
  parent
    .command('list')
    .aliases(['ls'])
    .description('List all contexts')
    .action(async (_options, command) => {
      const globalOpts = command.optsWithGlobals();
      const logger = globalOpts.logger as Logger;
      
      try {
        const contexts = await listContexts();
        
        if (contexts.length === 0) {
          console.log('No contexts found');
          return;
        }
        
        const table = new Table({
          head: ['CURRENT', 'NAME'],
          style: { head: ['cyan'] }
        });
        
        contexts.forEach((ctx) => {
          table.push([
            ctx.current ? chalk.green('*') : '',
            ctx.name
          ]);
        });
        
        console.log(table.toString());
      } catch (error) {
        logger.error(`Failed to list contexts: ${(error as Error).message}`);
        process.exit(1);
      }
    });
}
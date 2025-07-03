import { Command } from '@commander-js/extra-typings';
import { loadContext, getContext } from '../../config/context';
import { Logger } from '../../logger';
import { OutputFormatter, OutputFormat } from '../../output';
import chalk from 'chalk';

export function setupShowCommand(parent: Command) {
  parent
    .command('show [context-name]')
    .description('Show context configuration')
    .option('-o, --output <format>', 'Output format (json|yaml|table)', 'table')
    .action(async (contextName, options, command) => {
      const globalOpts = command.optsWithGlobals();
      const logger = globalOpts.logger as Logger;
      
      try {
        if (contextName) {
          // Show specific context
          const context = await getContext(contextName);
          if (!context) {
            logger.error(`Context "${contextName}" not found`);
            process.exit(1);
          }
          
          if (options.output === 'table') {
            console.log(chalk.bold(`Context: ${contextName}`));
            console.log('');
            Object.entries(context).forEach(([key, value]) => {
              if (key !== 'name') {
                console.log(`  ${chalk.cyan(key)}: ${value || '-'}`);
              }
            });
          } else {
            OutputFormatter.print(context, options.output as OutputFormat);
          }
        } else {
          // Show current context
          const config = await loadContext();
          const currentContext = config.contexts[config.currentContext];
          
          if (options.output === 'table') {
            console.log(chalk.bold(`Current context: ${config.currentContext}`));
            console.log('');
            Object.entries(currentContext).forEach(([key, value]) => {
              if (key !== 'name') {
                console.log(`  ${chalk.cyan(key)}: ${value || '-'}`);
              }
            });
          } else {
            OutputFormatter.print({
              name: config.currentContext,
              ...currentContext
            }, options.output as OutputFormat);
          }
        }
      } catch (error) {
        logger.error(`Failed to show context: ${(error as Error).message}`);
        process.exit(1);
      }
    });
}
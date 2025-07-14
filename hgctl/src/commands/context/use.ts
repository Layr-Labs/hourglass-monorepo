import { Command } from '@commander-js/extra-typings';
import { setCurrentContext } from '../../config/context';
import { Logger } from '../../logger';

export function setupUseCommand(parent: Command) {
  parent
    .command('use <context-name>')
    .description('Switch to a different context')
    .action(async (contextName, _options, command) => {
      const globalOpts = command.optsWithGlobals();
      const logger = globalOpts.logger as Logger;
      
      try {
        await setCurrentContext(contextName);
        logger.info(`Switched to context "${contextName}"`);
      } catch (error) {
        logger.error(`Failed to switch context: ${(error as Error).message}`);
        process.exit(1);
      }
    });
}
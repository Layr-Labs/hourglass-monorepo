import { Command } from '@commander-js/extra-typings';
import { setupSetCommand } from './set';
import { setupShowCommand } from './show';
import { setupUseCommand } from './use';
import { setupListCommand } from './list';

export function setupContextCommands(parent: Command) {
  const contextCmd = parent
    .command('context')
    .description('Manage contexts');

  setupSetCommand(contextCmd);
  setupShowCommand(contextCmd);
  setupUseCommand(contextCmd);
  setupListCommand(contextCmd);
}
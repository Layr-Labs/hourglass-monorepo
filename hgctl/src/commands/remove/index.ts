import { Command } from '@commander-js/extra-typings';
import { setupPerformerCommand } from './performer';

export function setupRemoveCommands(parent: Command) {
  const removeCmd = parent
    .command('remove')
    .aliases(['rm', 'delete'])
    .description('Remove resources');

  setupPerformerCommand(removeCmd);
}
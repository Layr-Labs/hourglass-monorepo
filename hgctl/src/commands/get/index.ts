import { Command } from '@commander-js/extra-typings';
import { setupPerformersCommand } from './performers';
import { setupReleasesCommand } from './releases';

export function setupGetCommands(parent: Command) {
  const getCmd = parent
    .command('get')
    .description('Get resources');

  setupPerformersCommand(getCmd);
  setupReleasesCommand(getCmd);
}
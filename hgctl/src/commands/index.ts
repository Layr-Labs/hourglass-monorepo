import { Command } from '@commander-js/extra-typings';
import { setupGetCommands } from './get';
import { setupDeployCommands } from './deploy';
import { setupRemoveCommands } from './remove';
import { setupContextCommands } from './context';
import { setupCompletionCommand } from './completion';

export function setupCommands(program: Command) {
  setupGetCommands(program);
  setupDeployCommands(program);
  setupRemoveCommands(program);
  setupContextCommands(program);
  setupCompletionCommand(program);
}
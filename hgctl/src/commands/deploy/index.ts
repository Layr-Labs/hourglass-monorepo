import { Command } from '@commander-js/extra-typings';
import { setupArtifactCommand } from './artifact';

export function setupDeployCommands(parent: Command) {
  const deployCmd = parent
    .command('deploy')
    .description('Deploy resources');

  setupArtifactCommand(deployCmd);
}
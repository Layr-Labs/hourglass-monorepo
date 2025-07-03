import Table from 'cli-table3';
import yaml from 'yaml';
import chalk from 'chalk';
import { Performer } from '../client/executor';
import { Release } from '../client/contract';

export type OutputFormat = 'table' | 'json' | 'yaml';

export class OutputFormatter {
  static print(data: any, format: OutputFormat = 'table'): void {
    switch (format) {
      case 'json':
        console.log(JSON.stringify(data, null, 2));
        break;
      case 'yaml':
        console.log(yaml.stringify(data));
        break;
      case 'table':
        this.printTable(data);
        break;
    }
  }

  private static printTable(data: any): void {
    if (Array.isArray(data) && data.length === 0) {
      console.log('No data to display');
      return;
    }

    // Handle performers
    if (this.isPerformerArray(data)) {
      this.printPerformersTable(data);
    }
    // Handle releases
    else if (this.isReleaseArray(data)) {
      this.printReleasesTable(data);
    }
    // Handle single release
    else if (this.isRelease(data)) {
      this.printReleasesTable([data]);
    }
    // Default to JSON
    else {
      console.log(JSON.stringify(data, null, 2));
    }
  }

  private static isPerformerArray(data: any): data is Performer[] {
    return Array.isArray(data) && data.length > 0 && 'performer_id' in data[0];
  }

  private static isReleaseArray(data: any): data is Release[] {
    return Array.isArray(data) && data.length > 0 && 'artifacts' in data[0];
  }

  private static isRelease(data: any): data is Release {
    return data && 'artifacts' in data && !Array.isArray(data);
  }

  private static printPerformersTable(performers: Performer[]): void {
    const table = new Table({
      head: ['PERFORMER ID', 'AVS ADDRESS', 'STATUS', 'HEALTH', 'ARTIFACT'],
      style: { head: ['cyan'] }
    });

    performers.forEach((performer) => {
      const health = performer.application_healthy && performer.resource_healthy
        ? chalk.green('Healthy')
        : chalk.red('Unhealthy');
      
      const digest = performer.artifact_digest
        ? this.formatDigest(performer.artifact_digest)
        : '-';

      table.push([
        performer.performer_id,
        this.formatAddress(performer.avs_address),
        performer.status,
        health,
        digest
      ]);
    });

    console.log(table.toString());
  }

  private static printReleasesTable(releases: Release[]): void {
    const table = new Table({
      head: ['RELEASE ID', 'UPGRADE BY', 'ARTIFACTS'],
      style: { head: ['cyan'] },
      colWidths: [12, 25, 60]
    });

    releases.forEach((release) => {
      const upgradeBy = new Date(release.upgradeByTime * 1000).toLocaleString();
      
      if (release.artifacts.length === 0) {
        table.push([release.id, upgradeBy, '(no artifacts)']);
      } else {
        // First artifact
        const firstArtifact = release.artifacts[0];
        table.push([
          release.id,
          upgradeBy,
          `${this.formatDigest(firstArtifact.digest)} @ ${firstArtifact.registryUrl}`
        ]);
        
        // Additional artifacts
        release.artifacts.slice(1).forEach((artifact) => {
          table.push([
            '',
            '',
            `${this.formatDigest(artifact.digest)} @ ${artifact.registryUrl}`
          ]);
        });
      }
    });

    console.log(table.toString());
  }

  private static formatAddress(address: string): string {
    if (!address) return '-';
    if (address.length > 10) {
      return `${address.substring(0, 6)}...${address.substring(address.length - 4)}`;
    }
    return address;
  }

  private static formatDigest(digest: string): string {
    if (!digest) return '-';
    // If it's a bytes32 hex string, format it
    if (digest.startsWith('0x') && digest.length > 12) {
      return digest.substring(0, 12) + '...';
    }
    if (digest.length > 12) {
      return digest.substring(0, 12) + '...';
    }
    return digest;
  }
}
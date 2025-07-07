"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.OutputFormatter = void 0;
const cli_table3_1 = __importDefault(require("cli-table3"));
const yaml_1 = __importDefault(require("yaml"));
const chalk_1 = __importDefault(require("chalk"));
class OutputFormatter {
    static print(data, format = 'table') {
        switch (format) {
            case 'json':
                console.log(JSON.stringify(data, null, 2));
                break;
            case 'yaml':
                console.log(yaml_1.default.stringify(data));
                break;
            case 'table':
                this.printTable(data);
                break;
        }
    }
    static printTable(data) {
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
    static isPerformerArray(data) {
        return Array.isArray(data) && data.length > 0 && 'performer_id' in data[0];
    }
    static isReleaseArray(data) {
        return Array.isArray(data) && data.length > 0 && 'artifacts' in data[0];
    }
    static isRelease(data) {
        return data && 'artifacts' in data && !Array.isArray(data);
    }
    static printPerformersTable(performers) {
        const table = new cli_table3_1.default({
            head: ['PERFORMER ID', 'AVS ADDRESS', 'STATUS', 'HEALTH', 'ARTIFACT'],
            style: { head: ['cyan'] }
        });
        performers.forEach((performer) => {
            const health = performer.application_healthy && performer.resource_healthy
                ? chalk_1.default.green('Healthy')
                : chalk_1.default.red('Unhealthy');
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
    static printReleasesTable(releases) {
        const table = new cli_table3_1.default({
            head: ['RELEASE ID', 'UPGRADE BY', 'ARTIFACTS'],
            style: { head: ['cyan'] },
            colWidths: [12, 25, 60]
        });
        releases.forEach((release) => {
            const upgradeBy = new Date(release.upgradeByTime * 1000).toLocaleString();
            if (release.artifacts.length === 0) {
                table.push([release.id, upgradeBy, '(no artifacts)']);
            }
            else {
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
    static formatAddress(address) {
        if (!address)
            return '-';
        if (address.length > 10) {
            return `${address.substring(0, 6)}...${address.substring(address.length - 4)}`;
        }
        return address;
    }
    static formatDigest(digest) {
        if (!digest)
            return '-';
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
exports.OutputFormatter = OutputFormatter;
//# sourceMappingURL=index.js.map
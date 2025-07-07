"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.setupReleasesCommand = setupReleasesCommand;
const contract_1 = require("../../client/contract");
const output_1 = require("../../output");
function setupReleasesCommand(parent) {
    parent
        .command('releases <avs-address>')
        .description('List releases for an AVS')
        .option('--operator-set-id <id>', 'Operator set ID', '0')
        .option('--latest', 'Show only the latest release')
        .option('--current', 'Show only the current release')
        .option('--limit <number>', 'Maximum number of releases to show', '10')
        .option('-o, --output <format>', 'Output format (json|yaml|table)', 'table')
        .action(async (avsAddress, options, command) => {
        const globalOpts = command.optsWithGlobals();
        const logger = globalOpts.logger;
        const context = globalOpts.context;
        const operatorSetId = parseInt(options.operatorSetId) || context.operatorSetId || 0;
        if (!context.rpcUrl || !context.releaseManagerAddress) {
            logger.error('RPC URL and ReleaseManager address must be configured');
            logger.info('Use "hgctl context set" to configure these values');
            process.exit(1);
        }
        const client = (0, contract_1.createContractClient)(context, logger);
        try {
            if (options.latest) {
                logger.info(`Getting latest release for AVS ${avsAddress}, operator set ${operatorSetId}`);
                const release = await client.getLatestRelease(avsAddress, operatorSetId);
                if (!release) {
                    logger.warn('No releases found');
                    return;
                }
                output_1.OutputFormatter.print(release, options.output);
            }
            else if (options.current) {
                logger.info(`Getting current release for AVS ${avsAddress}, operator set ${operatorSetId}`);
                const release = await client.getCurrentRelease(avsAddress, operatorSetId);
                if (!release) {
                    logger.warn('No current release found');
                    return;
                }
                output_1.OutputFormatter.print(release, options.output);
            }
            else {
                const limit = parseInt(options.limit) || 10;
                logger.info(`Listing releases for AVS ${avsAddress}, operator set ${operatorSetId} (limit: ${limit})`);
                const releases = await client.getReleases(avsAddress, operatorSetId, limit);
                if (releases.length === 0) {
                    logger.warn('No releases found');
                    return;
                }
                output_1.OutputFormatter.print(releases, options.output);
            }
        }
        catch (error) {
            logger.error(`Failed to get releases: ${error.message}`);
            process.exit(1);
        }
    });
}
//# sourceMappingURL=releases.js.map
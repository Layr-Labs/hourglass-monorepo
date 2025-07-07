"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.setupPerformersCommand = setupPerformersCommand;
const executor_1 = require("../../client/executor");
const output_1 = require("../../output");
function setupPerformersCommand(parent) {
    parent
        .command('performers [avs-address]')
        .description('List all performers')
        .option('--avs-address <address>', 'Filter by AVS address')
        .option('-o, --output <format>', 'Output format (json|yaml|table)', 'table')
        .action(async (avsAddressArg, options, command) => {
        const globalOpts = command.optsWithGlobals();
        const logger = globalOpts.logger;
        const context = globalOpts.context;
        // Determine AVS address from args, options, or context
        const avsAddress = avsAddressArg || options.avsAddress || context.avsAddress;
        const client = (0, executor_1.createExecutorClient)(context, logger);
        try {
            if (avsAddress) {
                logger.info(`Listing performers for AVS: ${avsAddress}`);
            }
            else {
                logger.info('Listing all performers');
            }
            const performers = await client.listPerformers(avsAddress);
            output_1.OutputFormatter.print(performers, options.output);
        }
        catch (error) {
            logger.error(`Failed to list performers: ${error.message}`);
            process.exit(1);
        }
        finally {
            client.close();
        }
    });
}
//# sourceMappingURL=performers.js.map
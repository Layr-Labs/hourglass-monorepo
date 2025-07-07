"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.setupArtifactCommand = setupArtifactCommand;
const executor_1 = require("../../client/executor");
const ora_1 = __importDefault(require("ora"));
function setupArtifactCommand(parent) {
    parent
        .command('artifact <avs-address> <digest>')
        .description('Deploy an artifact for an AVS')
        .option('--registry-url <url>', 'Registry URL', 'docker.io')
        .action(async (avsAddress, digest, options, command) => {
        const globalOpts = command.optsWithGlobals();
        const logger = globalOpts.logger;
        const context = globalOpts.context;
        const spinner = (0, ora_1.default)('Deploying artifact...').start();
        const client = (0, executor_1.createExecutorClient)(context, logger);
        try {
            logger.debug(`Deploying artifact ${digest} for AVS ${avsAddress}`);
            const response = await client.deployArtifact({
                avsAddress,
                digest,
                registryUrl: options.registryUrl
            });
            if (response.success) {
                spinner.succeed('Artifact deployed successfully');
                if (response.performerId) {
                    console.log(`Performer ID: ${response.performerId}`);
                }
            }
            else {
                spinner.fail(`Deployment failed: ${response.message || 'Unknown error'}`);
                process.exit(1);
            }
        }
        catch (error) {
            spinner.fail(`Failed to deploy artifact: ${error.message}`);
            logger.debug(error.stack || '');
            process.exit(1);
        }
        finally {
            client.close();
        }
    });
}
//# sourceMappingURL=artifact.js.map
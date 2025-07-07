#!/usr/bin/env node
"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const extra_typings_1 = require("@commander-js/extra-typings");
const commands_1 = require("./commands");
const logger_1 = require("./logger");
const context_1 = require("./config/context");
const client_1 = require("./telemetry/client");
async function main() {
    const program = new extra_typings_1.Command()
        .name('hgctl')
        .description('Hourglass Control - CLI for managing AVS deployments')
        .version('1.0.0')
        .option('-v, --verbose', 'Enable verbose logging')
        .option('--executor-address <address>', 'Executor gRPC server address')
        .option('--rpc-url <url>', 'Ethereum RPC URL')
        .option('--release-manager-address <address>', 'ReleaseManager contract address')
        .option('--context <name>', 'Context to use')
        .hook('preAction', async (thisCommand, actionCommand) => {
        // Setup logger
        const logger = (0, logger_1.setupLogger)(thisCommand.opts().verbose);
        actionCommand.setOptionValue('logger', logger);
        // Load context
        const config = await (0, context_1.loadContext)();
        const contextName = thisCommand.opts().context || config.currentContext || 'default';
        const context = config.contexts[contextName] || config.contexts.default;
        // Apply overrides
        if (thisCommand.opts().executorAddress) {
            context.executorAddress = thisCommand.opts().executorAddress;
        }
        if (thisCommand.opts().rpcUrl) {
            context.rpcUrl = thisCommand.opts().rpcUrl;
        }
        if (thisCommand.opts().releaseManagerAddress) {
            context.releaseManagerAddress = thisCommand.opts().releaseManagerAddress;
        }
        actionCommand.setOptionValue('context', context);
        // Setup telemetry
        await (0, client_1.setupTelemetry)(actionCommand);
    });
    (0, commands_1.setupCommands)(program);
    await program.parseAsync(process.argv);
}
main().catch(console.error);
//# sourceMappingURL=index.js.map
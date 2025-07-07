"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.setupSetCommand = setupSetCommand;
const context_1 = require("../../config/context");
function setupSetCommand(parent) {
    parent
        .command('set [context-name]')
        .description('Set context configuration values')
        .option('--executor-address <address>', 'Executor gRPC server address')
        .option('--avs-address <address>', 'Default AVS address')
        .option('--operator-set-id <id>', 'Default operator set ID')
        .option('--network-id <id>', 'Network ID')
        .option('--rpc-url <url>', 'Ethereum RPC URL')
        .option('--release-manager-address <address>', 'ReleaseManager contract address')
        .action(async (contextName, options, command) => {
        const globalOpts = command.optsWithGlobals();
        const logger = globalOpts.logger;
        const name = contextName || 'default';
        // Build updates object from provided options
        const updates = {};
        if (options.executorAddress)
            updates.executorAddress = options.executorAddress;
        if (options.avsAddress)
            updates.avsAddress = options.avsAddress;
        if (options.operatorSetId)
            updates.operatorSetId = parseInt(options.operatorSetId);
        if (options.networkId)
            updates.networkId = parseInt(options.networkId);
        if (options.rpcUrl)
            updates.rpcUrl = options.rpcUrl;
        if (options.releaseManagerAddress)
            updates.releaseManagerAddress = options.releaseManagerAddress;
        if (Object.keys(updates).length === 0) {
            logger.error('No configuration values provided');
            console.log('\nAvailable options:');
            console.log('  --executor-address <address>');
            console.log('  --avs-address <address>');
            console.log('  --operator-set-id <id>');
            console.log('  --network-id <id>');
            console.log('  --rpc-url <url>');
            console.log('  --release-manager-address <address>');
            process.exit(1);
        }
        try {
            await (0, context_1.updateContext)(name, updates);
            logger.info(`Context "${name}" updated successfully`);
            // Show what was updated
            Object.entries(updates).forEach(([key, value]) => {
                console.log(`  ${key}: ${value}`);
            });
        }
        catch (error) {
            logger.error(`Failed to update context: ${error.message}`);
            process.exit(1);
        }
    });
}
//# sourceMappingURL=set.js.map
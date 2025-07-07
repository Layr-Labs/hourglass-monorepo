"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.setupPerformerCommand = setupPerformerCommand;
const executor_1 = require("../../client/executor");
const prompts_1 = __importDefault(require("prompts"));
const ora_1 = __importDefault(require("ora"));
function setupPerformerCommand(parent) {
    parent
        .command('performer <performer-id>')
        .description('Remove a performer')
        .option('-f, --force', 'Skip confirmation prompt')
        .action(async (performerId, options, command) => {
        const globalOpts = command.optsWithGlobals();
        const logger = globalOpts.logger;
        const context = globalOpts.context;
        // Confirm deletion unless --force is used
        if (!options.force) {
            const response = await (0, prompts_1.default)({
                type: 'confirm',
                name: 'confirm',
                message: `Are you sure you want to remove performer ${performerId}?`,
                initial: false
            });
            if (!response.confirm) {
                console.log('Removal cancelled');
                return;
            }
        }
        const spinner = (0, ora_1.default)(`Removing performer ${performerId}...`).start();
        const client = (0, executor_1.createExecutorClient)(context, logger);
        try {
            await client.removePerformer(performerId);
            spinner.succeed('Performer removed successfully');
        }
        catch (error) {
            spinner.fail(`Failed to remove performer: ${error.message}`);
            logger.debug(error.stack || '');
            process.exit(1);
        }
        finally {
            client.close();
        }
    });
}
//# sourceMappingURL=performer.js.map
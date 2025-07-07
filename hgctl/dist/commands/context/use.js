"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.setupUseCommand = setupUseCommand;
const context_1 = require("../../config/context");
function setupUseCommand(parent) {
    parent
        .command('use <context-name>')
        .description('Switch to a different context')
        .action(async (contextName, _options, command) => {
        const globalOpts = command.optsWithGlobals();
        const logger = globalOpts.logger;
        try {
            await (0, context_1.setCurrentContext)(contextName);
            logger.info(`Switched to context "${contextName}"`);
        }
        catch (error) {
            logger.error(`Failed to switch context: ${error.message}`);
            process.exit(1);
        }
    });
}
//# sourceMappingURL=use.js.map
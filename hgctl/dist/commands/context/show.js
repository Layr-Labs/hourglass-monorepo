"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.setupShowCommand = setupShowCommand;
const context_1 = require("../../config/context");
const output_1 = require("../../output");
const chalk_1 = __importDefault(require("chalk"));
function setupShowCommand(parent) {
    parent
        .command('show [context-name]')
        .description('Show context configuration')
        .option('-o, --output <format>', 'Output format (json|yaml|table)', 'table')
        .action(async (contextName, options, command) => {
        const globalOpts = command.optsWithGlobals();
        const logger = globalOpts.logger;
        try {
            if (contextName) {
                // Show specific context
                const context = await (0, context_1.getContext)(contextName);
                if (!context) {
                    logger.error(`Context "${contextName}" not found`);
                    process.exit(1);
                }
                if (options.output === 'table') {
                    console.log(chalk_1.default.bold(`Context: ${contextName}`));
                    console.log('');
                    Object.entries(context).forEach(([key, value]) => {
                        if (key !== 'name') {
                            console.log(`  ${chalk_1.default.cyan(key)}: ${value || '-'}`);
                        }
                    });
                }
                else {
                    output_1.OutputFormatter.print(context, options.output);
                }
            }
            else {
                // Show current context
                const config = await (0, context_1.loadContext)();
                const currentContext = config.contexts[config.currentContext];
                if (options.output === 'table') {
                    console.log(chalk_1.default.bold(`Current context: ${config.currentContext}`));
                    console.log('');
                    Object.entries(currentContext).forEach(([key, value]) => {
                        if (key !== 'name') {
                            console.log(`  ${chalk_1.default.cyan(key)}: ${value || '-'}`);
                        }
                    });
                }
                else {
                    output_1.OutputFormatter.print({
                        name: config.currentContext,
                        ...currentContext
                    }, options.output);
                }
            }
        }
        catch (error) {
            logger.error(`Failed to show context: ${error.message}`);
            process.exit(1);
        }
    });
}
//# sourceMappingURL=show.js.map
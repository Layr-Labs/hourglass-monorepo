"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.setupListCommand = setupListCommand;
const context_1 = require("../../config/context");
const cli_table3_1 = __importDefault(require("cli-table3"));
const chalk_1 = __importDefault(require("chalk"));
function setupListCommand(parent) {
    parent
        .command('list')
        .aliases(['ls'])
        .description('List all contexts')
        .action(async (_options, command) => {
        const globalOpts = command.optsWithGlobals();
        const logger = globalOpts.logger;
        try {
            const contexts = await (0, context_1.listContexts)();
            if (contexts.length === 0) {
                console.log('No contexts found');
                return;
            }
            const table = new cli_table3_1.default({
                head: ['CURRENT', 'NAME'],
                style: { head: ['cyan'] }
            });
            contexts.forEach((ctx) => {
                table.push([
                    ctx.current ? chalk_1.default.green('*') : '',
                    ctx.name
                ]);
            });
            console.log(table.toString());
        }
        catch (error) {
            logger.error(`Failed to list contexts: ${error.message}`);
            process.exit(1);
        }
    });
}
//# sourceMappingURL=list.js.map
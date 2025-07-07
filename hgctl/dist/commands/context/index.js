"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.setupContextCommands = setupContextCommands;
const set_1 = require("./set");
const show_1 = require("./show");
const use_1 = require("./use");
const list_1 = require("./list");
function setupContextCommands(parent) {
    const contextCmd = parent
        .command('context')
        .description('Manage contexts');
    (0, set_1.setupSetCommand)(contextCmd);
    (0, show_1.setupShowCommand)(contextCmd);
    (0, use_1.setupUseCommand)(contextCmd);
    (0, list_1.setupListCommand)(contextCmd);
}
//# sourceMappingURL=index.js.map
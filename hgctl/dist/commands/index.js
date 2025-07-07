"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.setupCommands = setupCommands;
const get_1 = require("./get");
const deploy_1 = require("./deploy");
const remove_1 = require("./remove");
const context_1 = require("./context");
const completion_1 = require("./completion");
function setupCommands(program) {
    (0, get_1.setupGetCommands)(program);
    (0, deploy_1.setupDeployCommands)(program);
    (0, remove_1.setupRemoveCommands)(program);
    (0, context_1.setupContextCommands)(program);
    (0, completion_1.setupCompletionCommand)(program);
}
//# sourceMappingURL=index.js.map
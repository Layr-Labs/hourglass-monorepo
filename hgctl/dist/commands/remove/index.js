"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.setupRemoveCommands = setupRemoveCommands;
const performer_1 = require("./performer");
function setupRemoveCommands(parent) {
    const removeCmd = parent
        .command('remove')
        .aliases(['rm', 'delete'])
        .description('Remove resources');
    (0, performer_1.setupPerformerCommand)(removeCmd);
}
//# sourceMappingURL=index.js.map
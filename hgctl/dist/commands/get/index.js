"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.setupGetCommands = setupGetCommands;
const performers_1 = require("./performers");
const releases_1 = require("./releases");
function setupGetCommands(parent) {
    const getCmd = parent
        .command('get')
        .description('Get resources');
    (0, performers_1.setupPerformersCommand)(getCmd);
    (0, releases_1.setupReleasesCommand)(getCmd);
}
//# sourceMappingURL=index.js.map
"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.setupDeployCommands = setupDeployCommands;
const artifact_1 = require("./artifact");
function setupDeployCommands(parent) {
    const deployCmd = parent
        .command('deploy')
        .description('Deploy resources');
    (0, artifact_1.setupArtifactCommand)(deployCmd);
}
//# sourceMappingURL=index.js.map
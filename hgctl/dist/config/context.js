"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.loadContext = loadContext;
exports.saveContext = saveContext;
exports.getContext = getContext;
exports.setCurrentContext = setCurrentContext;
exports.updateContext = updateContext;
exports.deleteContext = deleteContext;
exports.listContexts = listContexts;
const fs_1 = require("fs");
const path_1 = __importDefault(require("path"));
const os_1 = __importDefault(require("os"));
const yaml_1 = __importDefault(require("yaml"));
const CONFIG_PATH = path_1.default.join(os_1.default.homedir(), '.hgctl', 'config.yaml');
async function loadContext() {
    try {
        const data = await fs_1.promises.readFile(CONFIG_PATH, 'utf-8');
        const config = yaml_1.default.parse(data);
        // Ensure default context exists
        if (!config.contexts.default) {
            config.contexts.default = {
                executorAddress: 'executor:9090'
            };
        }
        return config;
    }
    catch (error) {
        // Return default config if file doesn't exist
        return {
            currentContext: 'default',
            contexts: {
                default: {
                    executorAddress: 'executor:9090'
                }
            }
        };
    }
}
async function saveContext(config) {
    const dir = path_1.default.dirname(CONFIG_PATH);
    await fs_1.promises.mkdir(dir, { recursive: true });
    await fs_1.promises.writeFile(CONFIG_PATH, yaml_1.default.stringify(config));
}
async function getContext(name) {
    const config = await loadContext();
    return config.contexts[name] || null;
}
async function setCurrentContext(name) {
    const config = await loadContext();
    if (!config.contexts[name]) {
        throw new Error(`Context "${name}" does not exist`);
    }
    config.currentContext = name;
    await saveContext(config);
}
async function updateContext(name, updates) {
    const config = await loadContext();
    if (!config.contexts[name]) {
        config.contexts[name] = {
            executorAddress: 'executor:9090'
        };
    }
    config.contexts[name] = {
        ...config.contexts[name],
        ...updates
    };
    await saveContext(config);
}
async function deleteContext(name) {
    if (name === 'default') {
        throw new Error('Cannot delete the default context');
    }
    const config = await loadContext();
    if (!config.contexts[name]) {
        throw new Error(`Context "${name}" does not exist`);
    }
    delete config.contexts[name];
    // If we deleted the current context, switch to default
    if (config.currentContext === name) {
        config.currentContext = 'default';
    }
    await saveContext(config);
}
async function listContexts() {
    const config = await loadContext();
    return Object.keys(config.contexts).map(name => ({
        name,
        current: name === config.currentContext
    }));
}
//# sourceMappingURL=context.js.map
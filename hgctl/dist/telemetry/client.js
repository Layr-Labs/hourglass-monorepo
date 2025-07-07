"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.setupTelemetry = setupTelemetry;
exports.trackEvent = trackEvent;
exports.trackError = trackError;
const posthog_node_1 = require("posthog-node");
let posthog = null;
// Initialize PostHog if API key is provided
const POSTHOG_API_KEY = process.env.HGCTL_POSTHOG_API_KEY;
if (POSTHOG_API_KEY) {
    posthog = new posthog_node_1.PostHog(POSTHOG_API_KEY, {
        host: 'https://app.posthog.com',
        flushAt: 1,
        flushInterval: 0
    });
}
async function setupTelemetry(command) {
    if (!posthog) {
        return;
    }
    const commandPath = getCommandPath(command);
    const startTime = Date.now();
    // Track command execution
    posthog.capture({
        distinctId: getAnonymousId(),
        event: 'command_executed',
        properties: {
            command: commandPath,
            version: '1.0.0',
            platform: process.platform,
            node_version: process.version
        }
    });
    // Track command completion on exit
    process.on('exit', (code) => {
        if (posthog) {
            const duration = Date.now() - startTime;
            posthog.capture({
                distinctId: getAnonymousId(),
                event: 'command_completed',
                properties: {
                    command: commandPath,
                    exit_code: code,
                    duration_ms: duration
                }
            });
            posthog.flush();
        }
    });
}
function getCommandPath(command) {
    const parts = [];
    let current = command;
    while (current && current.name && current.name() !== 'hgctl') {
        parts.unshift(current.name());
        current = current.parent;
    }
    return parts.join(' ');
}
function getAnonymousId() {
    // Generate a consistent anonymous ID based on machine characteristics
    const crypto = require('crypto');
    const data = `${process.platform}-${process.arch}-${os_1.default.hostname()}`;
    return crypto.createHash('sha256').update(data).digest('hex').substring(0, 16);
}
// Expose telemetry methods for custom events
function trackEvent(event, properties) {
    if (posthog) {
        posthog.capture({
            distinctId: getAnonymousId(),
            event,
            properties
        });
    }
}
function trackError(error, context) {
    if (posthog) {
        posthog.capture({
            distinctId: getAnonymousId(),
            event: 'error_occurred',
            properties: {
                error_message: error.message,
                error_stack: error.stack,
                ...context
            }
        });
    }
}
const os_1 = __importDefault(require("os"));
//# sourceMappingURL=client.js.map
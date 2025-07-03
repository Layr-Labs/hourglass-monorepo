import { Command } from '@commander-js/extra-typings';
import { PostHog } from 'posthog-node';

let posthog: PostHog | null = null;

// Initialize PostHog if API key is provided
const POSTHOG_API_KEY = process.env.HGCTL_POSTHOG_API_KEY;

if (POSTHOG_API_KEY) {
  posthog = new PostHog(POSTHOG_API_KEY, {
    host: 'https://app.posthog.com',
    flushAt: 1,
    flushInterval: 0
  });
}

export async function setupTelemetry(command: any): Promise<void> {
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

function getCommandPath(command: any): string {
  const parts: string[] = [];
  let current = command;
  
  while (current && current.name && current.name() !== 'hgctl') {
    parts.unshift(current.name());
    current = current.parent;
  }
  
  return parts.join(' ');
}

function getAnonymousId(): string {
  // Generate a consistent anonymous ID based on machine characteristics
  const crypto = require('crypto');
  const data = `${process.platform}-${process.arch}-${os.hostname()}`;
  return crypto.createHash('sha256').update(data).digest('hex').substring(0, 16);
}

// Expose telemetry methods for custom events
export function trackEvent(event: string, properties?: Record<string, any>): void {
  if (posthog) {
    posthog.capture({
      distinctId: getAnonymousId(),
      event,
      properties
    });
  }
}

export function trackError(error: Error, context?: Record<string, any>): void {
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

import os from 'os';
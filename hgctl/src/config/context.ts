import { promises as fs } from 'fs';
import path from 'path';
import os from 'os';
import yaml from 'yaml';

export interface Context {
  name?: string;
  executorAddress: string;
  avsAddress?: string;
  operatorSetId?: number;
  networkId?: number;
  rpcUrl?: string;
  releaseManagerAddress?: string;
}

export interface Config {
  currentContext: string;
  contexts: Record<string, Context>;
}

const CONFIG_PATH = path.join(os.homedir(), '.hgctl', 'config.yaml');

export async function loadContext(): Promise<Config> {
  try {
    const data = await fs.readFile(CONFIG_PATH, 'utf-8');
    const config = yaml.parse(data) as Config;
    
    // Ensure default context exists
    if (!config.contexts.default) {
      config.contexts.default = {
        executorAddress: 'executor:9090'
      };
    }
    
    return config;
  } catch (error) {
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

export async function saveContext(config: Config): Promise<void> {
  const dir = path.dirname(CONFIG_PATH);
  await fs.mkdir(dir, { recursive: true });
  await fs.writeFile(CONFIG_PATH, yaml.stringify(config));
}

export async function getContext(name: string): Promise<Context | null> {
  const config = await loadContext();
  return config.contexts[name] || null;
}

export async function setCurrentContext(name: string): Promise<void> {
  const config = await loadContext();
  if (!config.contexts[name]) {
    throw new Error(`Context "${name}" does not exist`);
  }
  config.currentContext = name;
  await saveContext(config);
}

export async function updateContext(name: string, updates: Partial<Context>): Promise<void> {
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

export async function deleteContext(name: string): Promise<void> {
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

export async function listContexts(): Promise<{ name: string; current: boolean }[]> {
  const config = await loadContext();
  return Object.keys(config.contexts).map(name => ({
    name,
    current: name === config.currentContext
  }));
}
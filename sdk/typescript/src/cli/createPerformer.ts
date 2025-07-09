#!/usr/bin/env node

// CLI tool for scaffolding new Hourglass performers with TypeChain integration

import * as fs from 'fs';
import * as path from 'path';
import { execSync } from 'child_process';
import * as readline from 'readline';

interface ProjectConfig {
  name: string;
  description: string;
  useTypeChain: boolean;
  contractsPath?: string;
  port: number;
  author: string;
}

/**
 * CLI for creating new Hourglass performers
 */
class CreatePerformerCLI {
  private rl: readline.Interface;

  constructor() {
    this.rl = readline.createInterface({
      input: process.stdin,
      output: process.stdout,
    });
  }

  /**
   * Main CLI entry point
   */
  async run(args: string[]): Promise<void> {
    console.log('🏗️  Hourglass Performer Scaffolding Tool');
    console.log('━'.repeat(50));

    try {
      const config = await this.collectProjectConfig(args);
      const projectPath = await this.createProject(config);
      await this.installDependencies(projectPath);
      
      if (config.useTypeChain && config.contractsPath) {
        await this.setupTypeChain(projectPath, config.contractsPath);
      }

      await this.createExampleFiles(projectPath, config);
      this.showSuccessMessage(projectPath, config);
    } catch (error) {
      console.error('❌ Error creating performer:', error);
      process.exit(1);
    } finally {
      this.rl.close();
    }
  }

  /**
   * Collect project configuration from user
   */
  private async collectProjectConfig(args: string[]): Promise<ProjectConfig> {
    const config: ProjectConfig = {
      name: '',
      description: '',
      useTypeChain: false,
      port: 8080,
      author: 'AVS Developer',
    };

    // Check for project name in arguments
    if (args.length > 0) {
      config.name = args[0];
    } else {
      config.name = await this.ask('Project name: ');
    }
    
    // Ensure name is not empty
    if (!config.name.trim()) {
      throw new Error('Project name is required');
    }

    // Validate project name
    if (!config.name || !/^[a-zA-Z0-9-_]+$/.test(config.name)) {
      throw new Error('Invalid project name. Use alphanumeric characters, hyphens, and underscores only.');
    }

    config.description = await this.ask(`Description (optional): `) || `${config.name} AVS performer`;
    config.author = await this.ask('Author: ') || config.author;

    const portInput = await this.ask('Port (8080): ');
    if (portInput) {
      const port = parseInt(portInput, 10);
      if (isNaN(port) || port < 1 || port > 65535) {
        throw new Error('Invalid port number');
      }
      config.port = port;
    }

    const useTypeChainInput = await this.ask('Use TypeChain for Solidity contract integration? (y/N): ');
    config.useTypeChain = useTypeChainInput.toLowerCase() === 'y' || useTypeChainInput.toLowerCase() === 'yes';

    if (config.useTypeChain) {
      config.contractsPath = await this.ask('Path to contract ABIs (./contracts): ') || './contracts';
    }

    return config;
  }

  /**
   * Create project directory and basic structure
   */
  private async createProject(config: ProjectConfig): Promise<string> {
    const projectPath = path.resolve(process.cwd(), config.name);

    if (fs.existsSync(projectPath)) {
      throw new Error(`Directory ${config.name} already exists`);
    }

    console.log(`\n📁 Creating project structure in ${projectPath}`);
    
    // Create directory structure
    fs.mkdirSync(projectPath, { recursive: true });
    fs.mkdirSync(path.join(projectPath, 'src'), { recursive: true });
    fs.mkdirSync(path.join(projectPath, 'src', 'workers'), { recursive: true });
    fs.mkdirSync(path.join(projectPath, 'src', 'types'), { recursive: true });

    if (config.useTypeChain && config.contractsPath) {
      fs.mkdirSync(path.join(projectPath, config.contractsPath), { recursive: true });
      fs.mkdirSync(path.join(projectPath, 'typechain-types'), { recursive: true });
    }

    // Create package.json
    const packageJson = this.generatePackageJson(config);
    fs.writeFileSync(
      path.join(projectPath, 'package.json'),
      JSON.stringify(packageJson, null, 2)
    );

    // Create TypeScript config
    const tsConfig = this.generateTsConfig(config);
    fs.writeFileSync(
      path.join(projectPath, 'tsconfig.json'),
      JSON.stringify(tsConfig, null, 2)
    );

    // Create .gitignore
    const gitignore = this.generateGitignore(config);
    fs.writeFileSync(path.join(projectPath, '.gitignore'), gitignore);

    // Create README
    const readme = this.generateReadme(config);
    fs.writeFileSync(path.join(projectPath, 'README.md'), readme);

    return projectPath;
  }

  /**
   * Install npm dependencies
   */
  private async installDependencies(projectPath: string): Promise<void> {
    console.log('\n📦 Installing dependencies...');
    
    try {
      execSync('npm install', { 
        cwd: projectPath, 
        stdio: 'inherit' 
      });
    } catch (error) {
      throw new Error('Failed to install dependencies');
    }
  }

  /**
   * Set up TypeChain configuration
   */
  private async setupTypeChain(projectPath: string, contractsPath: string): Promise<void> {
    console.log('\n⚙️  Setting up TypeChain...');

    // Create typechain config
    const typechainConfig = {
      files: [`${contractsPath}/**/*.json`],
      target: 'ethers-v6',
      outDir: 'typechain-types',
    };

    fs.writeFileSync(
      path.join(projectPath, 'typechain.config.json'),
      JSON.stringify(typechainConfig, null, 2)
    );

    // Create sample ABI if contracts directory is empty
    const fullContractsPath = path.join(projectPath, contractsPath);
    if (fs.readdirSync(fullContractsPath).length === 0) {
      const sampleAbi = this.generateSampleAbi();
      fs.writeFileSync(
        path.join(fullContractsPath, 'SampleContract.json'),
        JSON.stringify(sampleAbi, null, 2)
      );
    }
  }

  /**
   * Create example files based on configuration
   */
  private async createExampleFiles(projectPath: string, config: ProjectConfig): Promise<void> {
    console.log('\n📝 Creating example files...');

    // Create main performer file
    const mainPerformer = config.useTypeChain 
      ? this.generateSolidityPerformer(config)
      : this.generateBasicPerformer(config);
    
    fs.writeFileSync(
      path.join(projectPath, 'src', 'performer.ts'),
      mainPerformer
    );

    // Create Docker files
    const dockerfile = this.generateDockerfile(config);
    fs.writeFileSync(path.join(projectPath, 'Dockerfile'), dockerfile);

    const dockerCompose = this.generateDockerCompose(config);
    fs.writeFileSync(path.join(projectPath, 'docker-compose.yml'), dockerCompose);

    // Create example worker
    if (config.useTypeChain) {
      const solidityWorker = this.generateSolidityWorkerExample(config);
      fs.writeFileSync(
        path.join(projectPath, 'src', 'workers', 'sampleWorker.ts'),
        solidityWorker
      );
    } else {
      const basicWorker = this.generateBasicWorkerExample(config);
      fs.writeFileSync(
        path.join(projectPath, 'src', 'workers', 'sampleWorker.ts'),
        basicWorker
      );
    }
  }

  /**
   * Generate package.json content
   */
  private generatePackageJson(config: ProjectConfig): any {
    const basePackage: any = {
      name: config.name,
      version: '1.0.0',
      description: config.description,
      main: 'dist/performer.js',
      scripts: {
        build: 'tsc',
        'build:watch': 'tsc --watch',
        dev: 'ts-node-dev --respawn --transpile-only src/performer.ts',
        start: 'node dist/performer.js',
        'docker:build': `docker build -t ${config.name} .`,
        'docker:run': `docker run -p ${config.port}:${config.port} ${config.name}`,
      },
      keywords: ['hourglass', 'avs', 'performer', 'eigenlayer'],
      author: config.author,
      license: 'MIT',
      dependencies: {
        '@hourglass/performer': '^0.1.0',
      },
      devDependencies: {
        '@types/node': '^20.10.5',
        typescript: '^5.3.3',
        'ts-node-dev': '^2.0.0',
      },
    };

    if (config.useTypeChain) {
      basePackage.scripts['typechain'] = 'typechain --target ethers-v6 --out-dir typechain-types \'contracts/**/*.json\'';
      basePackage.scripts['prebuild'] = 'npm run typechain';
      basePackage.dependencies['ethers'] = '^6.8.0';
      basePackage.devDependencies['typechain'] = '^8.3.0';
      basePackage.devDependencies['@typechain/ethers-v6'] = '^0.5.0';
    }

    return basePackage;
  }

  /**
   * Generate TypeScript configuration
   */
  private generateTsConfig(config: ProjectConfig): any {
    return {
      compilerOptions: {
        target: 'ES2020',
        module: 'commonjs',
        lib: ['ES2020'],
        outDir: './dist',
        rootDir: './src',
        strict: true,
        esModuleInterop: true,
        skipLibCheck: true,
        forceConsistentCasingInFileNames: true,
        declaration: true,
        declarationMap: true,
        sourceMap: true,
        resolveJsonModule: true,
        typeRoots: ['./node_modules/@types', './typechain-types'],
      },
      include: ['src/**/*'],
      exclude: ['node_modules', 'dist'],
    };
  }

  /**
   * Generate .gitignore content
   */
  private generateGitignore(config: ProjectConfig): string {
    let gitignore = `# Dependencies
node_modules/
npm-debug.log*
yarn-debug.log*
yarn-error.log*

# Build output
dist/
*.tsbuildinfo

# Environment files
.env
.env.local
.env.development.local
.env.test.local
.env.production.local

# Logs
logs/
*.log

# IDE
.vscode/
.idea/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db
`;

    if (config.useTypeChain) {
      gitignore += `
# TypeChain generated files
typechain-types/
`;
    }

    return gitignore;
  }

  /**
   * Generate README content
   */
  private generateReadme(config: ProjectConfig): string {
    return `# ${config.name}

${config.description}

## Quick Start

\`\`\`bash
# Install dependencies
npm install

# Build the project
npm run build

# Run in development mode
npm run dev

# Run in production
npm start
\`\`\`

## Docker

\`\`\`bash
# Build Docker image
npm run docker:build

# Run container
npm run docker:run
\`\`\`

${config.useTypeChain ? `## TypeChain Integration

This project uses TypeChain for type-safe Solidity contract integration.

\`\`\`bash
# Generate TypeChain types from contract ABIs
npm run typechain

# The types will be generated in ./typechain-types/
\`\`\`

Add your contract ABIs to the \`${config.contractsPath}\` directory and run \`npm run typechain\` to generate TypeScript types.

` : ''}## Architecture

- **Performer**: Main gRPC server handling task execution
- **Workers**: Task processing logic${config.useTypeChain ? ' with Solidity contract integration' : ''}
- **Health & Monitoring**: Built-in health checks and metrics collection

## Development

The performer server will automatically restart when you make changes to the source code in development mode.

## Production Deployment

1. Build the project: \`npm run build\`
2. Start the server: \`npm start\`
3. The server will listen on port ${config.port}

## Configuration

Configure your performer by modifying the server configuration in \`src/performer.ts\`.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## License

MIT
`;
  }

  /**
   * Generate sample ABI for TypeChain
   */
  private generateSampleAbi(): any {
    return {
      abi: [
        {
          inputs: [
            { name: 'taskId', type: 'bytes32' },
            { name: 'data', type: 'bytes' },
            { name: 'user', type: 'address' },
          ],
          name: 'processTask',
          outputs: [
            { name: 'result', type: 'bytes' },
            { name: 'success', type: 'bool' },
          ],
          stateMutability: 'nonpayable',
          type: 'function',
        },
        {
          inputs: [
            { name: 'value', type: 'uint256' },
          ],
          name: 'square',
          outputs: [
            { name: 'result', type: 'uint256' },
          ],
          stateMutability: 'pure',
          type: 'function',
        },
      ],
      bytecode: '0x608060405234801561001057600080fd5b50...',
    };
  }

  /**
   * Generate basic performer implementation
   */
  private generateBasicPerformer(config: ProjectConfig): string {
    return `import { BaseWorker } from '@hourglass/performer';

/**
 * ${config.name} performer
 */
class ${this.toPascalCase(config.name)}Performer extends BaseWorker {
  async handleSimpleTask(input: any) {
    // TODO: Implement your AVS logic here
    // Example: Process a number and return its square
    const result = typeof input === 'number' ? input * input : 42;
    
    return result;
  }
}

// One-line server startup
new ${this.toPascalCase(config.name)}Performer().start(${config.port});
`;
  }

  /**
   * Generate Solidity performer implementation
   */
  private generateSolidityPerformer(config: ProjectConfig): string {
    return `import { SolidityWorker } from '@hourglass/performer';

// TODO: Import your contract types from typechain-types
// import type { MyContract } from './typechain-types';

/**
 * ${config.name} performer with TypeChain integration
 */
class ${this.toPascalCase(config.name)}Performer extends SolidityWorker<any, 'processTask'> {
  async handleSolidityTask(params: { taskId: string; data: Uint8Array; user: string }) {
    // TODO: Implement your AVS logic here
    // params are automatically typed based on your contract ABI
    
    const { taskId, data, user } = params;
    
    // Example: Process the data and return result
    const processedData = new TextEncoder().encode(
      \`Processed: \${new TextDecoder().decode(data)}\`
    );
    
    return {
      result: processedData,
      success: true,
    };
  }
}

// One-line server startup
new ${this.toPascalCase(config.name)}Performer().start(${config.port});
`;
  }

  /**
   * Generate basic worker example
   */
  private generateBasicWorkerExample(config: ProjectConfig): string {
    return `import { BaseWorker } from '@hourglass/performer';
import { TaskRequest, TaskResponse } from '@hourglass/performer';

export class SampleWorker extends BaseWorker {
  async handleTask(task: TaskRequest): Promise<TaskResponse> {
    try {
      // Parse the task payload
      const input = this.parsePayload(task.payload);
      
      // Process the task - example: square a number
      const result = typeof input === 'number' ? input * input : 42;
      
      // Return the response
      return this.createResponse(task.taskId, result);
    } catch (error) {
      throw new Error(\`Task processing failed: \${error instanceof Error ? error.message : 'Unknown error'}\`);
    }
  }

  async validateTask(task: TaskRequest): Promise<void> {
    await super.validateTask(task);
    
    // Add custom validation logic here
    if (!task.payload || task.payload.length === 0) {
      throw new Error('Task payload is required');
    }
  }
}
`;
  }

  /**
   * Generate Solidity worker example
   */
  private generateSolidityWorkerExample(config: ProjectConfig): string {
    return `import { SolidityWorker } from '@hourglass/performer';
import { SampleContract } from '../typechain-types/SampleContract';

// Import the contract ABI
const contractAbi = require('../../contracts/SampleContract.json').abi;

export class SampleSolidityWorker extends SolidityWorker<SampleContract, 'processTask'> {
  constructor() {
    super({
      abi: contractAbi,
      functionName: 'processTask',
      autoDetectPayload: true,
      strictMode: false,
    });
  }

  async handleSolidityTask(
    params: { taskId: string; data: Uint8Array; user: string }
  ): Promise<{ result: Uint8Array; success: boolean }> {
    try {
      // Access typed parameters from the contract
      const { taskId, data, user } = params;
      
      console.log(\`Processing task \${taskId} for user \${user}\`);
      
      // Simulate some processing
      const processedData = new TextEncoder().encode(
        \`Processed: \${new TextDecoder().decode(data)}\`
      );
      
      return {
        result: processedData,
        success: true,
      };
    } catch (error) {
      return {
        result: new TextEncoder().encode(\`Error: \${error instanceof Error ? error.message : 'Unknown error'}\`),
        success: false,
      };
    }
  }
}

// Alternative: Simple worker using utility function
export class SimpleSolidityWorker extends SolidityWorker<SampleContract, 'square'> {
  constructor() {
    super({
      abi: contractAbi,
      functionName: 'square',
      autoDetectPayload: true,
    });
  }

  async handleSolidityTask(params: { value: bigint }): Promise<{ result: bigint }> {
    const { value } = params;
    
    return {
      result: value * value,
    };
  }
}
`;
  }

  /**
   * Generate Dockerfile
   */
  private generateDockerfile(config: ProjectConfig): string {
    return `FROM node:18-alpine

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci --only=production

# Copy source code
COPY . .

# Build the application
RUN npm run build

# Expose port
EXPOSE ${config.port}

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \\
  CMD node -e "const grpc = require('@grpc/grpc-js'); const client = new grpc.Client('localhost:${config.port}', grpc.credentials.createInsecure()); client.close();"

# Run the application
CMD ["npm", "start"]
`;
  }

  /**
   * Generate docker-compose.yml
   */
  private generateDockerCompose(config: ProjectConfig): string {
    return `version: '3.8'

services:
  ${config.name}:
    build: .
    ports:
      - "${config.port}:${config.port}"
    environment:
      - NODE_ENV=production
      - PORT=${config.port}
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "nc", "-z", "localhost", "${config.port}"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s
`;
  }

  /**
   * Show success message with next steps
   */
  private showSuccessMessage(projectPath: string, config: ProjectConfig): void {
    console.log('\n🎉 Project created successfully!');
    console.log('━'.repeat(50));
    console.log(`📁 Location: ${projectPath}`);
    console.log(`🚀 Port: ${config.port}`);
    console.log(`📄 TypeChain: ${config.useTypeChain ? 'Enabled' : 'Disabled'}`);
    
    console.log('\n🏃 Next steps:');
    console.log(`  cd ${config.name}`);
    
    if (config.useTypeChain) {
      console.log('  # Add your contract ABIs to ./contracts/');
      console.log('  npm run typechain');
    }
    
    console.log('  npm run dev');
    
    console.log('\n📚 Available commands:');
    console.log('  npm run build     - Build for production');
    console.log('  npm run dev       - Start development server');
    console.log('  npm start         - Start production server');
    console.log('  npm run docker:build - Build Docker image');
    console.log('  npm run docker:run   - Run Docker container');
    
    if (config.useTypeChain) {
      console.log('  npm run typechain - Generate TypeChain types');
    }
    
    console.log('\n🔗 Resources:');
    console.log('  • Documentation: https://docs.hourglass.io');
    console.log('  • Examples: https://github.com/Layr-Labs/hourglass-examples');
    console.log('  • Support: https://discord.gg/hourglass');
  }

  /**
   * Prompt user for input
   */
  private ask(question: string): Promise<string> {
    return new Promise((resolve) => {
      this.rl.question(question, resolve);
    });
  }

  /**
   * Convert kebab-case to PascalCase
   */
  private toPascalCase(str: string): string {
    return str.split('-').map(word => word.charAt(0).toUpperCase() + word.slice(1)).join('');
  }
}

// CLI entry point
if (require.main === module) {
  const cli = new CreatePerformerCLI();
  const args = process.argv.slice(2);
  cli.run(args).catch(console.error);
}

export default CreatePerformerCLI;
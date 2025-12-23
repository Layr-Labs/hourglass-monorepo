# Hourglass TypeScript Performer SDK

Build EigenLayer AVS performers in TypeScript with minimal setup and maximum type safety.

## 🚀 Quick Start

### 1. Basic Performer

```typescript
// performer.ts
import { BaseWorker } from '@layr-labs/hourglass-performer';

class MyBasicPerformer extends BaseWorker {
  async handleSimpleTask(input: any) {
    // Your AVS logic here
    return input * input; // Example: square the input
  }
}

// One-line server startup
new MyBasicPerformer().start();
```

### 2. Solidity Contract Integration

```typescript
// performer.ts
import { SolidityWorker } from '@layr-labs/hourglass-performer';

class MySolidityPerformer extends SolidityWorker<any, 'processTask'> {
  async handleSolidityTask(params: { amount: bigint; user: string }) {
    // Fully typed based on your contract ABI
    const result = params.amount * 2n;
    return { result };
  }
}

new MySolidityPerformer().start();
```

### 3. Deploy with Docker

```bash
# Build your performer
npm run build

# Build Docker image
npm run docker:build

# Run container
npm run docker:run
```

## 📋 Table of Contents

- [Installation](#installation)
- [Core Concepts](#core-concepts)
- [API Reference](#api-reference)
- [TypeChain Integration](#typechain-integration)
- [Environment Configuration](#environment-configuration)
- [Docker Deployment](#docker-deployment)
- [Examples](#examples)
- [Troubleshooting](#troubleshooting)
- [Monitoring](#monitoring)
- [Development Workflow](#development-workflow)
- [Contributing](#contributing)

## 📦 Installation

```bash
npm install @layr-labs/hourglass-performer
```

### Package Usage

When using the SDK as a package dependency:

```typescript
import { PerformerServer, IWorker } from '@layr-labs/hourglass-performer';

// Implement your worker
class MyWorker implements IWorker {
  async execute(payload: Uint8Array): Promise<Uint8Array> {
    // Your task logic here
    return new Uint8Array(Buffer.from('Hello from worker!'));
  }
}

// Create and start the server
const worker = new MyWorker();
const server = new PerformerServer(worker, { port: 8080 });

async function main() {
  await server.start();
  console.log('Performer server started on port 8080');
}

main().catch(console.error);
```

#### Protobuf File Resolution

The SDK automatically resolves protobuf files from multiple possible locations:

1. **Development environment**: `proto/performer.proto` relative to source
2. **Package installation**: `node_modules/@layr-labs/hourglass-performer/proto/performer.proto`
3. **Alternative paths**: Various fallback locations

If you encounter protobuf resolution errors:

```bash
# Verify proto files are included
ls node_modules/@layr-labs/hourglass-performer/proto/

# Or use the utility function
import { resolveProtoPath } from '@layr-labs/hourglass-performer';

try {
  const protoPath = resolveProtoPath('performer.proto');
  console.log('Proto file found at:', protoPath);
} catch (error) {
  console.error('Proto file not found:', error.message);
}
```

## 🔧 Core Concepts

### Worker Types

#### BaseWorker
For simple task processing without contract interaction:

```typescript
import { BaseWorker } from '@layr-labs/hourglass-performer';

class SimpleWorker extends BaseWorker {
  async handleSimpleTask(input: any) {
    // Process input and return result
    return processInput(input);
  }
}
```

#### SolidityWorker
For type-safe Solidity contract interaction:

```typescript
import { SolidityWorker } from '@layr-labs/hourglass-performer';

class ContractWorker extends SolidityWorker<MyContract, 'processTask'> {
  async handleSolidityTask(params: ProcessTaskParams) {
    // Fully typed parameters from your contract
    return { result: params.amount * 2n };
  }
}
```

### Server Architecture

The SDK provides a complete gRPC server implementation:

- **Automatic Configuration**: No manual gRPC setup required
- **Health Checks**: Built-in health monitoring
- **Metrics Collection**: Performance monitoring
- **Graceful Shutdown**: Clean shutdown handling
- **Error Handling**: Comprehensive error management

## 📚 API Reference

### BaseWorker

#### Methods

##### `handleSimpleTask(input: any): Promise<any>`
Override this method to implement your task processing logic.

```typescript
class MyWorker extends BaseWorker {
  async handleSimpleTask(input: any) {
    // Your processing logic
    return result;
  }
}
```

##### `start(port?: number): Promise<void>`
Start the performer server.

```typescript
const worker = new MyWorker();
await worker.start(8080); // Optional port override
```

##### `validateTask(task: TaskRequest): Promise<void>`
Custom task validation (optional override).

```typescript
class MyWorker extends BaseWorker {
  async validateTask(task: TaskRequest) {
    await super.validateTask(task);
    // Additional validation logic
  }
}
```

### SolidityWorker

#### Constructor Options

```typescript
interface SolidityWorkerConfig {
  abi: any[];                    // Contract ABI
  functionName: string;          // Function to handle
  autoDetectPayload?: boolean;   // Auto-detect payload format
  strictMode?: boolean;          // Strict validation mode
}
```

#### Methods

##### `handleSolidityTask(params: T): Promise<R>`
Type-safe contract function handling.

```typescript
class ContractWorker extends SolidityWorker<MyContract, 'processTask'> {
  async handleSolidityTask(params: ProcessTaskParams): Promise<ProcessTaskReturn> {
    // Fully typed based on contract ABI
    return { result: params.amount * 2n };
  }
}
```

### PerformerServer

#### Configuration

```typescript
interface PerformerServerConfig {
  port?: number;     // Server port (default: 8080)
  timeout?: number;  // Request timeout (default: 10000ms)
  debug?: boolean;   // Debug mode (default: false)
}
```

#### Advanced Usage

```typescript
import { PerformerServer } from '@layr-labs/hourglass-performer';

const server = new PerformerServer(worker, {
  port: 8080,
  timeout: 15000,
  debug: true,
});

await server.start();
```

## 🔗 TypeChain Integration

### Setup

1. **Install TypeChain** (included in scaffolded projects):
```bash
npm install --save-dev typechain @typechain/ethers-v6
```

2. **Add contract ABIs** to `./contracts/`:
```bash
cp MyContract.json contracts/
```

3. **Generate TypeScript types**:
```bash
npm run typechain
```

### Usage

```typescript
// Import generated types
import type { MyContract } from './typechain-types';

// Use with SolidityWorker
class TypedWorker extends SolidityWorker<MyContract, 'processTask'> {
  async handleSolidityTask(params: ProcessTaskParams): Promise<ProcessTaskReturn> {
    // Full IntelliSense and type safety
    const { taskId, amount, user } = params;
    
    // Type-safe return
    return {
      result: amount * 2n,
      success: true,
    };
  }
}
```

### Configuration

```typescript
// typechain.config.json
{
  "files": ["contracts/**/*.json"],
  "target": "ethers-v6",
  "outDir": "typechain-types"
}
```

## ⚙️ Environment Configuration

### Environment Variables

```bash
# Server configuration
PORT=8080
HOST=0.0.0.0
TIMEOUT=10000

# Application configuration
NODE_ENV=production
DEBUG=false
LOG_LEVEL=info

# Performance
MAX_CONCURRENT_TASKS=10
TASK_TIMEOUT=30000

# TypeChain
TYPECHAIN_ENABLED=true
CONTRACTS_PATH=./contracts

# Security
CORS_ENABLED=true
CORS_ORIGINS=https://yourdomain.com
```

### Configuration Loading

```typescript
import { loadEnvironmentConfig } from '@layr-labs/hourglass-performer';

const config = loadEnvironmentConfig();
console.log(`Server will run on port ${config.port}`);
```

### Environment Files

```bash
# Development
.env.development

# Production
.env.production

# Local development
.env.local
```

## 🐳 Docker Deployment

### Basic Deployment

```dockerfile
FROM node:24-alpine

WORKDIR /app

# Copy package files
COPY package*.json ./
RUN npm ci --only=production

# Copy built application
COPY dist/ ./dist/

# Run performer
CMD ["node", "dist/performer.js"]
```

### Docker Compose

```yaml
version: '3.8'

services:
  performer:
    build: .
    ports:
      - "8080:8080"
    environment:
      - NODE_ENV=production
      - LOG_LEVEL=info
    volumes:
      - ./contracts:/app/contracts
    restart: unless-stopped
```

### Build and Run

```bash
# Build
npm run docker:build

# Run
npm run docker:run

# Development with Docker Compose
npm run docker:dev
```

## 🎯 Examples

### Basic Number Processing

```typescript
import { BaseWorker } from '@layr-labs/hourglass-performer';

class NumberProcessor extends BaseWorker {
  async handleSimpleTask(input: number) {
    return {
      original: input,
      squared: input * input,
      doubled: input * 2,
      timestamp: Date.now(),
    };
  }
}

new NumberProcessor().start();
```

### ERC-20 Token Processing

```typescript
import { SolidityWorker } from '@layr-labs/hourglass-performer';
import type { ERC20 } from './typechain-types';

class TokenProcessor extends SolidityWorker<ERC20, 'transfer'> {
  async handleSolidityTask(params: { to: string; amount: bigint }) {
    const { to, amount } = params;
    
    // Validate transfer
    if (amount <= 0n) {
      throw new Error('Invalid amount');
    }
    
    // Process transfer logic
    return {
      success: true,
      txHash: '0x...',
    };
  }
}

new TokenProcessor().start();
```

### Data Aggregation

```typescript
import { BaseWorker } from '@layr-labs/hourglass-performer';

class DataAggregator extends BaseWorker {
  async handleSimpleTask(data: { values: number[] }) {
    const { values } = data;
    
    return {
      sum: values.reduce((a, b) => a + b, 0),
      avg: values.reduce((a, b) => a + b, 0) / values.length,
      min: Math.min(...values),
      max: Math.max(...values),
      count: values.length,
    };
  }
}

new DataAggregator().start();
```

## 🔍 Troubleshooting

### Common Issues

#### Port Already in Use
```bash
# Check what's using the port
lsof -i :8080

# Use a different port
PORT=8081 npm run dev
```

#### TypeChain Generation Fails
```bash
# Ensure contract ABIs are valid JSON
npm run typechain -- --show-stack-traces
```

#### Container Won't Start
```bash
# Check container logs
docker logs <container-id>

# Run in interactive mode
docker run -it my-performer sh
```

#### "Could not find performer.proto file"

**Solution**: Ensure the package was installed correctly and includes proto files:

```bash
# Reinstall the package
npm uninstall @layr-labs/hourglass-performer
npm install @layr-labs/hourglass-performer

# Verify proto files exist
ls node_modules/@layr-labs/hourglass-performer/proto/
```

#### gRPC Service Definition Not Found

**Solution**: This typically happens when protobuf files can't be loaded. Use the protobuf troubleshooting steps above.

#### Module Resolution in Monorepo

**Solution**: If using in a monorepo, ensure proper module resolution:

```typescript
// In your tsconfig.json
{
  "compilerOptions": {
    "moduleResolution": "node",
    "esModuleInterop": true,
    "allowSyntheticDefaultImports": true
  }
}
```

### Debug Mode

Enable debug logging:

```bash
DEBUG=true LOG_LEVEL=debug npm run dev
```

### Health Check

```bash
# Check performer health
curl http://localhost:8080/health

# Expected response
{
  "status": "READY_FOR_TASK"
}
```

## 📊 Monitoring

### Built-in Metrics

The SDK includes comprehensive monitoring:

- **Task Metrics**: Execution time, success/failure rates
- **Health Monitoring**: System health checks
- **Performance Tracking**: Memory usage, request counts
- **Error Tracking**: Error rates and patterns

### Accessing Metrics

```bash
# Metrics endpoint (if enabled)
curl http://localhost:9090/metrics

# Health status
curl http://localhost:8080/health
```

## 🔄 Development Workflow

### Local Development

```bash
# Start development server with hot reload
npm run dev

# Run tests
npm test

# Type checking
npm run build

# Lint code
npm run lint
```

### Testing

```bash
# Unit tests
npm test

# Watch mode
npm run test:watch

# Coverage
npm run test:coverage
```

### Production Build

```bash
# Build for production
npm run build

# Start production server
npm start
```

## 🤝 Contributing

### Development Setup

```bash
# Clone repository
git clone https://github.com/Layr-Labs/hourglass-monorepo

# Navigate to TypeScript SDK
cd sdk/typescript

# Install dependencies
npm install

# Build the project
npm run build

# Run tests
npm test
```

### Code Style

- Use TypeScript strict mode
- Follow ESLint configuration
- Write tests for new features
- Document public APIs

### Pull Request Process

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests and documentation
5. Submit a pull request

## 📝 License

MIT License - see [LICENSE](LICENSE) for details.

## 🔗 Resources

- **Documentation**: [https://docs.hourglass.io](https://docs.hourglass.io)
- **Examples**: [https://github.com/Layr-Labs/hourglass-examples](https://github.com/Layr-Labs/hourglass-examples)
- **Discord**: [https://discord.gg/hourglass](https://discord.gg/hourglass)
- **GitHub**: [https://github.com/Layr-Labs/hourglass-monorepo](https://github.com/Layr-Labs/hourglass-monorepo)

## 📈 Version History

- **v0.1.0**: Initial release with basic performer support
- **v0.2.0**: Added TypeChain integration
- **v0.3.0**: Docker deployment support
- **v1.0.0**: Stable release with full feature set

---

Built with ❤️ by the Hourglass team for the EigenLayer ecosystem.
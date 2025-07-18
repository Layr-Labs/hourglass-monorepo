# Troubleshooting Guide

Common issues and solutions for the Hourglass TypeScript Performer SDK.

## Table of Contents

- [Installation Issues](#installation-issues)
- [Build Issues](#build-issues)
- [Runtime Issues](#runtime-issues)
- [TypeChain Issues](#typechain-issues)
- [Docker Issues](#docker-issues)
- [Performance Issues](#performance-issues)
- [Networking Issues](#networking-issues)
- [Development Issues](#development-issues)
- [Debugging Tips](#debugging-tips)
- [Getting Help](#getting-help)

## Installation Issues

### NPM Installation Fails

#### Problem
```bash
npm install @hourglass/performer
# Error: Unable to resolve dependency tree
```

#### Solution
```bash
# Clear npm cache
npm cache clean --force

# Delete node_modules and package-lock.json
rm -rf node_modules package-lock.json

# Reinstall
npm install

# Or use yarn
yarn install
```

### CLI Tool Not Found

#### Problem
```bash
npx @hourglass/create-performer my-avs
# Error: command not found
```

#### Solution
```bash
# Install globally first
npm install -g @hourglass/performer

# Or use full package name
npx @hourglass/create-performer my-avs

# Or specify exact version
npx @hourglass/performer@latest my-avs
```

### Permission Errors

#### Problem
```bash
# Error: EACCES: permission denied
```

#### Solution
```bash
# Use npm prefix to install in user directory
npm config set prefix ~/.npm-global
export PATH=~/.npm-global/bin:$PATH

# Or use yarn
yarn global add @hourglass/performer

# Or use sudo (not recommended)
sudo npm install -g @hourglass/performer
```

## Build Issues

### TypeScript Compilation Errors

#### Problem
```bash
npm run build
# Error: Cannot find module '@hourglass/performer'
```

#### Solution
```bash
# Ensure the package is installed
npm install @hourglass/performer

# Check tsconfig.json paths
{
  "compilerOptions": {
    "moduleResolution": "node",
    "esModuleInterop": true
  }
}

# Clear build cache
npm run clean
npm run build
```

### Protobuf Generation Fails

#### Problem
```bash
npm run generate:proto
# Error: protoc: command not found
```

#### Solution
```bash
# Install protobuf compiler
# macOS
brew install protobuf

# Ubuntu/Debian
sudo apt-get install protobuf-compiler

# Or use Docker
docker run --rm -v $(pwd):/workspace -w /workspace \
  namely/protoc-all -f proto/performer.proto -l typescript
```

### Module Resolution Issues

#### Problem
```typescript
// Error: Cannot resolve module
import { BaseWorker } from '@hourglass/performer';
```

#### Solution
```typescript
// Use relative imports for local development
import { BaseWorker } from '../src/worker/iWorker';

// Or check package.json main/types fields
{
  "main": "dist/index.js",
  "types": "dist/index.d.ts"
}
```

## Runtime Issues

### Server Won't Start

#### Problem
```bash
npm start
# Error: Port 8080 is already in use
```

#### Solution
```bash
# Check what's using the port
lsof -i :8080
netstat -tulnp | grep :8080

# Kill the process
kill -9 <PID>

# Or use different port
PORT=8081 npm start

# Or in code
new MyWorker().start(8081);
```

### gRPC Connection Issues

#### Problem
```bash
# Error: 14 UNAVAILABLE: No connection established
```

#### Solution
```typescript
// Ensure server is listening on correct interface
const server = new PerformerServer(worker, {
  port: 8080,
  host: '0.0.0.0', // Listen on all interfaces
});

// Check firewall settings
sudo ufw allow 8080

// Test connection
grpcurl -plaintext localhost:8080 list
```

### Task Execution Fails

#### Problem
```bash
# Error: Task processing failed: Unknown error
```

#### Solution
```typescript
// Add proper error handling
class MyWorker extends BaseWorker {
  async handleSimpleTask(input: any) {
    try {
      return await this.processInput(input);
    } catch (error) {
      console.error('Processing failed:', error);
      
      // Return error information
      return {
        success: false,
        error: error.message,
        timestamp: Date.now(),
      };
    }
  }
}

// Enable debug logging
DEBUG=true LOG_LEVEL=debug npm start
```

### Memory Issues

#### Problem
```bash
# Error: JavaScript heap out of memory
```

#### Solution
```bash
# Increase heap size
node --max-old-space-size=4096 dist/performer.js

# Or set environment variable
export NODE_OPTIONS="--max-old-space-size=4096"

# Check for memory leaks
node --inspect dist/performer.js
```

## TypeChain Issues

### TypeChain Generation Fails

#### Problem
```bash
npm run typechain
# Error: Cannot find any files matching pattern
```

#### Solution
```bash
# Check contracts directory exists
ls -la contracts/

# Ensure ABI files are present
ls -la contracts/*.json

# Verify ABI format
cat contracts/MyContract.json | jq .abi

# Run with debug output
npm run typechain -- --show-stack-traces
```

### Invalid ABI Files

#### Problem
```bash
# Error: Invalid ABI: Expected array
```

#### Solution
```bash
# Check ABI format - should be array or object with abi property
# Correct format 1 (array):
[
  {
    "name": "myFunction",
    "type": "function",
    "inputs": [...],
    "outputs": [...]
  }
]

# Correct format 2 (object):
{
  "abi": [...],
  "bytecode": "0x..."
}

# Validate JSON
cat contracts/MyContract.json | jq .
```

### TypeChain Types Not Found

#### Problem
```typescript
// Error: Cannot find module './typechain-types'
import type { MyContract } from './typechain-types';
```

#### Solution
```bash
# Generate types first
npm run typechain

# Check output directory
ls -la typechain-types/

# Ensure tsconfig includes typechain types
{
  "compilerOptions": {
    "typeRoots": ["./node_modules/@types", "./typechain-types"]
  }
}
```

### Function Signature Mismatch

#### Problem
```typescript
// Error: Property 'myFunction' does not exist on type
class Worker extends SolidityWorker<MyContract, 'myFunction'> {
  // Function not found in contract
}
```

#### Solution
```typescript
// Check function exists in ABI
console.log(this.getAllFunctions());

// Use correct function name
class Worker extends SolidityWorker<MyContract, 'processTask'> {
  // Use exact function name from ABI
}

// Or check function info
const info = this.getFunctionInfo('processTask');
console.log(info);
```

## Docker Issues

### Docker Build Fails

#### Problem
```bash
docker build -t my-performer .
# Error: npm install failed
```

#### Solution
```dockerfile
# Use specific Node version
FROM node:24-alpine

# Install build dependencies
RUN apk add --no-cache python3 make g++

# Copy package files first
COPY package*.json ./
RUN npm ci --only=production

# Then copy source
COPY dist/ ./dist/
```

### Container Won't Start

#### Problem
```bash
docker run my-performer
# Error: Cannot find module
```

#### Solution
```bash
# Check if build output exists
docker run -it my-performer ls -la dist/

# Ensure main entry point is correct
{
  "main": "dist/performer.js"
}

# Check container logs
docker logs <container-id>

# Run interactively for debugging
docker run -it my-performer sh
```

### Port Mapping Issues

#### Problem
```bash
# Container runs but can't connect
docker run -p 8080:8080 my-performer
curl localhost:8080 # Connection refused
```

#### Solution
```bash
# Check if container is listening on all interfaces
# In your code:
const server = new PerformerServer(worker, {
  host: '0.0.0.0', // Not 'localhost'
  port: 8080,
});

# Check Docker networking
docker network ls
docker inspect <container-id>
```

### Health Check Fails

#### Problem
```bash
# Container marked as unhealthy
docker ps # Shows unhealthy status
```

#### Solution
```dockerfile
# Fix health check command
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD node -e "const http = require('http'); \
        const options = { hostname: 'localhost', port: 8080, path: '/health' }; \
        const req = http.request(options, (res) => { \
            process.exit(res.statusCode === 200 ? 0 : 1); \
        }); \
        req.on('error', () => process.exit(1)); \
        req.end();"
```

## Performance Issues

### Slow Task Processing

#### Problem
```bash
# Tasks taking too long to process
```

#### Solution
```typescript
// Add timeout configuration
const server = new PerformerServer(worker, {
  timeout: 30000, // 30 seconds
});

// Monitor processing time
class TimedWorker extends BaseWorker {
  async handleSimpleTask(input: any) {
    const start = Date.now();
    
    try {
      const result = await this.processInput(input);
      const duration = Date.now() - start;
      
      console.log(`Task completed in ${duration}ms`);
      return result;
    } catch (error) {
      console.error(`Task failed after ${Date.now() - start}ms`);
      throw error;
    }
  }
}
```

### High Memory Usage

#### Problem
```bash
# Memory usage keeps growing
```

#### Solution
```typescript
// Enable memory monitoring
const server = new PerformerServer(worker, {
  // Enable health checks
}, {
  // Task processor config
  enableMetrics: true,
}, {
  // Health config
  checkInterval: 10000,
});

// Add memory health provider
import { MemoryHealthProvider } from '@hourglass/performer';
server.addHealthProvider(new MemoryHealthProvider(256)); // 256MB limit
```

### Connection Pooling Issues

#### Problem
```bash
# Too many connections
```

#### Solution
```typescript
// Configure connection limits
const server = new PerformerServer(worker, {
  timeout: 10000,
}, {
  maxConcurrentTasks: 10, // Limit concurrent tasks
});

// Monitor active connections
setInterval(() => {
  const usage = process.memoryUsage();
  console.log('Memory usage:', usage);
}, 30000);
```

## Networking Issues

### gRPC Connection Refused

#### Problem
```bash
grpcurl -plaintext localhost:8080 list
# Error: connection refused
```

#### Solution
```bash
# Check if server is running
netstat -tulnp | grep :8080

# Check firewall
sudo ufw status
sudo ufw allow 8080

# Test with different host
grpcurl -plaintext 0.0.0.0:8080 list
```

### DNS Resolution Issues

#### Problem
```bash
# Cannot resolve hostname
```

#### Solution
```bash
# Use IP address instead
grpcurl -plaintext 127.0.0.1:8080 list

# Check DNS
nslookup localhost
dig localhost

# Check /etc/hosts
cat /etc/hosts
```

### TLS/SSL Issues

#### Problem
```bash
# Certificate errors
```

#### Solution
```bash
# Use plaintext for development
grpcurl -plaintext localhost:8080 list

# For production, ensure proper certificates
# Configure TLS in server if needed
```

## Development Issues

### Hot Reload Not Working

#### Problem
```bash
# Changes don't trigger restart
npm run dev
```

#### Solution
```bash
# Check if ts-node-dev is installed
npm list ts-node-dev

# Install if missing
npm install --save-dev ts-node-dev

# Or use nodemon
npm install --save-dev nodemon
nodemon --exec ts-node src/performer.ts
```

### IDE Type Errors

#### Problem
```typescript
// TypeScript errors in IDE but builds fine
```

#### Solution
```bash
# Restart TypeScript server in IDE
# VS Code: Ctrl+Shift+P -> "TypeScript: Restart TS Server"

# Check tsconfig.json
{
  "compilerOptions": {
    "strict": true,
    "skipLibCheck": true
  }
}

# Update IDE TypeScript version
npm install -g typescript
```

### Test Failures

#### Problem
```bash
npm test
# Tests failing unexpectedly
```

#### Solution
```bash
# Run tests with debug output
npm test -- --verbose

# Run specific test
npm test -- --testNamePattern="MyWorker"

# Check test environment
NODE_ENV=test npm test
```

## Debugging Tips

### Enable Debug Logging

```bash
# Environment variables
DEBUG=true
LOG_LEVEL=debug
NODE_ENV=development

# Run with debug
DEBUG=true npm start
```

### Use Built-in Diagnostics

```typescript
// Get diagnostic report
const report = server.getDiagnosticReport();
console.log(report);

// Check health summary
const health = server.getHealthSummary();
console.log(health);

// Monitor metrics
const metrics = server.getMetricsSummary();
console.log(metrics);
```

### Trace Task Execution

```typescript
class DebugWorker extends BaseWorker {
  async handleSimpleTask(input: any) {
    console.log('Input received:', input);
    
    try {
      const result = await this.processInput(input);
      console.log('Result:', result);
      return result;
    } catch (error) {
      console.error('Error:', error);
      throw error;
    }
  }
}
```

### Use Node.js Inspector

```bash
# Start with inspector
node --inspect dist/performer.js

# Connect with Chrome DevTools
# Navigate to chrome://inspect
```

### Monitor System Resources

```bash
# Check system resources
top -p $(pgrep node)
htop

# Monitor logs
tail -f logs/performer.log

# Check disk space
df -h
```

## Getting Help

### Before Asking for Help

1. **Check the logs** - Enable debug logging and check for error messages
2. **Read the error message** - Most errors contain helpful information
3. **Check the documentation** - Review the API documentation and examples
4. **Search existing issues** - Check GitHub issues for similar problems
5. **Create a minimal reproduction** - Isolate the problem in a simple example

### Information to Include

When reporting issues, please include:

1. **Version information**:
   ```bash
   node --version
   npm --version
   npm list @hourglass/performer
   ```

2. **System information**:
   ```bash
   uname -a
   docker --version
   ```

3. **Error messages** (full stack traces)
4. **Configuration files** (tsconfig.json, package.json, etc.)
5. **Steps to reproduce**
6. **Expected vs actual behavior**

### Community Resources

- **Documentation**: [https://docs.hourglass.io](https://docs.hourglass.io)
- **GitHub Issues**: [https://github.com/Layr-Labs/hourglass-monorepo/issues](https://github.com/Layr-Labs/hourglass-monorepo/issues)
- **Discord**: [https://discord.gg/hourglass](https://discord.gg/hourglass)
- **Examples**: [https://github.com/Layr-Labs/hourglass-examples](https://github.com/Layr-Labs/hourglass-examples)

### Creating Bug Reports

Use this template for bug reports:

```markdown
## Bug Report

### Description
Brief description of the issue

### Steps to Reproduce
1. Step 1
2. Step 2
3. Step 3

### Expected Behavior
What should happen

### Actual Behavior
What actually happens

### Environment
- Node.js version: 
- npm version: 
- Package version: 
- Operating System: 

### Additional Context
Any other relevant information
```

### Feature Requests

Use this template for feature requests:

```markdown
## Feature Request

### Problem
What problem does this solve?

### Solution
Proposed solution

### Alternatives
Alternative solutions considered

### Additional Context
Any other relevant information
```

---

Still having issues? Don't hesitate to reach out to the community for help!
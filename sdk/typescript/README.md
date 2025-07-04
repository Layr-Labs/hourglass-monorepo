# Hourglass TypeScript Performer SDK

A TypeScript SDK for building Hourglass AVS performers with minimal setup.

## Quick Start

```bash
npm install @hourglass/performer
```

```typescript
import { PerformerServer } from '@hourglass/performer';

class MyPerformer extends PerformerServer {
  async validateTask(task: TaskRequest): Promise<void> {
    // Validation logic
  }
  
  async handleTask(task: TaskRequest): Promise<TaskResponse> {
    const input = this.parsePayload(task.payload);
    const result = input * input; // Square the number
    return this.createResponse(task.taskId, result);
  }
}

const server = new MyPerformer({ port: 8080 });
server.start();
```

## Development

```bash
# Install dependencies
npm install

# Build the project
npm run build

# Run in development mode
npm run dev

# Run tests
npm test

# Lint code
npm run lint
```

## Project Structure

- `src/server/` - Core server implementation
- `src/worker/` - Worker interface and utilities
- `src/types/` - TypeScript type definitions
- `src/utils/` - Utility functions
- `src/examples/` - Example implementations

## License

MIT
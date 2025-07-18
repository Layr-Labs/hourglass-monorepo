# Hourglass Performer SDK Package Usage

This document explains how to use the Hourglass Performer SDK when installed as an npm package.

## Installation

```bash
npm install @layr-labs/hourglass-performer
```

## Basic Usage

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

## Protobuf File Resolution

The SDK automatically resolves protobuf files from multiple possible locations:

1. **Development environment**: `proto/performer.proto` relative to source
2. **Package installation**: `node_modules/@layr-labs/hourglass-performer/proto/performer.proto`
3. **Alternative paths**: Various fallback locations

### Troubleshooting Protobuf Issues

If you encounter protobuf resolution errors:

1. **Verify proto files are included**:
   ```bash
   ls node_modules/@layr-labs/hourglass-performer/proto/
   ```

2. **Manual proto path resolution**:
   ```typescript
   import { resolveProtoPath } from '@layr-labs/hourglass-performer';
   
   try {
     const protoPath = resolveProtoPath('performer.proto');
     console.log('Proto file found at:', protoPath);
   } catch (error) {
     console.error('Proto file not found:', error.message);
   }
   ```

3. **Custom proto path** (if needed):
   ```typescript
   // Advanced usage - you can extend the PerformerServer if needed
   // and override the proto loading logic
   ```

## Package Contents

The published package includes:

- `dist/` - Compiled JavaScript and TypeScript declarations
- `proto/` - Protocol buffer definitions
- `generated/` - Generated protobuf code
- Documentation files

## Common Issues and Solutions

### Issue: "Could not find performer.proto file"

**Solution**: Ensure the package was installed correctly and includes proto files:

```bash
# Reinstall the package
npm uninstall @layr-labs/hourglass-performer
npm install @layr-labs/hourglass-performer

# Verify proto files exist
ls node_modules/@layr-labs/hourglass-performer/proto/
```

### Issue: gRPC Service Definition Not Found

**Solution**: This typically happens when protobuf files can't be loaded. Use the troubleshooting steps above.

### Issue: Module Resolution in Monorepo

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

## Examples

See the `examples/` directory (if included) or the main repository for complete examples.

## Support

For issues and questions:
- GitHub Issues: https://github.com/Layr-Labs/hourglass-monorepo/issues
- Documentation: [Main README](README.md)
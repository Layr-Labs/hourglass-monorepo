# Hourglass Performer Templates

This directory contains example templates for building Hourglass performers.

## Templates

### `basic-performer.ts`
Simple performer template for basic task processing without contract integration.

**Usage:**
```typescript
class MyBasicPerformer extends BaseWorker {
  async handleSimpleTask(input: any) {
    // Your AVS logic here
    return input * input;
  }
}

new MyBasicPerformer().start();
```

### `solidity-performer.ts`
Performer template with TypeChain integration for Solidity contract interaction.

**Usage:**
```typescript
class MyAVSPerformer extends SolidityWorker<MyContract, 'processTask'> {
  async handleSolidityTask(params: ProcessTaskParams) {
    // Your AVS logic with typed contract parameters
    return { result: params.amount * 2n };
  }
}

new MyAVSPerformer().start();
```

### `advanced-performer.ts`
Advanced template with full monitoring, health checks, and TypeChain integration.

**Features:**
- Comprehensive health monitoring
- Metrics collection
- Structured logging
- TypeChain integration
- Error handling

## Developer Experience

All templates follow the same simple pattern:
1. Extend the appropriate base class
2. Implement the handler method
3. Call `.start()` to run the server

This provides a minimal, zero-configuration developer experience while maintaining full flexibility for advanced use cases.
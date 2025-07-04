# TypeScript Performer Scaffold Implementation Plan

## Overview
Create a minimal TypeScript server scaffold that allows users to write Hourglass AVS performers in TypeScript with minimal setup, following the same architecture as the existing Go performer implementation.

## Architecture Analysis
Based on the Go performer in `ponos/pkg/performer/`, the TypeScript scaffold needs to:

1. **Implement PerformerService gRPC interface** with:
   - `ExecuteTask`: Main task processing endpoint
   - `HealthCheck`: Returns performer status
   - `StartSync`: Initialization endpoint

2. **Provide Worker Interface** similar to Go's `IWorker`:
   - `validateTask(task)`: Validate incoming task
   - `handleTask(task)`: Process task and return result

3. **Handle gRPC Communication**:
   - Receive `TaskRequest` with task_id and payload (bytes)
   - Return `TaskResponse` with task_id and result (bytes)

## Implementation Milestones

### Milestone 1: Project Setup & Dependencies ✅ COMPLETED
- [x] Create TypeScript project structure in `sdk/typescript/`
- [x] Set up package.json with dependencies:
  - [x] @grpc/grpc-js for gRPC server
  - [x] @grpc/proto-loader for protobuf loading
  - [x] TypeScript and build tools
  - [x] Winston for logging
- [x] Configure TypeScript compiler options
- [x] Set up build scripts and development workflow

**Completed Tasks:**
- Created project structure with `src/{types,server,worker,utils,examples}` directories
- Set up package.json with all required dependencies and dev tools
- Configured TypeScript with strict settings and ES2020 target
- Added ESLint, Jest, and ts-node-dev for development workflow
- Created .gitignore, README.md, and initial index.ts entry point

### Milestone 2: Protobuf Integration
- [ ] Identify protobuf definitions source (protocol-apis repository)
- [ ] Generate TypeScript types from protobuf definitions
- [ ] Create type definitions for:
  - [ ] PerformerService interface
  - [ ] TaskRequest and TaskResponse messages
  - [ ] HealthCheckRequest/Response
  - [ ] StartSyncRequest/Response
- [ ] Set up protobuf generation in build pipeline

### Milestone 3: Core Server Implementation
- [ ] Create base PerformerServer class
- [ ] Implement gRPC server setup and configuration
- [ ] Create abstract Worker interface for user implementation
- [ ] Implement server lifecycle management:
  - [ ] Start/stop functionality
  - [ ] Graceful shutdown handling
  - [ ] Error handling and logging
- [ ] Add configuration management (port, timeouts, etc.)

### Milestone 4: Task Processing Framework
- [ ] Implement ExecuteTask handler:
  - [ ] Task validation pipeline
  - [ ] Task routing to user worker
  - [ ] Response formatting
  - [ ] Error handling and status codes
- [ ] Create task context and metadata handling
- [ ] Add timeout and cancellation support
- [ ] Implement result serialization/deserialization

### Milestone 5: Health & Monitoring
- [ ] Implement HealthCheck endpoint
- [ ] Add performer status management
- [ ] Create monitoring and metrics collection
- [ ] Implement structured logging
- [ ] Add debugging and troubleshooting tools

### Milestone 6: Developer Experience
- [ ] Create simple Worker base class/interface
- [ ] Provide payload parsing utilities
- [ ] Add development server with hot reload
- [ ] Create CLI tool for scaffolding new performers
- [ ] Add TypeScript decorators for common patterns

### Milestone 7: Example Implementation
- [ ] Create example performer (similar to Go demo)
- [ ] Implement number squaring logic
- [ ] Add comprehensive documentation
- [ ] Create Docker container template
- [ ] Add environment variable configuration

### Milestone 8: Integration & Testing
- [ ] Create integration tests with existing Ponos system
- [ ] Test with executor/aggregator communication
- [ ] Validate gRPC compatibility
- [ ] Performance benchmarking
- [ ] Docker deployment testing

### Milestone 9: Documentation & Templates
- [ ] Create README with quick start guide
- [ ] Write API documentation
- [ ] Create project templates
- [ ] Add troubleshooting guide
- [ ] Create migration guide from Go performers

### Milestone 10: Packaging & Distribution
- [ ] Prepare npm package
- [ ] Create Docker base image
- [ ] Set up CI/CD pipeline
- [ ] Version management and release process
- [ ] Integration with existing Hourglass tooling

## Success Criteria

### Minimal Viable Product (MVP)
- [ ] User can create new TypeScript performer with single command
- [ ] Basic task execution with validation
- [ ] gRPC server automatically configured
- [ ] Docker container builds successfully
- [ ] Works with existing Ponos executor

### Developer Experience Goals
- [ ] Less than 10 lines of code for basic performer
- [ ] Hot reload during development
- [ ] Comprehensive error messages
- [ ] Type safety throughout
- [ ] Zero configuration for common use cases

### Performance Targets
- [ ] Task processing latency < 100ms overhead
- [ ] Memory usage < 50MB base
- [ ] Container startup time < 5 seconds
- [ ] Supports concurrent task processing

## Example Usage Target

```typescript
// performer.ts
import { PerformerServer, TaskRequest, TaskResponse } from '@hourglass/performer';

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

// Start server
const server = new MyPerformer({ port: 8080 });
server.start();
```

## Technical Requirements

### Core Dependencies
- Node.js >= 18
- TypeScript >= 5.0
- @grpc/grpc-js for gRPC communication
- Protocol buffer definitions from protocol-apis

### Container Requirements
- Alpine-based Docker image
- gRPC server on port 8080
- Health check endpoint
- Graceful shutdown support

### Integration Points
- Compatible with existing Ponos executor
- Follows same gRPC protocol as Go performer
- Supports same task payload format
- Maintains same error handling patterns
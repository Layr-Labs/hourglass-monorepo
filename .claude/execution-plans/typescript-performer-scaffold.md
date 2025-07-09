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

### Milestone 2: Protobuf Integration ✅ COMPLETED
- [x] Identify protobuf definitions source (protocol-apis repository)
- [x] Generate TypeScript types from protobuf definitions
- [x] Create type definitions for:
  - [x] PerformerService interface
  - [x] TaskRequest and TaskResponse messages
  - [x] HealthCheckRequest/Response
  - [x] StartSyncRequest/Response
- [x] Set up protobuf generation in build pipeline

**Completed Tasks:**
- Identified external protobuf source at github.com/Layr-Labs/protocol-apis v1.14.0
- Created TypeScript type definitions in `src/types/performer.ts`
- Created protobuf schema in `proto/performer.proto`
- Added protobuf generation to build pipeline with grpc-tools
- Created utility functions for protobuf data conversion in `src/types/protobuf.ts`
- Updated package.json with protobuf generation scripts and dependencies

### Milestone 3: Core Server Implementation ✅ COMPLETED
- [x] Create base PerformerServer class
- [x] Implement gRPC server setup and configuration
- [x] Create abstract Worker interface for user implementation
- [x] Implement server lifecycle management:
  - [x] Start/stop functionality
  - [x] Graceful shutdown handling
  - [x] Error handling and logging
- [x] Add configuration management (port, timeouts, etc.)

**Completed Tasks:**
- Created comprehensive PerformerServer class in `src/server/PerformerServer.ts`
- Implemented gRPC server setup with protobuf loading and service registration
- Created IWorker interface and BaseWorker abstract class in `src/worker/IWorker.ts`
- Added EchoWorker as a simple example implementation
- Implemented full server lifecycle with start/stop and graceful shutdown
- Added comprehensive error handling with gRPC status codes
- Created logger utility with winston in `src/utils/logger.ts`
- Added configuration management with sensible defaults
- Created demo examples in `src/examples/` showing various usage patterns

### Milestone 4: Task Processing Framework ✅ COMPLETED
- [x] Implement ExecuteTask handler:
  - [x] Task validation pipeline
  - [x] Task routing to user worker
  - [x] Response formatting
  - [x] Error handling and status codes
- [x] Create task context and metadata handling
- [x] Add timeout and cancellation support
- [x] Implement result serialization/deserialization

**Completed Tasks:**
- Created comprehensive TaskContext system with execution metadata and cancellation support
- Built TaskProcessor with pipeline architecture for validation, routing, and response handling
- Implemented multiple serialization strategies (JSON, raw bytes, numbers)
- Added TaskPipeline with customizable validation stages
- Enhanced PerformerServer to use TaskProcessor for robust task execution
- Created task execution metrics collection and monitoring
- Added timeout and cancellation support with AbortSignal
- Built comprehensive error handling with proper gRPC status codes
- Created enhanced demo example showing advanced features

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
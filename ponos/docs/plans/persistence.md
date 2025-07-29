# Ponos Persistence Layer Implementation Plan

## Overview

This document outlines the implementation plan for adding data persistence to the Ponos aggregator and executor services to enable crash recovery and high availability.

## Goals

1. **Crash Recovery**: Both aggregator and executor can resume from where they left off after a crash
2. **No Lost Tasks**: Tasks are persisted before processing begins
3. **Efficient State Management**: Replace in-memory sync.Maps with persistent storage
4. **Clean Architecture**: Separate storage abstractions for aggregator and executor
5. **Testability**: In-memory implementation first, then BadgerDB

## Architecture Decisions

- **Separate Storage Interfaces**: Aggregator and Executor will have their own storage interfaces to avoid cross-process dependencies
- **BadgerDB**: Selected for its pure Go implementation, embedded nature, and production readiness
- **Phased Approach**: Implement in-memory stores first, refactor code, ensure tests pass, then implement BadgerDB

## Milestones

### Milestone 1: Storage Interface Design (1.1 - 1.4)

#### 1.1 Create Aggregator Storage Interface
- [x] Create `pkg/aggregator/storage/` directory structure
- [x] Define `AggregatorStore` interface in `pkg/aggregator/storage/storage.go`
- [x] Define data types:
  - [x] `TaskStatus` enum (pending, processing, completed, failed)
  - [x] `OperatorSetTaskConfig` type (moved from avsExecutionManager)
  - [x] `AvsConfig` type (moved from avsExecutionManager)
  - [x] `ChainBlockHeight` type for chain polling state
- [x] Add methods for:
  - [x] Chain polling state (GetLastProcessedBlock, SetLastProcessedBlock)
  - [x] Task management (SaveTask, GetTask, ListPendingTasks, UpdateTaskStatus, DeleteTask)
  - [x] Config caching (SaveOperatorSetConfig, GetOperatorSetConfig, SaveAVSConfig, GetAVSConfig)
  - [x] Lifecycle management (Close)

#### 1.2 Create Executor Storage Interface
- [x] Create `pkg/executor/storage/` directory structure
- [x] Define `ExecutorStore` interface in `pkg/executor/storage/storage.go`
- [x] Define data types:
  - [x] `PerformerState` struct (id, avs address, container id, status, artifact info, timestamps)
  - [x] `TaskInfo` struct for inflight tasks
  - [x] `DeploymentInfo` struct for deployment tracking
  - [x] `DeploymentStatus` enum (pending, deploying, running, failed)
- [x] Add methods for:
  - [x] Performer state (SavePerformerState, GetPerformerState, ListPerformerStates, DeletePerformerState)
  - [x] Task tracking (SaveInflightTask, GetInflightTask, ListInflightTasks, DeleteInflightTask)
  - [x] Deployment tracking (SaveDeployment, GetDeployment, UpdateDeploymentStatus)
  - [x] Lifecycle management (Close)

#### 1.3 Create Storage Error Types
- [x] Define common error types in `pkg/aggregator/storage/errors.go`:
  - [x] `ErrNotFound` for missing keys
  - [x] `ErrAlreadyExists` for duplicate keys
  - [x] `ErrStoreClosed` for operations on closed store
- [x] Define executor errors in `pkg/executor/storage/errors.go`

#### 1.4 Create Storage Tests Interfaces
- [x] Create `pkg/aggregator/storage/storage_test.go` with interface compliance tests
- [x] Create `pkg/executor/storage/storage_test.go` with interface compliance tests

### Milestone 2: In-Memory Implementation (2.1 - 2.4)

#### 2.1 Implement In-Memory Aggregator Store
- [x] Create `pkg/aggregator/storage/memory/memory.go`
- [x] Implement `InMemoryAggregatorStore` struct with:
  - [x] RWMutex for thread safety
  - [x] Maps for each data type (blocks, tasks, configs)
- [x] Implement all `AggregatorStore` interface methods:
  - [x] GetLastProcessedBlock with proper error handling
  - [x] SetLastProcessedBlock with atomic updates
  - [x] SaveTask with duplicate checking
  - [x] GetTask with not found errors
  - [x] ListPendingTasks with status filtering
  - [x] UpdateTaskStatus with validation
  - [x] DeleteTask with existence checking
  - [x] Config methods with proper key generation
  - [x] Close method to clear maps

#### 2.2 Implement In-Memory Executor Store
- [x] Create `pkg/executor/storage/memory/memory.go`
- [x] Implement `InMemoryExecutorStore` struct with:
  - [x] RWMutex for thread safety
  - [x] Maps for performers, tasks, deployments
- [x] Implement all `ExecutorStore` interface methods:
  - [x] Performer state methods with validation
  - [x] Inflight task tracking methods
  - [x] Deployment tracking methods
  - [x] Close method to clear maps

#### 2.3 Create In-Memory Store Tests
- [x] Create `pkg/aggregator/storage/memory/memory_test.go`:
  - [x] Test concurrent access scenarios
  - [x] Test all CRUD operations
  - [x] Test error conditions
  - [x] Benchmark performance
- [x] Create `pkg/executor/storage/memory/memory_test.go`:
  - [x] Test concurrent access scenarios
  - [x] Test all CRUD operations
  - [x] Test error conditions
  - [x] Benchmark performance

#### 2.4 Create Storage Factory Functions [SKIPPED]
- [x] SKIPPED - Callers will directly instantiate the specific store type they need
- [x] No factory pattern - use `memory.NewInMemoryAggregatorStore()` or `memory.NewInMemoryExecutorStore()` directly

### Milestone 3: Aggregator Refactoring (3.1 - 3.5)

#### 3.4 Update Aggregator Configuration
- [ ] Add storage configuration to `aggregatorConfig`:
  ```go
  type StorageConfig struct {
      Type string `json:"type" yaml:"type"` // "memory" or "badger"
      BadgerConfig *BadgerConfig `json:"badger,omitempty" yaml:"badger,omitempty"`
  }
  ```
- [ ] Update config validation
- [ ] Update example configurations

#### 3.3 Update Aggregator Initialization
- [ ] Modify `aggregator.Initialize()`:
  - [ ] Create storage instance based on config
  - [ ] Pass storage to EVMChainPoller instances
  - [ ] Pass storage to AvsExecutionManager instances
- [ ] Add recovery logic:
  - [ ] Load pending tasks from storage
  - [ ] Re-queue tasks to taskQueue
  - [ ] Log recovery statistics

#### 3.2 Refactor AvsExecutionManager
- [ ] Add `storage.AggregatorStore` field to `AvsExecutionManager` struct
- [ ] Update constructor to accept storage parameter
- [ ] Replace `operatorSetTaskConfigs` sync.Map:
  - [ ] Use storage for get/set operations
  - [ ] Update `getOrSetOperatorSetTaskConfig` method
- [ ] Replace `avsConfig` caching:
  - [ ] Remove mutex and in-memory field
  - [ ] Update `getOrSetAggregatorTaskConfig` method
- [ ] Update task handling:
  - [ ] Save task to storage in `processTask`
  - [ ] Update task status in `handleTask`
  - [ ] Mark task complete after submission

#### 3.1 Refactor EVMChainPoller
- [ ] Add `storage.AggregatorStore` field to `EVMChainPoller` struct
- [ ] Update `NewEVMChainPoller` to accept storage parameter
- [ ] Replace `lastObservedBlock` field usage:
  - [ ] Load last block on initialization
  - [ ] Save block after successful processing
  - [ ] Handle storage errors appropriately
- [ ] Update `Start()` method to recover from last processed block
- [ ] Add logging for storage operations

#### 3.5 Update Aggregator Tests
- [ ] Update all existing aggregator tests to use in-memory storage
- [ ] Add storage-specific test cases:
  - [ ] Test crash recovery scenarios
  - [ ] Test task persistence
  - [ ] Test config caching
- [ ] Ensure all tests pass

### Milestone 4: Executor Refactoring (4.1 - 4.5)

#### 4.1 Refactor Executor struct
- [ ] Add `storage.ExecutorStore` field to `Executor` struct
- [ ] Update `NewExecutor` to accept storage parameter
- [ ] Replace `inflightTasks` sync.Map:
  - [ ] Use storage for tracking inflight tasks
  - [ ] Update task submission handling
  - [ ] Clean up completed tasks

#### 4.2 Update Performer Management
- [ ] Save performer state after deployment:
  - [ ] In `Initialize()` after successful deployment
  - [ ] In `DeployArtifact()` handler
- [ ] Update `ListPerformers()` to include persisted state
- [ ] Update `RemovePerformer()` to delete from storage
- [ ] Add recovery logic in `Initialize()`:
  - [ ] Load performer states
  - [ ] Verify containers/pods still exist
  - [ ] Re-create missing performers

#### 4.3 Update Deployment Tracking
- [ ] Create deployment record in `DeployArtifact()`
- [ ] Update deployment status during deployment
- [ ] Track deployment history for troubleshooting

#### 4.4 Update Executor Configuration
- [ ] Add storage configuration to `executorConfig`
- [ ] Mirror aggregator's storage config structure
- [ ] Update validation and examples

#### 4.5 Update Executor Tests
- [ ] Update all existing executor tests to use in-memory storage
- [ ] Add storage-specific test cases:
  - [ ] Test performer state persistence
  - [ ] Test deployment tracking
  - [ ] Test recovery scenarios
- [ ] Ensure all tests pass

### Milestone 5: Integration Testing (5.1 - 5.3)

#### 5.1 Create Integration Test Suite
- [ ] Create `internal/tests/persistence/` directory
- [ ] Write end-to-end tests:
  - [ ] Aggregator crash and recovery
  - [ ] Executor crash and recovery
  - [ ] Task flow with persistence

#### 5.2 Add Benchmarks
- [ ] Benchmark in-memory store performance
- [ ] Compare with sync.Map baseline
- [ ] Document performance characteristics

#### 5.3 Update Demo
- [ ] Update demo configurations
- [ ] Add persistence directory setup
- [ ] Test full demo with persistence

### Milestone 6: BadgerDB Implementation (6.1 - 6.6)

#### 6.1 Add BadgerDB Dependency
- [ ] Add `github.com/dgraph-io/badger/v3` to go.mod
- [ ] Run `go mod tidy`

#### 6.2 Implement BadgerDB Aggregator Store
- [ ] Create `pkg/aggregator/storage/badger/badger.go`
- [ ] Implement `BadgerAggregatorStore` struct
- [ ] Design key schemas:
  - [ ] Chain blocks: `chain:{chainId}:lastBlock`
  - [ ] Tasks: `task:{taskId}`
  - [ ] Operator configs: `opset:{avsAddress}:{operatorSetId}`
  - [ ] AVS configs: `avs:{avsAddress}`
- [ ] Implement all interface methods with:
  - [ ] Proper serialization (JSON or protobuf)
  - [ ] Transaction support where needed
  - [ ] TTL for temporary data
- [ ] Add periodic garbage collection

#### 6.3 Implement BadgerDB Executor Store
- [ ] Create `pkg/executor/storage/badger/badger.go`
- [ ] Implement `BadgerExecutorStore` struct
- [ ] Design key schemas:
  - [ ] Performers: `performer:{performerId}`
  - [ ] Tasks: `task:{taskId}`
  - [ ] Deployments: `deployment:{deploymentId}`
- [ ] Implement all interface methods
- [ ] Add data migration utilities

#### 6.4 Create BadgerDB Tests
- [ ] Test BadgerDB implementations against interface tests
- [ ] Add BadgerDB-specific tests:
  - [ ] Test persistence across restarts
  - [ ] Test concurrent access
  - [ ] Test large data sets
  - [ ] Test compaction

#### 6.5 Update Factory Functions [SKIPPED]
- [x] SKIPPED - No factory pattern, callers will use `badger.NewBadgerAggregatorStore()` or `badger.NewBadgerExecutorStore()` directly
- [ ] BadgerDB constructors will accept configuration directly

#### 6.6 Performance Optimization
- [ ] Profile BadgerDB performance
- [ ] Tune BadgerDB options:
  - [ ] Value log settings
  - [ ] Compaction settings
  - [ ] Cache sizes
- [ ] Add metrics collection

### Milestone 7: Production Readiness (7.1 - 7.4)

#### 7.1 Add Monitoring
- [ ] Add storage metrics:
  - [ ] Operation latencies
  - [ ] Storage size
  - [ ] Error rates
- [ ] Integrate with existing metrics system

#### 7.2 Add Operational Tools
- [ ] Create storage inspection tool
- [ ] Create data export/import utilities
- [ ] Create storage migration tool

#### 7.3 Documentation
- [ ] Update README with persistence details
- [ ] Create operations guide
- [ ] Document recovery procedures
- [ ] Add troubleshooting guide

#### 7.4 Final Testing
- [ ] Run extended stability tests
- [ ] Test upgrade scenarios
- [ ] Validate production configurations
- [ ] Performance benchmarks

## Success Criteria

1. Both aggregator and executor can recover from crashes without losing state
2. No tasks are lost during normal operations or crashes
3. Performance impact is minimal (< 5% overhead)
4. Storage size grows predictably
5. All existing tests pass
6. New persistence-specific tests pass

## Timeline

- **Week 1**: Milestones 1-2 (Interfaces and In-Memory Implementation)
- **Week 2**: Milestone 3 (Aggregator Refactoring)
- **Week 3**: Milestone 4 (Executor Refactoring)
- **Week 4**: Milestones 5-6 (Integration Testing and BadgerDB)
- **Week 5**: Milestone 7 (Production Readiness)

## Risks and Mitigations

1. **Risk**: Performance regression
   - **Mitigation**: Benchmark throughout, start with in-memory implementation

2. **Risk**: Data corruption
   - **Mitigation**: Use transactions, add data validation, implement backups

3. **Risk**: Storage growth
   - **Mitigation**: Implement TTLs, periodic cleanup, monitoring

4. **Risk**: Complex migration
   - **Mitigation**: Phased approach, feature flags, rollback plan
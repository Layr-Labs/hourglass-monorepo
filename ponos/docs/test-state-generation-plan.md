# Test State Generation Implementation Plan

## Overview
This document tracks the implementation of a test state generation system for the Ponos integration tests. The goal is to pre-generate expensive blockchain setup operations, reducing test execution time by approximately 85% (from 30+ seconds to ~5 seconds).

## Problem Statement
Current integration tests perform repetitive, time-consuming setup operations including:
- Stake table transport (~6 seconds)
- Operator registration and peering (~3-4 seconds each)
- Stake delegation (~3-4 seconds)
- Generation reservations (~3-4 seconds)

## Solution Architecture

### Core Components

1. **Test State Generator Framework** (`/internal/harness/testDataGenerator.go`)
   - Central orchestrator for state generation
   - Registry pattern for test-specific generators
   - Manages anvil lifecycle and state dumps

2. **State Generator Interface**
   ```go
   type StateGenerator interface {
       GenerateState(ctx context.Context, l1Caller, l2Caller contractCaller.IContractCaller) error
       GetStateID() string
       GetDescription() string
   }
   ```

3. **File Organization**
   ```
   /internal/testData/
   ├── states/
   │   ├── aggregator/
   │   │   ├── bn254_ecdsa/
   │   │   │   ├── l1-state.json
   │   │   │   ├── l2-state.json
   │   │   │   └── metadata.json
   │   │   └── ecdsa_bn254/
   │   │       ├── l1-state.json
   │   │       ├── l2-state.json
   │   │       └── metadata.json
   │   └── [other-tests]/
   ```

## Implementation Phases

### Phase 1: Core Infrastructure ⏳
**Status:** Not Started

- [ ] Create `/internal/harness/testDataGenerator.go`
  - [ ] Define `StateGenerator` interface
  - [ ] Implement `TestDataGenerator` struct
  - [ ] Add registry for state generators
  - [ ] Implement anvil lifecycle management

- [ ] Create `/internal/harness/types.go`
  - [ ] Define `StateDefinition` struct
  - [ ] Create `GeneratedState` struct
  - [ ] Add configuration types

- [ ] Create `/internal/harness/anvil.go`
  - [ ] Anvil state dump wrapper functions
  - [ ] State loading utilities
  - [ ] Cleanup and validation functions

### Phase 2: Aggregator Test Migration ⏳
**Status:** Not Started

- [ ] Create `/pkg/aggregator/testStateGenerator.go`
  - [ ] Implement `AggregatorStateGenerator`
  - [ ] Extract setup logic from `aggregator_test.go`:
    - [ ] Operator registration (lines 400-500)
    - [ ] Stake table transport (line 573)
    - [ ] AVS configuration (lines 380-410)
    - [ ] Key registration (lines 420-480)
  - [ ] Support multiple signature configurations:
    - [ ] BN254_Aggregator_ECDSA_Executor
    - [ ] ECDSA_Aggregator_BN254_Executor

- [ ] Create `/cmd/generateTestState/main.go`
  - [ ] CLI interface for state generation
  - [ ] Progress reporting
  - [ ] Error handling and recovery

- [ ] Update Makefile
  - [ ] Add `generate-test-state` target
  - [ ] Include dependencies
  - [ ] Environment setup

### Phase 3: Test Integration ⏳
**Status:** Not Started

- [ ] Modify `/pkg/aggregator/aggregator_test.go`
  - [ ] Add state loading logic
  - [ ] Replace inline setup with pre-generated states
  - [ ] Add fallback for missing states
  - [ ] Maintain backward compatibility

- [ ] Create state generation documentation
  - [ ] Usage instructions
  - [ ] Regeneration guidelines
  - [ ] CI integration notes

### Phase 4: Extended Implementation ⏳
**Status:** Not Started

- [ ] Extend to other test files
- [ ] Add CI workflow for state regeneration
- [ ] Implement state versioning
- [ ] Add state validation tools

## Technical Details

### State Generation Flow
1. Start anvil instances with base state
2. Create contract callers for L1 and L2
3. Invoke test-specific `GenerateState` method
4. Perform all expensive setup operations
5. Dump anvil state to JSON files
6. Save metadata with configuration details
7. Clean up anvil instances

### Test Execution Flow (with pre-generated states)
1. Check for existing state files
2. Start anvil with `--load-state` flag
3. Skip expensive setup operations
4. Run test logic directly
5. Clean up anvil instances

### Key Benefits
- **85% reduction in test time** (30s → 5s)
- **Consistent test environments** across runs
- **Easier debugging** with known states
- **Extensible architecture** for other tests
- **CI-friendly** with regeneration capabilities

## Progress Tracking

### Current Status
- ✅ Analysis and planning complete
- ✅ Core infrastructure implemented
- ✅ Aggregator state generator created
- ✅ CLI tool and Makefile targets added
- 🔧 Fixing compilation issues
- ⏳ Testing not started
- 📝 Documentation in progress

### Completed Tasks
1. ✅ Created `/internal/harness/` infrastructure
2. ✅ Implemented `StateGenerator` interface and types
3. ✅ Built anvil state management utilities
4. ✅ Created aggregator state generator
5. ✅ Added `generate-test-state` make target
6. ✅ Implemented CLI tool at `/cmd/generateTestState/`

### Next Steps
1. Fix compilation issues in the implementation
2. Test state generation with actual anvil instances
3. Modify aggregator tests to use pre-generated states
4. Document usage and best practices

## Usage Instructions

### Generating Test States

The test state generation system provides several make targets for generating pre-computed blockchain states:

#### Generate All States
```bash
# Generate all test states (overwrites existing)
make generate-test-state

# Generate with verbose output for debugging
make generate-test-state-verbose
```

#### Generate Specific State
```bash
# Generate state for specific test configuration
make generate-test-state-specific SUITE=aggregator/bn254_ecdsa
make generate-test-state-specific SUITE=aggregator/ecdsa_bn254
```

#### Manual CLI Usage
```bash
# Set required environment variable
export HOURGLASS_TRANSPORT_BLS_KEY="0x89d7404a597f6210a673cd0d07ae3c43df0344d233a0957f7682afdd2922e3e0"

# Run the CLI tool directly
go run ./cmd/generateTestState/main.go -help
go run ./cmd/generateTestState/main.go -output ./internal/testData -overwrite
go run ./cmd/generateTestState/main.go -suite aggregator/bn254_ecdsa -verbose
```

### Prerequisites

Before generating states, ensure:
1. Anvil is installed and available in PATH
2. Base chain states exist in `/internal/testData/`:
   - `anvil-l1-state.json`
   - `anvil-l2-state.json`
   - `chain-config.json`
3. HOURGLASS_TRANSPORT_BLS_KEY environment variable is set

### Generated Files

States are organized in the following structure:
```
/internal/testData/states/
├── aggregator/
│   ├── bn254_agg_ecdsa_exec/
│   │   ├── l1-state.json      # L1 anvil state
│   │   ├── l2-state.json      # L2 anvil state
│   │   └── metadata.json      # State metadata
│   └── ecdsa_agg_bn254_exec/
│       ├── l1-state.json
│       ├── l2-state.json
│       └── metadata.json
```

### Using Pre-Generated States in Tests

Tests can load pre-generated states instead of performing setup:

```go
// Check if pre-generated state exists
statePath := filepath.Join(root, "internal/testData/states/aggregator", stateID)
if stateExists(statePath) {
    // Load state instead of setup
    l1Anvil := testUtils.StartL1AnvilWithState(statePath + "/l1-state.json")
    l2Anvil := testUtils.StartL2AnvilWithState(statePath + "/l2-state.json")
    // Skip expensive setup operations
} else {
    // Fall back to regular setup
    performExpensiveSetup()
}
```

### Regenerating States

States should be regenerated when:
- Smart contract code changes
- Test setup logic changes
- Dependencies are updated
- New test configurations are added

To regenerate:
```bash
# Force regeneration of all states
make generate-test-state

# Regenerate specific configuration
make generate-test-state-specific SUITE=aggregator/bn254_ecdsa
```

### Troubleshooting

Common issues and solutions:

1. **Anvil not found**: Ensure anvil is installed via Foundry
2. **State generation fails**: Check anvil ports (8545, 9545) are not in use
3. **Transport fails**: Verify HOURGLASS_TRANSPORT_BLS_KEY is set correctly
4. **Out of memory**: State generation requires ~2GB RAM per configuration

## Notes and Considerations

### Performance Metrics
- Current test execution: ~30 seconds
- Expected with pre-generated states: ~5 seconds
- State generation time: ~30 seconds (one-time cost)

### Maintenance
- States should be regenerated when:
  - Contract code changes
  - Test setup logic changes
  - Dependencies update
- Consider adding version tracking to states

### Future Enhancements
- Parallel state generation for multiple configurations
- State diffing tools for debugging
- Automatic state regeneration in CI
- State compression for repository size management

## References
- Aggregator test: `/pkg/aggregator/aggregator_test.go`
- Existing state generation: `/scripts/generateTestChainState.sh`
- Test utilities: `/internal/testUtils/testUtils.go`
- Contract caller interface: `/pkg/contractCaller/contractCaller.go`

---
*Last Updated: 2025-09-08*
*Author: System Implementation Plan*
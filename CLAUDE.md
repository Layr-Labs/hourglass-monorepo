# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Monorepo Structure

Hourglass is a framework for building EigenLayer AVSs (Actively Validated Services) organized as a monorepo with three main components:

### Core Components

**contracts/**: Solidity smart contracts using Foundry
- Core: TaskMailbox contract for cross-chain task coordination
- AVS: TaskAVSRegistrarBase for AVS registration and management
- Built on EigenLayer middleware for operator registration and slashing

**ponos/**: Go-based aggregation and execution layer
- Aggregator: Coordinates tasks across operators, aggregates responses
- Executor: Runs AVS workloads in Docker containers
- Multi-chain support for L1/L2 coordination

**demo/**: Working example with Docker Compose orchestration
- Shows complete Ponos stack with AVS performer
- Uses Docker-in-Docker for isolated AVS execution

## Development Commands

### Contracts (Foundry-based)
```bash
# From contracts/ directory
make build                    # Clean, build, generate Go bindings
make test                     # Run Foundry tests
forge fmt                     # Format Solidity code
forge snapshot                # Gas usage snapshots

# Deployment workflow (requires env vars)
make deploy-task-mailbox
make deploy-avs-l1-contracts AVS_ADDRESS=0x...
make setup-avs-l1 TASK_AVS_REGISTRAR_ADDRESS=0x...
make deploy-avs-l2-contracts
make register-operator        # Complex command with many parameters
```

### Ponos (Go-based)
```bash
# From ponos/ directory
make deps                     # Install all dependencies + tools
make all                      # Build executor, aggregator, bls-helper
make proto                    # Generate protobuf Go code
make test                     # Run integration tests
make lint                     # Run golangci-lint

# Cross-platform builds
make release                  # Build for multiple OS/arch combinations

# Development
make run/executor
make run/aggregator ARGS="--config config.yaml"

# Test environments
make anvil/start/l1          # Fork mainnet at specific block
make anvil/start/l2          # Fork Base at specific block
```

### Demo
```bash
# From demo/ directory
make build-container         # Build AVS performer container
docker compose up           # Start full stack

# Test task submission
grpcurl -plaintext -d '{"avsAddress": "0xavs1...", "taskId": "0xtask1...", "payload": "..."}' localhost:9090 eigenlayer.hourglass.v1.ExecutorService/SubmitTask
```

### Root Level
```bash
# Commit message enforcement
npm install                  # Install commitlint
# Enforces conventional commits: feat, fix, docs, style, refactor, test, chore, revert, perf
```

## Architecture Overview

### Cross-Chain Task Coordination
The system enables AVSs to operate across L1/L2 with coordinated execution:

1. **Task Creation**: AVS creates tasks via TaskMailbox on either L1 or L2
2. **Aggregator Monitoring**: Ponos aggregator watches both chains for TaskCreated events
3. **Task Distribution**: Aggregator distributes tasks to registered executors
4. **Execution**: Executors run AVS-specific Docker containers with task payload
5. **Response Aggregation**: Aggregator collects and verifies responses using BLS signatures
6. **Result Submission**: Final aggregated result submitted back to originating chain

### Key Architectural Patterns

**Docker-Based Isolation**: AVS workloads run in isolated containers with standardized interfaces, enabling arbitrary computation while maintaining security boundaries.

**Multi-Signature Verification**: Support for ECDSA, BLS BN254, and BLS381 schemes with pluggable verification components (BN254CertificateVerifier, etc).

**Cross-Chain State Synchronization**: TaskMailbox contracts on both L1/L2 maintain synchronized task state with configurable routing and verification rules.

**gRPC Communication**: All inter-service communication uses Protocol Buffers with generated type-safe interfaces.

**EigenLayer Integration**: Built on EigenLayer middleware for operator registration, slashing, and restaking economics.

### Configuration Architecture

**Aggregator Config**: Defines supported chains, AVS addresses, operator keys, response timeouts
**Executor Config**: Defines operator identity, Docker image manifests, AVS-specific settings  
**Contract Deployment**: Scripted deployment workflow with environment-specific parameters

### Testing Strategy

Integration tests require Docker containers and live blockchain forks (Anvil). The demo/ directory provides a complete working example with realistic task flows.

Contracts use standard Foundry testing with gas optimization reports.

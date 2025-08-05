# Ponos Overview

Ponos is a comprehensive framework for building and operating EigenLayer AVS (Actively Validated Services) workloads. It provides a robust infrastructure for coordinating tasks across multiple blockchain networks, executing containerized workloads, and aggregating cryptographically signed results.

## Core Components

Ponos consists of two primary components that work together to enable decentralized task execution:

### 1. Aggregator
The Aggregator acts as the coordinator and orchestrator of the AVS system. It monitors multiple blockchain networks for tasks, distributes them to operators, collects and verifies responses, and submits aggregated results back to the blockchain.

**Key Responsibilities:**
- Multi-chain task monitoring and discovery
- Operator set management and coordination
- Response collection and verification
- Signature aggregation (BLS, ECDSA)
- Result submission to blockchain
- State persistence and recovery

### 2. Executor
The Executor runs on operator nodes and is responsible for executing AVS-specific workloads in isolated Docker containers. It receives tasks from the Aggregator, manages container lifecycle, and returns signed responses.

**Key Responsibilities:**
- Task execution in Docker containers
- Container lifecycle management
- Resource isolation and limits
- Response signing with operator keys
- State management for running tasks
- Health monitoring and reporting

## High-Level Architecture

```mermaid
graph TB
    subgraph "Blockchain Networks"
        BC1[Ethereum L1]
        BC2[Base L2]
        BC3[Arbitrum L2]
    end

    subgraph "Ponos Aggregator"
        CP[Chain Poller]
        TM[Task Manager]
        RM[Response Manager]
        SA[Signature Aggregator]
        AS[Aggregator Storage]
    end

    subgraph "Ponos Executors"
        E1[Executor 1]
        E2[Executor 2]
        E3[Executor N]
        
        subgraph "Executor Components"
            PM[Performer Manager]
            CM[Container Manager]
            ES[Executor Storage]
        end
    end

    subgraph "AVS Containers"
        AC1[AVS Container 1]
        AC2[AVS Container 2]
        AC3[AVS Container N]
    end

    BC1 --> CP
    BC2 --> CP
    BC3 --> CP
    
    CP --> TM
    TM --> AS
    TM --> E1
    TM --> E2
    TM --> E3
    
    E1 --> PM
    PM --> CM
    CM --> AC1
    CM --> ES
    
    AC1 --> PM
    PM --> RM
    E2 --> RM
    E3 --> RM
    
    RM --> SA
    SA --> BC1
    SA --> BC2
    SA --> BC3

    style BC1 fill:#e1f5fe
    style BC2 fill:#e1f5fe
    style BC3 fill:#e1f5fe
    style CP fill:#fff3e0
    style TM fill:#fff3e0
    style RM fill:#fff3e0
    style SA fill:#fff3e0
    style AS fill:#fff3e0
    style E1 fill:#f3e5f5
    style E2 fill:#f3e5f5
    style E3 fill:#f3e5f5
    style PM fill:#f3e5f5
    style CM fill:#f3e5f5
    style ES fill:#f3e5f5
    style AC1 fill:#e8f5e9
    style AC2 fill:#e8f5e9
    style AC3 fill:#e8f5e9
```

## Sub-Components

### Aggregator Sub-Components

1. **Chain Poller**
   - Monitors configured blockchains for TaskCreated events
   - Manages block synchronization and reorg handling
   - Maintains last processed block for each chain

2. **Task Manager**
   - Validates and stores incoming tasks
   - Manages task lifecycle (pending → processing → completed)
   - Handles task distribution to operators

3. **Response Manager**
   - Collects operator responses via gRPC
   - Validates response signatures
   - Enforces response timeouts and SLAs

4. **Signature Aggregator**
   - Aggregates BLS signatures for efficiency
   - Verifies operator stakes meet consensus threshold
   - Prepares aggregated responses for submission

5. **Storage Layer**
   - Persists task state and chain sync progress
   - Supports in-memory and BadgerDB backends
   - Enables recovery after restarts

### Executor Sub-Components

1. **Performer Manager**
   - Manages AVS performer lifecycle
   - Routes tasks to appropriate containers
   - Handles response collection from containers

2. **Container Manager**
   - Docker container lifecycle management
   - Resource isolation and limits enforcement
   - Health monitoring and restart policies

3. **Signing Module**
   - Signs responses with operator keys
   - Supports multiple signature schemes (ECDSA, BLS)
   - Secure key management

4. **Storage Layer**
   - Tracks in-flight tasks
   - Stores performer state
   - Manages deployment information

## Task Flow Process

The following diagram illustrates the complete lifecycle of a task from creation to completion:

```mermaid
sequenceDiagram
    participant AVS as AVS Contract
    participant BC as Blockchain
    participant AG as Aggregator
    participant CP as Chain Poller
    participant TM as Task Manager
    participant EX as Executor
    participant PM as Performer Manager
    participant AC as AVS Container
    participant SA as Signature Aggregator

    AVS->>BC: Create Task (TaskMailbox)
    BC->>BC: Emit TaskCreated Event
    
    CP->>BC: Poll for Events
    BC-->>CP: TaskCreated Event
    CP->>TM: New Task Detected
    TM->>TM: Validate & Store Task
    
    TM->>EX: Distribute Task (gRPC)
    EX->>PM: Route to Performer
    PM->>AC: Execute Task
    
    AC->>AC: Process Task Logic
    AC-->>PM: Task Result
    PM->>PM: Sign Response
    PM-->>EX: Signed Response
    
    EX-->>AG: Submit Response (gRPC)
    AG->>AG: Validate Signature
    AG->>AG: Store Response
    
    AG->>SA: Aggregate Responses
    SA->>SA: Verify Consensus
    SA->>SA: Aggregate Signatures
    
    SA->>BC: Submit Aggregated Response
    BC->>AVS: Update Task Status
    AVS->>AVS: Task Completed

    Note over BC,AG: Multiple operators submit responses in parallel
    Note over SA: Waits for consensus threshold before aggregation
```

## Key Features

### Multi-Chain Support
- Simultaneous monitoring of multiple EVM chains
- Chain-specific configuration (confirmations, polling intervals)
- Unified task processing across L1 and L2 networks

### Flexible Execution Model
- Docker-based isolation for AVS workloads
- Support for various container types (server, one-off)
- Resource limits and GPU support

### Cryptographic Security
- Multiple signature scheme support (ECDSA, BLS)
- Secure key management
- Threshold-based consensus

### High Availability
- Persistent state storage
- Graceful recovery after restarts
- Concurrent task processing

### Production Ready
- Comprehensive metrics and monitoring
- Structured logging
- Health checks and alerting
- Performance optimization

## Deployment Models

### Standalone Deployment
- Single aggregator instance
- Multiple executor instances (one per operator)
- Suitable for single AVS operations

### Multi-AVS Deployment
- Shared aggregator for multiple AVSs
- Executors configured with multiple AVS containers
- Efficient resource utilization

### High Availability Deployment
- Active-passive aggregator configuration
- Load-balanced executor fleet
- Distributed storage backend

## Integration Points

### Smart Contract Integration
- TaskMailbox contract for task creation
- AVS-specific contracts for business logic
- EigenLayer middleware for operator management

### Operator Integration
- Standard gRPC interface for communication
- Flexible authentication mechanisms
- Metrics and monitoring endpoints

### AVS Developer Integration
- Simple container interface
- Environment variable configuration
- Standard logging and error handling

## Getting Started

To deploy Ponos for your AVS:

1. Deploy TaskMailbox contracts on target chains
2. Configure and deploy the Aggregator
3. Operators deploy Executors with AVS containers
4. Monitor task flow and system health

For detailed setup instructions, see the [Configuration Guides](./config/) and [Deployment Guide](./guides/deployment.md).
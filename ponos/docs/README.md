# Ponos Documentation

Welcome to the Ponos documentation. This guide will help you understand, deploy, and operate Ponos for your EigenLayer AVS.

## Quick Start

- **[Overview](./overview.md)** - High-level architecture and components
- **[Aggregator Config](./config/aggregator.md)** - Configure the aggregator service
- **[Executor Config](./config/executor.md)** - Configure the executor service

## Documentation Structure

### 📋 Overview
- **[System Overview](./overview.md)** - Complete system architecture, components, and task flow

### ⚙️ Configuration
- **[Aggregator Configuration](./config/aggregator.md)** - Detailed aggregator configuration guide
- **[Executor Configuration](./config/executor.md)** - Detailed executor configuration guide

### 🔧 Components
- **[Aggregator Deep Dive](./components/aggregator.md)** - Internal architecture of the aggregator
- **[Executor Deep Dive](./components/executor.md)** - Internal architecture of the executor

### 📚 Guides
- **[Operations Guide](./guides/operations-guide.md)** - Day-to-day operations and maintenance
- **[Recovery Procedures](./guides/recovery-procedures.md)** - Disaster recovery and backup procedures
- **[Troubleshooting Guide](./guides/troubleshooting-guide.md)** - Common issues and solutions
- **[Persistence Implementation](./guides/persistence-implementation-summary.md)** - Storage layer implementation details

### 🔗 Multi-Chain Documentation
- **[Certificate Creation](./multichain/certificate-creation-process.md)** - BLS certificate creation process
- **[Contract Call Sequencing](./multichain/contract-call-sequencing.md)** - Multi-chain contract interactions
- **[EigenLayer Integration](./multichain/eigenlayer-contract-call-sequencing.md)** - EigenLayer-specific flows
- **[Signature Flow](./multichain/signature-flow-documentation.md)** - Signature aggregation documentation

### 📊 Benchmarks & Performance
- **[Persistence Benchmarks](./benchmarks/persistence-benchmark-results.md)** - Storage layer performance analysis

### 📝 Plans & Design
- **[Persistence Design](./plans/persistence.md)** - Storage layer design document

## Common Tasks

### For AVS Developers
1. Start with the [Overview](./overview.md) to understand the system
2. Review the [Configuration Guides](./config/) for both components
3. Deploy using the [Operations Guide](./guides/operations-guide.md)

### For Operators
1. Review the [Executor Configuration](./config/executor.md)
2. Follow the [Operations Guide](./guides/operations-guide.md) for deployment
3. Reference [Troubleshooting](./guides/troubleshooting-guide.md) for issues

### For Aggregator Operators
1. Study the [Aggregator Configuration](./config/aggregator.md)
2. Understand [Multi-chain flows](./multichain/)
3. Implement monitoring from the [Operations Guide](./guides/operations-guide.md)

## Key Concepts

### Architecture
Ponos uses a distributed architecture with two main components:
- **Aggregator**: Coordinates tasks and manages operator responses
- **Executor**: Runs AVS workloads in isolated containers

### Task Flow
1. AVS creates task on blockchain
2. Aggregator detects and distributes task
3. Executors run task in containers
4. Responses are signed and returned
5. Aggregator aggregates and submits result

### Storage
Both components support pluggable storage backends:
- **Memory**: Fast, ephemeral (development)
- **BadgerDB**: Persistent, production-ready

### Security
- Operator key management
- Container isolation
- Signature verification
- TLS communication

## Getting Help

- Check the [Troubleshooting Guide](./guides/troubleshooting-guide.md)
- Review component logs
- Monitor metrics endpoints
- Contact support channels

## Contributing

When adding documentation:
1. Use clear, concise language
2. Include examples where helpful
3. Keep diagrams up-to-date
4. Test all command examples
5. Update this index
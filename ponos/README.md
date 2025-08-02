# Ponos - Aggregation and Execution Framework for EigenLayer AVS

Ponos is a comprehensive framework for building and operating EigenLayer AVS (Actively Validated Services) workloads. It provides a robust infrastructure for coordinating tasks across multiple blockchain networks, executing containerized workloads, and aggregating cryptographically signed results.

## Development

```bash
# install deps
make deps

# run the test suite
make test

# build the protos
make proto

# build all binaries
make all

# lint the project
make lint
```

## Documentation

### 📋 Overview
- **[System Overview](./docs/overview.md)** - Complete system architecture, components, and task flow

### ⚙️ Configuration
- **[Aggregator Configuration](./docs/config/aggregator.md)** - Detailed aggregator configuration guide
- **[Executor Configuration](./docs/config/executor.md)** - Detailed executor configuration guide

### 🔧 Components
- **[Aggregator Deep Dive](./docs/components/aggregator.md)** - Internal architecture of the aggregator
- **[Executor Deep Dive](./docs/components/executor.md)** - Internal architecture of the executor

### 📚 Guides
- **[Operations Guide](./docs/guides/operations-guide.md)** - Day-to-day operations and maintenance
- **[Recovery Procedures](./docs/guides/recovery-procedures.md)** - Disaster recovery and backup procedures
- **[Troubleshooting Guide](./docs/guides/troubleshooting-guide.md)** - Common issues and solutions
- **[Persistence Implementation](./docs/guides/persistence-implementation-summary.md)** - Storage layer implementation details

### 🔗 Multi-Chain Documentation
- **[Certificate Creation](./docs/multichain/certificate-creation-process.md)** - BLS certificate creation process
- **[Contract Call Sequencing](./docs/multichain/contract-call-sequencing.md)** - Multi-chain contract interactions
- **[EigenLayer Integration](./docs/multichain/eigenlayer-contract-call-sequencing.md)** - EigenLayer-specific flows
- **[Signature Flow](./docs/multichain/signature-flow-documentation.md)** - Signature aggregation documentation

### 📊 Benchmarks & Performance
- **[Persistence Benchmarks](./docs/benchmarks/persistence-benchmark-results.md)** - Storage layer performance analysis

### 📝 Plans & Design
- **[Persistence Design](./docs/plans/persistence.md)** - Storage layer design document

## Common Tasks

### For AVS Developers
1. Start with the [Overview](./docs/overview.md) to understand the system
2. Review the [Configuration Guides](./docs/config/) for both components
3. Deploy using the [Operations Guide](./docs/guides/operations-guide.md)

### For Operators
1. Review the [Executor Configuration](./docs/config/executor.md)
2. Follow the [Operations Guide](./docs/guides/operations-guide.md) for deployment
3. Reference [Troubleshooting](./docs/guides/troubleshooting-guide.md) for issues

### For Aggregator Operators
1. Study the [Aggregator Configuration](./docs/config/aggregator.md)
2. Understand [Multi-chain flows](./docs/multichain/)
3. Implement monitoring from the [Operations Guide](./docs/guides/operations-guide.md)

## Getting Help

- Check the [Troubleshooting Guide](./docs/guides/troubleshooting-guide.md)
- Review component logs
- Monitor metrics endpoints
- Contact support channels

## Contributing

When adding documentation:
1. Use clear, concise language
2. Include examples where helpful
3. Keep diagrams up-to-date
4. Test all command examples
5. Update the documentation index

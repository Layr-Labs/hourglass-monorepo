# hgctl (Go Implementation)

`hgctl` is a command-line interface for managing Hourglass AVS deployments. This Go implementation leverages ORAS (OCI Registry As Storage) for robust OCI artifact handling and provides support for EigenRuntime specifications.

## Features

- **EigenRuntime Support**: Fetch and deploy AVS components using EigenRuntime specifications
- **OCI Artifact Management**: Pull runtime specs from OCI registries using ORAS
- **Context Management**: Manage multiple environments and configurations
- **Translation**: Convert EigenRuntime specs to various orchestration formats
- **Native Performance**: Compiled Go binary for better performance

## Installation

```bash
# Build from source
make build

# Install to ~/bin
make install

# Or run directly
go run cmd/hgctl/main.go

# Quick test after build
./bin/hgctl --help
```

## Quick Start

```bash
# 1. Set up a context
hgctl context set --rpc-url http://localhost:8545 --release-manager 0x5678...

# 2. List available releases
hgctl get release 0x1234...

# 3. Deploy the latest release
hgctl deploy artifact 0x1234... --operator-set-id 0

# 4. Check deployed performers
hgctl get performer
```

## Usage

### List Releases

List all releases for an AVS:

```bash
# List releases for an AVS
hgctl get release 0x1234... --rpc-url http://localhost:8545 --release-manager 0x5678...

# With limit
hgctl get release 0x1234... --limit 5

# Output as JSON
hgctl get release 0x1234... -o json

# Output as YAML
hgctl get release 0x1234... -o yaml
```

### Describe a Release

Fetch detailed information about a release including its runtime specification:

```bash
# Describe a specific release
hgctl describe release 0x1234... 0 --operator-set-id 1 --rpc-url http://localhost:8545 --release-manager 0x5678...

# With verbose logging
hgctl describe release 0x1234... 0 --operator-set-id 1 -v

# Output as JSON
hgctl describe release 0x1234... 0 --operator-set-id 1 -o json
```

### Deploy Artifacts

Deploy AVS artifacts using EigenRuntime specifications:

```bash
# Deploy latest release
hgctl deploy artifact 0x1234... --operator-set-id 1

# Deploy specific version
hgctl deploy artifact 0x1234... --operator-set-id 1 --version 0

# Legacy mode (direct digest)
hgctl deploy artifact 0x1234... --legacy-digest sha256:abc123 --registry-url ghcr.io/example/avs
```

### Translate Runtime Specs

Convert EigenRuntime specifications to different formats:

```bash
# Translate to Docker Compose
hgctl translate compose -i runtime-spec.yaml -o docker-compose.yml

# Use stdin/stdout
cat runtime-spec.yaml | hgctl translate compose > docker-compose.yml
```

### Context Management

```bash
# List contexts
hgctl context list

# Use a specific context
hgctl context use production

# Set context values
hgctl context set rpcUrl http://mainnet.infura.io
```

### Output Formats

All commands support multiple output formats:

- `table` - Human-readable table format (default)
- `json` - JSON output for programmatic use
- `yaml` - YAML output

## Development

### Project Structure

```
hgctl-go/
├── cmd/hgctl/           # Main entry point
├── internal/
│   ├── commands/        # CLI command implementations
│   ├── client/          # OCI and contract clients
│   ├── config/          # Configuration management
│   ├── executor/        # gRPC client for Executor service
│   ├── logger/          # Logging utilities
│   ├── output/          # Output formatting
│   ├── runtime/         # EigenRuntime spec types
│   ├── telemetry/       # Usage telemetry
│   └── version/         # Version information
├── Makefile
├── go.mod
└── example-runtime-spec.yaml
```

### Key Components

1. **OCI Client** (`internal/client/oci.go`)
   - Uses ORAS v2 for pulling OCI artifacts
   - Supports anonymous authentication for public registries
   - Handles ghcr.io specific authentication

2. **Contract Client** (`internal/client/contract.go`)
   - Interfaces with ReleaseManager smart contract
   - Currently uses mock data (TODO: integrate actual bindings)

3. **Logger** (`internal/logger/logger.go`)
   - Zap-based logging with colored output
   - Matches TypeScript version's format

### Building

```bash
# Build for current platform
make build

# Run tests
make test

# Format code
make fmt

# Run linter
make lint

# Build for all platforms
make release
```

### Adding New Commands

1. Create action file in `internal/commands/`
2. Add command definition to `commands.go`
3. Export function from `commands.go`

Example:
```go
// In commands.go
func MyCommand() *cli.Command {
    return &cli.Command{
        Name:  "mycommand",
        Usage: "Description",
        Action: myCommandAction,
    }
}

// In mycommand.go
func myCommandAction(c *cli.Context) error {
    // Implementation
    return nil
}
```

## Migration from TypeScript

This Go implementation maintains compatibility with the TypeScript version:

- Same command structure and names
- Same context configuration (~/.hgctl/config.yaml)
- Same output formats
- Improved OCI artifact handling with ORAS

## Current Status

### In Progress
- 🔄 Ethereum contract bindings (currently using mock data)
- 🔄 Full E2E testing with real contracts
- 🔄 Private registry authentication

### Architecture
- Uses urfave/cli for command structure
- ORAS v2 for OCI artifact handling
- gRPC client for Executor service communication
- YAML-based configuration in ~/.hgctl/config.yaml

## License

Apache 2.0

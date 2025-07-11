# hgctl - Hourglass Control CLI

Command-line interface for managing AVS deployments in the Hourglass framework.

## Installation

From the hgctl directory:

```bash
npm install
npm run build
npm link  # For local development
```

## Usage

### Basic Commands

```bash
# List all performers
hgctl get performers

# List performers for a specific AVS
hgctl get performers 0x1234...

# List releases for an AVS
hgctl get releases 0x1234...

# Deploy an artifact
hgctl deploy artifact 0x1234... sha256:abcd...

# Remove a performer
hgctl remove performer performer-123
```

### Context Management

```bash
# Show current context
hgctl context show

# Set context values
hgctl context set --executor-address localhost:9090 --rpc-url http://localhost:8545

# Create and use a new context
hgctl context set production --executor-address prod.example.com:9090
hgctl context use production

# List all contexts
hgctl context list
```

### Output Formats

```bash
# JSON output
hgctl get performers -o json

# YAML output
hgctl get performers -o yaml

# Table output (default)
hgctl get performers -o table
```

### Shell Completion

```bash
# Bash
source <(hgctl completion bash)

# Zsh
source <(hgctl completion zsh)

# Fish
hgctl completion fish > ~/.config/fish/completions/hgctl.fish
```

## Configuration

Configuration is stored in `~/.hgctl/config.yaml`:

```yaml
currentContext: default
contexts:
  default:
    executorAddress: executor:9090
    rpcUrl: http://localhost:8545
    releaseManagerAddress: 0x...
  production:
    executorAddress: prod.example.com:9090
    rpcUrl: https://mainnet.infura.io/v3/...
    releaseManagerAddress: 0x...
```

## Development

```bash
# Run in development mode
npm run dev -- get performers

# Run tests
npm test

# Lint code
npm run lint

# Format code
npm run format
```

## Building

```bash
# Build TypeScript
npm run build

# Package for distribution
npm run package
```

This creates standalone executables for Linux, macOS, and Windows in the `dist/` directory.
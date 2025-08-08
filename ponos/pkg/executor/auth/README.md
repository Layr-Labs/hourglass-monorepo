# Executor Authentication Implementation

## Overview

This implementation adds server-generated challenge token-based authentication to the Executor Management service for the following RPCs:
- `DeployArtifact`
- `ListPerformers` 
- `RemovePerformer`

The `SubmitTask` RPC remains unauthenticated as requested.

## Authentication Flow

1. **Client requests a challenge token**: Client calls `GetChallengeToken` with their operator address
2. **Server generates token**: Server creates a single-use challenge token (Keccak256 hash of UUID) tied to the operator and returns it with expiration time
3. **Client signs request**: Client signs `Hash(challenge_token + method + request_payload)` with their private key
4. **Client sends authenticated request**: Request includes the `AuthSignature` with challenge token and signature
5. **Server verifies**: Server checks token validity and signature correctness

## Key Components

### Server-Side

- **ChallengeTokenManager** (`auth.go`): 
  - Manages challenge token generation using Keccak256 hash of UUID
  - Tracks all generated tokens (used and unused)
  - Validates single-use property and expiration
  - Automatically cleans up expired tokens
  - Provides statistics on token usage
- **Authentication verification** (`auth.go`): `verifyAuthentication` method validates requests
- **Handler updates** (`handlers.go`): Protected RPCs now verify authentication before processing

### Client-Side

- **AuthenticatedExecutorClient** (`authenticatedClient.go`): Wrapper that handles challenge token acquisition and request signing

## Security Features

- **Single-use tokens**: Each challenge token can only be used once, preventing replay attacks
- **UUID-based tokens**: Uses Keccak256 hash of UUID for unpredictable token generation
- **Time-bound tokens**: Challenge tokens expire after 5 minutes
- **Operator-bound tokens**: ChallengeTokenManager is initialized with a specific operator address
- **Token tracking**: All tokens remain tracked even after use for auditing
- **Cryptographic signatures**: Requests are signed with operator's private key

## Usage Example

```go
// Create authenticated client
client, err := executorClient.NewAuthenticatedExecutorClient(
    "localhost:9090",
    operatorAddress,
    operatorSigner,
    true, // insecure connection for testing
)

// Make authenticated request
resp, err := client.DeployArtifact(ctx, &executorV1.DeployArtifactRequest{
    AvsAddress:  "0xAVS123",
    Digest:      "sha256:abc123",
    RegistryUrl: "registry.example.com",
})
```

## Testing

Run the authentication tests:
```bash
go test ./pkg/executor -run TestChallengeTokenManager -v
go test ./pkg/executor -run TestVerifyAuthentication -v
```
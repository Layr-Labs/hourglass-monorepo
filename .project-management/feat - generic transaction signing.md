# Feature: Generic Transaction Signing

## Overview
The current `ContractCaller` implementation in `ponos/pkg/contractCaller/caller/caller.go` is tightly coupled to direct private key signing. This document outlines a plan to implement a generic signing interface that supports multiple signing methods while maintaining backward compatibility.

## Current Architecture Analysis

### Current State
- **Configuration**: `ContractCallerConfig` contains a `PrivateKey` string field
- **Signing Pattern**: All transaction signing flows follow the same pattern:
  1. `buildNoSendOptsWithPrivateKey(ctx)` → returns `*bind.TransactOpts` and `*ecdsa.PrivateKey`
  2. Contract method call with `noSendTxOpts` (creates unsigned transaction)
  3. `EstimateGasPriceAndLimitAndSendTx(ctx, from, tx, privateKey, operation)` → signs and sends transaction

### Key Signing Methods Identified
- `buildNoSendOptsWithPrivateKey(ctx)` - Lines 807-818
- `buildTxOps(ctx, pk)` - Lines 820-832
- `EstimateGasPriceAndLimitAndSendTx(ctx, from, tx, privateKey, operation)` - Referenced but not shown in file

### Transaction Signing Locations
The following methods currently use direct private key signing:
- `SubmitBN254TaskResult()` - Line 174
- `SubmitECDSATaskResult()` - Line 275
- `PublishMessageToInbox()` - Line 463
- `ConfigureAVSOperatorSet()` - Line 592
- `RegisterKeyWithKeyRegistrar()` - Line 631
- `createOperator()` - Line 712
- `registerOperatorWithAvs()` - Line 762
- `SetupTaskMailboxForAvs()` - Line 982
- `DelegateToOperator()` - Line 1026
- `ModifyAllocations()` - Line 1071
- `VerifyECDSACertificate()` - Line 1100

## Implementation Plan

### Phase 1: Design Generic Signing Interface

#### 1.1 Create Signing Interface
**File**: `ponos/pkg/transaction-signer/signer.go`

```go
package transactionSigner

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// TransactionSigner provides methods for signing Ethereum transactions
type TransactionSigner interface {
	// GetTransactOpts returns transaction options for creating unsigned transactions
	GetTransactOpts(ctx context.Context) (*bind.TransactOpts, error)

	// SignAndSendTransaction signs a transaction and sends it to the network
	SignAndSendTransaction(ctx context.Context, tx *types.Transaction) (*types.Receipt, error)

	// GetFromAddress returns the address that will be used for signing
	GetFromAddress() common.Address

	// EstimateGasPriceAndLimit estimates gas price and limit for a transaction
	EstimateGasPriceAndLimit(ctx context.Context, tx *types.Transaction) (*big.Int, uint64, error)
}
```

#### 1.2 Create Signing Context
**File**: `ponos/pkg/transaction-signer/context.go`

```go
package transactionSigner

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

// SigningContext provides common functionality for transaction signing
type SigningContext struct {
	ethClient *ethclient.Client
	logger    *zap.Logger
	chainID   *big.Int
}

// NewSigningContext creates a new signing context
func NewSigningContext(ethClient *ethclient.Client, logger *zap.Logger) (*SigningContext, error) {
	chainID, err := ethClient.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	return &SigningContext{
		ethClient: ethClient,
		logger:    logger,
		chainID:   chainID,
	}, nil
}

// EstimateGasPriceAndLimit provides common gas estimation logic
func (sc *SigningContext) EstimateGasPriceAndLimit(ctx context.Context, tx *types.Transaction) (*big.Int, uint64, error) {
	// Implementation for gas estimation logic
	// This would contain the current EstimateGasPriceAndLimitAndSendTx logic
	// extracted and made generic
	return nil, 0, nil
}
```

### Phase 2: Implement Direct Private Key Signer

#### 2.1 Private Key Signer Implementation
**File**: `ponos/pkg/transaction-signer/privateKeySigner.go`

```go
package transactionSigner

import (
	"context"
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	cryptoUtils "github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/crypto"
)

// PrivateKeySigner implements TransactionSigner using a private key
type PrivateKeySigner struct {
	*SigningContext
	privateKey  *ecdsa.PrivateKey
	fromAddress common.Address
}

// NewPrivateKeySigner creates a new private key signer
func NewPrivateKeySigner(privateKeyHex string, signingContext *SigningContext) (*PrivateKeySigner, error) {
	privateKey, err := cryptoUtils.StringToECDSAPrivateKey(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to get public key ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	return &PrivateKeySigner{
		SigningContext: signingContext,
		privateKey:     privateKey,
		fromAddress:    fromAddress,
	}, nil
}

// GetTransactOpts returns transaction options for creating unsigned transactions
func (pks *PrivateKeySigner) GetTransactOpts(ctx context.Context) (*bind.TransactOpts, error) {
	opts, err := bind.NewKeyedTransactorWithChainID(pks.privateKey, pks.chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %w", err)
	}
	opts.NoSend = true
	opts.Context = ctx
	return opts, nil
}

// SignAndSendTransaction signs a transaction and sends it to the network
func (pks *PrivateKeySigner) SignAndSendTransaction(ctx context.Context, tx *types.Transaction) (*types.Receipt, error) {
	// Sign the transaction
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(pks.chainID), pks.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send the transaction
	err = pks.ethClient.SendTransaction(ctx, signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	// Wait for receipt
	receipt, err := bind.WaitMined(ctx, pks.ethClient, signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for transaction receipt: %w", err)
	}

	return receipt, nil
}

// GetFromAddress returns the address that will be used for signing
func (pks *PrivateKeySigner) GetFromAddress() common.Address {
	return pks.fromAddress
}
```

### Phase 3: Implement Web3Signer Integration

#### 3.1 Web3Signer Implementation
**File**: `ponos/pkg/transaction-signer/web3Signer.go`

```go
package transactionSigner

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/web3signer"
)

// Web3Signer implements TransactionSigner using Web3Signer service
type Web3Signer struct {
	*SigningContext
	web3SignerClient *web3signer.Client
	fromAddress      common.Address
}

// NewWeb3Signer creates a new Web3Signer
func NewWeb3Signer(web3SignerClient *web3signer.Client, fromAddress common.Address, signingContext *SigningContext) *Web3Signer {
	return &Web3Signer{
		SigningContext:   signingContext,
		web3SignerClient: web3SignerClient,
		fromAddress:      fromAddress,
	}
}

// GetTransactOpts returns transaction options for creating unsigned transactions
func (w3s *Web3Signer) GetTransactOpts(ctx context.Context) (*bind.TransactOpts, error) {
	opts := &bind.TransactOpts{
		From:    w3s.fromAddress,
		Context: ctx,
		NoSend:  true,
		Signer:  w3s.signTransaction,
	}
	return opts, nil
}

// SignAndSendTransaction signs a transaction and sends it to the network
func (w3s *Web3Signer) SignAndSendTransaction(ctx context.Context, tx *types.Transaction) (*types.Receipt, error) {
	// Convert transaction to Web3Signer format
	txData := map[string]interface{}{
		"to":       tx.To().Hex(),
		"value":    hexutil.EncodeBig(tx.Value()),
		"gas":      hexutil.EncodeUint64(tx.Gas()),
		"gasPrice": hexutil.EncodeBig(tx.GasPrice()),
		"nonce":    hexutil.EncodeUint64(tx.Nonce()),
		"data":     hexutil.Encode(tx.Data()),
	}

	// Sign with Web3Signer
	signedTxHex, err := w3s.web3SignerClient.EthSignTransaction(ctx, w3s.fromAddress.Hex(), txData)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction with Web3Signer: %w", err)
	}

	// Parse signed transaction
	signedTxBytes, err := hexutil.Decode(signedTxHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode signed transaction: %w", err)
	}

	var signedTx types.Transaction
	err = signedTx.UnmarshalBinary(signedTxBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal signed transaction: %w", err)
	}

	// Send the transaction
	err = w3s.ethClient.SendTransaction(ctx, &signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	// Wait for receipt
	receipt, err := bind.WaitMined(ctx, w3s.ethClient, &signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for transaction receipt: %w", err)
	}

	return receipt, nil
}

// GetFromAddress returns the address that will be used for signing
func (w3s *Web3Signer) GetFromAddress() common.Address {
	return w3s.fromAddress
}

// signTransaction is a signing function for bind.TransactOpts
func (w3s *Web3Signer) signTransaction(addr common.Address, tx *types.Transaction) (*types.Transaction, error) {
	// This would be called by go-ethereum binding code
	// Implementation depends on specific requirements
	return nil, fmt.Errorf("direct signing not supported for Web3Signer")
}
```

### Phase 4: Update ContractCaller

#### 4.1 Update Configuration
**File**: `ponos/pkg/contractCaller/caller/caller.go`

```go
// Updated configuration structure
type ContractCallerConfig struct {
	// Deprecated: Use Signer instead
	PrivateKey          string
	AVSRegistrarAddress string
	TaskMailboxAddress  string
	KeyRegistrarAddress string

	// New signing interface
	Signer transactionsigner.TransactionSigner
}

// Updated ContractCaller structure
type ContractCaller struct {
	avsRegistrarCaller *TaskAVSRegistrarBase.TaskAVSRegistrarBaseCaller
	taskMailbox        *ITaskMailbox.ITaskMailbox
	allocationManager  *IAllocationManager.IAllocationManager
	delegationManager  *IDelegationManager.IDelegationManager
	crossChainRegistry *ICrossChainRegistry.ICrossChainRegistry
	keyRegistrar       *IKeyRegistrar.IKeyRegistrar
	ecdsaCertVerifier  *IECDSACertificateVerifier.IECDSACertificateVerifier
	ethclient          *ethclient.Client
	config             *ContractCallerConfig
	logger             *zap.Logger
	coreContracts      *config.CoreContractAddresses

	// New signing interface
	signer transactionsigner.TransactionSigner
}
```

#### 4.2 Update Constructor
```go
func NewContractCaller(
	cfg *ContractCallerConfig,
	ethclient *ethclient.Client,
	logger *zap.Logger,
) (*ContractCaller, error) {
	// ... existing contract initialization ...

	// Initialize signer
	var signer transactionsigner.TransactionSigner
	if cfg.Signer != nil {
		signer = cfg.Signer
	} else if cfg.PrivateKey != "" {
		// Backward compatibility: create private key signer
		signingContext, err := transactionsigner.NewSigningContext(ethclient, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create signing context: %w", err)
		}

		signer, err = transactionsigner.NewPrivateKeySigner(cfg.PrivateKey, signingContext)
		if err != nil {
			return nil, fmt.Errorf("failed to create private key signer: %w", err)
		}
	} else {
		return nil, fmt.Errorf("no signer provided")
	}

	return &ContractCaller{
		// ... existing fields ...
		signer: signer,
	}, nil
}
```

#### 4.3 Update Signing Methods
```go
// Replace buildNoSendOptsWithPrivateKey with generic version
func (cc *ContractCaller) buildTransactionOpts(ctx context.Context) (*bind.TransactOpts, error) {
	return cc.signer.GetTransactOpts(ctx)
}

// Replace EstimateGasPriceAndLimitAndSendTx with generic version
func (cc *ContractCaller) signAndSendTransaction(ctx context.Context, tx *types.Transaction, operation string) (*types.Receipt, error) {
	cc.logger.Sugar().Infow("Signing and sending transaction",
		zap.String("operation", operation),
		zap.String("from", cc.signer.GetFromAddress().Hex()),
		zap.String("to", tx.To().Hex()),
	)

	return cc.signer.SignAndSendTransaction(ctx, tx)
}
```

#### 4.4 Update All Transaction Methods
Replace all occurrences of:
- `buildNoSendOptsWithPrivateKey(ctx)` → `buildTransactionOpts(ctx)`
- `EstimateGasPriceAndLimitAndSendTx(ctx, from, tx, privateKey, operation)` → `signAndSendTransaction(ctx, tx, operation)`

### Phase 5: Configuration and Factory

#### 5.1 Signer Factory
**File**: `ponos/pkg/transaction-signer/factory.go`

```go
package transactionSigner

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/web3signer"
	"go.uber.org/zap"
)

// SignerConfig represents configuration for creating signers
type SignerConfig struct {
	Type string `yaml:"type"` // "private_key" or "web3signer"

	// Private key configuration
	PrivateKey string `yaml:"private_key,omitempty"`

	// Web3Signer configuration
	Web3SignerURL string `yaml:"web3signer_url,omitempty"`
	FromAddress   string `yaml:"from_address,omitempty"`
}

// CreateSigner creates a signer based on configuration
func CreateSigner(config *SignerConfig, ethClient *ethclient.Client, logger *zap.Logger) (TransactionSigner, error) {
	signingContext, err := NewSigningContext(ethClient, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create signing context: %w", err)
	}

	switch config.Type {
	case "private_key":
		return NewPrivateKeySigner(config.PrivateKey, signingContext)

	case "web3signer":
		web3SignerConfig := &web3signer.Config{
			BaseURL: config.Web3SignerURL,
		}
		web3SignerClient := web3signer.NewClient(web3SignerConfig, logger)

		fromAddress := common.HexToAddress(config.FromAddress)
		return NewWeb3Signer(web3SignerClient, fromAddress, signingContext), nil

	default:
		return nil, fmt.Errorf("unsupported signer type: %s", config.Type)
	}
}
```

### Phase 6: Testing Strategy

#### 6.1 Unit Tests
- Test each signer implementation independently
- Mock Web3Signer client for testing
- Test backward compatibility with existing private key configuration

#### 6.2 Integration Tests
- Test with real Web3Signer instance
- Test transaction signing and sending
- Test error handling and recovery

#### 6.3 Migration Tests
- Test migration from old configuration to new configuration
- Test that existing functionality continues to work

### Phase 7: Documentation and Migration

#### 7.1 Configuration Migration Guide
- Document how to migrate from `PrivateKey` to `Signer` configuration
- Provide examples for both private key and Web3Signer configurations
- Document backward compatibility guarantees

#### 7.2 API Documentation
- Document the new signing interface
- Provide usage examples
- Document error handling patterns

## Implementation Timeline

### Week 1: Core Interface Design
- Implement generic signing interface
- Create signing context
- Implement private key signer

### Week 2: Web3Signer Integration
- Implement Web3Signer signer
- Create signer factory
- Update configuration structures

### Week 3: ContractCaller Updates
- Update ContractCaller to use generic signer
- Maintain backward compatibility
- Update all transaction methods

### Week 4: Testing and Documentation
- Comprehensive testing
- Documentation
- Migration guide

## Benefits

1. **Flexibility**: Support for multiple signing methods
2. **Security**: Ability to use remote signing services like Web3Signer
3. **Maintainability**: Cleaner separation of concerns
4. **Extensibility**: Easy to add new signing methods
5. **Backward Compatibility**: Existing code continues to work

## Risks and Mitigations

1. **Breaking Changes**: Mitigated by maintaining backward compatibility
2. **Complexity**: Mitigated by clean interface design and good documentation
3. **Performance**: Mitigated by efficient implementation and testing
4. **Security**: Mitigated by proper error handling and validation

## Success Criteria

1. All existing functionality continues to work
2. New Web3Signer integration works correctly
3. Performance is not degraded
4. Code is well-tested and documented
5. Migration path is clear and straightforward

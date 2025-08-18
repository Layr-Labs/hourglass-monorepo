package client

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/IKeyRegistrar"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/IReleaseManager"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/contracts"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"

	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/IAllocationManager"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/IDelegationManager"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/IStrategyManager"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/middleware-bindings/ITaskAVSRegistrarBase"
)

// Release types
type OperatorSetRelease struct {
	Digest   string `json:"digest"`
	Registry string `json:"registry"`
}

type Release struct {
	ID                  string                        `json:"id"`
	OperatorSetReleases map[string]OperatorSetRelease `json:"operatorSetReleases"`
	UpgradeByTime       uint32                        `json:"upgradeByTime"`
}

// ReleaseArtifact represents an artifact in a release
type ReleaseArtifact struct {
	Digest       [32]byte
	RegistryName string
}

// ReleaseManagerRelease represents a release from the contract
type ReleaseManagerRelease struct {
	UpgradeByTime uint32
	Artifacts     []ReleaseArtifact
}

// AVSConfig represents the configuration for an AVS
type AVSConfig struct {
	AggregatorOperatorSetID uint32
	ExecutorOperatorSetIDs  []uint32
}

// ContractConfig contains all configuration needed to instantiate contracts
type ContractConfig struct {
	// Required addresses
	AVSAddress      string
	OperatorAddress string

	// Optional contract addresses (will use defaults if not provided)
	DelegationManager string
	AllocationManager string
	StrategyManager   string
	KeyRegistrar      string
	ReleaseManager    string
}

type ContractClient struct {
	ethClient         *ethclient.Client
	logger            logger.Logger
	privateKey        *ecdsa.PrivateKey
	chainID           *big.Int
	avsAddress        common.Address
	operatorAddress   common.Address
	allocationManager *IAllocationManager.IAllocationManager
	delegationManager *IDelegationManager.IDelegationManager
	strategyManager   *IStrategyManager.IStrategyManager
	keyRegistrar      *IKeyRegistrar.IKeyRegistrar
	releaseManager    *IReleaseManager.IReleaseManager
	contractConfig    *ContractConfig
}

// DefaultContractAddresses contains the default contract addresses for a chain
type DefaultContractAddresses struct {
	DelegationManager string
	AllocationManager string
	StrategyManager   string
	KeyRegistrar      string
	ReleaseManager    string
}

// getDefaultContractAddresses returns the default contract addresses for a given chain ID
func getDefaultContractAddresses(chainID uint64) (*DefaultContractAddresses, error) {
	switch chainID {
	case 11155111:
		return &DefaultContractAddresses{
			DelegationManager: "0xD4A7E1Bd8015057293f0D0A557088c286942e84b",
			AllocationManager: "0x42583067658071247ec8ce0a516a58f682002d07",
			StrategyManager:   "0x2E3D6c0744b10eb0A4e6F679F71554a39Ec47a5D",
			KeyRegistrar:      "0xA4dB30D08d8bbcA00D40600bee9F029984dB162a",
			ReleaseManager:    "0x59c8D715DCa616e032B744a753C017c9f3E16bf4",
		}, nil
	case 31337:
		return &DefaultContractAddresses{
			DelegationManager: "0xD4A7E1Bd8015057293f0D0A557088c286942e84b",
			AllocationManager: "0x42583067658071247ec8ce0a516a58f682002d07",
			StrategyManager:   "0xdfB5f6CE42aAA7830E94ECFCcAd411beF4d4D5b6",
			KeyRegistrar:      "0xA4dB30D08d8bbcA00D40600bee9F029984dB162a",
			ReleaseManager:    "0xd9Cb89F1993292dEC2F973934bC63B0f2A702776",
		}, nil
	default:
		return nil, fmt.Errorf("default contract addresses not found for chain")
	}
}

// NewContractClient creates a new contract client with the given configuration
func NewContractClient(rpcURL, privateKeyHex string, log logger.Logger, config *ContractConfig) (*ContractClient, error) {
	if rpcURL == "" {
		return nil, fmt.Errorf("RPC URL is required")
	}

	if config == nil {
		return nil, fmt.Errorf("contract config is required")
	}

	if config.AVSAddress == "" {
		return nil, fmt.Errorf("AVS address is required")
	}

	if config.OperatorAddress == "" {
		return nil, fmt.Errorf("operator address is required")
	}

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum node: %w", err)
	}

	// Parse private key if provided - don't fail if not provided
	var privateKey *ecdsa.PrivateKey
	if privateKeyHex != "" {
		privateKey, err = crypto.HexToECDSA(strings.TrimPrefix(privateKeyHex, "0x"))
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
	} else {
		log.Debug("Private key not configured - read-only mode enabled")
	}

	// Get chain ID
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	contractClient := &ContractClient{
		ethClient:       client,
		logger:          log,
		privateKey:      privateKey,
		chainID:         chainID,
		avsAddress:      common.HexToAddress(config.AVSAddress),
		operatorAddress: common.HexToAddress(config.OperatorAddress),
		contractConfig:  config,
	}

	// Get default addresses for this chain if not provided in config
	defaultAddresses, err := getDefaultContractAddresses(chainID.Uint64())
	if err != nil {
		return nil, fmt.Errorf("failed to get default contract addresses: %w", err)
	}

	// Delegation Manager
	delegationAddr := config.DelegationManager
	if delegationAddr == "" && defaultAddresses != nil {
		delegationAddr = defaultAddresses.DelegationManager
	}
	if delegationAddr != "" {
		contractClient.delegationManager, err = IDelegationManager.NewIDelegationManager(
			common.HexToAddress(delegationAddr),
			client,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create delegation manager at %s: %w", delegationAddr, err)
		}
		log.Debug("Initialized delegation manager", zap.String("address", delegationAddr))
	}

	// Allocation Manager
	allocationAddr := config.AllocationManager
	if allocationAddr == "" && defaultAddresses != nil {
		allocationAddr = defaultAddresses.AllocationManager
	}
	if allocationAddr != "" {
		contractClient.allocationManager, err = IAllocationManager.NewIAllocationManager(
			common.HexToAddress(allocationAddr),
			client,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create allocation manager at %s: %w", allocationAddr, err)
		}
		log.Debug("Initialized allocation manager", zap.String("address", allocationAddr))
	}

	// Strategy Manager
	strategyAddr := config.StrategyManager
	if strategyAddr == "" && defaultAddresses != nil {
		strategyAddr = defaultAddresses.StrategyManager
	}
	if strategyAddr != "" {
		contractClient.strategyManager, err = IStrategyManager.NewIStrategyManager(
			common.HexToAddress(strategyAddr),
			client,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create strategy manager at %s: %w", strategyAddr, err)
		}
		log.Debug("Initialized strategy manager", zap.String("address", strategyAddr))
	}

	// Key Registrar
	keyRegistrarAddr := config.KeyRegistrar
	if keyRegistrarAddr == "" && defaultAddresses != nil {
		keyRegistrarAddr = defaultAddresses.KeyRegistrar
	}
	if keyRegistrarAddr != "" {
		contractClient.keyRegistrar, err = IKeyRegistrar.NewIKeyRegistrar(
			common.HexToAddress(keyRegistrarAddr),
			client,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create key registrar at %s: %w", keyRegistrarAddr, err)
		}
		log.Debug("Initialized key registrar", zap.String("address", keyRegistrarAddr))
	}

	// Release Manager
	releaseManagerAddr := config.ReleaseManager
	if releaseManagerAddr == "" && defaultAddresses != nil {
		releaseManagerAddr = defaultAddresses.ReleaseManager
	}
	if releaseManagerAddr != "" {
		contractClient.releaseManager, err = IReleaseManager.NewIReleaseManager(
			common.HexToAddress(releaseManagerAddr),
			client,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create release manager at %s: %w", releaseManagerAddr, err)
		}
		log.Debug("Initialized release manager", zap.String("address", releaseManagerAddr))
	}

	return contractClient, nil
}

// GetRelease fetches a release from the ReleaseManager contract
func (c *ContractClient) GetRelease(
	ctx context.Context,
	operatorSetId uint32,
	releaseId *big.Int,
) (*ReleaseManagerRelease, error) {
	// Create operator set
	operatorSet := IReleaseManager.OperatorSet{Avs: c.avsAddress, Id: operatorSetId}

	// Get release from contract
	release, err := c.releaseManager.GetRelease(&bind.CallOpts{Context: ctx}, operatorSet, releaseId)
	if err != nil {
		return nil, fmt.Errorf("failed to get release: %w", err)
	}

	// Convert to our internal type
	artifacts := make([]ReleaseArtifact, len(release.Artifacts))
	for i, artifact := range release.Artifacts {
		artifacts[i] = ReleaseArtifact{
			Digest:       artifact.Digest,
			RegistryName: artifact.Registry,
		}
	}

	return &ReleaseManagerRelease{
		UpgradeByTime: release.UpgradeByTime,
		Artifacts:     artifacts,
	}, nil
}

// GetReleaseCount gets the next release ID for an operator set
func (c *ContractClient) GetReleaseCount(ctx context.Context, operatorSetId uint32) (*big.Int, error) {
	// Create operator set
	operatorSet := IReleaseManager.OperatorSet{Avs: c.avsAddress, Id: operatorSetId}

	// Get total releases
	totalReleases, err := c.releaseManager.GetTotalReleases(&bind.CallOpts{Context: ctx}, operatorSet)
	if err != nil {
		return nil, fmt.Errorf("failed to get total releases: %w", err)
	}

	return totalReleases, nil
}

// GetReleases fetches multiple releases organized by operator set
func (c *ContractClient) GetReleases(ctx context.Context, operatorSetIds []uint32, limit uint64) ([]*Release, error) {
	var releases []*Release

	// Get releases for each operator set separately
	for _, opSetId := range operatorSetIds {
		nextId, err := c.GetReleaseCount(ctx, opSetId)
		if err != nil {
			c.logger.Warn("Failed to get next release ID",
				zap.Uint32("operatorSetId", opSetId),
				zap.Error(err))
			continue
		}

		totalReleases := nextId.Int64()
		if totalReleases == 0 {
			continue
		}

		// Fetch releases in descending order (newest first)
		for i := totalReleases - 1; i >= totalReleases-int64(limit) && i >= 0; i-- {
			release, err := c.GetRelease(ctx, opSetId, big.NewInt(i))
			if err != nil {
				continue
			}

			// Create release entry for this operator set
			internalRelease := &Release{
				ID: fmt.Sprintf("%d", i),
				OperatorSetReleases: map[string]OperatorSetRelease{
					fmt.Sprintf("%d", opSetId): {
						Digest:   fmt.Sprintf("0x%x", release.Artifacts[0].Digest),
						Registry: release.Artifacts[0].RegistryName,
					},
				},
				UpgradeByTime: release.UpgradeByTime,
			}
			releases = append(releases, internalRelease)
		}
	}

	return releases, nil
}

// checkPrivateKey ensures a private key is configured for mutable operations
func (c *ContractClient) checkPrivateKey() error {
	if c.privateKey == nil {
		return fmt.Errorf("private key not configured - this operation requires a configured private key")
	}
	return nil
}

// buildTxOpts creates transaction options for signing
func (c *ContractClient) buildTxOpts(ctx context.Context) (*bind.TransactOpts, error) {
	if err := c.checkPrivateKey(); err != nil {
		return nil, err
	}

	opts, err := bind.NewKeyedTransactorWithChainID(c.privateKey, c.chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %w", err)
	}
	opts.Context = ctx
	return opts, nil
}

// Operator Management Methods

// RegisterAsOperator registers an address as an operator with EigenLayer
func (c *ContractClient) RegisterAsOperator(ctx context.Context, allocationDelay uint32, metadataURI string) error {
	if err := c.checkPrivateKey(); err != nil {
		return err
	}

	if c.delegationManager == nil {
		return fmt.Errorf("delegation manager not initialized")
	}

	opts, err := c.buildTxOpts(ctx)
	if err != nil {
		return err
	}

	// Register operator with zero address as delegation approver (for self-delegation)
	// The operator performing the registration will automatically be registered and self-delegated
	tx, err := c.delegationManager.RegisterAsOperator(opts, c.operatorAddress, allocationDelay, metadataURI)
	if err != nil {
		return fmt.Errorf("failed to register operator: %w", err)
	}

	// Wait for transaction
	receipt, err := bind.WaitMined(ctx, c.ethClient, tx)
	if err != nil {
		return fmt.Errorf("failed to wait for transaction: %w", err)
	}

	if receipt.Status == 0 {
		return fmt.Errorf("transaction reverted")
	}

	c.logger.Info("Successfully registered operator with EigenLayer",
		zap.String("address", c.operatorAddress.Hex()),
		zap.String("txHash", receipt.TxHash.Hex()),
	)

	return nil
}

// RegisterOperatorToAVS registers an operator to an AVS
func (c *ContractClient) RegisterOperatorToAVS(ctx context.Context, operatorSetIDs []uint32, data []byte) error {
	if err := c.checkPrivateKey(); err != nil {
		return err
	}

	if c.allocationManager == nil {
		return fmt.Errorf("allocation manager not initialized")
	}

	opts, err := c.buildTxOpts(ctx)
	if err != nil {
		return err
	}

	// Create registration parameters with the provided data
	registerParams := IAllocationManager.IAllocationManagerTypesRegisterParams{
		Avs:            c.avsAddress,
		OperatorSetIds: operatorSetIDs,
		Data:           data,
	}

	// Register for operator sets
	tx, err := c.allocationManager.RegisterForOperatorSets(opts, c.operatorAddress, registerParams)
	if err != nil {
		return fmt.Errorf("failed to register operator to AVS: %w", err)
	}

	// Wait for transaction
	receipt, err := bind.WaitMined(ctx, c.ethClient, tx)
	if err != nil {
		return fmt.Errorf("failed to wait for transaction: %w", err)
	}

	if receipt.Status == 0 {
		return fmt.Errorf("transaction reverted")
	}

	c.logger.Info("Successfully registered operator to AVS",
		zap.String("operator", c.operatorAddress.Hex()),
		zap.String("avs", c.avsAddress.Hex()),
		zap.String("txHash", receipt.TxHash.Hex()),
	)

	return nil
}

// DepositIntoStrategy deposits tokens into a strategy
func (c *ContractClient) DepositIntoStrategy(
	ctx context.Context,
	strategyAddress string,
	tokenAddress string,
	amount *big.Int,
) error {
	if err := c.checkPrivateKey(); err != nil {
		return err
	}

	if c.strategyManager == nil {
		return fmt.Errorf("strategy manager not initialized")
	}

	opts, err := c.buildTxOpts(ctx)
	if err != nil {
		return err
	}

	// Convert strategy address
	stratAddr := common.HexToAddress(strategyAddress)
	tokenAddr := common.HexToAddress(tokenAddress)
	erc20, err := c.getERC20(tokenAddr)
	if err != nil {
		return err
	}
	approveTx, err := erc20.Transact(opts, "approve", stratAddr, amount)
	if err != nil {
		return err
	}
	_, err = bind.WaitMined(ctx, c.ethClient, approveTx)
	if err != nil {
		return err
	}

	// Deposit into strategy
	depositTx, err := c.strategyManager.DepositIntoStrategy(opts, stratAddr, tokenAddr, amount)
	if err != nil {
		return fmt.Errorf("failed to deposit into strategy: %w", err)
	}

	// Wait for transaction
	depositReceipt, err := bind.WaitMined(ctx, c.ethClient, depositTx)
	if err != nil {
		return fmt.Errorf("failed to wait for transaction: %w", err)
	}

	if depositReceipt.Status == 0 {
		return fmt.Errorf("transaction reverted")
	}

	c.logger.Info("Successfully deposited into strategy",
		zap.String("strategy", strategyAddress),
		zap.String("amount", amount.String()),
		zap.String("txHash", depositReceipt.TxHash.Hex()),
	)

	return nil
}

// DelegateTo delegates stake to an operator
func (c *ContractClient) DelegateTo(ctx context.Context, operatorAddress string) error {
	if err := c.checkPrivateKey(); err != nil {
		return err
	}

	if c.delegationManager == nil {
		return fmt.Errorf("delegation manager not initialized")
	}

	opts, err := c.buildTxOpts(ctx)
	if err != nil {
		return err
	}

	// Convert operator address
	opAddr := common.HexToAddress(operatorAddress)

	// Delegate to operator
	tx, err := c.delegationManager.DelegateTo(opts, opAddr,
		IDelegationManager.ISignatureUtilsMixinTypesSignatureWithExpiry{
			Signature: []byte{},
			// TODO: parameterize this.
			Expiry: big.NewInt(0),
		},
		[32]byte{},
	)
	if err != nil {
		return fmt.Errorf("failed to delegate to operator: %w", err)
	}

	// Wait for transaction
	receipt, err := bind.WaitMined(ctx, c.ethClient, tx)
	if err != nil {
		return fmt.Errorf("failed to wait for transaction: %w", err)
	}

	if receipt.Status == 0 {
		return fmt.Errorf("transaction reverted")
	}

	c.logger.Info("Successfully delegated to operator",
		zap.String("operator", operatorAddress),
		zap.String("txHash", receipt.TxHash.Hex()),
	)

	return nil
}

// GetOperatorECDSAKeyRegistrationMessageHash gets the message hash for ECDSA key registration
func (c *ContractClient) GetOperatorECDSAKeyRegistrationMessageHash(
	ctx context.Context,
	operatorSetID uint32,
	keyAddress common.Address,
) ([32]byte, error) {
	if c.keyRegistrar == nil {
		return [32]byte{}, fmt.Errorf("key registrar not initialized")
	}

	operatorSet := IKeyRegistrar.OperatorSet{Avs: c.avsAddress, Id: operatorSetID}
	return c.keyRegistrar.GetECDSAKeyRegistrationMessageHash(
		&bind.CallOpts{Context: ctx},
		c.operatorAddress,
		operatorSet,
		keyAddress,
	)
}

// GetOperatorBN254KeyRegistrationMessageHash gets the message hash for BN254 key registration
func (c *ContractClient) GetOperatorBN254KeyRegistrationMessageHash(
	ctx context.Context,
	operatorSetID uint32,
	keyData []byte,
) ([32]byte, error) {
	if c.keyRegistrar == nil {
		return [32]byte{}, fmt.Errorf("key registrar not initialized")
	}

	operatorSet := IKeyRegistrar.OperatorSet{Avs: c.avsAddress, Id: operatorSetID}
	return c.keyRegistrar.GetBN254KeyRegistrationMessageHash(
		&bind.CallOpts{Context: ctx},
		c.operatorAddress,
		operatorSet,
		keyData,
	)
}

// RegisterECDSAKey registers an operator's ECDSA signing key with an AVS
func (c *ContractClient) RegisterECDSAKey(
	ctx context.Context,
	operatorSetID uint32,
	keyAddress common.Address,
	signature []byte,
) error {
	if err := c.checkPrivateKey(); err != nil {
		return err
	}

	if c.keyRegistrar == nil {
		return fmt.Errorf("key registrar not initialized")
	}

	opts, err := c.buildTxOpts(ctx)
	if err != nil {
		return fmt.Errorf("failed to build transaction options: %w", err)
	}

	// For ECDSA, keyData is just the address bytes
	keyData := keyAddress.Bytes()
	operatorSet := IKeyRegistrar.OperatorSet{Avs: c.avsAddress, Id: operatorSetID}

	tx, err := c.keyRegistrar.RegisterKey(opts, c.operatorAddress, operatorSet, keyData, signature)
	if err != nil {
		return fmt.Errorf("failed to register ECDSA key: %w", err)
	}

	// Wait for transaction
	receipt, err := bind.WaitMined(ctx, c.ethClient, tx)
	if err != nil {
		return fmt.Errorf("failed to wait for transaction: %w", err)
	}

	if receipt.Status == 0 {
		return fmt.Errorf("transaction reverted")
	}

	c.logger.Info("Successfully registered ECDSA key",
		zap.String("operator", c.operatorAddress.Hex()),
		zap.String("avs", c.avsAddress.Hex()),
		zap.Uint32("operatorSetId", operatorSetID),
		zap.String("keyAddress", keyAddress.Hex()),
		zap.String("txHash", receipt.TxHash.Hex()),
	)

	return nil
}

// RegisterKey registers an operator's signing key with an AVS (generic method for both ECDSA and BN254)
func (c *ContractClient) RegisterKey(
	ctx context.Context,
	operatorSetID uint32,
	keyType string,
	keyData []byte,
	signature []byte,
) error {
	if err := c.checkPrivateKey(); err != nil {
		return err
	}

	switch keyType {
	case "ecdsa":
		// For ECDSA, keyData should be 20 bytes (address)
		if len(keyData) != 20 {
			return fmt.Errorf("invalid ECDSA key data length: expected 20 bytes, got %d", len(keyData))
		}
		keyAddress := common.BytesToAddress(keyData)
		return c.RegisterECDSAKey(ctx, operatorSetID, keyAddress, signature)

	case "bn254":
		// For BN254, register directly with the raw key data
		opts, err := c.buildTxOpts(ctx)
		if err != nil {
			return fmt.Errorf("failed to build transaction options: %w", err)
		}

		operatorSet := IKeyRegistrar.OperatorSet{Avs: c.avsAddress, Id: operatorSetID}
		tx, err := c.keyRegistrar.RegisterKey(opts, c.operatorAddress, operatorSet, keyData, signature)
		if err != nil {
			return fmt.Errorf("failed to register BN254 key: %w", err)
		}

		receipt, err := bind.WaitMined(ctx, c.ethClient, tx)
		if err != nil {
			return fmt.Errorf("failed to wait for transaction: %w", err)
		}

		if receipt.Status == 0 {
			return fmt.Errorf("transaction reverted: %w", err)
		}

		c.logger.Info("Successfully registered BN254 key",
			zap.String("operator", c.operatorAddress.Hex()),
			zap.String("avs", c.avsAddress.Hex()),
			zap.Uint32("operatorSetId", operatorSetID),
			zap.String("txHash", receipt.TxHash.Hex()),
		)
		return nil

	default:
		return fmt.Errorf("unsupported key type: %s", keyType)
	}
}

// GenerateECDSAKeyRegistrationSignature generates an EIP-712 signature for ECDSA key registration
func (c *ContractClient) GenerateECDSAKeyRegistrationSignature(
	ctx context.Context,
	operatorSetID uint32,
	keyAddress common.Address,
) ([32]byte, error) {
	if c.keyRegistrar == nil {
		return [32]byte{}, fmt.Errorf("key registrar not initialized")
	}

	// Create operator set
	operatorSet := IKeyRegistrar.OperatorSet{
		Avs: c.avsAddress,
		Id:  operatorSetID,
	}

	// Get the message hash from the contract
	return c.keyRegistrar.GetECDSAKeyRegistrationMessageHash(
		&bind.CallOpts{Context: ctx},
		c.operatorAddress,
		operatorSet,
		keyAddress,
	)
}

// RegisterBN254Key registers an operator's BN254 signing key with an AVS
func (c *ContractClient) RegisterBN254Key(
	ctx context.Context,
	operatorSetID uint32,
	keyData []byte,
	signature []byte,
) error {
	if err := c.checkPrivateKey(); err != nil {
		return err
	}

	if c.keyRegistrar == nil {
		return fmt.Errorf("key registrar not initialized")
	}

	opts, err := c.buildTxOpts(ctx)
	if err != nil {
		return fmt.Errorf("failed to build transaction options: %w", err)
	}

	operatorSet := IKeyRegistrar.OperatorSet{Avs: c.avsAddress, Id: operatorSetID}

	tx, err := c.keyRegistrar.RegisterKey(opts, c.operatorAddress, operatorSet, keyData, signature)
	if err != nil {
		return fmt.Errorf("failed to register BN254 key: %w", err)
	}

	// Wait for transaction
	receipt, err := bind.WaitMined(ctx, c.ethClient, tx)
	if err != nil {
		return fmt.Errorf("failed to wait for transaction: %w", err)
	}

	if receipt.Status == 0 {
		return fmt.Errorf("transaction reverted")
	}

	c.logger.Info("Successfully registered BN254 key",
		zap.String("operator", c.operatorAddress.Hex()),
		zap.String("avs", c.avsAddress.Hex()),
		zap.Uint32("operatorSetId", operatorSetID),
		zap.String("txHash", receipt.TxHash.Hex()),
	)

	return nil
}

func (c *ContractClient) EncodeBN254KeyData(pubKey *bn254.PublicKey) ([]byte, error) {
	// Convert G1 point
	g1Point := &bn254.G1Point{
		G1Affine: pubKey.GetG1Point(),
	}
	g1Bytes, err := g1Point.ToPrecompileFormat()
	if err != nil {
		return nil, fmt.Errorf("public key not in correct subgroup: %w", err)
	}

	keyRegG1 := IKeyRegistrar.BN254G1Point{
		X: new(big.Int).SetBytes(g1Bytes[0:32]),
		Y: new(big.Int).SetBytes(g1Bytes[32:64]),
	}

	g2Point := bn254.NewZeroG2Point().AddPublicKey(pubKey)
	g2Bytes, err := g2Point.ToPrecompileFormat()
	if err != nil {
		return nil, fmt.Errorf("public key not in correct subgroup: %w", err)
	}
	// Convert to IKeyRegistrar G2 point format
	keyRegG2 := IKeyRegistrar.BN254G2Point{
		X: [2]*big.Int{
			new(big.Int).SetBytes(g2Bytes[0:32]),
			new(big.Int).SetBytes(g2Bytes[32:64]),
		},
		Y: [2]*big.Int{
			new(big.Int).SetBytes(g2Bytes[64:96]),
			new(big.Int).SetBytes(g2Bytes[96:128]),
		},
	}

	return c.keyRegistrar.EncodeBN254KeyData(
		&bind.CallOpts{},
		keyRegG1,
		keyRegG2,
	)
}

// GenerateBN254KeyRegistrationSignature gets the message hash for BN254 key registration
func (c *ContractClient) GenerateBN254KeyRegistrationSignature(
	ctx context.Context,
	operatorSetID uint32,
	keyData []byte,
) ([32]byte, error) {
	if c.keyRegistrar == nil {
		return [32]byte{}, fmt.Errorf("key registrar not initialized")
	}

	// Create operator set
	operatorSet := IKeyRegistrar.OperatorSet{
		Avs: c.avsAddress,
		Id:  operatorSetID,
	}

	// Get the message hash from the contract
	return c.keyRegistrar.GetBN254KeyRegistrationMessageHash(
		&bind.CallOpts{Context: ctx},
		c.operatorAddress,
		operatorSet,
		keyData,
	)
}

// SetAllocationDelay sets the allocation delay for an operator
func (c *ContractClient) SetAllocationDelay(ctx context.Context, delay uint32) error {
	if err := c.checkPrivateKey(); err != nil {
		return err
	}

	if c.allocationManager == nil {
		return fmt.Errorf("allocation manager not initialized")
	}

	opts, err := c.buildTxOpts(ctx)
	if err != nil {
		return err
	}

	// Set allocation delay
	tx, err := c.allocationManager.SetAllocationDelay(opts, c.operatorAddress, delay)
	if err != nil {
		return fmt.Errorf("failed to set allocation delay: %w", err)
	}

	// Wait for transaction
	receipt, err := bind.WaitMined(ctx, c.ethClient, tx)
	if err != nil {
		return fmt.Errorf("failed to wait for transaction: %w", err)
	}

	if receipt.Status == 0 {
		return fmt.Errorf("transaction reverted")
	}

	c.logger.Info("Successfully set allocation delay",
		zap.String("operator", c.operatorAddress.Hex()),
		zap.Uint32("delay", delay),
		zap.String("txHash", receipt.TxHash.Hex()),
	)

	return nil
}

// ModifyAllocations modifies operator allocations to an AVS operator set
func (c *ContractClient) ModifyAllocations(ctx context.Context, operatorSetID uint32, strategyAddress string, magnitude uint64) error {
	if err := c.checkPrivateKey(); err != nil {
		return err
	}

	if c.allocationManager == nil {
		return fmt.Errorf("allocation manager not initialized")
	}

	opts, err := c.buildTxOpts(ctx)
	if err != nil {
		return err
	}

	// Get allocation delay first
	allocationDelay, err := c.allocationManager.GetAllocationDelay(&bind.CallOpts{Context: ctx}, c.operatorAddress)
	if err != nil {
		return fmt.Errorf("failed to get allocation delay: %w", err)
	}

	c.logger.Info("Retrieved allocation delay", zap.Any("allocationDelay", allocationDelay))

	// Create allocation parameters
	allocateParams := []IAllocationManager.IAllocationManagerTypesAllocateParams{
		{
			OperatorSet: IAllocationManager.OperatorSet{
				Avs: c.avsAddress,
				Id:  operatorSetID,
			},
			Strategies:    []common.Address{common.HexToAddress(strategyAddress)},
			NewMagnitudes: []uint64{magnitude},
		},
	}

	// Modify allocations
	tx, err := c.allocationManager.ModifyAllocations(opts, c.operatorAddress, allocateParams)
	if err != nil {
		c.logger.Error("Failed to modify allocation", zap.Error(err))
		return fmt.Errorf("failed to modify allocations: %w", err)
	}

	// Wait for transaction
	receipt, err := bind.WaitMined(ctx, c.ethClient, tx)
	if err != nil {
		return fmt.Errorf("failed to wait for transaction: %w", err)
	}

	if receipt.Status == 0 {
		return fmt.Errorf("transaction reverted")
	}

	c.logger.Info("Successfully modified allocations",
		zap.String("operator", c.operatorAddress.Hex()),
		zap.String("avs", c.avsAddress.Hex()),
		zap.Uint32("operatorSetId", operatorSetID),
		zap.String("txHash", receipt.TxHash.Hex()),
	)

	return nil
}

// GetAVSConfig fetches the operator set configuration from the AVS registrar contract
func (c *ContractClient) GetAVSConfig() (*AVSConfig, error) {
	avsRegistrarAddress, err := c.allocationManager.GetAVSRegistrar(&bind.CallOpts{}, c.avsAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get AVS registrar address: %w", err)
	}

	registrarCaller, err := ITaskAVSRegistrarBase.NewITaskAVSRegistrarBaseCaller(avsRegistrarAddress, c.ethClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create AVS registrar caller: %w", err)
	}

	avsConfig, err := registrarCaller.GetAvsConfig(&bind.CallOpts{})
	if err != nil {
		return nil, err
	}

	return &AVSConfig{
		AggregatorOperatorSetID: avsConfig.AggregatorOperatorSetId,
		ExecutorOperatorSetIDs:  avsConfig.ExecutorOperatorSetIds,
	}, nil
}

// CreateOperatorSets creates operator sets for an AVS
func (c *ContractClient) CreateOperatorSets(
	ctx context.Context,
	avsAddress string,
	operatorSetParams []IAllocationManager.IAllocationManagerTypesCreateSetParams,
) error {
	if err := c.checkPrivateKey(); err != nil {
		return err
	}

	if c.allocationManager == nil {
		return fmt.Errorf("allocation manager not initialized")
	}

	auth, err := c.buildTxOpts(ctx)
	if err != nil {
		return err
	}

	avsAddr := common.HexToAddress(avsAddress)
	tx, err := c.allocationManager.CreateOperatorSets(auth, avsAddr, operatorSetParams)
	if err != nil {
		return fmt.Errorf("failed to create operator sets: %w", err)
	}

	c.logger.Info("Create operator sets transaction sent",
		zap.String("tx", tx.Hash().Hex()),
		zap.String("avs", avsAddress),
		zap.Int("numSets", len(operatorSetParams)),
	)

	return nil
}

// GetOperatorSetMetadataURI gets the metadata URI for an operator set
func (c *ContractClient) GetOperatorSetMetadataURI(ctx context.Context, operatorSetID uint32) (string, error) {
	if c.releaseManager == nil {
		return "", fmt.Errorf("release manager not initialized")
	}

	operatorSet := IReleaseManager.OperatorSet{
		Avs: c.avsAddress,
		Id:  operatorSetID,
	}

	metadataURI, err := c.releaseManager.GetMetadataURI(&bind.CallOpts{Context: ctx}, operatorSet)
	if err != nil {
		return "", fmt.Errorf("failed to get metadata URI: %w", err)
	}

	return metadataURI, nil
}

func (c *ContractClient) Close() {
	c.ethClient.Close()
}

// getERC20 returns an ERC20 bound contract instance
func (c *ContractClient) getERC20(address common.Address) (*bind.BoundContract, error) {
	return contracts.NewERC20Contract(address, c.ethClient)
}

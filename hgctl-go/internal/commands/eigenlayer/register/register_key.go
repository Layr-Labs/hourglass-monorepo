package register

import (
	"encoding/json"
	"fmt"
	"github.com/Layr-Labs/crypto-libs/pkg/ecdsa"
	blskeystore "github.com/Layr-Labs/crypto-libs/pkg/keystore"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/client"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"os"
	"strings"

	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/middleware"
)

// RegisterKeyCommand returns the command for registering system keys
func RegisterKeyCommand() *cli.Command {
	return &cli.Command{
		Name:  "register-key",
		Usage: "Register system signing key with AVS",
		Description: `Register a system signing key with an AVS operator set.

This command registers the system keys configured in the context (ECDSA or BN254).
The operator ECDSA key is used to sign ECDSA registrations.
BN254 registrations use the BN254 key itself for signing.

Prerequisites:
- AVS address must be configured: hgctl context set --avs-address <address>
- Operator set ID must be configured: hgctl context set --operator-set-id <id>
- Operator signer must be configured: hgctl signer operator
- System signer must be configured: hgctl signer system
- For system keystores, SYSTEM_KEYSTORE_PASSWORD environment variable must be set

Usage:
  hgctl register-key`,
		Action: registerKeyAction,
	}
}

// KeyRegistrationParams holds common parameters for key registration
type KeyRegistrationParams struct {
	OperatorSetID uint32
	KeyType       string
	KeyData       []byte
	Signature     []byte
}

// KeyHandler interface for different key types
type KeyHandler interface {
	PrepareKeyData(c *cli.Context, contractClient *client.ContractClient) ([]byte, error)
	GenerateSignature(c *cli.Context, contractClient *client.ContractClient, operatorSetID uint32, keyData []byte) ([]byte, error)
	ValidateParams(c *cli.Context) error
}

// ECDSAKeyHandler handles ECDSA key registration
type ECDSAKeyHandler struct {
	log logger.Logger
	ctx *config.Context
}

// BN254KeyHandler handles BN254 key registration
type BN254KeyHandler struct {
	log        logger.Logger
	ctx        *config.Context
	privateKey *bn254.PrivateKey
}

func registerKeyAction(c *cli.Context) error {
	log := middleware.GetLogger(c)

	// Get contract client from middleware
	contractClient, err := middleware.GetContractClient(c)
	if err != nil {
		return fmt.Errorf("failed to get contract client: %w", err)
	}

	// Parse common parameters
	params, handler, err := parseKeyRegistrationParams(c, log)
	if err != nil {
		return err
	}

	// Validate parameters for specific key type
	if err := handler.ValidateParams(c); err != nil {
		return err
	}

	// Prepare key data
	keyData, err := handler.PrepareKeyData(c, contractClient)
	if err != nil {
		return fmt.Errorf("failed to prepare key data: %w", err)
	}
	params.KeyData = keyData

	// Generate signature if not provided
	if params.Signature == nil {
		signature, err := handler.GenerateSignature(c, contractClient, params.OperatorSetID, keyData)
		if err != nil {
			return fmt.Errorf("failed to generate signature: %w", err)
		}
		params.Signature = signature
	}

	// Register the key
	log.Info("Registering key",
		zap.Uint32("operatorSetId", params.OperatorSetID),
		zap.String("keyType", params.KeyType),
	)

	if err := contractClient.RegisterKey(
		c.Context,
		params.OperatorSetID,
		params.KeyType,
		params.KeyData,
		params.Signature,
	); err != nil {
		log.Error("Failed to register key", zap.Error(err))
		return fmt.Errorf("failed to register key: %w", err)
	}

	log.Info("Successfully registered key")
	return nil
}

// parseKeyRegistrationParams parses common parameters and returns appropriate handler
func parseKeyRegistrationParams(c *cli.Context, log logger.Logger) (*KeyRegistrationParams, KeyHandler, error) {
	// Get context
	currentCtx := c.Context.Value(config.ContextKey).(*config.Context)
	if currentCtx == nil {
		return nil, nil, fmt.Errorf("no context configured. Run: hgctl context use <name>")
	}

	if currentCtx.AVSAddress == "" {
		return nil, nil, fmt.Errorf("AVS address not configured. Run: hgctl context set --avs-address <address>")
	}

	if currentCtx.OperatorKeys == nil {
		return nil, nil, fmt.Errorf("operator signer not configured. Run: hgctl signer operator")
	}

	if currentCtx.SystemSignerKeys == nil {
		return nil, nil, fmt.Errorf("system signer not configured. Run: hgctl signer system")
	}

	// Get operator set ID from context
	operatorSetID := currentCtx.OperatorSetID

	// Detect key type from SystemSignerKeys
	var keyType string
	if currentCtx.SystemSignerKeys.ECDSA != nil {
		keyType = "ecdsa"
	} else if currentCtx.SystemSignerKeys.BN254 != nil {
		keyType = "bn254"
	} else {
		return nil, nil, fmt.Errorf("no system signing key configured. Run: hgctl signer system")
	}

	params := &KeyRegistrationParams{
		OperatorSetID: operatorSetID,
		KeyType:       keyType,
		Signature:     nil, // Will be generated
	}

	// Create appropriate handler with context
	var handler KeyHandler
	switch keyType {
	case "ecdsa":
		handler = &ECDSAKeyHandler{log: log, ctx: currentCtx}
	case "bn254":
		handler = &BN254KeyHandler{log: log, ctx: currentCtx}
	}

	log.Info("Using context configuration",
		zap.Uint32("operatorSetId", operatorSetID),
		zap.String("keyType", keyType),
		zap.String("avsAddress", currentCtx.AVSAddress))

	return params, handler, nil
}

// ECDSAKeyHandler implementation

func (h *ECDSAKeyHandler) ValidateParams(c *cli.Context) error {
	// Validate that system ECDSA is configured
	if h.ctx.SystemSignerKeys == nil || h.ctx.SystemSignerKeys.ECDSA == nil {
		return fmt.Errorf("system ECDSA key not configured. Run: hgctl signer system")
	}
	return nil
}

func (h *ECDSAKeyHandler) PrepareKeyData(c *cli.Context, _ *client.ContractClient) ([]byte, error) {
	// Get system ECDSA key address from context configuration
	systemECDSA := h.ctx.SystemSignerKeys.ECDSA
	var keyAddress string

	if systemECDSA.Keystore != nil {
		// Load keystore to get address
		password := os.Getenv("SYSTEM_KEYSTORE_PASSWORD")
		if password == "" {
			return nil, fmt.Errorf("SYSTEM_KEYSTORE_PASSWORD environment variable required for system ECDSA keystore")
		}

		// Find keystore path
		var keystorePath string
		for _, ks := range h.ctx.Keystores {
			if ks.Name == systemECDSA.Keystore.Name {
				keystorePath = ks.Path
				break
			}
		}

		if keystorePath == "" {
			return nil, fmt.Errorf("system ECDSA keystore '%s' not found in context", systemECDSA.Keystore.Name)
		}

		// Load keystore and extract address
		keystoreBytes, err := os.ReadFile(keystorePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read system ECDSA keystore: %w", err)
		}

		key, err := keystore.DecryptKey(keystoreBytes, password)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt system ECDSA keystore: %w", err)
		}
		keyAddress = key.Address.Hex()

	} else if systemECDSA.PrivateKey {
		// For system ECDSA private key mode, we need the address
		// The address should be provided via SYSTEM_ECDSA_ADDRESS environment variable
		keyAddress = os.Getenv("SYSTEM_ECDSA_ADDRESS")
		if keyAddress == "" {
			return nil, fmt.Errorf("SYSTEM_ECDSA_ADDRESS environment variable required when using system ECDSA private key")
		}

	} else if systemECDSA.RemoteSignerConfig != nil {
		// For remote signer, we need to get the address from the configuration
		// This would typically be stored in the RemoteSignerConfig
		return nil, fmt.Errorf("remote signer not yet supported for system keys")
	} else {
		return nil, fmt.Errorf("system ECDSA key configuration not recognized")
	}

	if keyAddress == "" {
		return nil, fmt.Errorf("could not determine system ECDSA key address")
	}

	h.log.Info("Using system ECDSA key", zap.String("address", keyAddress))
	addr := common.HexToAddress(keyAddress)
	return addr.Bytes(), nil
}

func (h *ECDSAKeyHandler) GenerateSignature(c *cli.Context, contractClient *client.ContractClient, operatorSetID uint32, keyData []byte) ([]byte, error) {
	h.log.Info("Generating signature for ECDSA key registration")

	// Get the private key from context
	currentCtx := c.Context.Value(config.ContextKey).(*config.Context)
	if currentCtx == nil {
		return nil, fmt.Errorf("no context configured")
	}

	privateKeyHex, err := config.GetOperatorPrivateKey(currentCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to get operator private key: %w", err)
	}

	privateKey, err := ecdsa.NewPrivateKeyFromHexString(strings.TrimPrefix(privateKeyHex, "0x"))
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	// Convert keyData back to address
	addr := common.BytesToAddress(keyData)

	// Get the message hash
	messageHash, err := contractClient.GenerateECDSAKeyRegistrationSignature(
		c.Context,
		operatorSetID,
		addr,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get message hash: %w", err)
	}

	rawSig, err := privateKey.Sign(messageHash[:])
	if err != nil {
		return nil, fmt.Errorf("failed to sign message hash: %w", err)
	}
	fmt.Printf("Sig: %+v\n", rawSig)

	signature, err := privateKey.SignAndPack(messageHash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign message hash: %w", err)
	}
	fmt.Printf("Signature: %s\n", hexutil.Encode(signature))

	// verify the signature
	sig, err := ecdsa.NewSignatureFromBytes(signature)
	if err != nil {
		return nil, fmt.Errorf("failed to create signature from bytes: %w", err)
	}
	valid, err := sig.Verify(privateKey.Public(), messageHash)
	if err != nil {
		return nil, fmt.Errorf("failed to verify signature: %w", err)
	}
	if !valid {
		return nil, fmt.Errorf("signature verification failed")
	}
	h.log.Sugar().Infow("Signature verified successfully")
	return signature, nil
}

// BN254KeyHandler implementation

func (h *BN254KeyHandler) ValidateParams(c *cli.Context) error {
	// Validate that system BN254 is configured
	if h.ctx.SystemSignerKeys == nil || h.ctx.SystemSignerKeys.BN254 == nil {
		return fmt.Errorf("system BN254 key not configured. Run: hgctl signer system")
	}
	return nil
}

func (h *BN254KeyHandler) PrepareKeyData(c *cli.Context, contractClient *client.ContractClient) ([]byte, error) {
	// Get system BN254 keystore configuration
	bn254KeyRef := h.ctx.SystemSignerKeys.BN254

	// Find keystore path
	var keystorePath string
	for _, ks := range h.ctx.Keystores {
		if ks.Name == bn254KeyRef.Name {
			keystorePath = ks.Path
			break
		}
	}

	if keystorePath == "" {
		return nil, fmt.Errorf("system BN254 keystore '%s' not found in context", bn254KeyRef.Name)
	}

	// Get password from environment
	password := os.Getenv("SYSTEM_KEYSTORE_PASSWORD")
	if password == "" {
		return nil, fmt.Errorf("SYSTEM_KEYSTORE_PASSWORD environment variable required for system BN254 keystore")
	}

	h.log.Info("Using system BN254 keystore", zap.String("keystore", bn254KeyRef.Name))
	return h.prepareKeyDataFromKeystore(contractClient, keystorePath, password)
}

func (h *BN254KeyHandler) prepareKeyDataFromKeystore(
	contractClient *client.ContractClient,
	keystorePath string,
	password string,
) ([]byte, error) {

	// Read keystore file
	keystoreBytes, err := os.ReadFile(keystorePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read keystore: %w", err)
	}

	// Decrypt keystore
	privateKeyBytes, err := deccryptBn254Keystore(keystoreBytes, password)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt keystore: %w", err)
	}

	// Create BN254 private key
	privateKey, err := bn254.NewPrivateKeyFromBytes(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create private key: %w", err)
	}

	// Store for signature generation
	h.privateKey = privateKey

	keyData, err := contractClient.EncodeBN254KeyData(privateKey.Public())
	if err != nil {
		return nil, fmt.Errorf("failed to encode key: %w", err)
	}

	return keyData, nil
}

func (h *BN254KeyHandler) GenerateSignature(
	c *cli.Context,
	contractClient *client.ContractClient,
	operatorSetID uint32,
	keyData []byte,
) ([]byte, error) {
	if h.privateKey == nil {
		return nil, fmt.Errorf("no private key available for signature generation")
	}

	h.log.Info("Generating BN254 signature")

	// Get message hash from contract
	messageHash, err := contractClient.GenerateBN254KeyRegistrationSignature(
		c.Context,
		operatorSetID,
		keyData,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get message hash: %w", err)
	}

	// Sign using Solidity-compatible method
	sig, err := h.privateKey.SignSolidityCompatible(messageHash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign: %w", err)
	}

	g1Point := &bn254.G1Point{
		G1Affine: sig.GetG1Point(),
	}
	signature, err := g1Point.ToPrecompileFormat()
	if err != nil {
		return nil, fmt.Errorf("signature not in correct subgroup: %w", err)
	}
	return signature, nil
}

// deccryptBn254Keystore decrypts a keystore file to get the private key
func deccryptBn254Keystore(keystoreData []byte, password string) ([]byte, error) {
	// Check if it's a BN254 keystore
	var testKeystore map[string]interface{}
	if err := json.Unmarshal(keystoreData, &testKeystore); err == nil {
		var keystoreFile blskeystore.EIP2335Keystore
		if err := json.Unmarshal(keystoreData, &keystoreFile); err != nil {
			return nil, fmt.Errorf("failed to unmarshal BN254 keystore: %w", err)
		}
		scheme, err := blskeystore.GetSigningSchemeForCurveType("bn254")
		if err != nil {
			return nil, fmt.Errorf("failed to get bn254 scheme: %w", err)
		}
		privateKey, err := keystoreFile.GetPrivateKey(password, scheme)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt BN254 private key: %w", err)
		}

		return privateKey.Bytes(), nil
	}

	// Try standard Ethereum keystore
	key, err := keystore.DecryptKey(keystoreData, password)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt keystore: %w", err)
	}

	return key.PrivateKey.D.Bytes(), nil
}

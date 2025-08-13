package register

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	blskeystore "github.com/Layr-Labs/crypto-libs/pkg/keystore"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/client"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"math/big"
	"os"
	"strings"

	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/middleware"
)

// RegisterKeyCommand returns the command for registering operator keys
func RegisterKeyCommand() *cli.Command {
	return &cli.Command{
		Name:  "register-key",
		Usage: "Register operator signing key with AVS",
		Description: `Register an operator's signing key with an AVS operator set.
This command supports both ECDSA and BN254 key types.

The AVS address and operator address must be configured in the context before running this command.

For ECDSA keys:
  hgctl register-key --operator-set-id 0 --key-type ecdsa --key-address 0x789...

For BN254 keys:
  hgctl register-key --operator-set-id 0 --key-type bn254 --key-data <hex-encoded-key>`,
		Flags: []cli.Flag{
			&cli.Uint64Flag{
				Name:     "operator-set-id",
				Usage:    "Operator set ID to register key for",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "key-type",
				Usage:    "Key type (ecdsa or bn254)",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "key-address",
				Usage: "ECDSA key address (required for ecdsa key type)",
			},
			&cli.StringFlag{
				Name:  "key-data",
				Usage: "BN254 key data in hex format (required for bn254 key type)",
			},
			&cli.StringFlag{
				Name:  "keystore-path",
				Usage: "Path to BN254 keystore file (alternative to --key-data)",
			},
			&cli.StringFlag{
				Name:  "password",
				Usage: "Password for BN254 keystore",
			},
			&cli.StringFlag{
				Name:  "signature",
				Usage: "Pre-signed signature in hex format (optional, will be generated if not provided)",
			},
		},
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
	PrepareKeyData(c *cli.Context) ([]byte, error)
	GenerateSignature(c *cli.Context, contractClient *client.ContractClient, operatorSetID uint32, keyData []byte) ([]byte, error)
	ValidateParams(c *cli.Context) error
}

// ECDSAKeyHandler handles ECDSA key registration
type ECDSAKeyHandler struct {
	log logger.Logger
}

// BN254KeyHandler handles BN254 key registration
type BN254KeyHandler struct {
	log        logger.Logger
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
	keyData, err := handler.PrepareKeyData(c)
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
	operatorSetID := uint32(c.Uint64("operator-set-id"))
	keyType := strings.ToLower(c.String("key-type"))
	signatureHex := c.String("signature")

	// Validate key type
	if keyType != "ecdsa" && keyType != "bn254" {
		return nil, nil, fmt.Errorf("invalid key type: %s (must be 'ecdsa' or 'bn254')", keyType)
	}

	// Parse signature if provided
	var signature []byte
	if signatureHex != "" {
		var err error
		signature, err = hex.DecodeString(strings.TrimPrefix(signatureHex, "0x"))
		if err != nil {
			return nil, nil, fmt.Errorf("invalid signature hex: %w", err)
		}
	}

	params := &KeyRegistrationParams{
		OperatorSetID: operatorSetID,
		KeyType:       keyType,
		Signature:     signature,
	}

	// Create appropriate handler
	var handler KeyHandler
	switch keyType {
	case "ecdsa":
		handler = &ECDSAKeyHandler{log: log}
	case "bn254":
		handler = &BN254KeyHandler{log: log}
	}

	return params, handler, nil
}

// ECDSAKeyHandler implementation

func (h *ECDSAKeyHandler) ValidateParams(c *cli.Context) error {
	keyAddress := c.String("key-address")
	if keyAddress == "" {
		return fmt.Errorf("key-address is required for ECDSA key type")
	}
	return nil
}

func (h *ECDSAKeyHandler) PrepareKeyData(c *cli.Context) ([]byte, error) {
	keyAddress := c.String("key-address")
	addr := common.HexToAddress(keyAddress)
	return addr.Bytes(), nil
}

func (h *ECDSAKeyHandler) GenerateSignature(c *cli.Context, contractClient *client.ContractClient, operatorSetID uint32, keyData []byte) ([]byte, error) {
	h.log.Info("Generating signature for ECDSA key registration")

	// Get the private key from environment
	privateKeyHex := os.Getenv("PRIVATE_KEY")
	if privateKeyHex == "" {
		return nil, fmt.Errorf("PRIVATE_KEY environment variable required to generate signature")
	}

	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(privateKeyHex, "0x"))
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

	// Sign the message hash
	signature, err := crypto.Sign(messageHash[:], privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign message: %w", err)
	}

	// Convert to Ethereum signature format (v = 27 or 28)
	if signature[64] < 27 {
		signature[64] += 27
	}

	h.log.Debug("Generated ECDSA signature", zap.String("signature", hex.EncodeToString(signature)))
	return signature, nil
}

// BN254KeyHandler implementation

func (h *BN254KeyHandler) ValidateParams(c *cli.Context) error {
	keyDataHex := c.String("key-data")
	keystorePath := c.String("keystore-path")

	if keyDataHex == "" && keystorePath == "" {
		return fmt.Errorf("either --key-data or --keystore-path required for BN254")
	}

	if keystorePath != "" && c.String("password") == "" {
		return fmt.Errorf("password required when using keystore")
	}

	return nil
}

func (h *BN254KeyHandler) PrepareKeyData(c *cli.Context) ([]byte, error) {
	keyDataHex := c.String("key-data")
	keystorePath := c.String("keystore-path")

	if keystorePath != "" {
		return h.prepareKeyDataFromKeystore(c, keystorePath)
	}

	// Use provided key data
	keyData, err := hex.DecodeString(strings.TrimPrefix(keyDataHex, "0x"))
	if err != nil {
		return nil, fmt.Errorf("invalid key data hex: %w", err)
	}

	if len(keyData) != 192 {
		return nil, fmt.Errorf("invalid BN254 key data length: expected 192 bytes, got %d", len(keyData))
	}

	return keyData, nil
}

func (h *BN254KeyHandler) prepareKeyDataFromKeystore(c *cli.Context, keystorePath string) ([]byte, error) {
	password := c.String("password")

	// Read keystore file
	keystoreBytes, err := os.ReadFile(keystorePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read keystore: %w", err)
	}

	// Decrypt keystore
	privateKeyBytes, err := decryptKeystore(keystoreBytes, password)
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

	// Generate public key and encode
	publicKey := privateKey.Public()
	keyData := encodeBN254KeyData(publicKey)

	h.log.Debug("Generated BN254 key data from keystore",
		zap.String("keyDataHex", hex.EncodeToString(keyData)))

	return keyData, nil
}

func (h *BN254KeyHandler) GenerateSignature(c *cli.Context, contractClient *client.ContractClient, operatorSetID uint32, keyData []byte) ([]byte, error) {
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

// encodeBN254KeyData encodes the public key for the contract
func encodeBN254KeyData(publicKey *bn254.PublicKey) []byte {
	g1 := publicKey.GetG1Point()
	g2 := publicKey.GetG2Point()

	// Convert points to big.Int format
	g1X := new(big.Int)
	g1Y := new(big.Int)
	g2X0 := new(big.Int)
	g2X1 := new(big.Int)
	g2Y0 := new(big.Int)
	g2Y1 := new(big.Int)

	// Extract coordinates from field elements
	g1.X.BigInt(g1X)
	g1.Y.BigInt(g1Y)
	g2.X.A0.BigInt(g2X0)
	g2.X.A1.BigInt(g2X1)
	g2.Y.A0.BigInt(g2Y0)
	g2.Y.A1.BigInt(g2Y1)

	// Pack as 32-byte padded values (matching Solidity abi.encode)
	data := make([]byte, 192) // 6 * 32 bytes
	g1X.FillBytes(data[0:32])
	g1Y.FillBytes(data[32:64])
	g2X0.FillBytes(data[64:96])
	g2X1.FillBytes(data[96:128])
	g2Y0.FillBytes(data[128:160])
	g2Y1.FillBytes(data[160:192])

	return data
}

// decryptKeystore decrypts a keystore file to get the private key
func decryptKeystore(keystoreData []byte, password string) ([]byte, error) {
	// Check if it's a BLS keystore
	var testKeystore map[string]interface{}
	if err := json.Unmarshal(keystoreData, &testKeystore); err == nil {
		if crypto, ok := testKeystore["crypto"].(map[string]interface{}); ok {
			if _, ok := crypto["kdf"].(map[string]interface{}); ok {
				var keystoreFile blskeystore.EIP2335Keystore
				if err := json.Unmarshal(keystoreData, &keystoreFile); err != nil {
					return nil, fmt.Errorf("failed to unmarshal BLS keystore: %w", err)
				}
				scheme, err := blskeystore.GetSigningSchemeForCurveType("bn254")
				if err != nil {
					return nil, fmt.Errorf("failed to get bn254 scheme: %w", err)
				}
				privateKey, err := keystoreFile.GetPrivateKey(password, scheme)
				if err != nil {
					return nil, fmt.Errorf("failed to decrypt BLS private key: %w", err)
				}

				return privateKey.Bytes(), nil
			}
		}
	}

	// Try standard Ethereum keystore
	key, err := keystore.DecryptKey(keystoreData, password)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt keystore: %w", err)
	}

	return key.PrivateKey.D.Bytes(), nil
}

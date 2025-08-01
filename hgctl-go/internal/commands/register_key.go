package commands

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
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
				Name:  "signature",
				Usage: "Pre-signed signature in hex format (optional, will be generated if not provided)",
			},
		},
		Action: registerKeyAction,
	}
}

func registerKeyAction(c *cli.Context) error {
	log := middleware.GetLogger(c)

	// Get contract client from middleware
	contractClient, err := middleware.GetContractClient(c)
	if err != nil {
		return fmt.Errorf("failed to get contract client: %w", err)
	}

	// Get parameters
	operatorSetID := uint32(c.Uint64("operator-set-id"))
	keyType := strings.ToLower(c.String("key-type"))
	signatureHex := c.String("signature")

	// Validate key type
	if keyType != "ecdsa" && keyType != "bn254" {
		return fmt.Errorf("invalid key type: %s (must be 'ecdsa' or 'bn254')", keyType)
	}

	// Get key data based on type
	var keyData []byte
	if keyType == "ecdsa" {
		keyAddress := c.String("key-address")
		if keyAddress == "" {
			return fmt.Errorf("key-address is required for ECDSA key type")
		}
		// For ECDSA, the key data is the address bytes
		addr := common.HexToAddress(keyAddress)
		keyData = addr.Bytes()
	} else { // bn254
		keyDataHex := c.String("key-data")
		if keyDataHex == "" {
			return fmt.Errorf("key-data is required for BN254 key type")
		}
		keyData, err = hex.DecodeString(strings.TrimPrefix(keyDataHex, "0x"))
		if err != nil {
			return fmt.Errorf("invalid key data hex: %w", err)
		}
	}

	// Parse signature if provided
	var signature []byte
	if signatureHex != "" {
		signature, err = hex.DecodeString(strings.TrimPrefix(signatureHex, "0x"))
		if err != nil {
			return fmt.Errorf("invalid signature hex: %w", err)
		}
	}

	log.Info("Registering key",
		zap.Uint32("operatorSetId", operatorSetID),
		zap.String("keyType", keyType),
	)

	// Register key
	if err := contractClient.RegisterKey(
		c.Context,
		operatorSetID,
		keyType,
		keyData,
		signature,
	); err != nil {
		return fmt.Errorf("failed to register key: %w", err)
	}

	log.Info("Successfully registered key")
	return nil
}

package delegate

import (
	"fmt"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/middleware"
)

// Command returns the command for delegating to an operator
func Command() *cli.Command {
	return &cli.Command{
		Name:  "delegate",
		Usage: "Self-delegate as an operator",
		Description: `Self-delegate to yourself as an operator.
This command allows an operator to delegate to themselves, which is required after registering as an operator.

The operator address must be configured in the context before running this command.`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "operator",
				Usage:    "Operator address to delegate to (defaults to configured operator address for self-delegation)",
				Required: false,
			},
		},
		Action: delegateAction,
	}
}

func delegateAction(c *cli.Context) error {
	log := middleware.GetLogger(c)

	// Get contract client from middleware
	contractClient, err := middleware.GetContractClient(c)
	if err != nil {
		return fmt.Errorf("failed to get contract client: %w", err)
	}

	// Get operator address parameter
	operatorAddress := c.String("operator")

	// If no operator address specified, use the configured operator address for self-delegation
	if operatorAddress == "" {
		privateKeyHex := os.Getenv("OPERATOR_PRIVATE_KEY")
		if privateKeyHex != "" {
			// Derive address from private key
			privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(privateKeyHex, "0x"))
			if err == nil {
				operatorAddress = crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
				log.Info("No operator address specified, using address derived from private key for self-delegation",
					zap.String("address", operatorAddress),
				)
			}
		}

		if operatorAddress == "" {
			return fmt.Errorf("operator address is required (use --operator flag or provide PRIVATE_KEY env var)")
		}
	}

	log.Info("Delegating to operator",
		zap.String("operator", operatorAddress),
	)

	// Delegate to operator (with empty signature for self-delegation)
	if err := contractClient.DelegateTo(c.Context, operatorAddress); err != nil {
		log.Error("Failed to delegate to operator", zap.Error(err))
		return fmt.Errorf("failed to delegate to operator: %w", err)
	}

	log.Info("Successfully delegated to operator")
	return nil
}

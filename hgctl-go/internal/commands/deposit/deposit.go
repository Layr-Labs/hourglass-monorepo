package deposit

import (
	"fmt"
	"math/big"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/middleware"
)

// Command returns the command for depositing into strategies
func Command() *cli.Command {
	return &cli.Command{
		Name:  "deposit",
		Usage: "Deposit tokens into a strategy",
		Description: `Deposit tokens into an EigenLayer strategy.
This command handles the deposit process for stakers to deposit tokens into strategies.

The operator address must be configured in the context before running this command.`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "strategy",
				Usage:    "Strategy contract address",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "amount",
				Usage:    "Amount to deposit (in wei or ether format, e.g., '1000000000000000000' or '1 ether')",
				Required: true,
			},
		},
		Action: depositAction,
	}
}

func depositAction(c *cli.Context) error {
	log := middleware.GetLogger(c)

	// Get parameters
	strategyAddress := c.String("strategy")
	amountStr := c.String("amount")

	// Parse amount
	amount, err := parseAmount(amountStr)
	if err != nil {
		return fmt.Errorf("invalid amount: %w", err)
	}

	// Get contract client from middleware
	contractClient, err := middleware.GetContractClient(c)
	if err != nil {
		return fmt.Errorf("failed to get contract client: %w", err)
	}

	log.Info("Depositing into strategy",
		zap.String("strategy", strategyAddress),
		zap.String("amount", amount.String()),
	)

	// Deposit into strategy
	if err := contractClient.DepositIntoStrategy(c.Context, strategyAddress, amount); err != nil {
		return fmt.Errorf("failed to deposit into strategy: %w", err)
	}

	log.Info("Successfully deposited into strategy")
	return nil
}

// parseAmount parses an amount string into wei
func parseAmount(amount string) (*big.Int, error) {
	// Check if it's in ether format (e.g., "1 ether", "0.5 ether")
	if len(amount) > 6 && amount[len(amount)-6:] == " ether" {
		etherStr := amount[:len(amount)-6]
		ether, ok := new(big.Float).SetString(etherStr)
		if !ok {
			return nil, fmt.Errorf("invalid ether amount: %s", etherStr)
		}
		wei := new(big.Float).Mul(ether, big.NewFloat(1e18))
		result, _ := wei.Int(nil)
		return result, nil
	}

	// Otherwise, assume it's already in wei
	result, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return nil, fmt.Errorf("invalid wei amount: %s", amount)
	}
	return result, nil
}

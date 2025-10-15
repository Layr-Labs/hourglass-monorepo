package rewards

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/middleware"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/output"
)

func SetClaimerCommand() *cli.Command {
	return &cli.Command{
		Name:  "set-claimer",
		Usage: "Set claimer address for an earner",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "earner-address",
				Usage: "Earner address (defaults to operator address from context)",
			},
			&cli.StringFlag{
				Name:     "claimer-address",
				Usage:    "Address that will be authorized to claim rewards",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "rewards-coordinator",
				Usage:    "RewardsCoordinator contract address",
				Required: true,
			},
		},
		Action: setClaimerAction,
	}
}

func setClaimerAction(c *cli.Context) error {
	currentCtx := c.Context.Value(config.ContextKey).(*config.Context)
	log := config.LoggerFromContext(c.Context)

	if currentCtx == nil {
		return fmt.Errorf("no context configured")
	}

	contractClient, err := middleware.GetContractClient(c)
	if err != nil {
		return fmt.Errorf("failed to get contract client: %w", err)
	}

	earnerAddress := c.String("earner-address")
	claimerAddress := c.String("claimer-address")
	rewardsCoordinator := c.String("rewards-coordinator")

	return executeSetClaimer(c.Context, contractClient, log, currentCtx, earnerAddress, claimerAddress, rewardsCoordinator)
}

func executeSetClaimer(ctx context.Context, contractClient interface {
	SetClaimerFor(context.Context, string, string, string) (string, error)
}, log interface{ Info(string, ...zap.Field) }, currentCtx *config.Context, earnerAddress, claimerAddress, rewardsCoordinator string) error {
	if earnerAddress == "" {
		earnerAddress = currentCtx.OperatorAddress
	}
	if earnerAddress == "" {
		return fmt.Errorf("earner address not provided (use --earner-address flag or set operator-address in context)")
	}

	log.Info("Setting claimer for earner",
		zap.String("earner", earnerAddress),
		zap.String("claimer", claimerAddress),
		zap.String("rewardsCoordinator", rewardsCoordinator))

	txHash, err := contractClient.SetClaimerFor(ctx, rewardsCoordinator, earnerAddress, claimerAddress)
	if err != nil {
		return fmt.Errorf("failed to set claimer: %w", err)
	}

	result := map[string]string{
		"earner":             earnerAddress,
		"claimer":            claimerAddress,
		"rewardsCoordinator": rewardsCoordinator,
		"txHash":             txHash,
	}

	log.Info("Successfully set claimer", zap.String("txHash", txHash))

	formatter := output.NewFormatter("json")
	return formatter.PrintJSON(result)
}

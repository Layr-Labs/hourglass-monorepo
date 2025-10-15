package rewards

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/client"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/middleware"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/output"
)

func ClaimCommand() *cli.Command {
	return &cli.Command{
		Name:  "claim",
		Usage: "Claim rewards for an earner address",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "earner-address",
				Usage: "Earner address to claim rewards for (defaults to operator address from context)",
			},
			&cli.StringFlag{
				Name:     "rewards-coordinator",
				Usage:    "RewardsCoordinator contract address",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "sidecar-endpoint",
				Usage: "Rewards sidecar API endpoint",
			},
		},
		Action: claimRewardsAction,
	}
}

func claimRewardsAction(c *cli.Context) error {
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
	if earnerAddress == "" {
		earnerAddress = currentCtx.OperatorAddress
	}
	if earnerAddress == "" {
		return fmt.Errorf("earner address not provided (use --earner-address flag or set operator-address in context)")
	}

	rewardsCoordinator := c.String("rewards-coordinator")

	sidecarURL := c.String("sidecar-url")
	if sidecarURL == "" {
		return fmt.Errorf("sidecar URL not provided (use --sidecar-url flag or set in context)")
	}

	log.Info("Fetching claim proof from rewards endpoint", zap.String("earner", earnerAddress))

	rewardsClient, err := client.NewRewardsClient(sidecarURL, log)
	if err != nil {
		return fmt.Errorf("failed to create rewards client: %w", err)
	}
	defer rewardsClient.Close()

	proof, err := rewardsClient.GetClaimProof(c.Context, earnerAddress)
	if err != nil {
		return fmt.Errorf("failed to get claim proof: %w", err)
	}

	log.Info("Submitting claim to contract",
		zap.String("earner", earnerAddress),
		zap.Int("tokenCount", len(proof.TokenLeaves)),
		zap.String("rewardsCoordinator", rewardsCoordinator))

	txHash, err := contractClient.ProcessClaim(c.Context, rewardsCoordinator, proof)
	if err != nil {
		return fmt.Errorf("failed to process claim: %w", err)
	}

	result := map[string]interface{}{
		"earner":             earnerAddress,
		"tokenCount":         len(proof.TokenLeaves),
		"rewardsCoordinator": rewardsCoordinator,
		"txHash":             txHash,
	}

	log.Info("Successfully processed claim", zap.String("txHash", txHash))

	formatter := output.NewFormatter("json")
	return formatter.PrintJSON(result)
}

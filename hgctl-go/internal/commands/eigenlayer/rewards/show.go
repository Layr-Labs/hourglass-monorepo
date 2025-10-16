package rewards

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/client"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/output"
)

func ShowCommand() *cli.Command {
	return &cli.Command{
		Name:      "show",
		Usage:     "Show rewards for an earner address",
		ArgsUsage: "[earner-address]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "earner-address",
				Usage: "Earner address to query rewards for",
			},
			&cli.StringFlag{
				Name:  "sidecar-url",
				Usage: "Rewards sidecar API URL",
			},
		},
		Action: showRewardsAction,
	}
}

func showRewardsAction(c *cli.Context) error {
	currentCtx := c.Context.Value(config.ContextKey).(*config.Context)
	log := config.LoggerFromContext(c.Context)

	if currentCtx == nil {
		return fmt.Errorf("no context configured")
	}

	earnerAddress := c.String("earner-address")
	if earnerAddress == "" && c.NArg() > 0 {
		earnerAddress = c.Args().Get(0)
	}

	sidecarURL := c.String("sidecar-url")
	if sidecarURL == "" {
		return fmt.Errorf("sidecar URL not provided (use --sidecar-url flag or set in context)")
	}

	rewardsClient, err := client.NewRewardsClient(sidecarURL, log)
	if err != nil {
		return fmt.Errorf("failed to create rewards client: %w", err)
	}
	defer rewardsClient.Close()

	return executeShowRewards(c.Context, rewardsClient, log, currentCtx, earnerAddress)
}

func executeShowRewards(ctx context.Context, rewardsClient interface {
	GetSummarizedRewards(context.Context, string) (*client.RewardsSummary, error)
}, log interface{ Info(string, ...zap.Field) }, currentCtx *config.Context, earnerAddress string) error {
	if earnerAddress == "" {
		earnerAddress = currentCtx.OperatorAddress
	}
	if earnerAddress == "" {
		return fmt.Errorf("earner address not provided (use --earner-address flag, arg, or set operator-address in context)")
	}

	log.Info("Fetching rewards", zap.String("earner", earnerAddress))

	summary, err := rewardsClient.GetSummarizedRewards(ctx, earnerAddress)
	if err != nil {
		return fmt.Errorf("failed to get rewards: %w", err)
	}

	if len(summary.Tokens) == 0 {
		log.Info("No rewards found for earner")
		return nil
	}

	log.Info("Found rewards", zap.Int("tokenCount", len(summary.Tokens)))

	formatter := output.NewFormatter("json")
	return formatter.PrintJSON(summary)
}

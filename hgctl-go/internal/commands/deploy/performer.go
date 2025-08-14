package deploy

import (
	"context"
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/client"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/middleware"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/runtime"
)

func performerCommand() *cli.Command {
	return &cli.Command{
		Name:      "performer",
		Usage:     "Deploy the AVS performer component",
		ArgsUsage: "",
		Description: `Deploy the performer component from a release.

The AVS address must be configured in the context before running this command.`,
		Flags: []cli.Flag{
			&cli.Uint64Flag{
				Name:  "operator-set-id",
				Usage: "Operator set ID",
			},
			&cli.StringFlag{
				Name:  "release-id",
				Usage: "Release ID to deploy (defaults to latest)",
			},
			&cli.StringSliceFlag{
				Name:  "env",
				Usage: "Set environment variables (can be used multiple times)",
			},
			&cli.StringFlag{
				Name:  "env-file",
				Usage: "Load environment variables from file",
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Validate configuration without starting the container",
			},
		},
		Action: deployPerformerAction,
	}
}

// PerformerDeployer handles performer-specific deployment logic
type PerformerDeployer struct {
	*PlatformDeployer
	dryRun bool
}

func deployPerformerAction(c *cli.Context) error {
	// Get context and validate
	currentCtx := c.Context.Value(config.ContextKey).(*config.Context)
	log := logger.FromContext(c.Context)

	if currentCtx == nil {
		return fmt.Errorf("no context configured")
	}

	if currentCtx.AVSAddress == "" {
		return fmt.Errorf("AVS address not configured. Run 'hgctl context set --avs-address <address>' first")
	}

	opSetId := uint32(c.Uint64("operator-set-id"))
	if opSetId == 0 {
		opSetId = currentCtx.OperatorSetID
	}

	if currentCtx.ExecutorAddress == "" {
		return fmt.Errorf("executor address not configured. Run 'hgctl context set --executor-address <address>' first")
	}

	// Get contract client
	contractClient, err := middleware.GetContractClient(c)
	if err != nil {
		return fmt.Errorf("failed to get contract client: %w", err)
	}

	// Create platform deployer
	platform := NewPlatformDeployer(
		currentCtx,
		log,
		contractClient,
		currentCtx.AVSAddress,
		opSetId,
		c.String("release-id"),
		c.String("env-file"),
		c.StringSlice("env"),
	)

	// Create performer deployer
	deployer := &PerformerDeployer{
		PlatformDeployer: platform,
		dryRun:           c.Bool("dry-run"),
	}

	return deployer.Deploy(c.Context)
}

// Deploy executes the performer deployment
func (d *PerformerDeployer) Deploy(ctx context.Context) error {
	// Step 1: Fetch runtime specification
	spec, err := d.FetchRuntimeSpec(ctx)
	if err != nil {
		return err
	}

	// Step 2: Extract performer component
	component, err := d.ExtractComponent(spec, "performer")
	if err != nil {
		return err
	}

	// Step 3: Prepare environment configuration
	cfg := d.PrepareEnvironmentConfig()

	// Step 4: Validate required environment variables from spec
	if err := ValidateComponentSpec(component, cfg.FinalEnvMap); err != nil {
		return err
	}

	// Step 5: Handle dry-run
	if d.dryRun {
		return d.handleDryRun(cfg, component.Registry, component.Digest)
	}

	// Step 6: Deploy via executor
	return d.deployViaExecutor(ctx, component, cfg)
}

// handleDryRun handles the dry-run scenario
func (d *PerformerDeployer) handleDryRun(cfg *DeploymentConfig, registry string, digest string) error {
	d.Log.Info("✅ Dry run successful - performer configuration is valid")

	// Display configuration summary
	d.Log.Info("Configuration:",
		zap.String("avsAddress", cfg.FinalEnvMap["AVS_ADDRESS"]),
		zap.String("executorAddress", d.Context.ExecutorAddress),
		zap.String("performerImage", registry),
		zap.String("performerDigest", digest))

	// Display environment variables that would be passed
	d.Log.Info("Environment variables to be passed:")
	for k, v := range cfg.FinalEnvMap {
		// Don't display sensitive values
		if strings.Contains(strings.ToLower(k), "private") ||
			strings.Contains(strings.ToLower(k), "password") ||
			strings.Contains(strings.ToLower(k), "secret") {
			d.Log.Info(fmt.Sprintf("  %s: <redacted>", k))
		} else {
			d.Log.Info(fmt.Sprintf("  %s: %s", k, v))
		}
	}

	fmt.Println("\n✅ Configuration is valid. Run without --dry-run to deploy.")
	return nil
}

// deployViaExecutor deploys the performer using the executor service
func (d *PerformerDeployer) deployViaExecutor(
	ctx context.Context,
	component *runtime.ComponentSpec,
	cfg *DeploymentConfig,
) error {
	// Create executor client
	executorClient, err := client.NewExecutorClient(d.Context.ExecutorAddress, d.Log)
	if err != nil {
		return fmt.Errorf("failed to create executor client: %w", err)
	}
	defer executorClient.Close()

	d.Log.Info("Deploying performer via executor",
		zap.String("executor", d.Context.ExecutorAddress),
		zap.String("avsAddress", cfg.FinalEnvMap["AVS_ADDRESS"]),
		zap.String("image", component.Registry),
		zap.String("digest", component.Digest),
	)

	// Deploy performer with environment variables
	deploymentID, err := executorClient.DeployPerformerWithEnv(
		ctx,
		cfg.FinalEnvMap["AVS_ADDRESS"],
		component.Digest,
		component.Registry,
		cfg.FinalEnvMap,
	)
	if err != nil {
		return fmt.Errorf("failed to deploy performer: %w", err)
	}

	// Print success message
	d.printSuccessMessage(deploymentID, cfg, component)
	return nil
}

// printSuccessMessage prints a user-friendly success message
func (d *PerformerDeployer) printSuccessMessage(deploymentID string, cfg *DeploymentConfig, component *runtime.ComponentSpec) {
	d.Log.Info("✅ Performer deployed successfully",
		zap.String("deploymentID", deploymentID),
		zap.String("avsAddress", cfg.FinalEnvMap["AVS_ADDRESS"]),
		zap.String("executor", d.Context.ExecutorAddress))

	fmt.Printf("\n✅ Performer deployed successfully\n")
	fmt.Printf("Deployment ID: %s\n", deploymentID)
	fmt.Printf("AVS Address: %s\n", cfg.FinalEnvMap["AVS_ADDRESS"])
	fmt.Printf("Executor: %s\n", d.Context.ExecutorAddress)
	fmt.Printf("Image: %s@%s\n", component.Registry, component.Digest)
	fmt.Printf("\nThe performer is now running on the executor.\n")
}

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
	currentCtx := c.Context.Value(config.ContextKey).(*config.Context)
	log := config.LoggerFromContext(c.Context)

	if currentCtx == nil {
		return fmt.Errorf("no context configured")
	}

	if currentCtx.AVSAddress == "" {
		return fmt.Errorf("AVS address not configured. Run 'hgctl context set --avs-address <address>' first")
	}

	if currentCtx.ExecutorAddress == "" {
		return fmt.Errorf("no executor address found in context. Please deploy an executor first using 'hgctl deploy executor' or set manually with 'hgctl context set --executor-address <address>'")
	}

	contractClient, err := middleware.GetContractClient(c)
	if err != nil {
		return fmt.Errorf("failed to get contract client: %w", err)
	}

	platform := NewPlatformDeployer(
		currentCtx,
		log,
		contractClient,
		currentCtx.AVSAddress,
		currentCtx.OperatorSetID,
		c.String("release-id"),
		c.String("env-file"),
		c.StringSlice("env"),
	)

	deployer := &PerformerDeployer{
		PlatformDeployer: platform,
		dryRun:           c.Bool("dry-run"),
	}

	return deployer.Deploy(c.Context)
}

// Deploy executes the performer deployment
func (d *PerformerDeployer) Deploy(ctx context.Context) error {
	spec, err := d.FetchRuntimeSpec(ctx)
	if err != nil {
		return err
	}

	component, err := d.ExtractComponent(spec, "performer")
	if err != nil {
		return err
	}

	cfg := &DeploymentConfig{
		Env: d.LoadEnvironmentVariables(),
	}

	if err := ValidateComponentSpec(component, cfg.Env); err != nil {
		return err
	}

	if d.dryRun {
		return d.handleDryRun(cfg, component.Registry, component.Digest)
	}

	return d.deployViaExecutor(ctx, component, cfg)
}

// deployViaExecutor deploys the performer using the executor service
func (d *PerformerDeployer) deployViaExecutor(
	ctx context.Context,
	component *runtime.ComponentSpec,
	cfg *DeploymentConfig,
) error {
	executorClient, err := client.NewExecutorClient(d.Context.ExecutorAddress, d.Log)
	if err != nil {
		return fmt.Errorf("failed to create executor client: %w", err)
	}
	defer executorClient.Close()

	d.Log.Info("Deploying performer via executor",
		zap.String("executor", d.Context.ExecutorAddress),
		zap.String("avsAddress", d.Context.AVSAddress),
		zap.String("image", component.Registry),
		zap.String("digest", component.Digest),
	)

	deploymentID, err := executorClient.DeployPerformerWithEnv(
		ctx,
		d.Context.AVSAddress,
		component.Digest,
		component.Registry,
		cfg.Env,
	)
	if err != nil {
		return fmt.Errorf("failed to deploy performer: %w", err)
	}

	d.printSuccessMessage(deploymentID, component)
	return nil
}

// handleDryRun handles the dry-run scenario
func (d *PerformerDeployer) handleDryRun(cfg *DeploymentConfig, registry string, digest string) error {
	d.Log.Info("✅ Dry run successful - performer configuration is valid")

	d.Log.Info("Configuration:",
		zap.String("avsAddress", d.Context.AVSAddress),
		zap.String("executorAddress", d.Context.ExecutorAddress),
		zap.String("performerImage", registry),
		zap.String("performerDigest", digest))

	d.Log.Info("Environment variables to be passed:")
	for k, v := range cfg.Env {
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

// printSuccessMessage prints a user-friendly success message
func (d *PerformerDeployer) printSuccessMessage(deploymentID string, component *runtime.ComponentSpec) {
	d.Log.Info("✅ Performer deployed successfully",
		zap.String("deploymentID", deploymentID),
		zap.String("avsAddress", d.Context.AVSAddress),
		zap.String("executor", d.Context.ExecutorAddress))

	fmt.Printf("\n✅ Performer deployed successfully\n")
	fmt.Printf("Deployment ID: %s\n", deploymentID)
	fmt.Printf("AVS Address: %s\n", d.Context.AVSAddress)
	fmt.Printf("Executor: %s\n", d.Context.ExecutorAddress)
	fmt.Printf("Image: %s@%s\n", component.Registry, component.Digest)
	fmt.Printf("\nThe performer is now running on the executor.\n")
	fmt.Printf("\nUseful commands:\n")
	fmt.Printf("  List performers:  hgctl get performer\n")
	fmt.Printf("  Remove performer: hgctl remove performer %s\n", deploymentID)
}

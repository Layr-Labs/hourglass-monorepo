package deploy

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/client"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
)

func artifactCommand() *cli.Command {
	return &cli.Command{
		Name:      "artifact",
		Usage:     "Deploy an AVS artifact",
		ArgsUsage: "<registry> <digest>",
		Flags:     []cli.Flag{},
		Action:    deployArtifactAction,
	}
}

func deployArtifactAction(c *cli.Context) error {
	if c.NArg() != 2 {
		return cli.ShowSubcommandHelp(c)
	}

	registryURL := c.Args().Get(0)
	digest := c.Args().Get(1)

	// Get context
	currentCtx := c.Context.Value("currentContext").(*config.Context)
	log := logger.FromContext(c.Context)

	if currentCtx == nil {
		return fmt.Errorf("no context configured")
	}

	log.Info("Deploying AVS artifact",
		zap.String("registry", registryURL),
		zap.String("digest", digest))

	// Pull the runtime spec from OCI registry
	log.Info("Pulling runtime spec from OCI registry...")
	ociClient := client.NewOCIClient(log)
	
	// Clean up the digest (remove 0x prefix if present)
	digestClean := strings.TrimPrefix(digest, "0x")
	if !strings.HasPrefix(digestClean, "sha256:") {
		digestClean = "sha256:" + digestClean
	}
	
	spec, err := ociClient.PullRuntimeSpec(c.Context, registryURL, digestClean)
	if err != nil {
		return fmt.Errorf("failed to pull runtime spec: %w", err)
	}

	// Find performer component in spec
	performerComponent, found := spec.Spec["performer"]
	if !found {
		return fmt.Errorf("performer component not found in runtime spec")
	}

	// Container name
	containerName := fmt.Sprintf("hgctl-performer-%s", digest[:12])
	
	// Build docker run command for the performer
	dockerArgs := []string{"run", "-d", "--name", containerName}

	// Add environment variables
	for _, env := range performerComponent.Env {
		dockerArgs = append(dockerArgs, "-e", fmt.Sprintf("%s=%s", env.Name, env.Value))
	}

	// Add ports
	for _, port := range performerComponent.Ports {
		dockerArgs = append(dockerArgs, "-p", fmt.Sprintf("%d:%d", port, port))
	}

	// Add image
	imageRef := fmt.Sprintf("%s@%s", performerComponent.Registry, performerComponent.Digest)
	dockerArgs = append(dockerArgs, imageRef)

	// Add command if specified
	if len(performerComponent.Command) > 0 {
		dockerArgs = append(dockerArgs, performerComponent.Command...)
	}

	// Execute docker command
	log.Info("Starting performer container...", zap.Strings("args", dockerArgs))
	
	cmd := exec.Command("docker", dockerArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start performer: %w", err)
	}

	log.Info("✅ AVS artifact deployed successfully",
		zap.String("container", containerName))

	return nil
}
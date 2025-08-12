package register

import (
	"fmt"
	"net/url"
	"runtime"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/middleware"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
)

// RegisterAVSCommand returns the command for registering an operator with an AVS
func RegisterAVSCommand() *cli.Command {
	return &cli.Command{
		Name:  "register-avs",
		Usage: "Register operator with an AVS",
		Description: `Register an operator with an AVS (Actively Validated Service).

This command registers the operator to the operator set configured in the context.

Prerequisites:
- AVS address must be configured: hgctl context set --avs-address <address>
- Operator set ID must be configured: hgctl context set --operator-set-id <id>
- Operator address must be configured: hgctl context set --operator-address <address>

Example:
  hgctl register-avs --socket https://operator.example.com:8080`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "socket",
				Usage:    "Operator socket endpoint (e.g., 'https://operator.example.com:8080')",
				Required: true,
			},
		},
		Action: registerAVSAction,
	}
}

func registerAVSAction(c *cli.Context) error {
	log := middleware.GetLogger(c)

	// Get context
	currentCtx := c.Context.Value(config.ContextKey).(*config.Context)
	if currentCtx == nil {
		return fmt.Errorf("no context configured. Run: hgctl context use <name>")
	}

	if currentCtx.AVSAddress == "" {
		return fmt.Errorf("AVS address not configured. Run: hgctl context set --avs-address <address>")
	}

	if currentCtx.OperatorAddress == "" {
		return fmt.Errorf("operator address not configured. Run: hgctl context set --operator-address <address>")
	}

	// Get contract client from middleware
	contractClient, err := middleware.GetContractClient(c)
	if err != nil {
		return fmt.Errorf("failed to get contract client: %w", err)
	}

	// Get operator set ID from context (as a single ID, converted to slice for compatibility)
	operatorSetIDs := []uint32{currentCtx.OperatorSetID}
	socket := c.String("socket")

	// Check if running on macOS and socket is localhost, translate to host.docker.internal
	if runtime.GOOS == "darwin" {
		socketToUse := translateLocalhostForDocker(socket, log)
		if socketToUse != socket {
			log.Debug("Translated localhost to host.docker.internal for Docker on macOS",
				zap.String("originalSocket", socket),
				zap.String("translatedSocket", socketToUse),
			)
			socket = socketToUse
		}
	}

	// ABI encode the socket string
	stringType, err := abi.NewType("string", "", nil)
	if err != nil {
		return fmt.Errorf("failed to create ABI type: %w", err)
	}
	socketData, err := abi.Arguments{{Type: stringType}}.Pack(socket)
	if err != nil {
		return fmt.Errorf("failed to ABI encode socket: %w", err)
	}

	log.Info("Registering operator with AVS",
		zap.Uint32("operatorSetId", currentCtx.OperatorSetID),
		zap.String("socket", socket),
		zap.String("avsAddress", currentCtx.AVSAddress),
		zap.String("operatorAddress", currentCtx.OperatorAddress),
	)

	// Register operator to AVS
	if err := contractClient.RegisterOperatorToAVS(c.Context, operatorSetIDs, socketData); err != nil {
		return fmt.Errorf("failed to register operator to AVS: %w", err)
	}

	log.Info("Successfully registered operator with AVS")
	return nil
}

// translateLocalhostForDocker translates localhost URLs to host.docker.internal for Docker on macOS
func translateLocalhostForDocker(socket string, log logger.Logger) string {
	// First check if it's a simple host:port format (not a URL)
	if !strings.Contains(socket, "://") {
		// Simple host:port format - do direct replacement
		if strings.Contains(socket, "localhost") || strings.Contains(socket, "127.0.0.1") {
			replaced := strings.ReplaceAll(socket, "localhost", "host.docker.internal")
			replaced = strings.ReplaceAll(replaced, "127.0.0.1", "host.docker.internal")
			// Also handle 127.x.x.x range
			if strings.HasPrefix(replaced, "127.") {
				parts := strings.SplitN(replaced, ":", 2)
				if len(parts) > 0 && strings.HasPrefix(parts[0], "127.") {
					parts[0] = "host.docker.internal"
					replaced = strings.Join(parts, ":")
				}
			}
			return replaced
		}
		return socket
	}

	// Parse as URL
	u, err := url.Parse(socket)
	if err != nil {
		// If we can't parse it as URL, return as-is
		log.Debug("Could not parse socket as URL, using as-is", zap.String("socket", socket))
		return socket
	}

	// Check if the host is localhost or 127.0.0.1
	if u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1" || strings.HasPrefix(u.Hostname(), "127.") {
		// Replace with host.docker.internal
		u.Host = strings.Replace(u.Host, u.Hostname(), "host.docker.internal", 1)
		return u.String()
	}

	// Return original if not localhost
	return socket
}

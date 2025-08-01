package commands

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/middleware"
)

// RegisterAVSCommand returns the command for registering an operator with an AVS
func RegisterAVSCommand() *cli.Command {
	return &cli.Command{
		Name:  "register-avs",
		Usage: "Register operator with an AVS",
		Description: `Register an operator with an AVS (Actively Validated Service).
This command handles the operator registration to specific operator sets within an AVS.

The operator address and AVS address must be configured in the context before running this command.

To discover available operator sets for an AVS, use:
  hgctl operator-set get

Example:
  hgctl register-avs --operator-set-ids 0,1 --socket https://operator.example.com:8080`,
		Flags: []cli.Flag{
			&cli.Uint64SliceFlag{
				Name:     "operator-set-ids",
				Usage:    "Operator set IDs to register for (can specify multiple)",
				Required: true,
			},
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

	// Get contract client from middleware
	contractClient, err := middleware.GetContractClient(c)
	if err != nil {
		return fmt.Errorf("failed to get contract client: %w", err)
	}

	// Get parameters
	operatorSetIDsUint64 := c.Uint64Slice("operator-set-ids")
	socket := c.String("socket")

	// Convert uint64 slice to uint32
	operatorSetIDs := make([]uint32, len(operatorSetIDsUint64))
	for i, id := range operatorSetIDsUint64 {
		operatorSetIDs[i] = uint32(id)
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
		zap.Any("operatorSetIds", operatorSetIDs),
		zap.String("socket", socket),
	)

	// Register operator to AVS
	if err := contractClient.RegisterOperatorToAVS(c.Context, operatorSetIDs, socketData); err != nil {
		return fmt.Errorf("failed to register operator to AVS: %w", err)
	}

	log.Info("Successfully registered operator with AVS")
	return nil
}

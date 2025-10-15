package appointee

import (
	"fmt"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/middleware"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

func ListPermissionsCommand() *cli.Command {
	return &cli.Command{
		Name:  "list-permissions",
		Usage: "List all permissions granted to an appointee",
		Description: `List all permissions that have been granted to a specific appointee address.

Flags:
--account-address   The account address that granted the permissions (defaults to operator address from context)
--appointee-address The address of the appointee to check permissions for

Usage:
  hgctl eigenlayer user appointee list-permissions --appointee-address 0x5678...
  hgctl eigenlayer user appointee list-permissions --account-address 0x1234... --appointee-address 0x5678...`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "account-address",
				Usage:    "Account address that granted the permissions (defaults to operator address from context)",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "appointee-address",
				Usage:    "Address of the appointee to check permissions for",
				Required: true,
			},
		},
		Action: listPermissionsAction,
	}
}

func listPermissionsAction(c *cli.Context) error {
	log := middleware.GetLogger(c)

	contractClient, err := middleware.GetContractClient(c)
	if err != nil {
		return fmt.Errorf("failed to get contract client: %w", err)
	}

	accountAddressStr := c.String("account-address")
	if accountAddressStr == "" {
		ctx, ok := c.Context.Value(config.ContextKey).(*config.Context)
		if !ok || ctx == nil {
			return fmt.Errorf("context not found")
		}
		if ctx.OperatorAddress == "" {
			return fmt.Errorf("operator address not set in context and --account-address not provided")
		}
		accountAddressStr = ctx.OperatorAddress
		log.Debug("Using operator address from context as account address",
			zap.String("accountAddress", accountAddressStr),
		)
	}

	if !common.IsHexAddress(accountAddressStr) {
		return fmt.Errorf("invalid account address: %s", accountAddressStr)
	}
	accountAddress := common.HexToAddress(accountAddressStr)

	appointeeAddressStr := c.String("appointee-address")
	if !common.IsHexAddress(appointeeAddressStr) {
		return fmt.Errorf("invalid appointee address: %s", appointeeAddressStr)
	}
	appointeeAddress := common.HexToAddress(appointeeAddressStr)

	log.Debug("Listing appointee permissions",
		zap.String("accountAddress", accountAddress.Hex()),
		zap.String("appointeeAddress", appointeeAddress.Hex()),
	)

	targets, selectors, err := contractClient.GetAppointeePermissions(c.Context, accountAddress, appointeeAddress)
	if err != nil {
		log.Error("Failed to get appointee permissions", zap.Error(err))
		return fmt.Errorf("failed to get appointee permissions: %w", err)
	}

	log.Info("Appointee permissions",
		zap.String("appointeeAddress", appointeeAddress.Hex()),
		zap.String("appointedBy", accountAddress.Hex()),
		zap.Int("count", len(targets)),
	)

	if len(targets) == 0 {
		log.Info("No permissions found")
	} else {
		for i := range targets {
			log.Info("Permission",
				zap.Int("index", i+1),
				zap.String("target", targets[i].Hex()),
				zap.String("selector", fmt.Sprintf("0x%x", selectors[i])),
			)
		}
	}

	return nil
}
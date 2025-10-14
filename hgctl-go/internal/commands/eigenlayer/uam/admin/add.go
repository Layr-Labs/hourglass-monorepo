package admin

import (
	"fmt"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/middleware"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

func AddPendingAdminCommand() *cli.Command {
	return &cli.Command{
		Name:  "add",
		Usage: "Add a pending admin for an account",
		Description: `Add a pending admin for an account in the PermissionController.

The transaction sender (operator) will add a pending admin for the specified account.
The pending admin must then accept the role using the accept command.

Prerequisites:
- Operator must have permission to add admins for the account
- Operator signer must be configured with a private key

Flags:
--account-address  The account address to add an admin for (defaults to operator address from context)
--admin-address    The address of the admin being appointed (required)

Usage:
  hgctl eigenlayer uam admin add --admin-address 0x5678...  # Uses operator address from context
  hgctl eigenlayer uam admin add --account-address 0x1234... --admin-address 0x5678...`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "account-address",
				Usage:    "Account address to add admin for (defaults to operator address from context)",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "admin-address",
				Usage:    "Address of the admin being appointed",
				Required: true,
			},
		},
		Action: addPendingAdminAction,
	}
}

func addPendingAdminAction(c *cli.Context) error {
	log := middleware.GetLogger(c)

	contractClient, err := middleware.GetContractClient(c)
	if err != nil {
		return fmt.Errorf("failed to get contract client: %w", err)
	}

	// Get account address - default to operator address from context if not provided
	accountAddressStr := c.String("account-address")
	if accountAddressStr == "" {
		// Get operator address from context
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

	// Get admin address
	adminAddressStr := c.String("admin-address")
	if !common.IsHexAddress(adminAddressStr) {
		return fmt.Errorf("invalid admin address: %s", adminAddressStr)
	}
	adminAddress := common.HexToAddress(adminAddressStr)

	log.Info("Adding pending admin",
		zap.String("accountAddress", accountAddress.Hex()),
		zap.String("adminAddress", adminAddress.Hex()),
	)

	if err := contractClient.AddPendingAdmin(c.Context, accountAddress, adminAddress); err != nil {
		log.Error("Failed to add pending admin", zap.Error(err))
		return fmt.Errorf("failed to add pending admin: %w", err)
	}

	log.Info("Successfully added pending admin",
		zap.String("accountAddress", accountAddress.Hex()),
		zap.String("adminAddress", adminAddress.Hex()),
	)

	return nil
}
package admin

import (
	"fmt"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/middleware"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

func RemoveAdminCommand() *cli.Command {
	return &cli.Command{
		Name:  "remove-admin",
		Usage: "Remove an admin from an account",
		Description: `Remove an admin from an account in the PermissionController.

Flags:
--account-address  The account address to remove admin from (defaults to operator address from context)
--admin-address    The address of the admin to remove (required)

Usage:
  hgctl eigenlayer user admin remove-admin --admin-address 0x5678...  # Uses operator address from context
  hgctl eigenlayer user admin remove-admin --account-address 0x1234... --admin-address 0x5678...`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "account-address",
				Usage:    "Account address to remove admin from (defaults to operator address from context)",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "admin-address",
				Usage:    "Address of the admin to remove",
				Required: true,
			},
		},
		Action: removeAdminAction,
	}
}

func removeAdminAction(c *cli.Context) error {
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

	adminAddressStr := c.String("admin-address")
	if !common.IsHexAddress(adminAddressStr) {
		return fmt.Errorf("invalid admin address: %s", adminAddressStr)
	}
	adminAddress := common.HexToAddress(adminAddressStr)

	if adminAddress == accountAddress {
		log.Warn("Self-removal detected: You are removing your own admin privileges",
			zap.String("address", adminAddress.Hex()),
		)
		fmt.Println("Warning: You are about to remove your own admin privileges.")
	}

	log.Debug("Removing admin",
		zap.String("accountAddress", accountAddress.Hex()),
		zap.String("adminAddress", adminAddress.Hex()),
	)

	err = contractClient.RemoveAdmin(c.Context, accountAddress, adminAddress)
	if err != nil {
		log.Error("Failed to remove admin", zap.Error(err))
		return fmt.Errorf("failed to remove admin: %w", err)
	}

	fmt.Printf("Successfully removed admin %s from account %s\n", adminAddress.Hex(), accountAddress.Hex())

	return nil
}

package admin

import (
	"fmt"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/middleware"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

func IsPendingCommand() *cli.Command {
	return &cli.Command{
		Name:  "is-pending",
		Usage: "Check if an address is a pending admin for an account",
		Description: `Check if an address is a pending admin for an account in the PermissionController.

Flags:
--account-address  The account address to check pending admin status for (defaults to operator address from context)
--admin-address    The address to check if it's a pending admin (required)

Usage:
  hgctl eigenlayer user admin is-pending --admin-address 0x5678...  # Uses operator address from context
  hgctl eigenlayer user admin is-pending --account-address 0x1234... --admin-address 0x5678...`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "account-address",
				Usage:    "Account address to check pending admin status for (defaults to operator address from context)",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "admin-address",
				Usage:    "Address to check if it's a pending admin",
				Required: true,
			},
		},
		Action: isPendingAction,
	}
}

func isPendingAction(c *cli.Context) error {
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

	log.Debug("Checking pending admin status",
		zap.String("accountAddress", accountAddress.Hex()),
		zap.String("adminAddress", adminAddress.Hex()),
	)

	isPending, err := contractClient.IsPendingAdmin(c.Context, accountAddress, adminAddress)
	if err != nil {
		log.Error("Failed to check pending admin status", zap.Error(err))
		return fmt.Errorf("failed to check pending admin status: %w", err)
	}

	if isPending {
		log.Info("Address is a pending admin",
			zap.String("accountAddress", accountAddress.Hex()),
			zap.String("adminAddress", adminAddress.Hex()),
		)
		fmt.Printf("✓ %s is a pending admin for account %s\n", adminAddress.Hex(), accountAddress.Hex())
	} else {
		log.Info("Address is not a pending admin",
			zap.String("accountAddress", accountAddress.Hex()),
			zap.String("adminAddress", adminAddress.Hex()),
		)
		fmt.Printf("✗ %s is not a pending admin for account %s\n", adminAddress.Hex(), accountAddress.Hex())
	}

	return nil
}

package admin

import (
	"fmt"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/middleware"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

func IsAdminCommand() *cli.Command {
	return &cli.Command{
		Name:  "is-admin",
		Usage: "Check if an address is an admin for an account",
		Description: `Check if a user address has admin privileges for an account in the PermissionController.

Flags:
--account-address  The account address to check admin status for (defaults to operator address from context)
--user-address     The address to check if it's an admin (required)

Usage:
  hgctl eigenlayer user admin is-admin --user-address 0x5678...  # Uses operator address from context
  hgctl eigenlayer user admin is-admin --account-address 0x1234... --user-address 0x5678...`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "account-address",
				Usage:    "Account address to check admin status for (defaults to operator address from context)",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "user-address",
				Usage:    "Address to check if it's an admin",
				Required: true,
			},
		},
		Action: isAdminAction,
	}
}

func isAdminAction(c *cli.Context) error {
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

	// Get user address
	userAddressStr := c.String("user-address")
	if !common.IsHexAddress(userAddressStr) {
		return fmt.Errorf("invalid user address: %s", userAddressStr)
	}
	userAddress := common.HexToAddress(userAddressStr)

	log.Debug("Checking admin status",
		zap.String("accountAddress", accountAddress.Hex()),
		zap.String("userAddress", userAddress.Hex()),
	)

	isAdmin, err := contractClient.IsAdmin(c.Context, accountAddress, userAddress)
	if err != nil {
		log.Error("Failed to check admin status", zap.Error(err))
		return fmt.Errorf("failed to check admin status: %w", err)
	}

	if isAdmin {
		log.Info("Address is an admin",
			zap.String("accountAddress", accountAddress.Hex()),
			zap.String("userAddress", userAddress.Hex()),
		)
		fmt.Printf("✓ %s is an admin for account %s\n", userAddress.Hex(), accountAddress.Hex())
	} else {
		log.Info("Address is not an admin",
			zap.String("accountAddress", accountAddress.Hex()),
			zap.String("userAddress", userAddress.Hex()),
		)
		fmt.Printf("✗ %s is not an admin for account %s\n", userAddress.Hex(), accountAddress.Hex())
	}

	return nil
}
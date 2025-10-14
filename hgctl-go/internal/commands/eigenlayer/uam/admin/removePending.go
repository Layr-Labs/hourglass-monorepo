package admin

import (
	"fmt"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/middleware"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

func RemovePendingCommand() *cli.Command {
	return &cli.Command{
		Name:  "remove-pending",
		Usage: "Remove a pending admin from an account",
		Description: `Remove a pending admin from an account in the PermissionController.

Flags:
--account-address  The account address to remove pending admin from (defaults to operator address from context)
--user-address     The address of the pending admin to remove (required)

Usage:
  hgctl eigenlayer user admin remove-pending --user-address 0x5678...  # Uses operator address from context
  hgctl eigenlayer user admin remove-pending --account-address 0x1234... --user-address 0x5678...`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "account-address",
				Usage:    "Account address to remove pending admin from (defaults to operator address from context)",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "user-address",
				Usage:    "Address of the pending admin to remove",
				Required: true,
			},
		},
		Action: removePendingAction,
	}
}

func removePendingAction(c *cli.Context) error {
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

	userAddressStr := c.String("user-address")
	if !common.IsHexAddress(userAddressStr) {
		return fmt.Errorf("invalid user address: %s", userAddressStr)
	}
	userAddress := common.HexToAddress(userAddressStr)

	log.Debug("Removing pending admin",
		zap.String("accountAddress", accountAddress.Hex()),
		zap.String("userAddress", userAddress.Hex()),
	)

	err = contractClient.RemovePendingAdmin(c.Context, accountAddress, userAddress)
	if err != nil {
		log.Error("Failed to remove pending admin", zap.Error(err))
		return fmt.Errorf("failed to remove pending admin: %w", err)
	}

	fmt.Printf("Successfully removed pending admin %s from account %s\n", userAddress.Hex(), accountAddress.Hex())

	return nil
}
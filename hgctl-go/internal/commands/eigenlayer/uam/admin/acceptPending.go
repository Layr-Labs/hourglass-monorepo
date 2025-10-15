package admin

import (
	"fmt"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/middleware"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

func AcceptAdminCommand() *cli.Command {
	return &cli.Command{
		Name:  "accept",
		Usage: "Accept admin privileges for an account",
		Description: `Accept admin privileges for an account in the PermissionController.

The transaction sender (operator) will accept pending admin privileges for the specified account.

Prerequisites:
- Operator must have a pending admin invitation for the account
- Operator signer must be configured with a private key

Flags:
--account-address  The account address whose admin privileges are being accepted (required)

Usage:
  hgctl eigenlayer user accept --account-address 0x1234...`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "account-address",
				Usage:    "Account address whose admin privileges are being accepted",
				Required: true,
			},
		},
		Action: acceptAdminAction,
	}
}

func acceptAdminAction(c *cli.Context) error {
	log := middleware.GetLogger(c)

	contractClient, err := middleware.GetContractClient(c)
	if err != nil {
		return fmt.Errorf("failed to get contract client: %w", err)
	}

	accountAddressStr := c.String("account-address")
	if !common.IsHexAddress(accountAddressStr) {
		return fmt.Errorf("invalid account address: %s", accountAddressStr)
	}
	accountAddress := common.HexToAddress(accountAddressStr)

	log.Info("Accepting admin role",
		zap.String("accountAddress", accountAddress.Hex()),
	)

	if err := contractClient.AcceptAdmin(c.Context, accountAddress); err != nil {
		log.Error("Failed to accept admin role", zap.Error(err))
		return fmt.Errorf("failed to accept admin role: %w", err)
	}

	log.Info("Successfully accepted admin role for account",
		zap.String("accountAddress", accountAddress.Hex()),
	)

	return nil
}

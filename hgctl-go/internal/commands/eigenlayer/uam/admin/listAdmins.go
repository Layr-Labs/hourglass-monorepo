package admin

import (
	"fmt"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/middleware"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

func ListAdminsCommand() *cli.Command {
	return &cli.Command{
		Name:  "list-admins",
		Usage: "List all admins for an account",
		Description: `List all admins for an account in the PermissionController.

Flags:
--account-address  The account address to list admins for (defaults to operator address from context)

Usage:
  hgctl eigenlayer user admin list-admins  # Uses operator address from context
  hgctl eigenlayer user admin list-admins --account-address 0x1234...`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "account-address",
				Usage:    "Account address to list admins for (defaults to operator address from context)",
				Required: false,
			},
		},
		Action: listAdminsAction,
	}
}

func listAdminsAction(c *cli.Context) error {
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

	log.Debug("Fetching admins",
		zap.String("accountAddress", accountAddress.Hex()),
	)

	admins, err := contractClient.GetAdmins(c.Context, accountAddress)
	if err != nil {
		log.Error("Failed to get admins", zap.Error(err))
		return fmt.Errorf("failed to get admins: %w", err)
	}

	if len(admins) == 0 {
		log.Info("Account has NO admins set",
			zap.String("accountAddress", accountAddress.Hex()),
		)
		log.Info("In this state, only the account itself can manage its permissions")
		log.Info("To add an admin, use: hgctl eigenlayer user admin add --admin-address <address>")
		return nil
	}

	log.Info("Account admins",
		zap.String("accountAddress", accountAddress.Hex()),
		zap.Int("count", len(admins)),
	)
	for i, admin := range admins {
		if admin == accountAddress {
			log.Info("Admin (self)",
				zap.Int("index", i+1),
				zap.String("address", admin.Hex()),
			)
		} else {
			log.Info("Admin",
				zap.Int("index", i+1),
				zap.String("address", admin.Hex()),
			)
		}
	}

	return nil
}

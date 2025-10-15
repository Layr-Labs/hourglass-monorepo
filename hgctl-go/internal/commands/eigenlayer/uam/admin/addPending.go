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

Once an admin is added and accepts, the account loses its default
admin privileges. If you want to retain admin access for your account address,
you should first add yourself as an admin before adding others.

The pending admin must accept the role using the 'accept' command before becoming an admin.

Prerequisites:
- If account has NO admins: Only the account itself can add admins
- If account has admins: Only existing admins can add admins
- Operator signer must be configured with a private key

Flags:
--account-address  The account address to add an admin for (defaults to operator address from context)
--admin-address    The address of the admin being appointed (required)

Usage:
  hgctl eigenlayer user admin add --admin-address 0x5678... 
  hgctl eigenlayer user admin add --account-address 0x1234... --admin-address 0x5678...`,
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

	if accountAddress == adminAddress {
		log.Info("Self-appointment detected: Adding yourself as an admin",
			zap.String("address", accountAddress.Hex()),
		)

		admins, err := contractClient.GetAdmins(c.Context, accountAddress)
		if err != nil {
			log.Warn("Could not check existing admins", zap.Error(err))
		} else if len(admins) == 0 {
			log.Info("This is recommended for first-time setup to retain control after adding other admins")
			fmt.Println("✓ Self-appointment detected - this is the recommended first step")
			fmt.Println("  After accepting, you can safely add other admins while retaining control")
		}
	}

	log.Info("Adding pending admin",
		zap.String("accountAddress", accountAddress.Hex()),
		zap.String("adminAddress", adminAddress.Hex()),
	)

	if err := contractClient.AddPendingAdmin(c.Context, accountAddress, adminAddress); err != nil {
		return fmt.Errorf("failed to add pending admin: %w", err)
	}

	log.Info("Successfully added pending admin",
		zap.String("accountAddress", accountAddress.Hex()),
		zap.String("adminAddress", adminAddress.Hex()),
	)

	return nil
}

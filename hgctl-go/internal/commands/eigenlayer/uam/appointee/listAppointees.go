package appointee

import (
	"fmt"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/middleware"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

func ListAppointeesCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List all appointees for a specific function permission",
		Description: `List all appointees that have been granted permission to call a specific function on a target contract.

Flags:
--account-address   The account address that owns the permission (defaults to operator address from context)
--contract-address  The target contract address
--selector          The function selector (e.g., 0x12345678)

Usage:
  hgctl eigenlayer user appointee list --contract-address 0xABCD... --selector 0x12345678
  hgctl eigenlayer user appointee list --account-address 0x1234... --contract-address 0xABCD... --selector 0x12345678`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "account-address",
				Usage:    "Account address that owns the permission (defaults to operator address from context)",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "contract-address",
				Usage:    "Target contract address",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "selector",
				Usage:    "Function selector (e.g., 0x12345678)",
				Required: true,
			},
		},
		Action: listAppointeesAction,
	}
}

func listAppointeesAction(c *cli.Context) error {
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

	contractAddressStr := c.String("contract-address")
	if !common.IsHexAddress(contractAddressStr) {
		return fmt.Errorf("invalid contract address: %s", contractAddressStr)
	}
	contractAddress := common.HexToAddress(contractAddressStr)

	selectorStr := c.String("selector")
	selector, err := parseSelector(selectorStr)
	if err != nil {
		return fmt.Errorf("invalid selector: %w", err)
	}

	log.Debug("Listing appointees",
		zap.String("accountAddress", accountAddress.Hex()),
		zap.String("contractAddress", contractAddress.Hex()),
		zap.String("selector", selectorStr),
	)

	appointees, err := contractClient.GetAppointees(c.Context, accountAddress, contractAddress, selector)
	if err != nil {
		log.Error("Failed to get appointees", zap.Error(err))
		return fmt.Errorf("failed to get appointees: %w", err)
	}

	log.Info("Appointees for contract and selector",
		zap.String("contractAddress", contractAddress.Hex()),
		zap.String("selector", selectorStr),
		zap.String("appointer", accountAddress.Hex()),
		zap.Int("count", len(appointees)),
	)

	if len(appointees) == 0 {
		log.Info("No appointees found")
	} else {
		for i, appointee := range appointees {
			log.Info("Appointee",
				zap.Int("index", i+1),
				zap.String("address", appointee.Hex()),
			)
		}
	}

	return nil
}

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
		Name:  "list-appointees",
		Usage: "List all appointees for a specific function permission",
		Description: `List all appointees that have been granted permission to call a specific function on a target contract.

Flags:
--account-address  The account address that owns the permission (defaults to operator address from context)
--target           The target contract address
--selector         The function selector (e.g., 0x12345678)

Usage:
  hgctl eigenlayer user appointee list-appointees --target 0xABCD... --selector 0x12345678
  hgctl eigenlayer user appointee list-appointees --account-address 0x1234... --target 0xABCD... --selector 0x12345678`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "account-address",
				Usage:    "Account address that owns the permission (defaults to operator address from context)",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "target",
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

	targetAddressStr := c.String("target")
	if !common.IsHexAddress(targetAddressStr) {
		return fmt.Errorf("invalid target address: %s", targetAddressStr)
	}
	targetAddress := common.HexToAddress(targetAddressStr)

	selectorStr := c.String("selector")
	selector, err := parseSelector(selectorStr)
	if err != nil {
		return fmt.Errorf("invalid selector: %w", err)
	}

	log.Debug("Listing appointees",
		zap.String("accountAddress", accountAddress.Hex()),
		zap.String("target", targetAddress.Hex()),
		zap.String("selector", selectorStr),
	)

	appointees, err := contractClient.GetAppointees(c.Context, accountAddress, targetAddress, selector)
	if err != nil {
		log.Error("Failed to get appointees", zap.Error(err))
		return fmt.Errorf("failed to get appointees: %w", err)
	}

	// Output format matching eigenlayer-cli
	fmt.Printf("Target, Selector and Appointer: %s, %s, %s\n",
		targetAddress.Hex(),
		selectorStr,
		accountAddress.Hex(),
	)
	fmt.Println("============================================================")

	if len(appointees) == 0 {
		fmt.Println("No appointees found")
	} else {
		for _, appointee := range appointees {
			fmt.Println(appointee.Hex())
		}
	}

	return nil
}
package appointee

import (
	"fmt"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/middleware"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

func SetAppointeeCommand() *cli.Command {
	return &cli.Command{
		Name:  "set",
		Usage: "Grant permission to an appointee to call a specific function",
		Description: `Grant permission for an appointee to call a specific function on a target contract.

Flags:
--account-address   The account address that grants the permission (defaults to operator address from context)
--appointee-address The address of the appointee to grant permission to
--contract-address  The target contract address
--selector          The function selector (e.g., 0x12345678)

Usage:
  hgctl eigenlayer user appointee set --appointee-address 0x5678... --contract-address 0xABCD... --selector 0x12345678
  hgctl eigenlayer user appointee set --account-address 0x1234... --appointee-address 0x5678... --contract-address 0xABCD... --selector 0x12345678`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "account-address",
				Usage:    "Account address that grants the permission (defaults to operator address from context)",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "appointee-address",
				Usage:    "Address of the appointee to grant permission to",
				Required: true,
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
		Action: setAppointeeAction,
	}
}

func setAppointeeAction(c *cli.Context) error {
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

	appointeeAddressStr := c.String("appointee-address")
	if !common.IsHexAddress(appointeeAddressStr) {
		return fmt.Errorf("invalid appointee address: %s", appointeeAddressStr)
	}
	appointeeAddress := common.HexToAddress(appointeeAddressStr)

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

	log.Debug("Setting appointee permission",
		zap.String("accountAddress", accountAddress.Hex()),
		zap.String("appointeeAddress", appointeeAddress.Hex()),
		zap.String("contractAddress", contractAddress.Hex()),
		zap.String("selector", selectorStr),
	)

	err = contractClient.SetAppointee(c.Context, accountAddress, appointeeAddress, contractAddress, selector)
	if err != nil {
		log.Error("Failed to set appointee", zap.Error(err))
		return fmt.Errorf("failed to set appointee: %w", err)
	}

	fmt.Printf("Successfully granted permission to appointee %s for account %s\n", appointeeAddress.Hex(), accountAddress.Hex())
	fmt.Printf("Contract: %s, Selector: %s\n", contractAddress.Hex(), selectorStr)

	return nil
}

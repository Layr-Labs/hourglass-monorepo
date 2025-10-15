package appointee

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/commands/middleware"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

func CanCallCommand() *cli.Command {
	return &cli.Command{
		Name:  "can-call",
		Usage: "Check if an appointee can call a specific function",
		Description: `Check if an appointee has permission to call a specific function on a target contract.

Flags:
--account-address   The account address that owns the permission (defaults to operator address from context)
--appointee-address The address of the appointee to check
--target            The target contract address
--selector          The function selector (e.g., 0x12345678)

Usage:
  hgctl eigenlayer user appointee can-call --appointee-address 0x5678... --target 0xABCD... --selector 0x12345678
  hgctl eigenlayer user appointee can-call --account-address 0x1234... --appointee-address 0x5678... --target 0xABCD... --selector 0x12345678`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "account-address",
				Usage:    "Account address that owns the permission (defaults to operator address from context)",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "appointee-address",
				Usage:    "Address of the appointee to check",
				Required: true,
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
		Action: canCallAction,
	}
}

func canCallAction(c *cli.Context) error {
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

	log.Debug("Checking canCall permission",
		zap.String("accountAddress", accountAddress.Hex()),
		zap.String("appointeeAddress", appointeeAddress.Hex()),
		zap.String("target", targetAddress.Hex()),
		zap.String("selector", selectorStr),
	)

	canCall, err := contractClient.CanCall(c.Context, accountAddress, appointeeAddress, targetAddress, selector)
	if err != nil {
		log.Error("Failed to check canCall", zap.Error(err))
		return fmt.Errorf("failed to check canCall: %w", err)
	}

	fmt.Printf("CanCall Result: %t\n", canCall)
	fmt.Printf("Target, Selector and Appointee: %s, %s, %s\n",
		targetAddress.Hex(),
		selectorStr,
		appointeeAddress.Hex(),
	)

	return nil
}

// parseSelector converts a selector string (e.g., "0x12345678") to a [4]byte array
func parseSelector(selectorStr string) ([4]byte, error) {
	var selector [4]byte

	// Remove 0x prefix if present
	selectorStr = strings.TrimPrefix(selectorStr, "0x")

	// Validate length
	if len(selectorStr) != 8 {
		return selector, fmt.Errorf("selector must be 8 hex characters (4 bytes), got %d characters", len(selectorStr))
	}

	// Decode hex string
	bytes, err := hex.DecodeString(selectorStr)
	if err != nil {
		return selector, fmt.Errorf("failed to decode hex string: %w", err)
	}

	// Copy to fixed-size array
	copy(selector[:], bytes)

	return selector, nil
}
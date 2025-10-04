package tools

import (
	"context"
	"crypto/ecdsa"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
	"go.uber.org/zap"
)

type Receipt struct {
	Status          hexutil.Uint64 `json:"status"`
	TransactionHash common.Hash    `json:"transactionHash"`
}

func TransferOwnership(logger *zap.Logger, rpcURL string, proxy common.Address, privateKey string) {
	ctx := context.Background()
	c, err := rpc.DialContext(ctx, rpcURL)
	if err != nil {
		logger.Error("failed to connect to rpc", zap.Error(err))
	}

	// Your private key - used only to derive the new owner address
	priv := mustKey(logger, privateKey)
	newOwner := crypto.PubkeyToAddress(priv.PublicKey)

	// ABI with owner() and transferOwnership(address)
	ownableABI := MustABI(logger, `[
	  {"inputs":[],"name":"owner","outputs":[{"type":"address"}],"stateMutability":"view","type":"function"},
	  {"inputs":[{"name":"newOwner","type":"address"}],"name":"transferOwnership","outputs":[],"stateMutability":"nonpayable","type":"function"}
	]`)

	// read current owner
	currOwner := readOwner(ctx, logger, c, ownableABI, proxy)
	logger.Info("Current owner:", zap.String("owner", currOwner.Hex()))

	// impersonate the current owner and fund it
	impersonate(ctx, logger, c, currOwner)
	defer stopImpersonate(ctx, c, currOwner)

	// pack transferOwnership(newOwner)
	calldata, err := ownableABI.Pack("transferOwnership", newOwner)
	if err != nil {
		logger.Error("failed to pack callData %w", zap.Error(err))
	}

	// send tx via eth_sendTransaction from the impersonated owner to the proxy
	tx := map[string]any{
		"from":  currOwner.Hex(),
		"to":    proxy.Hex(),
		"data":  hexutil.Encode(calldata),
		"value": "0x0",
	}
	var txHash common.Hash
	if err := c.CallContext(ctx, &txHash, "eth_sendTransaction", tx); err != nil {
		logger.Error("failed to send tx: %w", zap.Error(err))
	}

	// await for tx receipt
	MustWaitReceipt(ctx, logger, c, txHash)
	logger.Info("TransferOwnership tx:", zap.String("owner", txHash.Hex()))

	// verify
	newOwnerRead := readOwner(ctx, logger, c, ownableABI, proxy)
	logger.Info("New owner", zap.String("owner", newOwnerRead.Hex()))
}

func mustKey(logger *zap.Logger, hex string) *ecdsa.PrivateKey {
	if strings.HasPrefix(hex, "0x") || strings.HasPrefix(hex, "0X") {
		hex = hex[2:]
	}
	k, err := crypto.HexToECDSA(hex)
	if err != nil {
		logger.Error("invalid key: %w", zap.Error(err))
	}
	return k
}

func readOwner(ctx context.Context, logger *zap.Logger, c *rpc.Client, ab abi.ABI, proxy common.Address) common.Address {
	data, _ := ab.Pack("owner")
	call := map[string]any{"to": proxy.Hex(), "data": hexutil.Encode(data)}
	var out string
	if err := c.CallContext(ctx, &out, "eth_call", call, "latest"); err != nil {
		logger.Error("failed to call contract: %w", zap.Error(err))
	}
	b := common.FromHex(out)
	return common.BytesToAddress(b[len(b)-20:])
}

func impersonate(ctx context.Context, logger *zap.Logger, c *rpc.Client, who common.Address) {
	var ok bool
	if err := c.CallContext(ctx, &ok, "anvil_impersonateAccount", who.Hex()); err != nil {
		logger.Error("failed to impersonate: %w", zap.Error(err))
	}
	// fund so it can pay gas
	_ = c.CallContext(ctx, &ok, "anvil_setBalance", who.Hex(), "0x56BC75E2D63100000") // 100 ETH
}

func stopImpersonate(ctx context.Context, c *rpc.Client, who common.Address) {
	var ok bool
	_ = c.CallContext(ctx, &ok, "anvil_stopImpersonatingAccount", who.Hex())
}

func MustABI(logger *zap.Logger, s string) abi.ABI {
	a, err := abi.JSON(strings.NewReader(s))
	if err != nil {
		logger.Error("invalid abi: %w", zap.Error(err))
	}
	return a
}

func MustWaitReceipt(ctx context.Context, logger *zap.Logger, c *rpc.Client, h common.Hash) {
	var r Receipt
	for {
		_ = c.CallContext(ctx, &r, "eth_getTransactionReceipt", h)
		if r.TransactionHash != (common.Hash{}) {
			break
		}
		time.Sleep(150 * time.Millisecond)
	}
	if r.Status != 1 {
		// Get reason
		var trace map[string]any
		_ = c.CallContext(ctx, &trace, "debug_traceTransaction", h.Hex(), map[string]any{"disableStack": true})
		logger.Error("tx reverted. trace")
	}
}

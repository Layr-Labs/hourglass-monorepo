package client

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/signer"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"strings"
)

// Web3Signer implements the ISigner interface using a remote Web3Signer service.
// This implementation supports ECDSA signing only - BN254 operations are not supported
// by the Web3Signer protocol.
type Web3Signer struct {
	client      *Web3SignerClient
	fromAddress common.Address
	publicKey   string
	curveType   signer.CurveType
	logger      logger.Logger
}

// NewWeb3Signer creates a new Web3Signer that implements the ISigner interface.
// It only supports ECDSA curve type - attempting to use BN254 will result in errors.
// The publicKey parameter should be the hex-encoded public key (with or without 0x prefix)
// that corresponds to the fromAddress.
func NewWeb3Signer(
	client *Web3SignerClient,
	fromAddress common.Address,
	publicKey string,
	curveType signer.CurveType,
	logger logger.Logger,
) (signer.ISigner, error) {

	if curveType != signer.CurveTypeECDSA {
		return nil, fmt.Errorf("web3signer only supports ECDSA curve type, got %s", curveType)
	}

	if client == nil {
		return nil, fmt.Errorf("web3signer client cannot be nil")
	}

	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	if publicKey == "" {
		return nil, fmt.Errorf("publicKey cannot be empty")
	}

	// Clean up public key format - remove 0x prefix if present
	cleanPublicKey := strings.TrimPrefix(publicKey, "0x")

	logger.Sugar().Debug("Creating new Web3Signer",
		"fromAddress", fromAddress.Hex(),
		"publicKey", cleanPublicKey,
		"curveType", curveType,
	)

	return &Web3Signer{
		client:      client,
		fromAddress: fromAddress,
		publicKey:   cleanPublicKey,
		curveType:   curveType,
		logger:      logger,
	}, nil
}

// SignMessage signs arbitrary data using the Web3Signer REST API for raw ECDSA signing.
// This method uses the /api/v1/eth1/sign/{identifier} endpoint to perform generic
// ECDSA signing without Ethereum message prefixes, making it compatible with crypto-libs.
func (w3s *Web3Signer) SignMessage(data []byte) ([]byte, error) {
	if w3s.curveType != signer.CurveTypeECDSA {
		return nil, fmt.Errorf("web3signer only supports ECDSA curve type")
	}

	w3s.logger.Sugar().Debugw("Signing message with Web3Signer REST API",
		"fromAddress", w3s.fromAddress.Hex(),
		"dataLength", len(data),
	)

	// Use the REST API for raw ECDSA signing (compatible with crypto-libs)
	ctx := context.Background()
	sigHex, err := w3s.client.SignRaw(ctx, w3s.publicKey, data)
	if err != nil {
		w3s.logger.Sugar().Errorw("Failed to sign message with Web3Signer REST API",
			"error", err,
			"fromAddress", w3s.fromAddress.Hex(),
		)
		return nil, fmt.Errorf("failed to sign message with web3signer: %w", err)
	}

	// Convert hex signature back to bytes
	sigBytes, err := hexutil.Decode(sigHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode signature from web3signer: %w", err)
	}

	w3s.logger.Sugar().Debugw("Successfully signed message with Web3Signer REST API",
		"fromAddress", w3s.fromAddress.Hex(),
		"signatureLength", len(sigBytes),
		"signature", sigHex,
	)

	return sigBytes, nil
}

// SignMessageForSolidity signs data in a format compatible with Solidity verification.
// For ECDSA, this uses the same signing method as SignMessage since Web3Signer
// produces Ethereum-standard ECDSA signatures that are Solidity-compatible.
func (w3s *Web3Signer) SignMessageForSolidity(data []byte) ([]byte, error) {
	if w3s.curveType != signer.CurveTypeECDSA {
		return nil, fmt.Errorf("web3signer only supports ECDSA curve type")
	}

	w3s.logger.Sugar().Debugw("Signing message for Solidity with Web3Signer",
		"fromAddress", w3s.fromAddress.Hex(),
		"dataHash", "0x"+hex.EncodeToString(data[:]),
	)

	// For ECDSA, Web3Signer's eth_sign produces Solidity-compatible signatures
	// Convert [32]byte to []byte and use the same signing method
	return w3s.SignMessage(data[:])
}

// GetFromAddress returns the address used for signing operations.
// This is a convenience method for callers who need to know the signing address.
func (w3s *Web3Signer) GetFromAddress() common.Address {
	return w3s.fromAddress
}

// GetCurveType returns the curve type supported by this signer.
// This will always be CurveTypeECDSA for Web3Signer implementations.
func (w3s *Web3Signer) GetCurveType() signer.CurveType {
	return w3s.curveType
}

// SupportsRemoteSigning returns true since this is a remote signing implementation.
// This can be useful for callers who need to distinguish between local and remote signers.
func (w3s *Web3Signer) SupportsRemoteSigning() bool {
	return true
}

// Validate checks if the Web3Signer is properly configured and can communicate with the service.
// This performs a basic connectivity test by trying to list available accounts.
func (w3s *Web3Signer) Validate() error {
	ctx := context.Background()
	accounts, err := w3s.client.EthAccounts(ctx)
	if err != nil {
		return fmt.Errorf("failed to validate web3signer connection: %w", err)
	}

	// Check if our fromAddress is available in the Web3Signer service
	fromAddressLower := strings.ToLower(w3s.fromAddress.Hex())
	for _, account := range accounts {
		if strings.ToLower(account) == fromAddressLower {
			w3s.logger.Sugar().Debugw("Web3Signer validation successful",
				"fromAddress", w3s.fromAddress.Hex(),
				"availableAccounts", len(accounts),
			)
			return nil
		}
	}

	return fmt.Errorf("signing address %s not found in web3signer accounts: %v",
		w3s.fromAddress.Hex(), accounts)
}

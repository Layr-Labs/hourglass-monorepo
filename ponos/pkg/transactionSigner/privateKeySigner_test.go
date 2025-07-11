package transactionSigner

import (
	"testing"
)

func TestNewPrivateKeySigner(t *testing.T) {
	// Note: This test will fail because we need a real ethClient to get chainID
	// For now, we'll skip this test until we have a mock ethClient
	t.Skip("Skipping test that requires real ethClient for chainID")

	// This is what the test would look like with a real ethClient:
	// Generate a test private key
	// privateKey, err := crypto.GenerateKey()
	// require.NoError(t, err)
	// signer, err := NewPrivateKeySigner(privateKeyHex, ethClient, logger)
	// require.NoError(t, err)
	// assert.NotNil(t, signer)
	// assert.Equal(t, crypto.PubkeyToAddress(privateKey.PublicKey), signer.GetFromAddress())

	// Test invalid private key
	// _, err = NewPrivateKeySigner("invalid-key", ethClient, logger)
	// assert.Error(t, err)
	// assert.Contains(t, err.Error(), "failed to parse private key")
}

func TestPrivateKeySigner_GetTransactOpts(t *testing.T) {
	t.Skip("Skipping test that requires real ethClient for chainID")
}

func TestPrivateKeySigner_GetFromAddress(t *testing.T) {
	t.Skip("Skipping test that requires real ethClient for chainID")
}

func TestPrivateKeySigner_EstimateGasPriceAndLimit(t *testing.T) {
	t.Skip("Skipping test that requires real ethClient for chainID")
}

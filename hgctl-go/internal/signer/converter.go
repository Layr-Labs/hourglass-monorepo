package signer

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/accounts/keystore"
)

// ConvertKeystoreToPrivateKey converts an ECDSA keystore to a hex-encoded private key
func ConvertKeystoreToPrivateKey(keystorePath, password string) (string, error) {
	// Clean and validate the path
	keystorePath = filepath.Clean(keystorePath)
	
	// Read the keystore file
	keyStoreContents, err := os.ReadFile(keystorePath)
	if err != nil {
		return "", fmt.Errorf("failed to read keystore at %s: %w", keystorePath, err)
	}

	// Decrypt the keystore
	key, err := keystore.DecryptKey(keyStoreContents, password)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt keystore: %w", err)
	}

	// Convert to hex string with 0x prefix
	privateKeyHex := hex.EncodeToString(key.PrivateKey.D.Bytes())
	return "0x" + privateKeyHex, nil
}
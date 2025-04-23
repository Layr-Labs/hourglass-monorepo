package bn254

import (
	"os"
	"path/filepath"
	"testing"
)

func TestKeystore(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "bn254-keystore-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("SaveAndLoadKeystore", func(t *testing.T) {
		// Generate a key pair
		privateKey, publicKey, err := GenerateKeyPair()
		if err != nil {
			t.Fatalf("Failed to generate key pair: %v", err)
		}

		// Create keystore path
		keystorePath := filepath.Join(tempDir, "keystore.json")
		password := "test-password"

		// Save the private key to keystore
		err = SaveToKeystore(privateKey, keystorePath, password, nil)
		if err != nil {
			t.Fatalf("Failed to save private key to keystore: %v", err)
		}

		// Verify the file exists
		if _, err := os.Stat(keystorePath); os.IsNotExist(err) {
			t.Fatalf("Keystore file was not created: %v", err)
		}

		// Load the private key from keystore
		loadedPrivateKey, err := LoadFromKeystore(keystorePath, password)
		if err != nil {
			t.Fatalf("Failed to load private key from keystore: %v", err)
		}

		// Verify the loaded private key is valid and matches the original
		loadedPublicKey := loadedPrivateKey.Public()

		// Compare public keys as a way to verify private keys match
		if string(loadedPublicKey.Bytes()) != string(publicKey.Bytes()) {
			t.Errorf("Loaded public key doesn't match original public key")
		}

		// Test signing with loaded key
		message := []byte("test message for keystore")
		signature, err := loadedPrivateKey.Sign(message)
		if err != nil {
			t.Fatalf("Failed to sign with loaded private key: %v", err)
		}

		// Verify signature with original public key
		valid, err := signature.Verify(publicKey, message)
		if err != nil {
			t.Fatalf("Failed to verify signature from loaded private key: %v", err)
		}
		if !valid {
			t.Error("Signature verification with loaded private key failed")
		}
	})

	t.Run("InvalidPassword", func(t *testing.T) {
		// Generate a key pair
		privateKey, _, err := GenerateKeyPair()
		if err != nil {
			t.Fatalf("Failed to generate key pair: %v", err)
		}

		// Create keystore path
		keystorePath := filepath.Join(tempDir, "keystore-invalid-password.json")
		password := "correct-password"
		wrongPassword := "wrong-password"

		// Save the private key to keystore
		err = SaveToKeystore(privateKey, keystorePath, password, nil)
		if err != nil {
			t.Fatalf("Failed to save private key to keystore: %v", err)
		}

		// Try to load with wrong password
		_, err = LoadFromKeystore(keystorePath, wrongPassword)
		if err == nil {
			t.Error("Expected error when loading with wrong password, but got nil")
		}
	})

	t.Run("LightKeystoreOptions", func(t *testing.T) {
		// Generate a key pair
		privateKey, publicKey, err := GenerateKeyPair()
		if err != nil {
			t.Fatalf("Failed to generate key pair: %v", err)
		}

		// Create keystore path
		keystorePath := filepath.Join(tempDir, "keystore-light.json")
		password := "test-password"

		// Save the private key with light options (faster but less secure)
		err = SaveToKeystore(privateKey, keystorePath, password, LightKeystoreOptions())
		if err != nil {
			t.Fatalf("Failed to save private key to keystore with light options: %v", err)
		}

		// Load the private key from keystore
		loadedPrivateKey, err := LoadFromKeystore(keystorePath, password)
		if err != nil {
			t.Fatalf("Failed to load private key from keystore: %v", err)
		}

		// Verify the loaded private key is valid
		loadedPublicKey := loadedPrivateKey.Public()

		// Compare public keys
		if string(loadedPublicKey.Bytes()) != string(publicKey.Bytes()) {
			t.Errorf("Loaded public key doesn't match original public key")
		}
	})

	t.Run("GenerateRandomPassword", func(t *testing.T) {
		// Test default length (16)
		password, err := GenerateRandomPassword(0)
		if err != nil {
			t.Fatalf("Failed to generate random password: %v", err)
		}
		if len(password) < 16 {
			t.Errorf("Random password too short, expected at least 16 characters, got %d", len(password))
		}

		// Test specific length
		length := 32
		password, err = GenerateRandomPassword(length)
		if err != nil {
			t.Fatalf("Failed to generate random password: %v", err)
		}
		if len(password) != length {
			t.Errorf("Random password wrong length, expected %d characters, got %d", length, len(password))
		}

		// Test uniqueness (two random passwords should be different)
		password1, _ := GenerateRandomPassword(20)
		password2, _ := GenerateRandomPassword(20)
		if password1 == password2 {
			t.Error("Two randomly generated passwords were identical")
		}
	})

	t.Run("InvalidKeystoreFile", func(t *testing.T) {
		// Create an invalid keystore file
		invalidKeystorePath := filepath.Join(tempDir, "invalid-keystore.json")
		invalidContent := []byte(`{"version": 4, "crypto": {}, "uuid": "1234"}`) // Missing publicKey field
		err := os.WriteFile(invalidKeystorePath, invalidContent, 0600)
		if err != nil {
			t.Fatalf("Failed to create invalid keystore file: %v", err)
		}

		// Try to load the invalid keystore
		_, err = LoadFromKeystore(invalidKeystorePath, "any-password")
		if err == nil {
			t.Error("Expected error when loading invalid keystore, but got nil")
		}
	})
}

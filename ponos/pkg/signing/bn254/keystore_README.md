# BN254 Keystore

This package provides keystore functionality for BN254 private keys, following the Ethereum Web3 Secret Storage format.

## Features

- **Secure Storage**: Store private keys securely using the Ethereum Web3 Secret Storage format
- **Key Management**: Generate, save, and load BN254 private keys
- **Password Protection**: Protect keystores with strong password-based encryption
- **Configurable Security**: Choose between standard and light encryption parameters
- **CLI Tool**: Use the included command-line tool for common keystore operations

## Usage

### Programmatic Usage

```go
package main

import (
	"fmt"
	"log"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/bn254"
)

func main() {
	// Generate a new key pair
	privateKey, publicKey, err := bn254.GenerateKeyPair()
	if err != nil {
		log.Fatalf("Failed to generate key pair: %v", err)
	}

	// Save the private key to a keystore file
	keystorePath := "my_key.json"
	password := "my-strong-password"
	
	// Use default security parameters
	err = bn254.SaveToKeystore(privateKey, keystorePath, password, nil)
	if err != nil {
		log.Fatalf("Failed to save keystore: %v", err)
	}
	
	fmt.Printf("Key saved to %s\n", keystorePath)
	fmt.Printf("Public key: %x\n", publicKey.Bytes())
	
	// Later, load the private key from the keystore
	loadedPrivateKey, err := bn254.LoadFromKeystore(keystorePath, password)
	if err != nil {
		log.Fatalf("Failed to load keystore: %v", err)
	}
	
	// Use the loaded private key for signing
	message := []byte("Hello, world!")
	signature, err := loadedPrivateKey.Sign(message)
	if err != nil {
		log.Fatalf("Failed to sign message: %v", err)
	}
	
	// Verify the signature
	loadedPublicKey := loadedPrivateKey.Public()
	valid, err := signature.Verify(loadedPublicKey, message)
	if err != nil {
		log.Fatalf("Failed to verify signature: %v", err)
	}
	
	if valid {
		fmt.Println("Signature verification successful!")
	} else {
		fmt.Println("Signature verification failed!")
	}
}
```

### Command-Line Tool

A command-line tool is included for common keystore operations:

#### Building

```bash
# From the repository root
cd ponos/pkg/signing/bn254/cmd/keystoretool
go build -o keystoretool .
```

#### Generating a Keystore

```bash
# Generate a keystore with a password
./keystoretool generate --output my_key.json --password "my-strong-password"

# Generate a keystore with a random password (will be displayed)
./keystoretool generate --output my_key.json --random-password

# Generate a keystore with light encryption (faster but less secure)
./keystoretool generate --output my_key.json --password "my-password" --light
```

#### Inspecting a Keystore

```bash
# View information about a keystore (without decrypting the private key)
./keystoretool inspect my_key.json
```

#### Testing a Keystore

```bash
# Test a keystore by signing a test message
./keystoretool test my_key.json
# You will be prompted for the password
```

## Security Notes

1. **Password Strength**: Always use strong, unique passwords for keystores containing private keys.
2. **Backup Security**: Store backups of keystores securely; anyone with the keystore file and password can access the private key.
3. **Memory Management**: Be aware that private keys are loaded into memory when used; take appropriate precautions in your application.
4. **Light vs. Standard Encryption**: Light encryption is faster but less secure; use standard encryption for production environments.

## Implementation Details

This keystore implementation follows the Ethereum Web3 Secret Storage format (version 4), which includes:

- Scrypt key derivation function for password-based encryption
- AES-128-CTR symmetric encryption for the private key
- JSON structure with metadata including UUID and version

The implementation is compatible with other tools and libraries that support the Ethereum Web3 Secret Storage format, such as the Ethereum go-ethereum `keystore` package. 
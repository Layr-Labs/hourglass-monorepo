// Package signing_test provides tests for the signing package
package signing_test

import (
	"bytes"
	"testing"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/bls381"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/bn254"
)

// TestGenericSigningInterface tests the basic functionality of signing schemes through the generic interface
func TestGenericSigningInterface(t *testing.T) {
	schemes := map[string]signing.SigningScheme{
		"BLS381": bls381.NewScheme(),
		"BN254":  bn254.NewScheme(),
	}

	for name, scheme := range schemes {
		t.Run(name, func(t *testing.T) {
			// Test key generation
			privKey, pubKey, err := scheme.GenerateKeyPair()
			if err != nil {
				t.Fatalf("Failed to generate key pair: %v", err)
			}

			// Test message signing
			message := []byte("Hello, world!")
			sig, err := privKey.Sign(message)
			if err != nil {
				t.Fatalf("Failed to sign message: %v", err)
			}

			// Test signature verification
			valid, err := sig.Verify(pubKey, message)
			if err != nil {
				t.Fatalf("Failed to verify signature: %v", err)
			}
			if !valid {
				t.Error("Signature verification failed")
			}

			// Test serialization/deserialization
			privKeyBytes := privKey.Bytes()
			pubKeyBytes := pubKey.Bytes()
			sigBytes := sig.Bytes()

			// Deserialize
			recoveredPrivKey, err := scheme.NewPrivateKeyFromBytes(privKeyBytes)
			if err != nil {
				t.Fatalf("Failed to deserialize private key: %v", err)
			}

			recoveredPubKey, err := scheme.NewPublicKeyFromBytes(pubKeyBytes)
			if err != nil {
				t.Fatalf("Failed to deserialize public key: %v", err)
			}

			recoveredSig, err := scheme.NewSignatureFromBytes(sigBytes)
			if err != nil {
				t.Fatalf("Failed to deserialize signature: %v", err)
			}

			// Verify recovered objects
			if !bytes.Equal(recoveredPrivKey.Bytes(), privKeyBytes) {
				t.Error("Deserialized private key doesn't match original")
			}

			if !bytes.Equal(recoveredPubKey.Bytes(), pubKeyBytes) {
				t.Error("Deserialized public key doesn't match original")
			}

			valid, err = recoveredSig.Verify(recoveredPubKey, message)
			if err != nil {
				t.Fatalf("Failed to verify with deserialized objects: %v", err)
			}
			if !valid {
				t.Error("Verification with deserialized objects failed")
			}

			// Test seed-based key generation
			seed := []byte("a seed phrase that is at least 32 bytes long")

			// Generate two key pairs with the same seed - should be identical
			seedKey1, seedPub1, err := scheme.GenerateKeyPairFromSeed(seed)
			if err != nil {
				t.Fatalf("Failed to generate key pair from seed: %v", err)
			}

			seedKey2, seedPub2, err := scheme.GenerateKeyPairFromSeed(seed)
			if err != nil {
				t.Fatalf("Failed to generate second key pair from seed: %v", err)
			}

			// Check that keys from the same seed are identical
			if !bytes.Equal(seedKey1.Bytes(), seedKey2.Bytes()) {
				t.Error("Private keys from the same seed don't match")
			}

			if !bytes.Equal(seedPub1.Bytes(), seedPub2.Bytes()) {
				t.Error("Public keys from the same seed don't match")
			}

			// Test signing with seed-derived key
			seedSig, err := seedKey1.Sign(message)
			if err != nil {
				t.Fatalf("Failed to sign with seed-derived key: %v", err)
			}

			valid, err = seedSig.Verify(seedPub1, message)
			if err != nil {
				t.Fatalf("Failed to verify signature from seed-derived key: %v", err)
			}
			if !valid {
				t.Error("Verification with seed-derived key failed")
			}

			t.Run("Batch Operations", func(t *testing.T) {
				// Test batch operations
				numKeys := 3
				privKeys := make([]signing.PrivateKey, numKeys)
				pubKeys := make([]signing.PublicKey, numKeys)
				sigs := make([]signing.Signature, numKeys)

				// Generate multiple key pairs
				for i := 0; i < numKeys; i++ {
					var err error
					privKeys[i], pubKeys[i], err = scheme.GenerateKeyPair()
					if err != nil {
						t.Fatalf("Failed to generate key pair %d: %v", i, err)
					}

					// Sign the same message with different keys
					sigs[i], err = privKeys[i].Sign(message)
					if err != nil {
						t.Fatalf("Failed to sign message with key %d: %v", i, err)
					}
				}

				// Aggregate signatures
				aggSig, err := scheme.AggregateSignatures(sigs)
				if err != nil {
					t.Fatalf("Failed to aggregate signatures: %v", err)
				}

				// Test batch verify
				valid, err := scheme.BatchVerify(pubKeys, message, sigs)
				if err != nil {
					t.Fatalf("Failed to batch verify signatures: %v", err)
				}
				if !valid {
					t.Error("Batch verification failed")
				}

				// Test aggregate verify with same message
				messages := make([][]byte, numKeys)
				for i := 0; i < numKeys; i++ {
					messages[i] = message
				}

				valid, err = scheme.AggregateVerify(pubKeys, messages, aggSig)
				if err != nil {
					t.Fatalf("Failed to verify aggregate signature: %v", err)
				}
				if !valid {
					t.Error("Aggregate signature verification failed")
				}
			})
		})
	}
}

// TestEIP2333Support tests the EIP-2333 support
func TestEIP2333Support(t *testing.T) {
	schemes := map[string]signing.SigningScheme{
		"BLS381": bls381.NewScheme(),
		"BN254":  bn254.NewScheme(),
	}

	for name, scheme := range schemes {
		t.Run(name, func(t *testing.T) {
			seed := []byte("a seed phrase that is at least 32 bytes long")
			path := []uint32{3, 14, 15, 92}

			privKey, pubKey, err := scheme.GenerateKeyPairEIP2333(seed, path)

			if name == "BLS381" {
				// BLS381 should support EIP-2333
				if err != nil {
					t.Fatalf("BLS381 should support EIP-2333, but got error: %v", err)
				}

				if privKey == nil || pubKey == nil {
					t.Fatal("BLS381 EIP-2333 key generation returned nil keys")
				}

				// Verify keys work for signing
				message := []byte("test message for EIP-2333")
				sig, err := privKey.Sign(message)
				if err != nil {
					t.Fatalf("Failed to sign with EIP-2333 key: %v", err)
				}

				valid, err := sig.Verify(pubKey, message)
				if err != nil {
					t.Fatalf("Failed to verify signature from EIP-2333 key: %v", err)
				}
				if !valid {
					t.Error("Verification with EIP-2333 key failed")
				}

				// Test deterministic nature - same seed and path should produce same keys
				privKey2, pubKey2, err := scheme.GenerateKeyPairEIP2333(seed, path)
				if err != nil {
					t.Fatalf("Failed on second EIP-2333 key generation: %v", err)
				}

				if !bytes.Equal(privKey.Bytes(), privKey2.Bytes()) {
					t.Error("EIP-2333 private keys with same inputs don't match")
				}

				if !bytes.Equal(pubKey.Bytes(), pubKey2.Bytes()) {
					t.Error("EIP-2333 public keys with same inputs don't match")
				}

				// Different paths should produce different keys
				differentPath := []uint32{42, 42, 42, 42}
				privKey3, _, err := scheme.GenerateKeyPairEIP2333(seed, differentPath)
				if err != nil {
					t.Fatalf("Failed to generate key with different path: %v", err)
				}

				if bytes.Equal(privKey.Bytes(), privKey3.Bytes()) {
					t.Error("EIP-2333 keys with different paths should not match")
				}

			} else if name == "BN254" {
				// BN254 should not support EIP-2333
				if err == nil {
					t.Fatal("BN254 should not support EIP-2333, but no error was returned")
				}

				if err != signing.ErrUnsupportedOperation {
					t.Fatalf("Expected ErrUnsupportedOperation, but got: %v", err)
				}

				if privKey != nil || pubKey != nil {
					t.Fatal("BN254 EIP-2333 key generation should return nil keys")
				}
			}
		})
	}
}

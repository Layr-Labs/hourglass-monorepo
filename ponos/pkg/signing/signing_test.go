// Package signing_test provides tests for the signing package
package signing_test

import (
	"testing"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/bls381"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/bn254"
)

// TestGenericSigningInterface demonstrates using both BLS381 and BN254 implementations
// through the generic SigningScheme interface
func TestGenericSigningInterface(t *testing.T) {
	// Test with both implementations
	schemes := []struct {
		name   string
		scheme signing.SigningScheme
	}{
		{"BLS381", bls381.NewScheme()},
		{"BN254", bn254.NewScheme()},
	}

	for _, tc := range schemes {
		t.Run(tc.name, func(t *testing.T) {
			testSigningScheme(t, tc.scheme)
		})
	}
}

// testSigningScheme tests all the functionality of a signing scheme through the generic interface
func testSigningScheme(t *testing.T, scheme signing.SigningScheme) {
	// Generate a key pair
	privateKey, publicKey, err := scheme.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Sign a message
	message := []byte("Hello, generic interface!")
	signature, err := privateKey.Sign(message)
	if err != nil {
		t.Fatalf("Failed to sign message: %v", err)
	}

	// Verify the signature
	valid, err := signature.Verify(publicKey, message)
	if err != nil {
		t.Fatalf("Failed to verify signature: %v", err)
	}
	if !valid {
		t.Error("Signature verification failed")
	}

	// Test verification with wrong message
	wrongMessage := []byte("Wrong message")
	valid, err = signature.Verify(publicKey, wrongMessage)
	if err != nil {
		t.Fatalf("Failed to verify signature with wrong message: %v", err)
	}
	if valid {
		t.Error("Signature verification passed with wrong message")
	}

	// Test serialization and deserialization
	privateKeyBytes := privateKey.Bytes()
	publicKeyBytes := publicKey.Bytes()
	signatureBytes := signature.Bytes()

	// Deserialize private key
	recoveredPrivateKey, err := scheme.NewPrivateKeyFromBytes(privateKeyBytes)
	if err != nil {
		t.Fatalf("Failed to deserialize private key: %v", err)
	}

	// Deserialize public key
	recoveredPublicKey, err := scheme.NewPublicKeyFromBytes(publicKeyBytes)
	if err != nil {
		t.Fatalf("Failed to deserialize public key: %v", err)
	}

	// Deserialize signature
	recoveredSignature, err := scheme.NewSignatureFromBytes(signatureBytes)
	if err != nil {
		t.Fatalf("Failed to deserialize signature: %v", err)
	}

	// Verify the deserialized signature
	valid, err = recoveredSignature.Verify(recoveredPublicKey, message)
	if err != nil {
		t.Fatalf("Failed to verify deserialized signature: %v", err)
	}
	if !valid {
		t.Error("Deserialized signature verification failed")
	}

	// Use the recovered private key to sign a message
	_, err = recoveredPrivateKey.Sign(message)
	if err != nil {
		t.Fatalf("Failed to sign with recovered private key: %v", err)
	}

	// Test batch operations
	t.Run("Batch Operations", func(t *testing.T) {
		numSigners := 3
		privateKeys := make([]signing.PrivateKey, numSigners)
		publicKeys := make([]signing.PublicKey, numSigners)
		signatures := make([]signing.Signature, numSigners)

		// Generate multiple key pairs and signatures
		for i := 0; i < numSigners; i++ {
			var err error
			privateKeys[i], publicKeys[i], err = scheme.GenerateKeyPair()
			if err != nil {
				t.Fatalf("Failed to generate key pair %d: %v", i, err)
			}

			// Sign the same message
			signatures[i], err = privateKeys[i].Sign(message)
			if err != nil {
				t.Fatalf("Failed to sign message with key %d: %v", i, err)
			}
		}

		// Aggregate signatures
		aggSignature, err := scheme.AggregateSignatures(signatures)
		if err != nil {
			t.Fatalf("Failed to aggregate signatures: %v", err)
		}

		// Test batch verification
		valid, err := scheme.BatchVerify(publicKeys, message, signatures)
		if err != nil {
			t.Fatalf("Failed to verify batch signatures: %v", err)
		}
		if !valid {
			t.Error("Batch signature verification failed")
		}

		// Test aggregate verification with same message
		messages := [][]byte{message, message, message}
		valid, err = scheme.AggregateVerify(publicKeys, messages, aggSignature)
		if err != nil {
			t.Fatalf("Failed to verify aggregate signature: %v", err)
		}
		if !valid {
			t.Error("Aggregate signature verification failed")
		}

		// Test aggregate verification with different messages
		differentMessages := [][]byte{
			[]byte("Message 1"),
			[]byte("Message 2"),
			[]byte("Message 3"),
		}

		// Generate new signatures with different messages
		diffSignatures := make([]signing.Signature, numSigners)
		for i := 0; i < numSigners; i++ {
			diffSignatures[i], err = privateKeys[i].Sign(differentMessages[i])
			if err != nil {
				t.Fatalf("Failed to sign different message with key %d: %v", i, err)
			}
		}

		// Aggregate the different-message signatures
		diffAggSignature, err := scheme.AggregateSignatures(diffSignatures)
		if err != nil {
			t.Fatalf("Failed to aggregate different-message signatures: %v", err)
		}

		// Verify the aggregated signature against different messages
		valid, err = scheme.AggregateVerify(publicKeys, differentMessages, diffAggSignature)
		if err != nil {
			t.Fatalf("Failed to verify aggregate signature with different messages: %v", err)
		}
		if !valid {
			t.Error("Aggregate signature verification with different messages failed")
		}
	})
}

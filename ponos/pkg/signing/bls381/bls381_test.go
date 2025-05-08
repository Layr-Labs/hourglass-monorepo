package bls381

import (
	"bytes"
	"testing"
)

func Test_BLS381(t *testing.T) {
	t.Run("KeyGeneration", func(t *testing.T) {
		privateKey, publicKey, err := GenerateKeyPair()
		if err != nil {
			t.Fatalf("Failed to generate key pair: %v", err)
		}

		// Check that keys are not nil
		if privateKey == nil {
			t.Error("Generated private key is nil")
		}
		if publicKey == nil {
			t.Error("Generated public key is nil")
		}

		// Verify that the private key bytes are not empty
		if len(privateKey.Bytes()) == 0 {
			t.Error("Private key bytes are empty")
		}

		// Verify that the public key bytes are not empty
		if len(publicKey.Bytes()) == 0 {
			t.Error("Public key bytes are empty")
		}

		// Ensure Public() derives the correct public key
		derivedPublicKey := privateKey.Public()
		if !bytes.Equal(derivedPublicKey.Bytes(), publicKey.Bytes()) {
			t.Error("Derived public key doesn't match the generated public key")
		}
	})

	t.Run("KeyGenerationFromSeed", func(t *testing.T) {
		// Test with the same seed to ensure deterministic behavior
		seed := []byte("a seed phrase that is at least 32 bytes long")

		// Generate first key pair
		privateKey1, publicKey1, err := GenerateKeyPairFromSeed(seed)
		if err != nil {
			t.Fatalf("Failed to generate key pair from seed: %v", err)
		}

		// Generate second key pair with the same seed
		privateKey2, publicKey2, err := GenerateKeyPairFromSeed(seed)
		if err != nil {
			t.Fatalf("Failed to generate second key pair from seed: %v", err)
		}

		// Keys generated from the same seed should be identical
		if !bytes.Equal(privateKey1.Bytes(), privateKey2.Bytes()) {
			t.Error("Private keys generated from the same seed are not equal")
		}
		if !bytes.Equal(publicKey1.Bytes(), publicKey2.Bytes()) {
			t.Error("Public keys generated from the same seed are not equal")
		}

		// Test with a different seed
		differentSeed := []byte("a different seed phrase at least 32 bytes")
		privateKey3, publicKey3, err := GenerateKeyPairFromSeed(differentSeed)
		if err != nil {
			t.Fatalf("Failed to generate key pair from different seed: %v", err)
		}

		// Keys generated from different seeds should be different
		if bytes.Equal(privateKey1.Bytes(), privateKey3.Bytes()) {
			t.Error("Private keys generated from different seeds are equal")
		}
		if bytes.Equal(publicKey1.Bytes(), publicKey3.Bytes()) {
			t.Error("Public keys generated from different seeds are equal")
		}

		// Make sure keys can be used for signing and verification
		message := []byte("test message for seed-based keys")
		signature, err := privateKey1.Sign(message)
		if err != nil {
			t.Fatalf("Failed to sign with seed-based key: %v", err)
		}

		valid, err := signature.Verify(publicKey1, message)
		if err != nil {
			t.Fatalf("Failed to verify signature from seed-based key: %v", err)
		}
		if !valid {
			t.Error("Signature verification with seed-based key failed")
		}
	})

	t.Run("KeyGenerationEIP2333", func(t *testing.T) {
		// Test with the same seed and path to ensure deterministic behavior
		seed := []byte("a seed phrase that is at least 32 bytes long")
		path := []uint32{3, 14, 15, 92}

		// Generate first key pair
		privateKey1, publicKey1, err := GenerateKeyPairEIP2333(seed, path)
		if err != nil {
			t.Fatalf("Failed to generate key pair using EIP-2333: %v", err)
		}

		// Generate second key pair with the same seed and path
		privateKey2, publicKey2, err := GenerateKeyPairEIP2333(seed, path)
		if err != nil {
			t.Fatalf("Failed to generate second key pair using EIP-2333: %v", err)
		}

		// Keys generated from the same seed and path should be identical
		if !bytes.Equal(privateKey1.Bytes(), privateKey2.Bytes()) {
			t.Error("Private keys generated with the same parameters are not equal")
		}
		if !bytes.Equal(publicKey1.Bytes(), publicKey2.Bytes()) {
			t.Error("Public keys generated with the same parameters are not equal")
		}

		// Test with a different path
		differentPath := []uint32{42, 42, 42, 42}
		privateKey3, publicKey3, err := GenerateKeyPairEIP2333(seed, differentPath)
		if err != nil {
			t.Fatalf("Failed to generate key pair with different path: %v", err)
		}

		// Keys generated from the same seed but different paths should be different
		if bytes.Equal(privateKey1.Bytes(), privateKey3.Bytes()) {
			t.Error("Private keys generated with different paths are equal")
		}
		if bytes.Equal(publicKey1.Bytes(), publicKey3.Bytes()) {
			t.Error("Public keys generated with different paths are equal")
		}

		// Make sure keys can be used for signing and verification
		message := []byte("test message for EIP-2333 keys")
		signature, err := privateKey1.Sign(message)
		if err != nil {
			t.Fatalf("Failed to sign with EIP-2333 key: %v", err)
		}

		valid, err := signature.Verify(publicKey1, message)
		if err != nil {
			t.Fatalf("Failed to verify signature from EIP-2333 key: %v", err)
		}
		if !valid {
			t.Error("Signature verification with EIP-2333 key failed")
		}
	})

	t.Run("SerializationDeserialization", func(t *testing.T) {
		// Generate a key pair
		privateKey, publicKey, err := GenerateKeyPair()
		if err != nil {
			t.Fatalf("Failed to generate key pair: %v", err)
		}

		// Test private key serialization/deserialization
		privateKeyBytes := privateKey.Bytes()
		recoveredPrivateKey, err := NewPrivateKeyFromBytes(privateKeyBytes)
		if err != nil {
			t.Fatalf("Failed to deserialize private key: %v", err)
		}

		// Test that the recovered private key works for signing
		message := []byte("test message")
		signatureFromRecovered, err := recoveredPrivateKey.Sign(message)
		if err != nil {
			t.Fatalf("Failed to sign with recovered private key: %v", err)
		}

		// Test public key serialization/deserialization
		publicKeyBytes := publicKey.Bytes()
		deserializedPublicKey, err := NewPublicKeyFromBytes(publicKeyBytes)
		if err != nil {
			t.Fatalf("Failed to deserialize public key: %v", err)
		}

		// Generate a signature to test signature serialization/deserialization
		signature, err := privateKey.Sign(message)
		if err != nil {
			t.Fatalf("Failed to sign message: %v", err)
		}

		// Verify the signature from the recovered private key
		valid, err := signatureFromRecovered.Verify(publicKey, message)
		if err != nil {
			t.Fatalf("Failed to verify signature from recovered private key: %v", err)
		}
		if !valid {
			t.Error("Signature from recovered private key verification failed")
		}

		// Test signature serialization/deserialization
		signatureBytes := signature.Bytes()
		deserializedSignature, err := NewSignatureFromBytes(signatureBytes)
		if err != nil {
			t.Fatalf("Failed to deserialize signature: %v", err)
		}

		// Verify the deserialized signature
		valid, err = deserializedSignature.Verify(deserializedPublicKey, message)
		if err != nil {
			t.Fatalf("Failed to verify deserialized signature: %v", err)
		}
		if !valid {
			t.Error("Deserialized signature verification failed")
		}
	})

	t.Run("SignAndVerify", func(t *testing.T) {
		// Generate a key pair
		privateKey, publicKey, err := GenerateKeyPair()
		if err != nil {
			t.Fatalf("Failed to generate key pair: %v", err)
		}

		// Test signing a message
		message := []byte("Hello, world!")
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

		t.Run("VerifyWithWrongMessage", func(t *testing.T) {
			wrongMessage := []byte("Wrong message")
			valid, err = signature.Verify(publicKey, wrongMessage)
			if err != nil {
				t.Fatalf("Failed to verify signature with wrong message: %v", err)
			}
			if valid {
				t.Error("Signature verification passed with wrong message")
			}
		})

		t.Run("VerifyWithWrongKey", func(t *testing.T) {
			_, wrongPublicKey, err := GenerateKeyPair()
			if err != nil {
				t.Fatalf("Failed to generate wrong key pair: %v", err)
			}
			valid, err = signature.Verify(wrongPublicKey, message)
			if err != nil {
				t.Fatalf("Failed to verify signature with wrong key: %v", err)
			}
			if valid {
				t.Error("Signature verification passed with wrong key")
			}
		})
	})

	t.Run("AggregateSignatures", func(t *testing.T) {
		// Generate multiple key pairs
		numKeys := 3
		message := []byte("Hello, world!")
		privateKeys := make([]*PrivateKey, numKeys)
		publicKeys := make([]*PublicKey, numKeys)
		signatures := make([]*Signature, numKeys)

		for i := 0; i < numKeys; i++ {
			var err error
			privateKeys[i], publicKeys[i], err = GenerateKeyPair()
			if err != nil {
				t.Fatalf("Failed to generate key pair %d: %v", i, err)
			}

			// Sign the same message with different keys
			signatures[i], err = privateKeys[i].Sign(message)
			if err != nil {
				t.Fatalf("Failed to sign message with key %d: %v", i, err)
			}

			// Verify individual signatures
			valid, err := signatures[i].Verify(publicKeys[i], message)
			if err != nil {
				t.Fatalf("Failed to verify signature %d: %v", i, err)
			}
			if !valid {
				t.Errorf("Signature %d verification failed", i)
			}
		}

		// Aggregate signatures
		aggregatedSignature, err := AggregateSignatures(signatures)
		if err != nil {
			t.Fatalf("Failed to aggregate signatures: %v", err)
		}

		t.Run("BatchVerification", func(t *testing.T) {
			// Verify batch signature (all signers signed the same message)
			valid, err := BatchVerify(publicKeys, message, signatures)
			if err != nil {
				t.Fatalf("Failed to verify batch signatures: %v", err)
			}
			if !valid {
				t.Error("Batch signature verification failed")
			}
		})

		t.Run("AggregateVerificationWithSameMessage", func(t *testing.T) {
			// Verify aggregate signature against multiple public keys with same message
			valid, err := AggregateVerify(publicKeys, [][]byte{message, message, message}, aggregatedSignature)
			if err != nil {
				t.Fatalf("Failed to verify aggregate signature: %v", err)
			}
			if !valid {
				t.Error("Aggregate signature verification failed")
			}
		})
	})

	t.Run("AggregateVerifyWithDifferentMessages", func(t *testing.T) {
		// Generate multiple key pairs
		numKeys := 3
		messages := [][]byte{
			[]byte("Message 1"),
			[]byte("Message 2"),
			[]byte("Message 3"),
		}
		privateKeys := make([]*PrivateKey, numKeys)
		publicKeys := make([]*PublicKey, numKeys)
		signatures := make([]*Signature, numKeys)

		for i := 0; i < numKeys; i++ {
			var err error
			privateKeys[i], publicKeys[i], err = GenerateKeyPair()
			if err != nil {
				t.Fatalf("Failed to generate key pair %d: %v", i, err)
			}

			// Each key signs a different message
			signatures[i], err = privateKeys[i].Sign(messages[i])
			if err != nil {
				t.Fatalf("Failed to sign message %d: %v", i, err)
			}
		}

		// Aggregate signatures
		aggregatedSignature, err := AggregateSignatures(signatures)
		if err != nil {
			t.Fatalf("Failed to aggregate signatures: %v", err)
		}

		t.Run("CorrectMessages", func(t *testing.T) {
			// Verify aggregate signature with different messages
			valid, err := AggregateVerify(publicKeys, messages, aggregatedSignature)
			if err != nil {
				t.Fatalf("Failed to verify aggregate signature with different messages: %v", err)
			}
			if !valid {
				t.Error("Aggregate signature verification with different messages failed")
			}
		})

		t.Run("WrongMessages", func(t *testing.T) {
			// Try with wrong messages
			wrongMessages := [][]byte{
				[]byte("Wrong message 1"),
				[]byte("Message 2"),
				[]byte("Message 3"),
			}
			valid, err := AggregateVerify(publicKeys, wrongMessages, aggregatedSignature)
			if err != nil {
				t.Fatalf("Failed to verify aggregate signature with wrong messages: %v", err)
			}
			if valid {
				t.Error("Aggregate signature verification passed with wrong messages")
			}
		})
	})

	t.Run("EmptyAggregation", func(t *testing.T) {
		// Test aggregating empty set of signatures
		_, err := AggregateSignatures([]*Signature{})
		if err == nil {
			t.Error("Expected error when aggregating empty set of signatures, but got none")
		}
	})
}

func TestHashToG1(t *testing.T) {
	tests := []struct {
		name    string
		message []byte
	}{
		{
			name:    "empty message",
			message: []byte{},
		},
		{
			name:    "simple message",
			message: []byte("Hello, World!"),
		},
		{
			name:    "long message",
			message: []byte("This is a longer message with some special characters: !@#$%^&*()"),
		},
		{
			name:    "very long message",
			message: bytes.Repeat([]byte("a"), 1000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			point := hashToG1(tt.message)

			// Check that the point is not nil
			if point == nil {
				t.Fatal("hashToG1 returned nil point")
			}

			// Check that the point is on the curve
			if !point.IsOnCurve() {
				t.Error("hashToG1 returned point not on curve")
			}

			// Check that the point is in the correct subgroup
			if !point.IsInSubGroup() {
				t.Error("hashToG1 returned point not in subgroup")
			}
		})
	}
}

func TestPublicKeyG2(t *testing.T) {
	// Generate a key pair
	sk, pk, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Check that the public key is in G2
	if !pk.GetG2Point().IsOnCurve() {
		t.Error("Public key point is not on curve")
	}

	if !pk.GetG2Point().IsInSubGroup() {
		t.Error("Public key point is not in subgroup")
	}

	// Verify that the public key matches the private key
	expectedPk := sk.Public()
	if !bytes.Equal(pk.PointBytes, expectedPk.PointBytes) {
		t.Error("Public key does not match private key")
	}
}

func TestSignatureG1(t *testing.T) {
	// Generate a key pair
	sk, pk, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Sign a message
	message := []byte("test message")
	sig, err := sk.Sign(message)
	if err != nil {
		t.Fatalf("Failed to sign message: %v", err)
	}

	// Check that the signature is in G1
	if !sig.sig.IsOnCurve() {
		t.Error("Signature point is not on curve")
	}

	if !sig.sig.IsInSubGroup() {
		t.Error("Signature point is not in subgroup")
	}

	// Verify the signature
	valid, err := sig.Verify(pk, message)
	if err != nil {
		t.Fatalf("Failed to verify signature: %v", err)
	}
	if !valid {
		t.Error("Signature verification failed")
	}
}

func TestAggregateSignaturesG1(t *testing.T) {
	// Generate multiple key pairs
	numKeys := 3
	privateKeys := make([]*PrivateKey, numKeys)
	publicKeys := make([]*PublicKey, numKeys)
	signatures := make([]*Signature, numKeys)

	message := []byte("test message")

	for i := 0; i < numKeys; i++ {
		sk, pk, err := GenerateKeyPair()
		if err != nil {
			t.Fatalf("Failed to generate key pair %d: %v", i, err)
		}

		privateKeys[i] = sk
		publicKeys[i] = pk

		// Sign the message
		sig, err := sk.Sign(message)
		if err != nil {
			t.Fatalf("Failed to sign message with key %d: %v", i, err)
		}
		signatures[i] = sig
	}

	// Aggregate signatures
	aggSig, err := AggregateSignatures(signatures)
	if err != nil {
		t.Fatalf("Failed to aggregate signatures: %v", err)
	}

	// Check that the aggregated signature is in G1
	if !aggSig.sig.IsOnCurve() {
		t.Error("Aggregated signature point is not on curve")
	}

	if !aggSig.sig.IsInSubGroup() {
		t.Error("Aggregated signature point is not in subgroup")
	}

	// Verify the aggregated signature
	valid, err := BatchVerify(publicKeys, message, signatures)
	if err != nil {
		t.Fatalf("Failed to batch verify signatures: %v", err)
	}
	if !valid {
		t.Error("Batch signature verification failed")
	}
}

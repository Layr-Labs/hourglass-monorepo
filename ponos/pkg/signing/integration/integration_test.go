package integration

import (
	"bytes"
	"testing"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/bls381"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/bn254"
)

// Test_AllSignatureTypes tests both BLS381 and BN254 signature schemes through the generic interface
func Test_AllSignatureTypes(t *testing.T) {
	schemes := []struct {
		name   string
		scheme signing.SigningScheme
	}{
		{"BLS381", bls381.NewScheme()},
		{"BN254", bn254.NewScheme()},
	}

	for _, tc := range schemes {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("BasicSignAndVerify", func(t *testing.T) {
				testBasicSignAndVerify(t, tc.scheme)
			})
			t.Run("SerializationDeserialization", func(t *testing.T) {
				testSerializationDeserialization(t, tc.scheme)
			})
			t.Run("AggregateSignatures", func(t *testing.T) {
				testAggregateSignatures(t, tc.scheme)
			})
			t.Run("AggregateSignaturesWithDifferentMessages", func(t *testing.T) {
				testAggregateSignaturesWithDifferentMessages(t, tc.scheme)
			})
			t.Run("BatchVerification", func(t *testing.T) {
				testBatchVerification(t, tc.scheme)
			})
		})
	}
}

// testBasicSignAndVerify tests basic signature creation and verification
func testBasicSignAndVerify(t *testing.T, scheme signing.SigningScheme) {
	// Generate a key pair
	privateKey, publicKey, err := scheme.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Sign a message
	message := []byte("Hello, BLS signatures!")
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

	// Test verification with wrong key
	wrongPrivateKey, wrongPublicKey, err := scheme.GenerateKeyPair()
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

	// Create a signature with the wrong key
	wrongSignature, err := wrongPrivateKey.Sign(message)
	if err != nil {
		t.Fatalf("Failed to sign with wrong key: %v", err)
	}

	valid, err = wrongSignature.Verify(publicKey, message)
	if err != nil {
		t.Fatalf("Failed to verify wrong signature: %v", err)
	}
	if valid {
		t.Error("Wrong signature verification passed")
	}
}

// testSerializationDeserialization tests serialization and deserialization of keys and signatures
func testSerializationDeserialization(t *testing.T, scheme signing.SigningScheme) {
	// Generate a key pair
	privateKey, publicKey, err := scheme.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Sign a message
	message := []byte("Hello, BLS signatures!")
	signature, err := privateKey.Sign(message)
	if err != nil {
		t.Fatalf("Failed to sign message: %v", err)
	}

	// Serialize keys and signature
	privateKeyBytes := privateKey.Bytes()
	publicKeyBytes := publicKey.Bytes()
	signatureBytes := signature.Bytes()

	// Ensure serialized bytes are not empty
	if len(privateKeyBytes) == 0 {
		t.Error("Serialized private key is empty")
	}
	if len(publicKeyBytes) == 0 {
		t.Error("Serialized public key is empty")
	}
	if len(signatureBytes) == 0 {
		t.Error("Serialized signature is empty")
	}

	// Deserialize private key
	recoveredPrivKey, err := scheme.NewPrivateKeyFromBytes(privateKeyBytes)
	if err != nil {
		t.Fatalf("Failed to deserialize private key: %v", err)
	}

	// Deserialize public key
	recoveredPubKey, err := scheme.NewPublicKeyFromBytes(publicKeyBytes)
	if err != nil {
		t.Fatalf("Failed to deserialize public key: %v", err)
	}

	// Deserialize signature
	recoveredSignature, err := scheme.NewSignatureFromBytes(signatureBytes)
	if err != nil {
		t.Fatalf("Failed to deserialize signature: %v", err)
	}

	// Verify the deserialized public key bytes match the original
	recoveredPubKeyBytes := recoveredPubKey.Bytes()
	if !bytes.Equal(recoveredPubKeyBytes, publicKeyBytes) {
		t.Error("Deserialized public key bytes don't match original")
	}

	// Verify the deserialized signature bytes match the original
	recoveredSigBytes := recoveredSignature.Bytes()
	if !bytes.Equal(recoveredSigBytes, signatureBytes) {
		t.Error("Deserialized signature bytes don't match original")
	}

	// Verify recovered signature
	valid, err := recoveredSignature.Verify(recoveredPubKey, message)
	if err != nil {
		t.Fatalf("Failed to verify recovered signature: %v", err)
	}
	if !valid {
		t.Error("Recovered signature verification failed")
	}

	// Sign a new message with the recovered private key
	newMessage := []byte("A new message to sign!")
	newSignature, err := recoveredPrivKey.Sign(newMessage)
	if err != nil {
		t.Fatalf("Failed to sign with recovered private key: %v", err)
	}

	// Verify the new signature
	valid, err = newSignature.Verify(recoveredPubKey, newMessage)
	if err != nil {
		t.Fatalf("Failed to verify signature from recovered private key: %v", err)
	}
	if !valid {
		t.Error("Signature from recovered private key verification failed")
	}
}

// testAggregateSignatures tests signature aggregation with the same message
func testAggregateSignatures(t *testing.T, scheme signing.SigningScheme) {
	// Generate multiple key pairs
	numSigners := 3
	message := []byte("Hello, aggregate signatures!")
	privateKeys := make([]signing.PrivateKey, numSigners)
	publicKeys := make([]signing.PublicKey, numSigners)
	signatures := make([]signing.Signature, numSigners)

	for i := 0; i < numSigners; i++ {
		var err error
		privateKeys[i], publicKeys[i], err = scheme.GenerateKeyPair()
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
	aggregatedSignature, err := scheme.AggregateSignatures(signatures)
	if err != nil {
		t.Fatalf("Failed to aggregate signatures: %v", err)
	}

	// Verify aggregated signature with multiple public keys on the same message
	messages := [][]byte{message, message, message}
	valid, err := scheme.AggregateVerify(publicKeys, messages, aggregatedSignature)
	if err != nil {
		t.Fatalf("Failed to verify aggregated signature: %v", err)
	}
	if !valid {
		t.Error("Aggregated signature verification failed")
	}

	// Test with wrong message
	wrongMessages := [][]byte{message, message, []byte("wrong message")}
	valid, err = scheme.AggregateVerify(publicKeys, wrongMessages, aggregatedSignature)
	if err != nil {
		t.Fatalf("Failed to verify aggregated signature with wrong message: %v", err)
	}
	if valid {
		t.Error("Aggregated signature verification passed with wrong message")
	}

	// Test empty aggregation
	_, err = scheme.AggregateSignatures([]signing.Signature{})
	if err == nil {
		t.Error("Expected error when aggregating empty set of signatures, but got none")
	}
}

// testAggregateSignaturesWithDifferentMessages tests signature aggregation with different messages
func testAggregateSignaturesWithDifferentMessages(t *testing.T, scheme signing.SigningScheme) {
	// Generate multiple key pairs
	numSigners := 3
	messages := [][]byte{
		[]byte("Message 1"),
		[]byte("Message 2"),
		[]byte("Message 3"),
	}
	privateKeys := make([]signing.PrivateKey, numSigners)
	publicKeys := make([]signing.PublicKey, numSigners)
	signatures := make([]signing.Signature, numSigners)

	for i := 0; i < numSigners; i++ {
		var err error
		privateKeys[i], publicKeys[i], err = scheme.GenerateKeyPair()
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
	aggregatedSignature, err := scheme.AggregateSignatures(signatures)
	if err != nil {
		t.Fatalf("Failed to aggregate signatures: %v", err)
	}

	// Verify aggregate signature with different messages
	valid, err := scheme.AggregateVerify(publicKeys, messages, aggregatedSignature)
	if err != nil {
		t.Fatalf("Failed to verify aggregate signature with different messages: %v", err)
	}
	if !valid {
		t.Error("Aggregate signature verification with different messages failed")
	}

	// Try with wrong messages
	wrongMessages := [][]byte{
		[]byte("Wrong message 1"),
		messages[1],
		messages[2],
	}
	valid, err = scheme.AggregateVerify(publicKeys, wrongMessages, aggregatedSignature)
	if err != nil {
		t.Fatalf("Failed to verify aggregate signature with wrong messages: %v", err)
	}
	if valid {
		t.Error("Aggregate signature verification passed with wrong messages")
	}

	// Try with mismatched number of public keys and messages
	tooFewMessages := messages[:2]
	_, err = scheme.AggregateVerify(publicKeys, tooFewMessages, aggregatedSignature)
	if err == nil {
		t.Error("Expected error with mismatched number of public keys and messages, but got none")
	}
}

// testBatchVerification tests batch verification (all signers signed the same message)
func testBatchVerification(t *testing.T, scheme signing.SigningScheme) {
	// Generate multiple key pairs
	numSigners := 3
	message := []byte("Hello, batch verification!")
	privateKeys := make([]signing.PrivateKey, numSigners)
	publicKeys := make([]signing.PublicKey, numSigners)
	signatures := make([]signing.Signature, numSigners)

	for i := 0; i < numSigners; i++ {
		var err error
		privateKeys[i], publicKeys[i], err = scheme.GenerateKeyPair()
		if err != nil {
			t.Fatalf("Failed to generate key pair %d: %v", i, err)
		}

		// Sign the same message with different keys
		signatures[i], err = privateKeys[i].Sign(message)
		if err != nil {
			t.Fatalf("Failed to sign message with key %d: %v", i, err)
		}
	}

	// Verify using batch verification (all signers signed the same message)
	valid, err := scheme.BatchVerify(publicKeys, message, signatures)
	if err != nil {
		t.Fatalf("Failed to batch verify signatures: %v", err)
	}
	if !valid {
		t.Error("Batch verification failed")
	}

	// Test with wrong message
	wrongMessage := []byte("Wrong message")
	valid, err = scheme.BatchVerify(publicKeys, wrongMessage, signatures)
	if err != nil {
		t.Fatalf("Failed to batch verify signatures with wrong message: %v", err)
	}
	if valid {
		t.Error("Batch verification passed with wrong message")
	}

	// Test with a wrong signature
	wrongSignatures := make([]signing.Signature, numSigners)
	copy(wrongSignatures, signatures)
	// Replace one signature with a signature for a different message
	wrongSignatures[1], err = privateKeys[1].Sign(wrongMessage)
	if err != nil {
		t.Fatalf("Failed to sign wrong message: %v", err)
	}

	valid, err = scheme.BatchVerify(publicKeys, message, wrongSignatures)
	if err != nil {
		t.Fatalf("Failed to batch verify signatures with one wrong signature: %v", err)
	}
	if valid {
		t.Error("Batch verification passed with one wrong signature")
	}

	// Test with mismatched number of public keys and signatures
	tooFewSignatures := signatures[:2]
	_, err = scheme.BatchVerify(publicKeys, message, tooFewSignatures)
	if err == nil {
		t.Error("Expected error with mismatched number of public keys and signatures, but got none")
	}
}

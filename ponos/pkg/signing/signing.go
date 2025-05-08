// Package signing provides a generic interface for BLS signature schemes
package signing

import "errors"

// Common errors
var (
	ErrInvalidSignatureType = errors.New("invalid signature type")
	ErrInvalidPublicKeyType = errors.New("invalid public key type")
	ErrUnsupportedOperation = errors.New("operation not supported by this scheme")
)

type SolidityBN254PublicKey struct {
	G1Point []byte
}

// PrivateKey is the interface that all private key implementations must satisfy
type PrivateKey interface {
	// Sign signs a message and returns a signature
	Sign(message []byte) (Signature, error)

	// Public returns the corresponding public key
	Public() PublicKey

	// Bytes serializes the private key to bytes
	Bytes() []byte
}

// PublicKey is the interface that all public key implementations must satisfy
type PublicKey interface {
	// Bytes serializes the public key to bytes
	Bytes() []byte
}

// Signature is the interface that all signature implementations must satisfy
type Signature interface {
	// Verify verifies the signature against a message and public key
	Verify(publicKey PublicKey, message []byte) (bool, error)

	// Bytes serializes the signature to bytes
	Bytes() []byte
}

// SigningScheme represents a BLS signature scheme implementation
type SigningScheme interface {
	// GenerateKeyPair creates a new random private key and the corresponding public key
	GenerateKeyPair() (PrivateKey, PublicKey, error)

	// GenerateKeyPairFromSeed creates a deterministic private key and the corresponding public key from a seed
	GenerateKeyPairFromSeed(seed []byte) (PrivateKey, PublicKey, error)

	// GenerateKeyPairEIP2333 creates a deterministic private key and the corresponding public key using the EIP-2333 standard
	// This is specific to BLS12-381 implementations and may not be supported by all schemes
	GenerateKeyPairEIP2333(seed []byte, path []uint32) (PrivateKey, PublicKey, error)

	// NewPrivateKeyFromBytes creates a private key from bytes
	NewPrivateKeyFromBytes(data []byte) (PrivateKey, error)

	// NewPublicKeyFromBytes creates a public key from bytes
	NewPublicKeyFromBytes(data []byte) (PublicKey, error)

	// NewPublicKeyFromHexString creates a public key from a hex string
	NewPublicKeyFromHexString(hex string) (PublicKey, error)

	// NewSignatureFromBytes creates a signature from bytes
	NewSignatureFromBytes(data []byte) (Signature, error)

	// AggregateSignatures combines multiple signatures into a single signature
	AggregateSignatures(signatures []Signature) (Signature, error)

	// BatchVerify verifies multiple signatures in a single batch operation
	// Each signature corresponds to the same message signed by different public keys
	BatchVerify(publicKeys []PublicKey, message []byte, signatures []Signature) (bool, error)

	// AggregateVerify verifies an aggregated signature against multiple public keys and multiple messages
	AggregateVerify(publicKeys []PublicKey, messages [][]byte, aggSignature Signature) (bool, error)
}

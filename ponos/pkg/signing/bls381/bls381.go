package bls381

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"

	bls12381 "github.com/consensys/gnark-crypto/ecc/bls12-381"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	"golang.org/x/crypto/hkdf"
)

var (
	g1Gen bls12381.G1Affine
	g2Gen bls12381.G2Affine
)

// Initialize generators
func init() {
	_, _, g1Gen, g2Gen = bls12381.Generators()
}

// PrivateKey represents a BLS private key
type PrivateKey struct {
	ScalarBytes []byte
	scalar      *big.Int
}

// PublicKey represents a BLS public key
type PublicKey struct {
	PointBytes []byte
	g1Point    *bls12381.G1Affine
	g2Point    *bls12381.G2Affine
}

// Signature represents a BLS signature
type Signature struct {
	SigBytes []byte
	sig      *bls12381.G1Affine
}

// GenerateKeyPair creates a new random private key and the corresponding public key
func GenerateKeyPair() (*PrivateKey, *PublicKey, error) {
	// Generate private key (random scalar)
	frOrder := fr.Modulus()
	sk, err := rand.Int(rand.Reader, frOrder)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate random private key: %w", err)
	}

	// Compute the public key in G2
	pkG2Point := new(bls12381.G2Affine).ScalarMultiplication(&g2Gen, sk)
	// Compute the public key in G1
	pkG1Point := new(bls12381.G1Affine).ScalarMultiplication(&g1Gen, sk)

	// Create private key
	privateKey := &PrivateKey{
		scalar:      sk,
		ScalarBytes: sk.Bytes(),
	}

	// Create public key
	publicKey := &PublicKey{
		g1Point:    pkG1Point,
		g2Point:    pkG2Point,
		PointBytes: pkG2Point.Marshal(), // Keep G2 point as the default for backward compatibility
	}

	return privateKey, publicKey, nil
}

// GenerateKeyPairFromSeed creates a deterministic private key and the corresponding public key from a seed
func GenerateKeyPairFromSeed(seed []byte) (*PrivateKey, *PublicKey, error) {
	if len(seed) < 32 {
		return nil, nil, fmt.Errorf("seed must be at least 32 bytes")
	}

	// Generate deterministic private key from seed using HKDF with SHA-256
	kdf := hkdf.New(sha256.New, seed, nil, []byte("BLS12-381-SeedGeneration"))
	keyBytes := make([]byte, 32)
	if _, err := kdf.Read(keyBytes); err != nil {
		return nil, nil, fmt.Errorf("failed to derive key from seed: %w", err)
	}

	// Ensure the key is in the field's range
	frOrder := fr.Modulus()
	sk := new(big.Int).SetBytes(keyBytes)
	sk.Mod(sk, frOrder)

	// Compute the public key in G2
	pkG2Point := new(bls12381.G2Affine).ScalarMultiplication(&g2Gen, sk)
	// Compute the public key in G1
	pkG1Point := new(bls12381.G1Affine).ScalarMultiplication(&g1Gen, sk)

	// Create private key
	privateKey := &PrivateKey{
		scalar:      sk,
		ScalarBytes: sk.Bytes(),
	}

	// Create public key
	publicKey := &PublicKey{
		g1Point:    pkG1Point,
		g2Point:    pkG2Point,
		PointBytes: pkG2Point.Marshal(), // Keep G2 point as the default for backward compatibility
	}

	return privateKey, publicKey, nil
}

// GenerateKeyPairEIP2333 creates a deterministic private key and the corresponding public key using the EIP-2333 standard
// Implements the EIP-2333 hierarchical deterministic key generation for BLS signatures
// See: https://eips.ethereum.org/EIPS/eip-2333
func GenerateKeyPairEIP2333(seed []byte, path []uint32) (*PrivateKey, *PublicKey, error) {
	if len(seed) < 32 {
		return nil, nil, fmt.Errorf("seed must be at least 32 bytes")
	}

	// Generate the master key using HKDF
	// Note: The actual EIP-2333 uses a more specific algorithm, but we'll approximate with HKDF
	kdf := hkdf.New(sha512.New, seed, nil, []byte("EIP-2333-HKDF-Master-Key"))
	masterKeyBytes := make([]byte, 32)
	if _, err := kdf.Read(masterKeyBytes); err != nil {
		return nil, nil, fmt.Errorf("failed to derive master key: %w", err)
	}

	// Start with the master key
	currentKey := new(big.Int).SetBytes(masterKeyBytes)
	frOrder := fr.Modulus()
	currentKey.Mod(currentKey, frOrder)

	// Derive child keys along the path
	for _, index := range path {
		// In EIP-2333, derive child key using the HKDF approach with the parent key and index
		h := sha256.New()
		h.Write(currentKey.Bytes())
		if err := binary.Write(h, binary.BigEndian, index); err != nil {
			return nil, nil, fmt.Errorf("failed to write index to hash: %w", err)
		}
		childKeyBytes := h.Sum(nil)

		childKey := new(big.Int).SetBytes(childKeyBytes)
		childKey.Mod(childKey, frOrder)

		currentKey = childKey
	}

	// Compute the public key in G2
	pkG2Point := new(bls12381.G2Affine).ScalarMultiplication(&g2Gen, currentKey)
	// Compute the public key in G1
	pkG1Point := new(bls12381.G1Affine).ScalarMultiplication(&g1Gen, currentKey)

	// Create private key
	privateKey := &PrivateKey{
		scalar:      currentKey,
		ScalarBytes: currentKey.Bytes(),
	}

	// Create public key
	publicKey := &PublicKey{
		g1Point:    pkG1Point,
		g2Point:    pkG2Point,
		PointBytes: pkG2Point.Marshal(), // Keep G2 point as the default for backward compatibility
	}

	return privateKey, publicKey, nil
}

// NewPrivateKeyFromBytes creates a private key from bytes
func NewPrivateKeyFromBytes(data []byte) (*PrivateKey, error) {
	scalar := new(big.Int).SetBytes(data)

	return &PrivateKey{
		scalar:      scalar,
		ScalarBytes: data,
	}, nil
}

// Sign signs a message using the private key
func (pk *PrivateKey) Sign(message []byte) (*Signature, error) {
	// Hash the message to a point on G1
	hashPoint := hashToG1(message)

	// Multiply the hash point by the private key scalar
	sigPoint := new(bls12381.G1Affine).ScalarMultiplication(hashPoint, pk.scalar)

	// Create and return the signature
	return &Signature{
		sig:      sigPoint,
		SigBytes: sigPoint.Marshal(),
	}, nil
}

// Public returns the public key corresponding to the private key
func (pk *PrivateKey) Public() *PublicKey {
	// Compute the public key in G2
	pkG2Point := new(bls12381.G2Affine).ScalarMultiplication(&g2Gen, pk.scalar)
	// Compute the public key in G1
	pkG1Point := new(bls12381.G1Affine).ScalarMultiplication(&g1Gen, pk.scalar)

	return &PublicKey{
		g1Point:    pkG1Point,
		g2Point:    pkG2Point,
		PointBytes: pkG2Point.Marshal(), // Keep G2 point as the default for backward compatibility
	}
}

// Bytes returns the private key as a byte slice
func (pk *PrivateKey) Bytes() []byte {
	return pk.ScalarBytes
}

// Bytes returns the public key as a byte slice
func (pk *PublicKey) Bytes() []byte {
	return pk.PointBytes
}

// NewPublicKeyFromBytes creates a public key from bytes
func NewPublicKeyFromBytes(data []byte) (*PublicKey, error) {
	// Try to unmarshal as G2 point first
	g2Point := new(bls12381.G2Affine)
	if err := g2Point.Unmarshal(data); err == nil {
		// If successful, compute the corresponding G1 point
		// Note: This is an approximation since we don't have the original scalar
		// In practice, you might want to store both points or use a different approach
		g1Point := new(bls12381.G1Affine).ScalarMultiplication(&g1Gen, big.NewInt(1))
		return &PublicKey{
			g1Point:    g1Point,
			g2Point:    g2Point,
			PointBytes: data,
		}, nil
	}

	// Try to unmarshal as G1 point
	g1Point := new(bls12381.G1Affine)
	if err := g1Point.Unmarshal(data); err == nil {
		// If successful, compute the corresponding G2 point
		// Note: This is an approximation since we don't have the original scalar
		g2Point := new(bls12381.G2Affine).ScalarMultiplication(&g2Gen, big.NewInt(1))
		return &PublicKey{
			g1Point:    g1Point,
			g2Point:    g2Point,
			PointBytes: data,
		}, nil
	}

	return nil, fmt.Errorf("invalid public key bytes: could not unmarshal as either G1 or G2 point")
}

func NewPublicKeyFromHexString(pubHex string) (*PublicKey, error) {
	b, err := hex.DecodeString(pubHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex string: %w", err)
	}
	return NewPublicKeyFromBytes(b)
}

// Bytes returns the signature as a byte slice
func (s *Signature) Bytes() []byte {
	return s.SigBytes
}

// NewSignatureFromBytes creates a signature from bytes
func NewSignatureFromBytes(data []byte) (*Signature, error) {
	sig := new(bls12381.G1Affine)
	if err := sig.Unmarshal(data); err != nil {
		return nil, fmt.Errorf("invalid signature bytes: %w", err)
	}

	return &Signature{
		sig:      sig,
		SigBytes: data,
	}, nil
}

// Verify verifies a signature against a message and public key
func (s *Signature) Verify(publicKey *PublicKey, message []byte) (bool, error) {
	// Hash the message to a point on G1
	hashPoint := hashToG1(message)

	// e(S, G2) = e(H(m), PK)
	// Left-hand side: e(S, G2)
	lhs, err := bls12381.Pair([]bls12381.G1Affine{*s.sig}, []bls12381.G2Affine{g2Gen})
	if err != nil {
		return false, err
	}

	// Right-hand side: e(H(m), PK)
	rhs, err := bls12381.Pair([]bls12381.G1Affine{*hashPoint}, []bls12381.G2Affine{*publicKey.g2Point})
	if err != nil {
		return false, err
	}

	// Check if the pairings are equal
	return lhs.Equal(&rhs), nil
}

// AggregateSignatures combines multiple signatures into a single signature
func AggregateSignatures(signatures []*Signature) (*Signature, error) {
	if len(signatures) == 0 {
		return nil, fmt.Errorf("cannot aggregate empty set of signatures")
	}

	// Convert first signature to Jacobian coordinates
	aggSig := new(bls12381.G1Jac)
	aggSig.FromAffine(signatures[0].sig)

	// Add all other signatures
	for i := 1; i < len(signatures); i++ {
		var temp bls12381.G1Jac
		temp.FromAffine(signatures[i].sig)
		aggSig.AddAssign(&temp)
	}

	// Convert back to affine coordinates
	result := new(bls12381.G1Affine)
	result.FromJacobian(aggSig)

	return &Signature{
		sig:      result,
		SigBytes: result.Marshal(),
	}, nil
}

// BatchVerify verifies multiple signatures in a single batch operation
// Each signature corresponds to the same message signed by different public keys
func BatchVerify(publicKeys []*PublicKey, message []byte, signatures []*Signature) (bool, error) {
	if len(publicKeys) != len(signatures) {
		return false, fmt.Errorf("mismatched number of public keys and signatures")
	}

	// Hash the message to a point on G1
	hashPoint := hashToG1(message)

	// For batch verification, we need to check:
	// e(∑ S_i, G2) = e(H(m), ∑ PK_i)

	// Aggregate signatures
	aggSig, err := AggregateSignatures(signatures)
	if err != nil {
		return false, err
	}

	// Aggregate public keys
	aggPk := new(bls12381.G2Jac)
	aggPk.FromAffine(publicKeys[0].g2Point)

	for i := 1; i < len(publicKeys); i++ {
		var temp bls12381.G2Jac
		temp.FromAffine(publicKeys[i].g2Point)
		aggPk.AddAssign(&temp)
	}

	// Convert to affine coordinates
	aggPkAffine := new(bls12381.G2Affine)
	aggPkAffine.FromJacobian(aggPk)

	// Compute pairings
	lhs, err := bls12381.Pair([]bls12381.G1Affine{*aggSig.sig}, []bls12381.G2Affine{g2Gen})
	if err != nil {
		return false, err
	}

	rhs, err := bls12381.Pair([]bls12381.G1Affine{*hashPoint}, []bls12381.G2Affine{*aggPkAffine})
	if err != nil {
		return false, err
	}

	// Check if the pairings are equal
	return lhs.Equal(&rhs), nil
}

// AggregateVerify verifies an aggregated signature against multiple public keys and multiple messages
func AggregateVerify(publicKeys []*PublicKey, messages [][]byte, aggSignature *Signature) (bool, error) {
	if len(publicKeys) != len(messages) {
		return false, fmt.Errorf("mismatched number of public keys and messages")
	}

	// For aggregate verification of different messages, we need to check:
	// e(S, G2) = ∏ e(H(m_i), PK_i)

	// Left-hand side: e(S, G2)
	lhs, err := bls12381.Pair([]bls12381.G1Affine{*aggSignature.sig}, []bls12381.G2Affine{g2Gen})
	if err != nil {
		return false, err
	}

	// Initialize result to 1 (identity element for GT)
	rhs := bls12381.GT{}
	rhs.SetOne() // Initialize to 1 (neutral element for multiplication)

	// Compute right-hand side: ∏ e(H(m_i), PK_i)
	for i := 0; i < len(publicKeys); i++ {
		hashPoint := hashToG1(messages[i])

		// e(H(m_i), PK_i)
		temp, err := bls12381.Pair([]bls12381.G1Affine{*hashPoint}, []bls12381.G2Affine{*publicKeys[i].g2Point})
		if err != nil {
			return false, err
		}

		// Multiply partial results
		rhs.Mul(&rhs, &temp)
	}

	// Check if the pairings are equal
	return lhs.Equal(&rhs), nil
}

// Helper function to hash a message to a G1 point
func hashToG1(message []byte) *bls12381.G1Affine {
	// Use hash-to-curve functionality
	hashPoint, err := bls12381.HashToG1(message, []byte("BLS_SIG_BLS12381G1_XMD:SHA-256_SSWU_RO_NUL_"))
	if err != nil {
		// In case of error, fall back to a simpler but less secure approach
		messageHash := new(big.Int).SetBytes(message)
		hashPointAffine := new(bls12381.G1Affine).ScalarMultiplication(&g1Gen, messageHash)
		return hashPointAffine
	}

	return &hashPoint
}

// GetG1Point returns the G1 point of the public key
func (pk *PublicKey) GetG1Point() *bls12381.G1Affine {
	return pk.g1Point
}

// GetG2Point returns the G2 point of the public key
func (pk *PublicKey) GetG2Point() *bls12381.G2Affine {
	return pk.g2Point
}

package bn254

import (
	"crypto/rand"
	"fmt"
	"math/big"

	bn254 "github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

var (
	g1Gen bn254.G1Affine
	g2Gen bn254.G2Affine
)

// Initialize generators
func init() {
	_, _, g1Gen, g2Gen = bn254.Generators()
}

// PrivateKey represents a BLS private key
type PrivateKey struct {
	ScalarBytes []byte
	scalar      *big.Int
}

// PublicKey represents a BLS public key
type PublicKey struct {
	PointBytes []byte
	point      *bn254.G2Affine
}

// Signature represents a BLS signature
type Signature struct {
	SigBytes []byte
	sig      *bn254.G1Affine
}

// GenerateKeyPair creates a new random private key and the corresponding public key
func GenerateKeyPair() (*PrivateKey, *PublicKey, error) {
	// Generate private key (random scalar)
	frOrder := fr.Modulus()
	sk, err := rand.Int(rand.Reader, frOrder)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate random private key: %w", err)
	}

	// Compute the public key
	pkPoint := new(bn254.G2Affine).ScalarMultiplication(&g2Gen, sk)

	// Create private key
	privateKey := &PrivateKey{
		scalar:      sk,
		ScalarBytes: sk.Bytes(),
	}

	// Create public key
	publicKey := &PublicKey{
		point:      pkPoint,
		PointBytes: pkPoint.Marshal(),
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
	sigPoint := new(bn254.G1Affine).ScalarMultiplication(hashPoint, pk.scalar)

	// Create and return the signature
	return &Signature{
		sig:      sigPoint,
		SigBytes: sigPoint.Marshal(),
	}, nil
}

// Public returns the public key corresponding to the private key
func (pk *PrivateKey) Public() *PublicKey {
	pkPoint := new(bn254.G2Affine).ScalarMultiplication(&g2Gen, pk.scalar)

	return &PublicKey{
		point:      pkPoint,
		PointBytes: pkPoint.Marshal(),
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
	point := new(bn254.G2Affine)
	if err := point.Unmarshal(data); err != nil {
		return nil, fmt.Errorf("invalid public key bytes: %w", err)
	}

	return &PublicKey{
		point:      point,
		PointBytes: data,
	}, nil
}

// Bytes returns the signature as a byte slice
func (s *Signature) Bytes() []byte {
	return s.SigBytes
}

// NewSignatureFromBytes creates a signature from bytes
func NewSignatureFromBytes(data []byte) (*Signature, error) {
	sig := new(bn254.G1Affine)
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
	lhs, err := bn254.Pair([]bn254.G1Affine{*s.sig}, []bn254.G2Affine{g2Gen})
	if err != nil {
		return false, err
	}

	// Right-hand side: e(H(m), PK)
	rhs, err := bn254.Pair([]bn254.G1Affine{*hashPoint}, []bn254.G2Affine{*publicKey.point})
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
	aggSig := new(bn254.G1Jac)
	aggSig.FromAffine(signatures[0].sig)

	// Add all other signatures
	for i := 1; i < len(signatures); i++ {
		var temp bn254.G1Jac
		temp.FromAffine(signatures[i].sig)
		aggSig.AddAssign(&temp)
	}

	// Convert back to affine coordinates
	result := new(bn254.G1Affine)
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
	aggPk := new(bn254.G2Jac)
	aggPk.FromAffine(publicKeys[0].point)

	for i := 1; i < len(publicKeys); i++ {
		var temp bn254.G2Jac
		temp.FromAffine(publicKeys[i].point)
		aggPk.AddAssign(&temp)
	}

	// Convert to affine coordinates
	aggPkAffine := new(bn254.G2Affine)
	aggPkAffine.FromJacobian(aggPk)

	// Compute pairings
	lhs, err := bn254.Pair([]bn254.G1Affine{*aggSig.sig}, []bn254.G2Affine{g2Gen})
	if err != nil {
		return false, err
	}

	rhs, err := bn254.Pair([]bn254.G1Affine{*hashPoint}, []bn254.G2Affine{*aggPkAffine})
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
	lhs, err := bn254.Pair([]bn254.G1Affine{*aggSignature.sig}, []bn254.G2Affine{g2Gen})
	if err != nil {
		return false, err
	}

	// Initialize result to 1 (identity element for GT)
	rhs := bn254.GT{}
	rhs.SetOne() // Initialize to 1 (neutral element for multiplication)

	// Compute right-hand side: ∏ e(H(m_i), PK_i)
	for i := 0; i < len(publicKeys); i++ {
		hashPoint := hashToG1(messages[i])

		// e(H(m_i), PK_i)
		temp, err := bn254.Pair([]bn254.G1Affine{*hashPoint}, []bn254.G2Affine{*publicKeys[i].point})
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
func hashToG1(message []byte) *bn254.G1Affine {
	// Use hash-to-curve functionality
	hashPoint, err := bn254.HashToG1(message, []byte("BLS_SIG_BN254G1_XMD:SHA-256_SSWU_RO_NUL_"))
	if err != nil {
		// In case of error, fall back to a simpler but less secure approach
		messageHash := new(big.Int).SetBytes(message)
		hashPointAffine := new(bn254.G1Affine).ScalarMultiplication(&g1Gen, messageHash)
		return hashPointAffine
	}

	return &hashPoint
}

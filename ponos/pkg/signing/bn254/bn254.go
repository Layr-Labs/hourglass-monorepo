package bn254

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fp"

	"github.com/Layr-Labs/hourglass-monorepo/contracts/pkg/bindings/ITaskAVSRegistrar"

	bn254 "github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"golang.org/x/crypto/hkdf"
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
	point      *bn254.G1Affine
}

// Signature represents a BLS signature
type Signature struct {
	SigBytes []byte
	Sig      *bn254.G2Affine
}

func (s *Signature) G2Point() *bn254.G2Jac {
	return new(bn254.G2Jac).FromAffine(s.Sig)
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
	pkPoint := new(bn254.G1Affine).ScalarMultiplication(&g1Gen, sk)

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

// GenerateKeyPairFromSeed creates a deterministic private key and the corresponding public key from a seed
func GenerateKeyPairFromSeed(seed []byte) (*PrivateKey, *PublicKey, error) {
	if len(seed) < 32 {
		return nil, nil, fmt.Errorf("seed must be at least 32 bytes")
	}

	// Generate deterministic private key from seed using HKDF with SHA-256
	kdf := hkdf.New(sha256.New, seed, nil, []byte("BN254-SeedGeneration"))
	keyBytes := make([]byte, 32)
	if _, err := kdf.Read(keyBytes); err != nil {
		return nil, nil, fmt.Errorf("failed to derive key from seed: %w", err)
	}

	// Ensure the key is in the field's range
	frOrder := fr.Modulus()
	sk := new(big.Int).SetBytes(keyBytes)
	sk.Mod(sk, frOrder)

	// Compute the public key
	pkPoint := new(bn254.G1Affine).ScalarMultiplication(&g1Gen, sk)

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
	// Hash the message to a point on G2
	hashPoint := hashToG2(message)

	// Multiply the hash point by the private key scalar
	sigPoint := new(bn254.G2Affine).ScalarMultiplication(hashPoint, pk.scalar)

	// Create and return the signature
	return &Signature{
		Sig:      sigPoint,
		SigBytes: sigPoint.Marshal(),
	}, nil
}

// Public returns the public key corresponding to the private key
func (pk *PrivateKey) Public() *PublicKey {
	pkPoint := new(bn254.G1Affine).ScalarMultiplication(&g1Gen, pk.scalar)

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

// NewPublicKeyFromSolidity creates a public key from a Solidity G1 point
func NewPublicKeyFromSolidity(g1 ITaskAVSRegistrar.BN254G1Point) (*PublicKey, error) {
	// Create a new PublicKey struct
	pubKey := &PublicKey{}

	// Create a new G1Affine point
	pubKey.point = new(bn254.G1Affine)

	// Set the X coordinate
	pubKey.point.X.SetBigInt(g1.X)

	// Set the Y coordinate
	pubKey.point.Y.SetBigInt(g1.Y)

	// Marshal the point to bytes to fill the PointBytes field
	pointBytes := pubKey.point.Marshal()
	pubKey.PointBytes = pointBytes

	return pubKey, nil
}

// NewPublicKeyFromBytes creates a public key from bytes
func NewPublicKeyFromBytes(data []byte) (*PublicKey, error) {
	point := new(bn254.G1Affine)
	if err := point.Unmarshal(data); err != nil {
		return nil, fmt.Errorf("invalid public key bytes: %w", err)
	}

	return &PublicKey{
		point:      point,
		PointBytes: data,
	}, nil
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
	sig := new(bn254.G2Affine)
	if err := sig.Unmarshal(data); err != nil {
		return nil, fmt.Errorf("invalid signature bytes: %w", err)
	}

	return &Signature{
		Sig:      sig,
		SigBytes: data,
	}, nil
}

// Verify verifies a signature against a message and public key
func (s *Signature) Verify(publicKey *PublicKey, message []byte) (bool, error) {
	// Hash the message to a point on G2
	hashPoint := hashToG2(message)

	// e(PK, H(m)) = e(G1, S)
	// Left-hand side: e(PK, H(m))
	lhs, err := bn254.Pair([]bn254.G1Affine{*publicKey.point}, []bn254.G2Affine{*hashPoint})
	if err != nil {
		return false, err
	}

	// Right-hand side: e(G1, S)
	rhs, err := bn254.Pair([]bn254.G1Affine{g1Gen}, []bn254.G2Affine{*s.Sig})
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
	aggSig := new(bn254.G2Jac)
	aggSig.FromAffine(signatures[0].Sig)

	// Add all other signatures
	for i := 1; i < len(signatures); i++ {
		var temp bn254.G2Jac
		temp.FromAffine(signatures[i].Sig)
		aggSig.AddAssign(&temp)
	}

	// Convert back to affine coordinates
	result := new(bn254.G2Affine)
	result.FromJacobian(aggSig)

	return &Signature{
		Sig:      result,
		SigBytes: result.Marshal(),
	}, nil
}

// BatchVerify verifies multiple signatures in a single batch operation
func BatchVerify(publicKeys []*PublicKey, message []byte, signatures []*Signature) (bool, error) {
	if len(publicKeys) != len(signatures) {
		return false, fmt.Errorf("mismatched number of public keys and signatures")
	}

	// Hash the message to a point on G2
	hashPoint := hashToG2(message)

	// For batch verification, we need to check:
	// e(∑ PK_i, H(m)) = e(G1, ∑ S_i)

	// Aggregate public keys
	aggPk := new(bn254.G1Jac)
	aggPk.FromAffine(publicKeys[0].point)

	for i := 1; i < len(publicKeys); i++ {
		var temp bn254.G1Jac
		temp.FromAffine(publicKeys[i].point)
		aggPk.AddAssign(&temp)
	}

	// Convert to affine coordinates
	aggPkAffine := new(bn254.G1Affine)
	aggPkAffine.FromJacobian(aggPk)

	// Aggregate signatures
	aggSig, err := AggregateSignatures(signatures)
	if err != nil {
		return false, err
	}

	// Compute pairings
	lhs, err := bn254.Pair([]bn254.G1Affine{*aggPkAffine}, []bn254.G2Affine{*hashPoint})
	if err != nil {
		return false, err
	}

	rhs, err := bn254.Pair([]bn254.G1Affine{g1Gen}, []bn254.G2Affine{*aggSig.Sig})
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
	// e(∑ PK_i, H(m_i)) = e(G1, S)

	// Aggregate public keys
	aggPk := new(bn254.G1Jac)
	aggPk.FromAffine(publicKeys[0].point)

	for i := 1; i < len(publicKeys); i++ {
		var temp bn254.G1Jac
		temp.FromAffine(publicKeys[i].point)
		aggPk.AddAssign(&temp)
	}

	// Convert to affine coordinates
	aggPkAffine := new(bn254.G1Affine)
	aggPkAffine.FromJacobian(aggPk)

	// Initialize result to 1 (identity element for GT)
	rhs := bn254.GT{}
	rhs.SetOne() // Initialize to 1 (neutral element for multiplication)

	// Compute right-hand side: ∏ e(PK_i, H(m_i))
	for i := 0; i < len(publicKeys); i++ {
		hashPoint := hashToG2(messages[i])

		// e(PK_i, H(m_i))
		temp, err := bn254.Pair([]bn254.G1Affine{*publicKeys[i].point}, []bn254.G2Affine{*hashPoint})
		if err != nil {
			return false, err
		}

		// Multiply partial results
		rhs.Mul(&rhs, &temp)
	}

	// Left-hand side: e(G1, S)
	lhs, err := bn254.Pair([]bn254.G1Affine{g1Gen}, []bn254.G2Affine{*aggSignature.Sig})
	if err != nil {
		return false, err
	}

	// Check if the pairings are equal
	return lhs.Equal(&rhs), nil
}

// Helper function to hash a message to a G2 point
func hashToG2(message []byte) *bn254.G2Affine {
	// Use hash-to-curve functionality
	hashPoint, err := bn254.HashToG2(message, []byte("BLS_SIG_BN254G2_XMD:SHA-256_SSWU_RO_NUL_"))
	if err != nil {
		// In case of error, fall back to a simpler but less secure approach
		messageHash := new(big.Int).SetBytes(message)
		hashPointAffine := new(bn254.G2Affine).ScalarMultiplication(&g2Gen, messageHash)
		return hashPointAffine
	}

	return &hashPoint
}

// AggregatePublicKeys combines multiple public keys into a single aggregated public key.
func AggregatePublicKeys(pubKeys []*PublicKey) (*PublicKey, error) {
	if len(pubKeys) == 0 {
		return nil, fmt.Errorf("cannot aggregate empty set of public keys")
	}

	// Start with the first public key in Jacobian coordinates
	aggPk := new(bn254.G1Jac)
	aggPk.FromAffine(pubKeys[0].point)

	// Add all other public keys
	for i := 1; i < len(pubKeys); i++ {
		var temp bn254.G1Jac
		temp.FromAffine(pubKeys[i].point)
		aggPk.AddAssign(&temp)
	}

	// Convert back to affine coordinates
	result := new(bn254.G1Affine)
	result.FromJacobian(aggPk)

	return &PublicKey{
		point:      result,
		PointBytes: result.Marshal(),
	}, nil
}

func newFpElement(x *big.Int) fp.Element {
	var p fp.Element
	p.SetBigInt(x)
	return p
}

type G1Point struct {
	*bn254.G1Affine
}

func NewG1Point(x, y *big.Int) *G1Point {
	return &G1Point{
		&bn254.G1Affine{
			X: newFpElement(x),
			Y: newFpElement(y),
		},
	}
}

func NewZeroG1Point() *G1Point {
	return NewG1Point(big.NewInt(0), big.NewInt(0))
}

// Add another G1 point to this one
func (p *G1Point) Add(p2 *G1Point) *G1Point {
	p.G1Affine.Add(p.G1Affine, p2.G1Affine)
	return p
}

// Sub another G1 point from this one
func (p *G1Point) Sub(p2 *G1Point) *G1Point {
	p.G1Affine.Sub(p.G1Affine, p2.G1Affine)
	return p
}

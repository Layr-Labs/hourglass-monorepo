package bn254

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fp"

	"github.com/Layr-Labs/hourglass-monorepo/contracts/pkg/bindings/ITaskAVSRegistrar"

	bn254 "github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"golang.org/x/crypto/hkdf"
)

// Error types for precompile compatibility
var (
	ErrInvalidPointFormat = errors.New("invalid point format for precompile")
	ErrPointNotInSubgroup = errors.New("point not in correct subgroup")
	ErrInvalidFieldOrder  = errors.New("number not in valid field order")
)

// FieldModulus is the BN254 field modulus
var FieldModulus = func() *big.Int {
	n, _ := new(big.Int).SetString("21888242871839275222246405745257275088696311157297823662689037894645226208583", 10)
	return n
}()

// Precompile format constants
const (
	G1PointSize = 64  // 32 bytes for x, 32 bytes for y
	G2PointSize = 128 // 64 bytes for x, 64 bytes for y
)

// ValidateFieldOrder checks if a number is in the correct field
func ValidateFieldOrder(n *big.Int) bool {
	return n.Cmp(FieldModulus) < 0
}

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
	g1Point    *bn254.G1Affine
	g2Point    *bn254.G2Affine
}

// Signature represents a BLS signature
type Signature struct {
	SigBytes []byte
	sig      *bn254.G1Affine
}

func (s *Signature) GetG1Point() *bn254.G1Affine {
	return s.sig
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
	pkG2Point := new(bn254.G2Affine).ScalarMultiplication(&g2Gen, sk)
	// Compute the public key in G1
	pkG1Point := new(bn254.G1Affine).ScalarMultiplication(&g1Gen, sk)

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
	kdf := hkdf.New(sha256.New, seed, nil, []byte("BN254-SeedGeneration"))
	keyBytes := make([]byte, 32)
	if _, err := kdf.Read(keyBytes); err != nil {
		return nil, nil, fmt.Errorf("failed to derive key from seed: %w", err)
	}

	// Ensure the key is in the field's range
	frOrder := fr.Modulus()
	sk := new(big.Int).SetBytes(keyBytes)
	sk.Mod(sk, frOrder)

	// Compute the public key in G2
	pkG2Point := new(bn254.G2Affine).ScalarMultiplication(&g2Gen, sk)
	// Compute the public key in G1
	pkG1Point := new(bn254.G1Affine).ScalarMultiplication(&g1Gen, sk)

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
	hashPoint, err := hashToG1(message)
	if err != nil {
		return nil, err
	}

	// Multiply the hash point by the private key scalar
	sigPoint := new(bn254.G1Affine).ScalarMultiplication(hashPoint, pk.scalar)

	// Create and return the signature
	return &Signature{
		sig:      sigPoint,
		SigBytes: sigPoint.Marshal(),
	}, nil
}

func (pk *PrivateKey) SignG1Point(hashPoint *bn254.G1Affine) (*Signature, error) {
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
	// Compute the public key in G2
	pkG2Point := new(bn254.G2Affine).ScalarMultiplication(&g2Gen, pk.scalar)
	// Compute the public key in G1
	pkG1Point := new(bn254.G1Affine).ScalarMultiplication(&g1Gen, pk.scalar)

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

// NewPublicKeyFromSolidity creates a public key from a Solidity G1 and G2 points
func NewPublicKeyFromSolidity(g1 ITaskAVSRegistrar.BN254G1Point, g2 ITaskAVSRegistrar.BN254G2Point) (*PublicKey, error) {
	// Create a new PublicKey struct
	pubKey := &PublicKey{}

	// Create and set G1 point
	pubKey.g1Point = new(bn254.G1Affine)
	pubKey.g1Point.X.SetBigInt(g1.X)
	pubKey.g1Point.Y.SetBigInt(g1.Y)

	// Validate G1 point is in correct subgroup
	if !pubKey.g1Point.IsInSubGroup() {
		return nil, ErrPointNotInSubgroup
	}

	// Create and set G2 point
	pubKey.g2Point = new(bn254.G2Affine)
	// Note: Contract stores coordinates as [X.A1, X.A0, Y.A1, Y.A0]
	pubKey.g2Point.X.A1.SetBigInt(g2.X[0])
	pubKey.g2Point.X.A0.SetBigInt(g2.X[1])
	pubKey.g2Point.Y.A1.SetBigInt(g2.Y[0])
	pubKey.g2Point.Y.A0.SetBigInt(g2.Y[1])

	// Validate G2 point is in correct subgroup
	if !pubKey.g2Point.IsInSubGroup() {
		return nil, ErrPointNotInSubgroup
	}

	// Marshal the G2 point to bytes to fill the PointBytes field
	pointBytes := pubKey.g2Point.Marshal()
	pubKey.PointBytes = pointBytes

	return pubKey, nil
}

// NewPublicKeyFromBytes creates a public key from bytes
func NewPublicKeyFromBytes(data []byte) (*PublicKey, error) {
	// Try to unmarshal as G2 point first
	g2Point := new(bn254.G2Affine)
	if err := g2Point.Unmarshal(data); err == nil {
		// If successful, compute the corresponding G1 point
		// Note: This is an approximation since we don't have the original scalar
		g1Point := new(bn254.G1Affine).ScalarMultiplication(&g1Gen, big.NewInt(1))
		return &PublicKey{
			g1Point:    g1Point,
			g2Point:    g2Point,
			PointBytes: data,
		}, nil
	}

	// Try to unmarshal as G1 point
	g1Point := new(bn254.G1Affine)
	if err := g1Point.Unmarshal(data); err == nil {
		// If successful, compute the corresponding G2 point
		// Note: This is an approximation since we don't have the original scalar
		g2Point := new(bn254.G2Affine).ScalarMultiplication(&g2Gen, big.NewInt(1))
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

// Add adds another signature to this one
func (s *Signature) Add(other *Signature) *Signature {
	if other == nil || other.sig == nil {
		return s
	}
	s.sig.Add(s.sig, other.sig)
	s.SigBytes = s.sig.Marshal()
	return s
}

// Sub subtracts another signature from this one
func (s *Signature) Sub(other *Signature) *Signature {
	if other == nil || other.sig == nil {
		return s
	}
	s.sig.Sub(s.sig, other.sig)
	s.SigBytes = s.sig.Marshal()
	return s
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
	hashPoint, err := hashToG1(message)
	if err != nil {
		return false, err
	}

	// e(S, G2) = e(H(m), PK)
	// Left-hand side: e(S, G2)
	lhs, err := bn254.Pair([]bn254.G1Affine{*s.sig}, []bn254.G2Affine{g2Gen})
	if err != nil {
		return false, err
	}

	// Right-hand side: e(H(m), PK)
	rhs, err := bn254.Pair([]bn254.G1Affine{*hashPoint}, []bn254.G2Affine{*publicKey.g2Point})
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
func BatchVerify(publicKeys []*PublicKey, message []byte, signatures []*Signature) (bool, error) {
	if len(publicKeys) != len(signatures) {
		return false, fmt.Errorf("mismatched number of public keys and signatures")
	}

	// Hash the message to a point on G1
	hashPoint, err := hashToG1(message)
	if err != nil {
		return false, err
	}

	// For batch verification, we need to check:
	// e(∑ S_i, G2) = e(H(m), ∑ PK_i)

	// Aggregate signatures
	aggSig, err := AggregateSignatures(signatures)
	if err != nil {
		return false, err
	}

	// Aggregate public keys
	aggPk := new(bn254.G2Jac)
	aggPk.FromAffine(publicKeys[0].g2Point)

	for i := 1; i < len(publicKeys); i++ {
		var temp bn254.G2Jac
		temp.FromAffine(publicKeys[i].g2Point)
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
		hashPoint, err := hashToG1(messages[i])
		if err != nil {
			return false, err
		}

		// e(H(m_i), PK_i)
		temp, err := bn254.Pair([]bn254.G1Affine{*hashPoint}, []bn254.G2Affine{*publicKeys[i].g2Point})
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
func hashToG1(message []byte) (*bn254.G1Affine, error) {
	// Use hash-to-curve functionality with the standardized domain separator
	hashPoint, err := bn254.HashToG1(message, []byte("BLS_SIG_BN254G1_XMD:SHA-256_SSWU_RO_NUL_"))
	if err != nil {
		return nil, fmt.Errorf("failed to hash message to G1: %w", err)
	}

	// Verify the point is in the correct subgroup
	if !hashPoint.IsInSubGroup() {
		return nil, ErrPointNotInSubgroup
	}

	return &hashPoint, nil
}

// AggregatePublicKeys combines multiple public keys into a single aggregated public key.
func AggregatePublicKeys(pubKeys []*PublicKey) (*PublicKey, error) {
	if len(pubKeys) == 0 {
		return nil, fmt.Errorf("cannot aggregate empty set of public keys")
	}

	// Start with the first public key in Jacobian coordinates
	aggPk := new(bn254.G2Jac)
	aggPk.FromAffine(pubKeys[0].g2Point)

	// Add all other public keys
	for i := 1; i < len(pubKeys); i++ {
		var temp bn254.G2Jac
		temp.FromAffine(pubKeys[i].g2Point)
		aggPk.AddAssign(&temp)
	}

	// Convert back to affine coordinates
	result := new(bn254.G2Affine)
	result.FromJacobian(aggPk)

	return &PublicKey{
		g2Point:    result,
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

// ToPrecompileFormat converts a G1 point to the format expected by the Ethereum precompile
func (p *G1Point) ToPrecompileFormat() ([]byte, error) {
	if !p.IsInSubGroup() {
		return nil, ErrPointNotInSubgroup
	}
	return p.Marshal(), nil
}

// FromPrecompileFormat creates a G1 point from the Ethereum precompile format
func G1PointFromPrecompileFormat(data []byte) (*G1Point, error) {
	if len(data) != G1PointSize {
		return nil, fmt.Errorf("%w: expected %d bytes, got %d", ErrInvalidPointFormat, G1PointSize, len(data))
	}
	point := new(bn254.G1Affine)
	if err := point.Unmarshal(data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal G1 point: %w", err)
	}
	if !point.IsInSubGroup() {
		return nil, ErrPointNotInSubgroup
	}
	return &G1Point{point}, nil
}

// Sub another G1 point from this one
func (p *G1Point) Sub(p2 *G1Point) *G1Point {
	p.G1Affine.Sub(p.G1Affine, p2.G1Affine)
	return p
}

// AddPublicKey adds the G1 point from a public key to this point
func (p *G1Point) AddPublicKey(pk *PublicKey) *G1Point {
	if pk.g1Point == nil {
		return p
	}
	p.G1Affine.Add(p.G1Affine, pk.g1Point)
	return p
}

// GetG1Point returns the G1 point of the public key
func (pk *PublicKey) GetG1Point() *bn254.G1Affine {
	return pk.g1Point
}

// GetG2Point returns the G2 point of the public key
func (pk *PublicKey) GetG2Point() *bn254.G2Affine {
	return pk.g2Point
}

type G2Point struct {
	*bn254.G2Affine
}

func NewG2Point(x0, x1, y0, y1 *big.Int) *G2Point {
	return &G2Point{
		&bn254.G2Affine{
			X: struct {
				A0 fp.Element
				A1 fp.Element
			}{
				A0: newFpElement(x0),
				A1: newFpElement(x1),
			},
			Y: struct {
				A0 fp.Element
				A1 fp.Element
			}{
				A0: newFpElement(y0),
				A1: newFpElement(y1),
			},
		},
	}
}

func NewZeroG2Point() *G2Point {
	return NewG2Point(
		big.NewInt(0), big.NewInt(0), // X coordinates
		big.NewInt(0), big.NewInt(0), // Y coordinates
	)
}

// Add another G2 point to this one
func (p *G2Point) Add(p2 *G2Point) *G2Point {
	p.G2Affine.Add(p.G2Affine, p2.G2Affine)
	return p
}

// ToPrecompileFormat converts a G2 point to the format expected by the Ethereum precompile
func (p *G2Point) ToPrecompileFormat() ([]byte, error) {
	if !p.IsInSubGroup() {
		return nil, ErrPointNotInSubgroup
	}
	return p.Marshal(), nil
}

// FromPrecompileFormat creates a G2 point from the Ethereum precompile format
func G2PointFromPrecompileFormat(data []byte) (*G2Point, error) {
	if len(data) != G2PointSize {
		return nil, fmt.Errorf("%w: expected %d bytes, got %d", ErrInvalidPointFormat, G2PointSize, len(data))
	}
	point := new(bn254.G2Affine)
	if err := point.Unmarshal(data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal G2 point: %w", err)
	}
	if !point.IsInSubGroup() {
		return nil, ErrPointNotInSubgroup
	}
	return &G2Point{point}, nil
}

// Sub another G2 point from this one
func (p *G2Point) Sub(p2 *G2Point) *G2Point {
	p.G2Affine.Sub(p.G2Affine, p2.G2Affine)
	return p
}

// AddPublicKey adds the G2 point from a public key to this point
func (p *G2Point) AddPublicKey(pk *PublicKey) *G2Point {
	if pk.g2Point == nil {
		return p
	}
	p.G2Affine.Add(p.G2Affine, pk.g2Point)
	return p
}

// Sub subtracts another public key from this one
func (pk *PublicKey) Sub(other *PublicKey) *PublicKey {
	if other == nil || other.g2Point == nil {
		return pk
	}
	pk.g2Point.Sub(pk.g2Point, other.g2Point)
	pk.PointBytes = pk.g2Point.Marshal()
	return pk
}

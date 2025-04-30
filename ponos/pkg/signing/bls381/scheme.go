package bls381

import (
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing"
)

// Ensure Scheme implements the SigningScheme interface
var _ signing.SigningScheme = (*Scheme)(nil)

// Scheme implements the SigningScheme interface for BLS381
type Scheme struct{}

// NewScheme creates a new BLS381 signing scheme
func NewScheme() *Scheme {
	return &Scheme{}
}

// GenerateKeyPair creates a new random private key and the corresponding public key
func (s *Scheme) GenerateKeyPair() (signing.PrivateKey, signing.PublicKey, error) {
	privKey, pubKey, err := GenerateKeyPair()
	if err != nil {
		return nil, nil, err
	}

	return &privateKeyAdapter{privKey}, &publicKeyAdapter{pubKey}, nil
}

// GenerateKeyPairFromSeed creates a deterministic private key and the corresponding public key from a seed
func (s *Scheme) GenerateKeyPairFromSeed(seed []byte) (signing.PrivateKey, signing.PublicKey, error) {
	privKey, pubKey, err := GenerateKeyPairFromSeed(seed)
	if err != nil {
		return nil, nil, err
	}

	return &privateKeyAdapter{privKey}, &publicKeyAdapter{pubKey}, nil
}

// GenerateKeyPairEIP2333 creates a deterministic private key and the corresponding public key using the EIP-2333 standard
func (s *Scheme) GenerateKeyPairEIP2333(seed []byte, path []uint32) (signing.PrivateKey, signing.PublicKey, error) {
	privKey, pubKey, err := GenerateKeyPairEIP2333(seed, path)
	if err != nil {
		return nil, nil, err
	}

	return &privateKeyAdapter{privKey}, &publicKeyAdapter{pubKey}, nil
}

// NewPrivateKeyFromBytes creates a private key from bytes
func (s *Scheme) NewPrivateKeyFromBytes(data []byte) (signing.PrivateKey, error) {
	privKey, err := NewPrivateKeyFromBytes(data)
	if err != nil {
		return nil, err
	}
	return &privateKeyAdapter{privKey}, nil
}

// NewPublicKeyFromBytes creates a public key from bytes
func (s *Scheme) NewPublicKeyFromBytes(data []byte) (signing.PublicKey, error) {
	pubKey, err := NewPublicKeyFromBytes(data)
	if err != nil {
		return nil, err
	}
	return &publicKeyAdapter{pubKey}, nil
}

// NewSignatureFromBytes creates a signature from bytes
func (s *Scheme) NewSignatureFromBytes(data []byte) (signing.Signature, error) {
	sig, err := NewSignatureFromBytes(data)
	if err != nil {
		return nil, err
	}
	return &signatureAdapter{sig}, nil
}

func (s *Scheme) NewPublicKeyFromHexString(hex string) (signing.PublicKey, error) {
	pubKey, err := NewPublicKeyFromHexString(hex)
	if err != nil {
		return nil, err
	}
	return &publicKeyAdapter{pubKey}, nil
}

// AggregateSignatures combines multiple signatures into a single signature
func (s *Scheme) AggregateSignatures(signatures []signing.Signature) (signing.Signature, error) {
	// Convert generic signatures to BLS381 specific signatures
	bls381Sigs := make([]*Signature, len(signatures))
	for i, sig := range signatures {
		bls381Sig, ok := sig.(*signatureAdapter)
		if !ok {
			return nil, signing.ErrInvalidSignatureType
		}
		bls381Sigs[i] = bls381Sig.sig
	}

	result, err := AggregateSignatures(bls381Sigs)
	if err != nil {
		return nil, err
	}

	return &signatureAdapter{result}, nil
}

// BatchVerify verifies multiple signatures in a single batch operation
func (s *Scheme) BatchVerify(publicKeys []signing.PublicKey, message []byte, signatures []signing.Signature) (bool, error) {
	// Convert generic public keys to BLS381 specific public keys
	bls381PubKeys := make([]*PublicKey, len(publicKeys))
	for i, pubKey := range publicKeys {
		bls381PubKey, ok := pubKey.(*publicKeyAdapter)
		if !ok {
			return false, signing.ErrInvalidPublicKeyType
		}
		bls381PubKeys[i] = bls381PubKey.pk
	}

	// Convert generic signatures to BLS381 specific signatures
	bls381Sigs := make([]*Signature, len(signatures))
	for i, sig := range signatures {
		bls381Sig, ok := sig.(*signatureAdapter)
		if !ok {
			return false, signing.ErrInvalidSignatureType
		}
		bls381Sigs[i] = bls381Sig.sig
	}

	return BatchVerify(bls381PubKeys, message, bls381Sigs)
}

// AggregateVerify verifies an aggregated signature against multiple public keys and multiple messages
func (s *Scheme) AggregateVerify(publicKeys []signing.PublicKey, messages [][]byte, aggSignature signing.Signature) (bool, error) {
	// Convert generic public keys to BLS381 specific public keys
	bls381PubKeys := make([]*PublicKey, len(publicKeys))
	for i, pubKey := range publicKeys {
		bls381PubKey, ok := pubKey.(*publicKeyAdapter)
		if !ok {
			return false, signing.ErrInvalidPublicKeyType
		}
		bls381PubKeys[i] = bls381PubKey.pk
	}

	// Convert generic signature to BLS381 specific signature
	bls381Sig, ok := aggSignature.(*signatureAdapter)
	if !ok {
		return false, signing.ErrInvalidSignatureType
	}

	return AggregateVerify(bls381PubKeys, messages, bls381Sig.sig)
}

// Adapter types for implementing the generic interfaces

// privateKeyAdapter adapts the BLS381 private key to the generic interface
type privateKeyAdapter struct {
	pk *PrivateKey
}

// Sign implements the signing.PrivateKey interface
func (a *privateKeyAdapter) Sign(message []byte) (signing.Signature, error) {
	sig, err := a.pk.Sign(message)
	if err != nil {
		return nil, err
	}
	return &signatureAdapter{sig}, nil
}

// Public implements the signing.PrivateKey interface
func (a *privateKeyAdapter) Public() signing.PublicKey {
	return &publicKeyAdapter{a.pk.Public()}
}

// Bytes implements the signing.PrivateKey interface
func (a *privateKeyAdapter) Bytes() []byte {
	return a.pk.Bytes()
}

// publicKeyAdapter adapts the BLS381 public key to the generic interface
type publicKeyAdapter struct {
	pk *PublicKey
}

// Bytes implements the signing.PublicKey interface
func (a *publicKeyAdapter) Bytes() []byte {
	return a.pk.Bytes()
}

// signatureAdapter adapts the BLS381 signature to the generic interface
type signatureAdapter struct {
	sig *Signature
}

// Verify implements the signing.Signature interface
func (a *signatureAdapter) Verify(publicKey signing.PublicKey, message []byte) (bool, error) {
	bls381PubKey, ok := publicKey.(*publicKeyAdapter)
	if !ok {
		return false, signing.ErrInvalidPublicKeyType
	}

	return a.sig.Verify(bls381PubKey.pk, message)
}

// Bytes implements the signing.Signature interface
func (a *signatureAdapter) Bytes() []byte {
	return a.sig.Bytes()
}

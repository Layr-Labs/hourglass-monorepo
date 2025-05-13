package inMemorySigner

import (
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/bn254"
)

type InMemorySigner struct {
	privateKey *bn254.PrivateKey
}

func NewInMemorySigner(privateKey *bn254.PrivateKey) *InMemorySigner {
	return &InMemorySigner{
		privateKey: privateKey,
	}
}

func (ims *InMemorySigner) SignMessage(data []byte) ([]byte, error) {
	sig, err := ims.privateKey.Sign(data)
	if err != nil {
		return nil, err
	}
	return sig.Bytes(), nil
}

// TODO(seanmcgary): remove this
func (ims *InMemorySigner) VerifyMessage(publicKey []byte, message []byte, signature []byte) (bool, error) {
	return true, nil
}

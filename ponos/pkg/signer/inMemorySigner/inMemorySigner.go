package inMemorySigner

import (
	"fmt"
	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	"github.com/Layr-Labs/crypto-libs/pkg/ecdsa"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
)

type InMemorySigner struct {
	privateKey interface{}
	curveType  config.CurveType
}

func NewInMemorySigner(privateKey interface{}, curveType config.CurveType) *InMemorySigner {
	return &InMemorySigner{
		privateKey: privateKey,
		curveType:  curveType,
	}
}

func (ims *InMemorySigner) SignMessage(data []byte) ([]byte, error) {
	if ims.curveType == config.CurveTypeBN254 {
		pk := ims.privateKey.(*bn254.PrivateKey)
		sig, err := pk.Sign(data)
		if err != nil {
			return nil, err
		}
		return sig.Bytes(), nil
	}
	if ims.curveType == config.CurveTypeECDSA {
		pk := ims.privateKey.(*ecdsa.PrivateKey)
		sig, err := pk.Sign(data)
		if err != nil {
			return nil, err
		}
		return sig.Bytes(), nil
	}
	return nil, fmt.Errorf("SignMessage is not implemented for curve type %s", ims.curveType)
}

func (ims *InMemorySigner) SignMessageForSolidity(data [32]byte) ([]byte, error) {
	if ims.curveType == config.CurveTypeBN254 {
		pk := ims.privateKey.(*bn254.PrivateKey)
		sig, err := pk.SignSolidityCompatible(data)
		if err != nil {
			return nil, err
		}
		return sig.Bytes(), nil
	}
	if ims.curveType == config.CurveTypeECDSA {
		pk := ims.privateKey.(*ecdsa.PrivateKey)
		sig, err := pk.Sign(data[:])
		if err != nil {
			return nil, err
		}
		return sig.Bytes(), nil
	}
	return nil, fmt.Errorf("SignMessageForSolidity is not implemented for curve type %s", ims.curveType)
}

package signer

import (
	"fmt"
	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	"github.com/Layr-Labs/crypto-libs/pkg/ecdsa"
	"github.com/ethereum/go-ethereum/crypto"
)

type InMemorySigner struct {
	privateKey interface{}
	curveType  CurveType
}

func NewInMemorySigner(privateKey interface{}, curveType CurveType) *InMemorySigner {
	return &InMemorySigner{
		privateKey: privateKey,
		curveType:  curveType,
	}
}

func (ims *InMemorySigner) SignMessage(data []byte) ([]byte, error) {
	hashedData := getKeccak256Digest(data)

	if ims.curveType == CurveTypeBN254 {
		pk := ims.privateKey.(*bn254.PrivateKey)
		sig, err := pk.Sign(hashedData[:])
		if err != nil {
			return nil, err
		}
		return sig.Bytes(), nil
	}
	if ims.curveType == CurveTypeECDSA {
		pk := ims.privateKey.(*ecdsa.PrivateKey)
		sig, err := pk.Sign(hashedData[:])
		if err != nil {
			return nil, err
		}
		return sig.Bytes(), nil
	}
	return nil, fmt.Errorf("SignMessage is not implemented for curve type %s", ims.curveType)
}

func (ims *InMemorySigner) SignMessageForSolidity(data []byte) ([]byte, error) {

	var hashedData [32]byte
	if len(data) == 32 {
		copy(hashedData[:], data)
	} else {
		hashedData = crypto.Keccak256Hash(data)
	}

	if ims.curveType == CurveTypeBN254 {
		pk := ims.privateKey.(*bn254.PrivateKey)
		sig, err := pk.SignSolidityCompatible(hashedData)
		if err != nil {
			return nil, err
		}
		return sig.Bytes(), nil
	}
	if ims.curveType == CurveTypeECDSA {
		pk := ims.privateKey.(*ecdsa.PrivateKey)
		sig, err := pk.Sign(hashedData[:])
		if err != nil {
			return nil, err
		}
		return sig.Bytes(), nil
	}
	return nil, fmt.Errorf("SignMessageForSolidity is not implemented for curve type %s", ims.curveType)
}

func getKeccak256Digest(input []byte) [32]byte {
	digest := crypto.Keccak256(input)
	return [32]byte(digest)
}

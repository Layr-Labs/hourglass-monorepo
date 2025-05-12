package crypto

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"strings"
)

func StringToECDSAPrivateKey(pk string) (*ecdsa.PrivateKey, error) {
	pk = strings.TrimPrefix(pk, "0x")

	privateKey, err := crypto.HexToECDSA(pk)
	if err != nil {
		return nil, fmt.Errorf("failed to convert hex string to ECDSA private key: %v", err)
	}
	return privateKey, nil
}

func DeriveAddress(pk *ecdsa.PrivateKey) common.Address {
	return crypto.PubkeyToAddress(pk.PublicKey)
}

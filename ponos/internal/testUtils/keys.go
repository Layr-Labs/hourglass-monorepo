package testUtils

import (
	"fmt"
	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	cryptoLibsEcdsa "github.com/Layr-Labs/crypto-libs/pkg/ecdsa"
	"github.com/Layr-Labs/crypto-libs/pkg/keystore"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/ethereum/go-ethereum/common"
	"testing"
)

type WrappedKeyPair struct {
	PrivateKey interface{}    // This can be *bn254.PrivateKey or *ecdsa.PrivateKeyConfig
	PublicKey  interface{}    // This can be *bn254.PublicKey or *ecdsa.PublicKey
	Address    common.Address // This can be a string or a common.Address type
}

func GetKeysForCurveType(t *testing.T, curve config.CurveType, chainConfig *ChainConfig) (*WrappedKeyPair, *WrappedKeyPair, config.CurveType, error) {
	if curve == config.CurveTypeBN254 {
		aggPrivateKey, aggPublicKey, err := bn254.GenerateKeyPair()
		if err != nil {
			t.Fatalf("Failed to generate key pair: %v", err)
		}

		execPrivateKey, execPublicKey, err := bn254.GenerateKeyPair()
		if err != nil {
			t.Fatalf("Failed to generate key pair: %v", err)
		}
		return &WrappedKeyPair{
				PrivateKey: aggPrivateKey,
				PublicKey:  aggPublicKey,
			}, &WrappedKeyPair{
				PrivateKey: execPrivateKey,
				PublicKey:  execPublicKey,
			}, curve, nil
	}
	if curve == config.CurveTypeECDSA {
		aggPrivateKey, err := cryptoLibsEcdsa.NewPrivateKeyFromHexString(chainConfig.OperatorAccountPrivateKey)
		if err != nil {
			t.Fatalf("Failed to parse key pair: %v", err)
		}
		derivedAggAddress, err := aggPrivateKey.DeriveAddress()
		if err != nil {
			t.Fatalf("Failed to derive address: %v", err)
		}

		execPrivateKey, err := cryptoLibsEcdsa.NewPrivateKeyFromHexString(chainConfig.ExecOperatorAccountPk)
		if err != nil {
			t.Fatalf("Failed to generate key pair: %v", err)
		}
		derivedExecAddress, err := execPrivateKey.DeriveAddress()
		if err != nil {
			t.Fatalf("Failed to derive address: %v", err)
		}
		return &WrappedKeyPair{
				PrivateKey: aggPrivateKey,
				Address:    derivedAggAddress,
				PublicKey:  chainConfig.OperatorAccountPublicKey,
			}, &WrappedKeyPair{
				PrivateKey: execPrivateKey,
				Address:    derivedExecAddress,
				PublicKey:  chainConfig.ExecOperatorAccountPublicKey,
			}, curve, nil
	}
	return nil, nil, curve, fmt.Errorf("unsupported curve type: %s", curve)
}

func ParseKeysFromConfig(
	operatorConfig *config.OperatorConfig,
	curveType config.CurveType,
) (*bn254.PrivateKey, *cryptoLibsEcdsa.PrivateKey, interface{}, error) {
	var genericExecutorSigningKey interface{}
	var bn254PrivateSigningKey *bn254.PrivateKey
	var ecdsaPrivateSigningKey *cryptoLibsEcdsa.PrivateKey
	var err error

	if curveType == config.CurveTypeBN254 {
		storedKeys, err := keystore.ParseKeystoreJSON(operatorConfig.SigningKeys.BLS.Keystore)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to parse keystore JSON: %w", err)
		}

		bn254PrivateSigningKey, err = storedKeys.GetBN254PrivateKey(operatorConfig.SigningKeys.BLS.Password)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to get private key: %w", err)
		}
		genericExecutorSigningKey = bn254PrivateSigningKey
	} else if curveType == config.CurveTypeECDSA {
		ecdsaPrivateSigningKey, err = cryptoLibsEcdsa.NewPrivateKeyFromHexString(operatorConfig.SigningKeys.ECDSA.PrivateKey)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to get ECDSA private key: %w", err)
		}
		genericExecutorSigningKey = ecdsaPrivateSigningKey
	} else {
		return nil, nil, nil, fmt.Errorf("unsupported curve type: %s", curveType)
	}
	return bn254PrivateSigningKey, ecdsaPrivateSigningKey, genericExecutorSigningKey, nil
}

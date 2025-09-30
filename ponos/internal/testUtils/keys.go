package testUtils

import (
	"fmt"
	"testing"

	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	cryptoLibsEcdsa "github.com/Layr-Labs/crypto-libs/pkg/ecdsa"
	"github.com/Layr-Labs/crypto-libs/pkg/keystore"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/ethereum/go-ethereum/common"
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

func GetKeysForCurveTypeFromChainConfig(
	t *testing.T,
	aggCurveType config.CurveType,
	execCurveType config.CurveType,
	chainConfig *ChainConfig,
) (*WrappedKeyPair, []*WrappedKeyPair, error) {
	// Generate aggregator keys
	var aggKeys *WrappedKeyPair
	if aggCurveType == config.CurveTypeBN254 {
		aggPrivateKey, aggPublicKey, err := bn254.GenerateKeyPair()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate aggregator BN254 key pair: %w", err)
		}
		aggKeys = &WrappedKeyPair{
			PrivateKey: aggPrivateKey,
			PublicKey:  aggPublicKey,
		}
	} else if aggCurveType == config.CurveTypeECDSA {
		aggPrivateKey, err := cryptoLibsEcdsa.NewPrivateKeyFromHexString(chainConfig.OperatorAccountPrivateKey)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse aggregator ECDSA key: %w", err)
		}
		derivedAggAddress, err := aggPrivateKey.DeriveAddress()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to derive aggregator address: %w", err)
		}
		aggKeys = &WrappedKeyPair{
			PrivateKey: aggPrivateKey,
			Address:    derivedAggAddress,
			PublicKey:  chainConfig.OperatorAccountPublicKey,
		}
	} else {
		return nil, nil, fmt.Errorf("unsupported aggregator curve type: %s", aggCurveType)
	}

	// Generate executor keys for 4 operators
	execKeys := make([]*WrappedKeyPair, 4)

	if execCurveType == config.CurveTypeBN254 {
		for i := 0; i < 4; i++ {
			execPrivateKey, execPublicKey, err := bn254.GenerateKeyPair()
			if err != nil {
				return nil, nil, fmt.Errorf("failed to generate executor %d BN254 key pair: %w", i+1, err)
			}
			execKeys[i] = &WrappedKeyPair{
				PrivateKey: execPrivateKey,
				PublicKey:  execPublicKey,
			}
		}
	} else if execCurveType == config.CurveTypeECDSA {
		// Load keys from chain config for 4 executors
		execPrivateKeyHexes := []string{
			chainConfig.ExecOperatorAccountPk,
			chainConfig.ExecOperator2AccountPk,
			chainConfig.ExecOperator3AccountPk,
			chainConfig.ExecOperator4AccountPk,
		}
		execPublicKeys := []string{
			chainConfig.ExecOperatorAccountPublicKey,
			chainConfig.ExecOperator2AccountPublicKey,
			chainConfig.ExecOperator3AccountPublicKey,
			chainConfig.ExecOperator4AccountPublicKey,
		}

		for i := 0; i < 4; i++ {
			execPrivateKey, err := cryptoLibsEcdsa.NewPrivateKeyFromHexString(execPrivateKeyHexes[i])
			if err != nil {
				return nil, nil, fmt.Errorf("failed to parse executor %d ECDSA key: %w", i+1, err)
			}
			derivedExecAddress, err := execPrivateKey.DeriveAddress()
			if err != nil {
				return nil, nil, fmt.Errorf("failed to derive executor %d address: %w", i+1, err)
			}
			execKeys[i] = &WrappedKeyPair{
				PrivateKey: execPrivateKey,
				Address:    derivedExecAddress,
				PublicKey:  execPublicKeys[i],
			}
		}
	} else {
		return nil, nil, fmt.Errorf("unsupported executor curve type: %s", execCurveType)
	}

	return aggKeys, execKeys, nil
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

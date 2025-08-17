package signer

import (
	"fmt"
	"github.com/Layr-Labs/crypto-libs/pkg/ecdsa"
	"github.com/Layr-Labs/crypto-libs/pkg/keystore"
	gethkeystore "github.com/ethereum/go-ethereum/accounts/keystore"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

type ISigner interface {
	SignMessage(data []byte) ([]byte, error)
	SignMessageForSolidity(data []byte) ([]byte, error)
}

type Signers struct {
	ECDSASigner ISigner
	BLSSigner   ISigner
}

type SigningKeys struct {
	BN254 *KeystoreReference `json:"bn254"`
	ECDSA *ECDSAKeyConfig    `json:"ecdsa"`
}

type ECDSAKeyConfig struct {
	RemoteSignerConfig *RemoteSignerReference `json:"remoteSignerConfig" yaml:"remoteSignerConfig,omitempty"`
	Keystore           *KeystoreReference     `json:"keystore" yaml:"keystore,omitempty"`
	PrivateKey         bool                   `json:"privateKey" yaml:"privateKey,omitempty"`
}

type SigningKey struct {
	Keystore     string `json:"keystore"`
	KeystoreFile string `json:"keystoreFile"`
	Password     string `json:"password"`
}

type RemoteSignerConfig struct {
	Url         string `json:"url" yaml:"url"`
	CACert      string `json:"caCert" yaml:"caCert"`
	Cert        string `json:"cert" yaml:"cert"`
	Key         string `json:"key" yaml:"key"`
	FromAddress string `json:"fromAddress" yaml:"fromAddress"`
	PublicKey   string `json:"publicKey" yaml:"publicKey"`
}

type KeystoreReference struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
	Type string `yaml:"type"`
}

type RemoteSignerReference struct {
	Name           string `yaml:"name"`
	ConfigPath     string `yaml:"configPath,omitempty"`
	Url            string `yaml:"url,omitempty"`
	CACertPath     string `yaml:"caCertPath,omitempty"`
	ClientCertPath string `yaml:"clientCertPath,omitempty"`
	ClientKeyPath  string `yaml:"clientKeyPath,omitempty"`
}

type CurveType string

func (c CurveType) String() string {
	return string(c)
}
func (c CurveType) Uint8() (uint8, error) {
	return ConvertCurveTypeToSolidityEnum(c)
}

const (
	CurveTypeUnknown CurveType = "unknown"
	CurveTypeECDSA   CurveType = "ecdsa"
	CurveTypeBN254   CurveType = "bn254" // BN254 is the only supported curve type for now
)

func ConvertCurveTypeToSolidityEnum(curveType CurveType) (uint8, error) {
	switch curveType {
	case CurveTypeUnknown:
		return 0, nil
	case CurveTypeECDSA:
		return 1, nil
	case CurveTypeBN254:
		return 2, nil
	default:
		return 0, fmt.Errorf("unsupported curve type: %s", curveType)
	}
}

func ConvertSolidityEnumToCurveType(enumValue uint8) (CurveType, error) {
	switch enumValue {
	case 0:
		return CurveTypeUnknown, nil
	case 1:
		return CurveTypeECDSA, nil
	case 2:
		return CurveTypeBN254, nil
	default:
		return "", fmt.Errorf("unsupported curve type enum value: %d", enumValue)
	}
}

func LoadKeystoreSigner(ks *KeystoreReference) (ISigner, error) {
	// Find the keystore reference
	keystorePath := ks.Path
	password := os.Getenv("KEYSTORE_PASSWORD")

	switch ks.Type {
	case "bn254":

		storedKeys, err := keystore.LoadKeystoreFile(keystorePath)
		if err != nil {
			return nil, fmt.Errorf("failed to load keystore file '%s': %w", keystorePath, err)
		}
		privateKey, err := storedKeys.GetBN254PrivateKey(password)
		if err != nil {
			return nil, fmt.Errorf("failed to get BN254 private key: %w", err)
		}
		return NewInMemorySigner(privateKey, CurveTypeBN254), nil

	case "ecdsa":

		keyStoreContents, err := os.ReadFile(filepath.Clean(keystorePath))
		if err != nil {
			return nil, err
		}

		key, err := gethkeystore.DecryptKey(keyStoreContents, password)
		if err != nil {
			return nil, err
		}
		return NewInMemorySigner(key.PrivateKey, CurveTypeECDSA), nil

	default:
		return nil, fmt.Errorf("unsupported keystore type: %s", ks.Type)
	}
}

func LoadWeb3SignerConfig(keys *RemoteSignerReference) (*RemoteSignerConfig, error) {
	if keys == nil {
		return nil, fmt.Errorf("remote signer not configured")
	}

	var signerConfig *RemoteSignerConfig

	if keys.Url == "" {
		return nil, fmt.Errorf("url is required for remote signer")
	}
	signerConfig.Url = keys.Url

	if keys.ConfigPath != "" {
		configPath := keys.ConfigPath
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read web3signer config: %w", err)
		}

		signerConfig = &RemoteSignerConfig{}
		if err := yaml.Unmarshal(data, signerConfig); err != nil {
			return nil, fmt.Errorf("failed to parse web3signer config: %w", err)
		}
		return signerConfig, nil
	}

	if keys.CACertPath != "" {
		caCert, err := loadCertFile(keys.CACertPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load CA cert: %w", err)
		}
		signerConfig.CACert = caCert
	}

	if keys.ClientCertPath != "" {
		clientCert, err := loadCertFile(keys.ClientCertPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load client cert: %w", err)
		}
		signerConfig.Cert = clientCert
	}

	if keys.ClientKeyPath != "" {
		clientKey, err := loadCertFile(keys.ClientKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load client key: %w", err)
		}
		signerConfig.Key = clientKey
	}

	return signerConfig, nil
}

func LoadPrivateKeySigner() (ISigner, error) {
	privateKey := os.Getenv("PRIVATE_KEY")
	if privateKey == "" {
		return nil, fmt.Errorf("private key not found")
	}

	ecdsaPk, err := ecdsa.NewPrivateKeyFromHexString(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create ECDSA private key: %w", err)
	}

	return NewInMemorySigner(ecdsaPk, CurveTypeECDSA), nil
}

func loadCertFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

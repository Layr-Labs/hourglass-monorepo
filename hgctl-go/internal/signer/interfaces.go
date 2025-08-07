package signer

import (
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
)

// SignerType represents the type of signer configuration
type SignerType string

const (
	SignerTypeWeb3Signer SignerType = "web3signer"
	SignerTypeKeystore   SignerType = "keystore"
	SignerTypePrivateKey SignerType = "privatekey"
)

// SignerConfig represents the configuration for a signer
type SignerConfig struct {
	Type SignerType

	// For Web3Signer
	Web3SignerURL       string
	Web3SignerPublicKey string
	Web3SignerCA        string
	Web3SignerCert      string
	Web3SignerKey       string

	// For Keystore
	KeystorePath     string
	KeystoreContent  string
	KeystorePassword string

	// For PrivateKey
	PrivateKey string
}

// SignerResolver determines which signer configuration to use based on context
type SignerResolver interface {
	// ResolveSignerConfig determines the appropriate signer configuration
	// based on the context and available credentials
	ResolveSignerConfig(ctx *config.Context, signerType string, password string) (*SignerConfig, error)
}

// PasswordProvider provides passwords for keystores
type PasswordProvider interface {
	// GetPassword returns a password for the given keystore name
	// It may prompt the user or read from stdin
	GetPassword(keystoreName string) (string, error)
}

// ConfigBuilder builds configuration files from templates and signer configs
type ConfigBuilder interface {
	// BuildExecutorConfig builds an executor configuration file
	BuildExecutorConfig(signerConfigs map[string]*SignerConfig, envVars map[string]string) ([]byte, error)

	// BuildAggregatorConfig builds an aggregator configuration file
	BuildAggregatorConfig(signerConfigs map[string]*SignerConfig, envVars map[string]string) ([]byte, error)
}

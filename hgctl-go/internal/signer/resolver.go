package signer

import (
	"fmt"
	"os"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
)

// DefaultSignerResolver implements SignerResolver with the preference order:
// 1. Web3Signer
// 2. Keystore
// 3. Private Key (from environment variable)
type DefaultSignerResolver struct {
	passwordProvider PasswordProvider
}

func NewDefaultSignerResolver(passwordProvider PasswordProvider) *DefaultSignerResolver {
	return &DefaultSignerResolver{
		passwordProvider: passwordProvider,
	}
}

func (r *DefaultSignerResolver) ResolveSignerConfig(ctx *config.Context, signerType string, password string) (*SignerConfig, error) {
	// Determine the key name based on signer type (BLS or ECDSA)
	var keyName string
	switch signerType {
	case "BLS", "bls":
		keyName = "BLS"
	case "ECDSA", "ecdsa":
		keyName = "ECDSA"
	default:
		return nil, fmt.Errorf("unsupported signer type: %s", signerType)
	}

	// 1. Check for Web3Signer configuration
	for _, ws := range ctx.Web3Signers {
		// Check if this web3signer is for the requested key type
		// This could be based on naming convention or configuration
		if ws.Name == keyName || ws.Name == fmt.Sprintf("%s_SIGNER", keyName) {
			config := &SignerConfig{
				Type: SignerTypeWeb3Signer,
			}

			// Read the web3signer URL from config file or environment
			web3SignerURL := os.Getenv(fmt.Sprintf("%s_WEB3SIGNER_URL", keyName))
			if web3SignerURL == "" && ws.ConfigPath != "" {
				// TODO: Parse web3signer config file to get URL
				// For now, we'll require it from environment
				return nil, fmt.Errorf("web3signer URL not found in environment for %s", keyName)
			}
			config.Web3SignerURL = web3SignerURL

			// Get public key from environment
			config.Web3SignerPublicKey = os.Getenv(fmt.Sprintf("%s_WEB3SIGNER_PUBLIC_KEY", keyName))

			// Load TLS certificates if configured
			if ws.CACertPath != "" {
				caContent, err := os.ReadFile(ws.CACertPath)
				if err != nil {
					return nil, fmt.Errorf("failed to read CA cert: %w", err)
				}
				config.Web3SignerCA = string(caContent)
			}

			if ws.ClientCertPath != "" {
				certContent, err := os.ReadFile(ws.ClientCertPath)
				if err != nil {
					return nil, fmt.Errorf("failed to read client cert: %w", err)
				}
				config.Web3SignerCert = string(certContent)
			}

			if ws.ClientKeyPath != "" {
				keyContent, err := os.ReadFile(ws.ClientKeyPath)
				if err != nil {
					return nil, fmt.Errorf("failed to read client key: %w", err)
				}
				config.Web3SignerKey = string(keyContent)
			}

			return config, nil
		}
	}

	// 2. Check for Keystore configuration
	for _, ks := range ctx.Keystores {
		// Match keystore by type and name
		if (signerType == "BLS" && ks.Type == "bn254") ||
			(signerType == "ECDSA" && ks.Type == "ecdsa") {
			// Check if this is the primary keystore for this type
			if ks.Name == keyName || ks.Name == fmt.Sprintf("%s_KEY", keyName) {
				config := &SignerConfig{
					Type:         SignerTypeKeystore,
					KeystorePath: ks.Path,
				}

				// Load keystore content
				keystoreContent, err := os.ReadFile(ks.Path)
				if err != nil {
					return nil, fmt.Errorf("failed to read keystore file: %w", err)
				}
				config.KeystoreContent = string(keystoreContent)

				// Use provided password or get from provider
				if password != "" {
					config.KeystorePassword = password
				} else if r.passwordProvider != nil {
					pwd, err := r.passwordProvider.GetPassword(ks.Name)
					if err != nil {
						return nil, fmt.Errorf("failed to get keystore password: %w", err)
					}
					config.KeystorePassword = pwd
				} else {
					// Try empty password as default
					config.KeystorePassword = ""
				}

				return config, nil
			}
		}
	}

	// 3. Check for Private Key in environment
	privateKeyEnv := fmt.Sprintf("%s_PRIVATE_KEY", keyName)
	if privateKey := os.Getenv(privateKeyEnv); privateKey != "" {
		return &SignerConfig{
			Type:       SignerTypePrivateKey,
			PrivateKey: privateKey,
		}, nil
	}

	return nil, fmt.Errorf("no signer configuration found for %s", keyName)
}

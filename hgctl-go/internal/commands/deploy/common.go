package deploy

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
)

// injectFileContentsAsEnvVars reads keystore and certificate files and injects them as environment variables
func injectFileContentsAsEnvVars(dockerArgs []string, contextDir string, log logger.Logger) []string {
	fileEnvMappings := map[string]string{
		// BLS keystore
		"operator.bls.keystore.json": "BLS_KEYSTORE_CONTENT",
		// ECDSA keystore  
		"operator.ecdsa.keystore.json": "ECDSA_KEYSTORE_CONTENT",
		// Web3 signer certificates for BLS
		"web3signer-bls-ca.crt": "WEB3_SIGNER_BLS_CA_CERT_CONTENT",
		"web3signer-bls-client.crt": "WEB3_SIGNER_BLS_CLIENT_CERT_CONTENT",
		"web3signer-bls-client.key": "WEB3_SIGNER_BLS_CLIENT_KEY_CONTENT",
		// Web3 signer certificates for ECDSA
		"web3signer-ecdsa-ca.crt": "WEB3_SIGNER_ECDSA_CA_CERT_CONTENT",
		"web3signer-ecdsa-client.crt": "WEB3_SIGNER_ECDSA_CLIENT_CERT_CONTENT",
		"web3signer-ecdsa-client.key": "WEB3_SIGNER_ECDSA_CLIENT_KEY_CONTENT",
	}
	
	for fileName, envVar := range fileEnvMappings {
		filePath := filepath.Join(contextDir, fileName)
		if content, err := os.ReadFile(filePath); err == nil {
			// File exists, inject its content
			dockerArgs = append(dockerArgs, "-e", fmt.Sprintf("%s=%s", envVar, string(content)))
			log.Debug("Injected file content as environment variable", 
				zap.String("file", fileName),
				zap.String("envVar", envVar))
		}
	}
	
	return dockerArgs
}
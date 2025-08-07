package templates

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/runtime"
)

//go:embed aggregator-config.yaml
var aggregatorConfigTemplate string

//go:embed executor-config.yaml
var executorConfigTemplate string

// ConfigGenerator handles the generation of configuration files from templates
type ConfigGenerator struct {
	contextDir string
	envMap     map[string]string
}

// NewConfigGenerator creates a new config generator
func NewConfigGenerator(contextDir string, envMap map[string]string) *ConfigGenerator {
	return &ConfigGenerator{
		contextDir: contextDir,
		envMap:     envMap,
	}
}

// GenerateAggregatorConfig generates aggregator configuration from template
func (g *ConfigGenerator) GenerateAggregatorConfig(outputPath string) error {
	// Check required environment variables
	required := []string{"OPERATOR_ADDRESS", "L1_CHAIN_ID", "L1_RPC_URL"}
	missing := []string{}

	for _, key := range required {
		if g.getEnvValue(key) == "" {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %v", missing)
	}

	return g.generateConfig("aggregator", aggregatorConfigTemplate, outputPath)
}

// GenerateExecutorConfig generates executor configuration from template
func (g *ConfigGenerator) GenerateExecutorConfig(outputPath string) error {
	// Check required environment variables
	required := []string{"OPERATOR_ADDRESS", "L1_CHAIN_ID", "L1_RPC_URL", "PERFORMER_REGISTRY"}
	missing := []string{}

	for _, key := range required {
		if g.getEnvValue(key) == "" {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %v", missing)
	}

	return g.generateConfig("executor", executorConfigTemplate, outputPath)
}

// generateConfig generates a configuration file from a template
func (g *ConfigGenerator) generateConfig(name, templateContent, outputPath string) error {
	// Create custom template functions
	funcMap := template.FuncMap{
		"env": func(key string) string {
			return g.getEnvValue(key)
		},
		"envDefault": func(key, defaultValue string) string {
			if val := g.getEnvValue(key); val != "" {
				return val
			}
			return defaultValue
		},
		"isTrue": func(key string) bool {
			val := g.getEnvValue(key)
			return val == "true" || val == "1" || val == "yes"
		},
		"indent": func(spaces int, v string) string {
			pad := ""
			for i := 0; i < spaces; i++ {
				pad += " "
			}
			return pad + strings.ReplaceAll(v, "\n", "\n"+pad)
		},
	}

	// Parse template
	tmpl, err := template.New(name).Funcs(funcMap).Parse(templateContent)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Create output directory if needed
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create output file
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	// Execute template
	if err := tmpl.Execute(f, g.envMap); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

// getEnvValue gets environment variable value with substitution
func (g *ConfigGenerator) getEnvValue(key string) string {
	// First check the provided envMap
	if val, ok := g.envMap[key]; ok {
		return val
	}
	// Then check OS environment
	return os.Getenv(key)
}

// CopyKeystoreFiles copies keystore and certificate files from context directory to destination
func (g *ConfigGenerator) CopyKeystoreFiles(destDir string) error {
	// List of files to copy - includes all possible naming variations
	files := []string{
		// Keystore files
		"operator.bls.keystore.json",
		"operator.ecdsa.keystore.json",
		// Web3 signer certificates for BLS
		"web3signer-bls-ca.crt",
		"web3signer-bls-client.crt",
		"web3signer-bls-client.key",
		// Web3 signer certificates for ECDSA
		"web3signer-ecdsa-ca.crt",
		"web3signer-ecdsa-client.crt",
		"web3signer-ecdsa-client.key",
		// Generic naming (backwards compatibility)
		"web3signer-ca.crt",
		"web3signer-client.crt",
		"web3signer-client.key",
	}

	// Create destination directory
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create keystore directory: %w", err)
	}

	// Copy each file if it exists
	for _, file := range files {
		srcPath := filepath.Join(g.contextDir, file)
		destPath := filepath.Join(destDir, file)

		// Check if source file exists
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			// File doesn't exist, skip it
			continue
		}

		// Read source file
		data, err := os.ReadFile(srcPath)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", file, err)
		}

		// Write to destination
		if err := os.WriteFile(destPath, data, 0600); err != nil {
			return fmt.Errorf("failed to write %s: %w", file, err)
		}
	}

	return nil
}

// LoadContextEnv loads environment variables from .hgctl/{context}/config.env
func LoadContextEnv(contextDir string) (map[string]string, error) {
	envFile := filepath.Join(contextDir, "config.env")

	// Check if file exists
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		// No env file, return empty map
		return make(map[string]string), nil
	}

	return runtime.LoadEnvFile(envFile)
}

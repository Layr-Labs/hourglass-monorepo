package templates

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/signer"
)

// ConfigBuilder implements the signer.ConfigBuilder interface
type ConfigBuilder struct{}

// NewConfigBuilder creates a new ConfigBuilder
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{}
}

// BuildExecutorConfig builds an executor configuration file
func (b *ConfigBuilder) BuildExecutorConfig(signerConfigs map[string]*signer.SignerConfig, envVars map[string]string) ([]byte, error) {
	funcMap := b.createFuncMap(signerConfigs, envVars)

	tmpl, err := template.New("executor").Funcs(funcMap).Parse(executorConfigTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse executor template: %w", err)
	}

	var buf bytes.Buffer
	data := struct {
		Env           map[string]string
		SignerConfigs map[string]*signer.SignerConfig
	}{
		Env:           envVars,
		SignerConfigs: signerConfigs,
	}

	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute executor template: %w", err)
	}

	return buf.Bytes(), nil
}

// BuildAggregatorConfig builds an aggregator configuration file
func (b *ConfigBuilder) BuildAggregatorConfig(signerConfigs map[string]*signer.SignerConfig, envVars map[string]string) ([]byte, error) {
	funcMap := b.createFuncMap(signerConfigs, envVars)

	tmpl, err := template.New("aggregator").Funcs(funcMap).Parse(aggregatorConfigTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse aggregator template: %w", err)
	}

	var buf bytes.Buffer
	data := struct {
		Env           map[string]string
		SignerConfigs map[string]*signer.SignerConfig
	}{
		Env:           envVars,
		SignerConfigs: signerConfigs,
	}

	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute aggregator template: %w", err)
	}

	return buf.Bytes(), nil
}

// createFuncMap creates the template function map with the provided context
func (b *ConfigBuilder) createFuncMap(signerConfigs map[string]*signer.SignerConfig, envVars map[string]string) template.FuncMap {
	return template.FuncMap{
		"env": func(key string) string {
			if val, ok := envVars[key]; ok {
				return val
			}
			return os.Getenv(key)
		},
		"envDefault": func(key, defaultValue string) string {
			if val, ok := envVars[key]; ok && val != "" {
				return val
			}
			if value := os.Getenv(key); value != "" {
				return value
			}
			return defaultValue
		},
		"required": func(key string) (string, error) {
			if val, ok := envVars[key]; ok && val != "" {
				return val, nil
			}
			value := os.Getenv(key)
			if value == "" {
				return "", fmt.Errorf("required environment variable %s is not set", key)
			}
			return value, nil
		},
		"hasWeb3Signer": func(signerType string) bool {
			if cfg, ok := signerConfigs[signerType]; ok {
				return cfg.Type == signer.SignerTypeWeb3Signer
			}
			return false
		},
		"hasKeystore": func(signerType string) bool {
			if cfg, ok := signerConfigs[signerType]; ok {
				return cfg.Type == signer.SignerTypeKeystore
			}
			return false
		},
		"hasPrivateKey": func(signerType string) bool {
			if cfg, ok := signerConfigs[signerType]; ok {
				return cfg.Type == signer.SignerTypePrivateKey
			}
			return false
		},
		"getSignerConfig": func(signerType string) *signer.SignerConfig {
			return signerConfigs[signerType]
		},
		"isTrue": func(key string) bool {
			val := ""
			if v, ok := envVars[key]; ok {
				val = v
			} else {
				val = os.Getenv(key)
			}
			return val == "true" || val == "1" || val == "yes"
		},
		"indent": func(spaces int, text string) string {
			padding := strings.Repeat(" ", spaces)
			lines := strings.Split(text, "\n")
			for i, line := range lines {
				if line != "" {
					lines[i] = padding + line
				}
			}
			return strings.Join(lines, "\n")
		},
	}
}

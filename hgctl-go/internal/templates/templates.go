package templates

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"text/template"
)

//go:embed aggregator-config.yaml
var aggregatorConfigTemplate string

//go:embed executor-config.yaml
var executorConfigTemplate string

// BuildExecutorConfig builds an executor configuration file
func BuildExecutorConfig(envVars map[string]string) ([]byte, error) {
	funcMap := createFuncMap(envVars)

	tmpl, err := template.New("executor").Funcs(funcMap).Parse(executorConfigTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse executor template: %w", err)
	}

	var buf bytes.Buffer
	data := struct {
		Env map[string]string
	}{
		Env: envVars,
	}

	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute executor template: %w", err)
	}

	return buf.Bytes(), nil
}

// BuildAggregatorConfig builds an aggregator configuration file
func BuildAggregatorConfig(envVars map[string]string) ([]byte, error) {
	funcMap := createFuncMap(envVars)

	tmpl, err := template.New("aggregator").Funcs(funcMap).Parse(aggregatorConfigTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse aggregator template: %w", err)
	}

	var buf bytes.Buffer
	data := struct {
		Env map[string]string
	}{
		Env: envVars,
	}

	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute aggregator template: %w", err)
	}

	return buf.Bytes(), nil
}

// createFuncMap creates the template function map with the provided context
func createFuncMap(envVars map[string]string) template.FuncMap {
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
	}
}

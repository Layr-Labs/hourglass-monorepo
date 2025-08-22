package middleware

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

// SecretsBeforeFunc loads environment variables from the secrets file configured in the context.
// This should run early in the middleware chain, before any middleware that needs secrets.
func SecretsBeforeFunc(c *cli.Context) error {
	log := GetLogger(c)

	// Get current context
	currentCtx, ok := c.Context.Value(config.ContextKey).(*config.Context)
	if !ok || currentCtx == nil {
		log.Debug("No context configured, skipping secrets loading")
		return nil
	}

	// Check if EnvSecretsPath is configured
	if currentCtx.EnvSecretsPath == "" {
		log.Debug("No secrets path configured in context")
		return nil
	}

	// Expand the path (handle ~/ for home directory)
	secretsPath := expandPath(currentCtx.EnvSecretsPath)

	// Check if file exists
	if _, err := os.Stat(secretsPath); os.IsNotExist(err) {
		log.Warn("Secrets file does not exist",
			zap.String("path", secretsPath),
			zap.String("originalPath", currentCtx.EnvSecretsPath))
		return nil // Don't fail, just warn
	}

	// Load the secrets file
	log.Debug("Loading secrets from file", zap.String("path", secretsPath))

	envVars, err := loadEnvFile(secretsPath)
	if err != nil {
		return fmt.Errorf("failed to load secrets from %s: %w", secretsPath, err)
	}

	// Set environment variables
	for key, value := range envVars {
		// Only set if not already set (allow command-line overrides)
		if os.Getenv(key) == "" {
			if err := os.Setenv(key, value); err != nil {
				log.Warn("Failed to set environment variable",
					zap.String("key", key),
					zap.Error(err))
				continue
			}
			log.Debug("Set environment variable from secrets",
				zap.String("key", key),
				zap.Bool("hasValue", value != ""))
		} else {
			log.Debug("Environment variable already set, skipping",
				zap.String("key", key))
		}
	}

	log.Debug("Loaded secrets from file",
		zap.String("path", secretsPath),
		zap.Int("count", len(envVars)))

	return nil
}

// loadEnvFile reads a .env file and returns a map of key-value pairs
func loadEnvFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	envVars := make(map[string]string)
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split on first = sign
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			// Skip malformed lines but don't fail
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove surrounding quotes if present
		value = strings.Trim(value, `"'`)

		// Validate key
		if key == "" {
			continue
		}

		// Handle special cases for multiline values or escaped characters if needed
		// For now, keep it simple
		envVars[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return envVars, nil
}

// expandPath expands ~ to the user's home directory
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			// If we can't get home dir, return original path
			return path
		}
		return filepath.Join(homeDir, path[2:])
	}
	return path
}

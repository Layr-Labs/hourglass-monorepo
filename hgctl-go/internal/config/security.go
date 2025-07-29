package config

import "strings"

// List of environment variable names that should never be persisted to config
var secretVariablePatterns = []string{
    "PRIVATE_KEY",
    "PASSWORD",
    "SECRET",
    "TOKEN",
    "API_KEY",
    "CERT",
    "KEY_PATH",
}

// IsSecretVariable checks if an environment variable name appears to contain secrets
func IsSecretVariable(name string) bool {
    upperName := strings.ToUpper(name)
    
    for _, pattern := range secretVariablePatterns {
        if strings.Contains(upperName, pattern) {
            return true
        }
    }
    
    return false
}

// FilterSecrets removes any environment variables that appear to be secrets
func FilterSecrets(vars map[string]string) map[string]string {
    filtered := make(map[string]string)
    
    for k, v := range vars {
        if !IsSecretVariable(k) {
            filtered[k] = v
        }
    }
    
    return filtered
}
package runtime

import (
	"os"
	"regexp"
	"strings"
)

// SubstituteEnvVars replaces environment variable placeholders in a string
// Supports formats: ${VAR} and ${VAR:default}
// Priority: provided envMap > OS environment > default value
func SubstituteEnvVars(value string, envMap map[string]string) string {
	// Regular expression to match ${VAR} or ${VAR:default}
	re := regexp.MustCompile(`\$\{([A-Z0-9_]+)(?::([^}]*))?\}`)

	return re.ReplaceAllStringFunc(value, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}

		varName := parts[1]
		defaultValue := ""
		if len(parts) > 2 {
			defaultValue = parts[2]
		}

		// First check provided envMap
		if val, ok := envMap[varName]; ok && val != "" {
			return val
		}

		// Then check OS environment
		if val := os.Getenv(varName); val != "" {
			return val
		}

		// Finally use default value
		return defaultValue
	})
}

// SubstituteComponentEnv replaces environment variables in a component spec
func SubstituteComponentEnv(component *ComponentSpec, envMap map[string]string) {
	for i := range component.Env {
		component.Env[i].Value = SubstituteEnvVars(component.Env[i].Value, envMap)
	}

	// Also substitute in command if needed
	for i := range component.Command {
		component.Command[i] = SubstituteEnvVars(component.Command[i], envMap)
	}
}

// LoadEnvFile reads environment variables from a file
// Format: KEY=VALUE (one per line, # for comments)
func LoadEnvFile(filename string) (map[string]string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	env := make(map[string]string)
	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// Remove surrounding quotes if present
			if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
				value = value[1 : len(value)-1]
			}

			env[key] = value
		}
	}

	return env, nil
}

// MergeEnvMaps merges multiple environment variable maps
// Later maps override earlier ones
func MergeEnvMaps(maps ...map[string]string) map[string]string {
	result := make(map[string]string)

	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}

	return result
}

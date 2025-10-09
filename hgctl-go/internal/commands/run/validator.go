package run

import (
	"fmt"
	"strings"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/runtime"
)

// ValidateComponentSpec validates that all required environment variables from the spec are present
func ValidateComponentSpec(component *runtime.ComponentSpec, envMap map[string]string) error {
	var missing []string

	// Check all environment variables marked as required in the spec
	for _, envVar := range component.Env {
		if envVar.Required {
			if value, exists := envMap[envVar.Name]; !exists || value == "" {
				missing = append(missing, envVar.Name)
			} else {
				envVar.Value = value
			}
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables:\n  - %s", strings.Join(missing, "\n  - "))
	}

	return nil
}

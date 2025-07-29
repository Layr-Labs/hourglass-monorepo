package deploy

import (
	"fmt"
	"strings"
)

// AggregatorRequirements - bare minimum environment variables required to deploy aggregator
var AggregatorRequirements = []string{
	"OPERATOR_ADDRESS",
	"OPERATOR_PRIVATE_KEY",
	"L1_CHAIN_ID",
	"L1_RPC_URL",
	"AVS_ADDRESS",
	"KEYSTORE_NAME",
	"KEYSTORE_PASSWORD",
}

// ExecutorRequirements - bare minimum environment variables required to deploy executor
var ExecutorRequirements = []string{
	"OPERATOR_ADDRESS",
	"OPERATOR_PRIVATE_KEY",
	"L1_CHAIN_ID",
	"L1_RPC_URL",
	"AVS_ADDRESS",
	"KEYSTORE_NAME",
	"KEYSTORE_PASSWORD",
}

// PerformerRequirements - bare minimum environment variables required to deploy performer
// Note: Performer requirements are mostly dynamic and come from the runtime spec
var PerformerRequirements = []string{
	"AVS_ADDRESS", // This is set by the system
}

// ValidateAggregator checks if all required environment variables are present
func ValidateAggregator(envMap map[string]string) error {
	var missing []string

	for _, req := range AggregatorRequirements {
		if envMap[req] == "" {
			missing = append(missing, req)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables:\n  - %s", strings.Join(missing, "\n  - "))
	}

	return nil
}

// ValidateExecutor checks if all required environment variables are present
func ValidateExecutor(envMap map[string]string) error {
	var missing []string

	for _, req := range ExecutorRequirements {
		if envMap[req] == "" {
			missing = append(missing, req)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables:\n  - %s", strings.Join(missing, "\n  - "))
	}

	return nil
}

// ValidatePerformer checks if all required environment variables are present
func ValidatePerformer(envMap map[string]string) error {
	var missing []string

	for _, req := range PerformerRequirements {
		if envMap[req] == "" {
			missing = append(missing, req)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables:\n  - %s", strings.Join(missing, "\n  - "))
	}

	return nil
}

package deploy

import (
	"strings"
	"testing"
)

func TestValidateAggregator(t *testing.T) {
	tests := []struct {
		name    string
		envMap  map[string]string
		wantErr bool
		missing []string
	}{
		{
			name: "all required variables present",
			envMap: map[string]string{
				"OPERATOR_ADDRESS":     "0x123",
				"OPERATOR_PRIVATE_KEY": "0xabc",
				"L1_CHAIN_ID":          "1",
				"L1_RPC_URL":           "https://eth.llamarpc.com",
				"AVS_ADDRESS":          "0x456",
				"KEYSTORE_NAME":        "test-keystore",
				"KEYSTORE_PASSWORD":    "test-password",
			},
			wantErr: false,
		},
		{
			name: "missing OPERATOR_ADDRESS",
			envMap: map[string]string{
				"OPERATOR_PRIVATE_KEY": "0xabc",
				"L1_CHAIN_ID":          "1",
				"L1_RPC_URL":           "https://eth.llamarpc.com",
				"AVS_ADDRESS":          "0x456",
				"KEYSTORE_NAME":        "test-keystore",
				"KEYSTORE_PASSWORD":    "test-password",
			},
			wantErr: true,
			missing: []string{"OPERATOR_ADDRESS"},
		},
		{
			name: "missing multiple variables",
			envMap: map[string]string{
				"L1_CHAIN_ID": "1",
				"L1_RPC_URL":  "https://eth.llamarpc.com",
			},
			wantErr: true,
			missing: []string{"OPERATOR_ADDRESS", "OPERATOR_PRIVATE_KEY", "AVS_ADDRESS", "KEYSTORE_NAME", "KEYSTORE_PASSWORD"},
		},
		{
			name:    "empty env map",
			envMap:  map[string]string{},
			wantErr: true,
			missing: []string{"OPERATOR_ADDRESS", "OPERATOR_PRIVATE_KEY", "L1_CHAIN_ID", "L1_RPC_URL", "AVS_ADDRESS", "KEYSTORE_NAME", "KEYSTORE_PASSWORD"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAggregator(tt.envMap)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAggregator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && len(tt.missing) > 0 {
				errStr := err.Error()
				for _, m := range tt.missing {
					if !strings.Contains(errStr, m) {
						t.Errorf("Expected error to contain %s, but got: %s", m, errStr)
					}
				}
			}
		})
	}
}

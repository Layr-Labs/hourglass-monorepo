package contracts

import (
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestNewContractStore(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		wantErr bool
	}{
		{
			name:    "empty environment",
			envVars: map[string]string{},
			wantErr: false,
		},
		{
			name: "with contract addresses",
			envVars: map[string]string{
				"TASKAVSREGISTRAR": "0x1234567890123456789012345678901234567890",
				"HELLOWORLDL1":     "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			},
			wantErr: false,
		},
		{
			name: "with mixed case environment variables",
			envVars: map[string]string{
				"TaskAVSRegistrar": "0x1234567890123456789012345678901234567890",
				"helloWorldL1":     "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Clearenv()

			// Set test environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			store, err := NewContractStore()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewContractStore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && store == nil {
				t.Error("NewContractStore() returned nil store")
			}
		})
	}
}

func TestContractStore_GetContract(t *testing.T) {
	tests := []struct {
		name         string
		envVars      map[string]string
		contractName string
		wantAddress  string
		wantErr      bool
	}{
		{
			name: "get existing contract",
			envVars: map[string]string{
				"HELLOWORLDL1": "0x1234567890123456789012345678901234567890",
			},
			contractName: "HELLOWORLDL1",
			wantAddress:  "0x1234567890123456789012345678901234567890",
			wantErr:      false,
		},
		{
			name: "get existing contract with lowercase input",
			envVars: map[string]string{
				"HELLOWORLDL1": "0x1234567890123456789012345678901234567890",
			},
			contractName: "helloworldl1",
			wantAddress:  "0x1234567890123456789012345678901234567890",
			wantErr:      false,
		},
		{
			name: "get existing contract with mixed case input",
			envVars: map[string]string{
				"HELLOWORLDL1": "0x1234567890123456789012345678901234567890",
			},
			contractName: "HelloWorldL1",
			wantAddress:  "0x1234567890123456789012345678901234567890",
			wantErr:      false,
		},
		{
			name:         "get non-existent contract",
			envVars:      map[string]string{},
			contractName: "NONEXISTENT",
			wantAddress:  "",
			wantErr:      true,
		},
		{
			name: "get contract with invalid address",
			envVars: map[string]string{
				"INVALID": "not-a-valid-address",
			},
			contractName: "INVALID",
			wantAddress:  "",
			wantErr:      true,
		},
		{
			name: "get contract with zero address",
			envVars: map[string]string{
				"ZEROCONTRACT": "0x0000000000000000000000000000000000000000",
			},
			contractName: "ZEROCONTRACT",
			wantAddress:  "0x0000000000000000000000000000000000000000",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Clearenv()

			// Set test environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			store, err := NewContractStore()
			if err != nil {
				t.Fatalf("Failed to create contract store: %v", err)
			}

			addr, err := store.GetContract(tt.contractName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetContract() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				expectedAddr := common.HexToAddress(tt.wantAddress)
				if addr != expectedAddr {
					t.Errorf("GetContract() = %v, want %v", addr, expectedAddr)
				}
			}
		})
	}
}

func TestContractStore_GetTaskAVSRegistrar(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		wantAddress string
		wantErr     bool
	}{
		{
			name: "get task AVS registrar",
			envVars: map[string]string{
				"TASKAVSREGISTRAR": "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			},
			wantAddress: "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			wantErr:     false,
		},
		{
			name:        "task AVS registrar not set",
			envVars:     map[string]string{},
			wantAddress: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Clearenv()

			// Set test environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			store, err := NewContractStore()
			if err != nil {
				t.Fatalf("Failed to create contract store: %v", err)
			}

			addr, err := store.GetTaskAVSRegistrar()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTaskAVSRegistrar() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				expectedAddr := common.HexToAddress(tt.wantAddress)
				if addr != expectedAddr {
					t.Errorf("GetTaskAVSRegistrar() = %v, want %v", addr, expectedAddr)
				}
			}
		})
	}
}

func TestContractStore_GetTaskMailbox(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		wantAddress string
		wantErr     bool
	}{
		{
			name: "get task mailbox",
			envVars: map[string]string{
				"TASKMAILBOX": "0xfeedfeedfeedfeedfeedfeedfeedfeedfeedfeed",
			},
			wantAddress: "0xfeedfeedfeedfeedfeedfeedfeedfeedfeedfeed",
			wantErr:     false,
		},
		{
			name:        "task mailbox not set",
			envVars:     map[string]string{},
			wantAddress: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Clearenv()

			// Set test environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			store, err := NewContractStore()
			if err != nil {
				t.Fatalf("Failed to create contract store: %v", err)
			}

			addr, err := store.GetTaskMailbox()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTaskMailbox() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				expectedAddr := common.HexToAddress(tt.wantAddress)
				if addr != expectedAddr {
					t.Errorf("GetTaskMailbox() = %v, want %v", addr, expectedAddr)
				}
			}
		})
	}
}

func TestContractStore_ListContracts(t *testing.T) {
	tests := []struct {
		name         string
		envVars      map[string]string
		wantContains []string
		wantCount    int
	}{
		{
			name:         "empty environment",
			envVars:      map[string]string{},
			wantContains: []string{},
			wantCount:    0,
		},
		{
			name: "single contract",
			envVars: map[string]string{
				"CONTRACT1": "0x1234567890123456789012345678901234567890",
			},
			wantContains: []string{"CONTRACT1"},
			wantCount:    1,
		},
		{
			name: "multiple contracts",
			envVars: map[string]string{
				"CONTRACT1":        "0x1234567890123456789012345678901234567890",
				"CONTRACT2":        "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
				"TASKAVSREGISTRAR": "0xfeedfeedfeedfeedfeedfeedfeedfeedfeedfeed",
			},
			wantContains: []string{"CONTRACT1", "CONTRACT2", "TASKAVSREGISTRAR"},
			wantCount:    3,
		},
		{
			name: "skip invalid addresses",
			envVars: map[string]string{
				"VALID":   "0x1234567890123456789012345678901234567890",
				"INVALID": "not-an-address",
				"ANOTHER": "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			},
			wantContains: []string{"VALID", "ANOTHER"},
			wantCount:    2,
		},
		{
			name: "skip non-contract env vars",
			envVars: map[string]string{
				"CONTRACT1":        "0x1234567890123456789012345678901234567890",
				"PATH":             "/usr/bin",
				"HOME":             "/home/user",
				"TASKAVSREGISTRAR": "0xfeedfeedfeedfeedfeedfeedfeedfeedfeedfeed",
			},
			wantContains: []string{"CONTRACT1", "TASKAVSREGISTRAR"},
			wantCount:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Clearenv()

			// Set test environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			store, err := NewContractStore()
			if err != nil {
				t.Fatalf("Failed to create contract store: %v", err)
			}

			contracts := store.ListContracts()

			if len(contracts) != tt.wantCount {
				t.Errorf("ListContracts() returned %d contracts, want %d", len(contracts), tt.wantCount)
			}

			// Check that expected contracts are present
			contractMap := make(map[string]bool)
			for _, c := range contracts {
				contractMap[c] = true
			}

			for _, want := range tt.wantContains {
				if !contractMap[want] {
					t.Errorf("ListContracts() missing expected contract %s", want)
				}
			}
		})
	}
}

func TestContractStore_getAddress(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		envVarName  string
		wantAddress string
		wantErr     bool
	}{
		{
			name: "valid address",
			envVars: map[string]string{
				"TEST": "0x1234567890123456789012345678901234567890",
			},
			envVarName:  "TEST",
			wantAddress: "0x1234567890123456789012345678901234567890",
			wantErr:     false,
		},
		{
			name: "address without 0x prefix",
			envVars: map[string]string{
				"TEST": "1234567890123456789012345678901234567890",
			},
			envVarName:  "TEST",
			wantAddress: "0x1234567890123456789012345678901234567890",
			wantErr:     false,
		},
		{
			name:        "non-existent env var",
			envVars:     map[string]string{},
			envVarName:  "NOTFOUND",
			wantAddress: "",
			wantErr:     true,
		},
		{
			name: "invalid hex address",
			envVars: map[string]string{
				"INVALID": "invalid-hex-string",
			},
			envVarName:  "INVALID",
			wantAddress: "",
			wantErr:     true,
		},
		{
			name: "empty address",
			envVars: map[string]string{
				"EMPTY": "",
			},
			envVarName:  "EMPTY",
			wantAddress: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Clearenv()

			// Set test environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			store, err := NewContractStore()
			if err != nil {
				t.Fatalf("Failed to create contract store: %v", err)
			}

			addr, err := store.getAddress(tt.envVarName)
			if (err != nil) != tt.wantErr {
				t.Errorf("getAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				expectedAddr := common.HexToAddress(tt.wantAddress)
				if addr != expectedAddr {
					t.Errorf("getAddress() = %v, want %v", addr, expectedAddr)
				}
			}
		})
	}
}

// TestContractStore_ThreadSafety tests that the ContractStore is thread-safe
func TestContractStore_ThreadSafety(t *testing.T) {
	// Set up test environment
	os.Clearenv()
	os.Setenv("CONTRACT1", "0x1234567890123456789012345678901234567890")
	os.Setenv("CONTRACT2", "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd")
	os.Setenv("CONTRACT3", "0xfeedfeedfeedfeedfeedfeedfeedfeedfeedfeed")

	store, err := NewContractStore()
	if err != nil {
		t.Fatalf("Failed to create contract store: %v", err)
	}

	// Run concurrent reads
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_, _ = store.GetContract("CONTRACT1")
				_, _ = store.GetContract("CONTRACT2")
				_, _ = store.GetContract("CONTRACT3")
				_ = store.ListContracts()
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

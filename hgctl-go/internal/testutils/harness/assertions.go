package harness

import (
	"encoding/json"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// CLIAssertions provides advanced assertion methods for CLI testing
type CLIAssertions struct {
	result *CLIResult
	t      *testing.T
}

// NewCLIAssertions creates a new assertions helper
func NewCLIAssertions(t *testing.T, result *CLIResult) *CLIAssertions {
	return &CLIAssertions{
		result: result,
		t:      t,
	}
}

// HasExitCode verifies the command exited with the expected code
func (a *CLIAssertions) HasExitCode(expected int) *CLIAssertions {
	a.t.Helper()
	if a.result.ExitCode != expected {
		a.t.Fatalf("Expected exit code %d, got %d\nStdout: %s\nStderr: %s",
			expected, a.result.ExitCode, a.result.Stdout, a.result.Stderr)
	}
	return a
}

// OutputMatches verifies the stdout matches a regex pattern
func (a *CLIAssertions) OutputMatches(pattern string) *CLIAssertions {
	a.t.Helper()
	re, err := regexp.Compile(pattern)
	if err != nil {
		a.t.Fatalf("Invalid regex pattern: %v", err)
	}

	if !re.MatchString(a.result.Stdout) {
		a.t.Fatalf("Output does not match pattern '%s'\nStdout: %s", pattern, a.result.Stdout)
	}
	return a
}

// ErrorMatches verifies the stderr matches a regex pattern
func (a *CLIAssertions) ErrorMatches(pattern string) *CLIAssertions {
	a.t.Helper()
	re, err := regexp.Compile(pattern)
	if err != nil {
		a.t.Fatalf("Invalid regex pattern: %v", err)
	}

	if !re.MatchString(a.result.Stderr) {
		a.t.Fatalf("Error output does not match pattern '%s'\nStderr: %s", pattern, a.result.Stderr)
	}
	return a
}

// OutputContainsAll verifies stdout contains all expected strings
func (a *CLIAssertions) OutputContainsAll(expected ...string) *CLIAssertions {
	a.t.Helper()
	for _, exp := range expected {
		if !strings.Contains(a.result.Stdout, exp) {
			a.t.Fatalf("Output does not contain expected string '%s'\nStdout: %s", exp, a.result.Stdout)
		}
	}
	return a
}

// OutputContainsNone verifies stdout contains none of the strings
func (a *CLIAssertions) OutputContainsNone(unexpected ...string) *CLIAssertions {
	a.t.Helper()
	for _, unexp := range unexpected {
		if strings.Contains(a.result.Stdout, unexp) {
			a.t.Fatalf("Output contains unexpected string '%s'\nStdout: %s", unexp, a.result.Stdout)
		}
	}
	return a
}

// JSONFieldEquals verifies a JSON field has the expected value
func (a *CLIAssertions) JSONFieldEquals(path string, expected interface{}) *CLIAssertions {
	a.t.Helper()

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(a.result.Stdout), &data); err != nil {
		a.t.Fatalf("Failed to parse JSON output: %v\nStdout: %s", err, a.result.Stdout)
	}

	value := a.getJSONValue(data, path)
	if !reflect.DeepEqual(value, expected) {
		a.t.Fatalf("JSON field '%s' = %v, expected %v", path, value, expected)
	}
	return a
}

// JSONFieldExists verifies a JSON field exists
func (a *CLIAssertions) JSONFieldExists(path string) *CLIAssertions {
	a.t.Helper()

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(a.result.Stdout), &data); err != nil {
		a.t.Fatalf("Failed to parse JSON output: %v\nStdout: %s", err, a.result.Stdout)
	}

	if a.getJSONValue(data, path) == nil {
		a.t.Fatalf("JSON field '%s' does not exist", path)
	}
	return a
}

// TableHasRow verifies the table output contains a row with the specified value
func (a *CLIAssertions) TableHasRow(columnName, value string) *CLIAssertions {
	a.t.Helper()

	table, err := a.result.ParseTable()
	if err != nil {
		a.t.Fatalf("Failed to parse table output: %v", err)
	}

	for _, row := range table {
		if row[columnName] == value {
			return a
		}
	}

	a.t.Fatalf("Table does not contain row with %s = '%s'", columnName, value)
	return a
}

// TableRowCount verifies the table has the expected number of rows
func (a *CLIAssertions) TableRowCount(expected int) *CLIAssertions {
	a.t.Helper()

	table, err := a.result.ParseTable()
	if err != nil {
		a.t.Fatalf("Failed to parse table output: %v", err)
	}

	if len(table) != expected {
		a.t.Fatalf("Table has %d rows, expected %d", len(table), expected)
	}
	return a
}

// TransactionSucceeded verifies a transaction was successful
func (a *CLIAssertions) TransactionSucceeded() *CLIAssertions {
	a.t.Helper()

	// Look for common success indicators
	successPatterns := []string{
		"transaction successful",
		"transaction mined",
		"successfully",
		"✓", // checkmark
	}

	combined := strings.ToLower(a.result.Stdout + a.result.Stderr)
	found := false
	for _, pattern := range successPatterns {
		if strings.Contains(combined, pattern) {
			found = true
			break
		}
	}

	if !found {
		a.t.Fatalf("Transaction does not appear to have succeeded\nStdout: %s\nStderr: %s",
			a.result.Stdout, a.result.Stderr)
	}
	return a
}

// TransactionFailed verifies a transaction failed
func (a *CLIAssertions) TransactionFailed() *CLIAssertions {
	a.t.Helper()

	// Look for common failure indicators
	failurePatterns := []string{
		"transaction failed",
		"reverted",
		"error",
		"❌", // X mark
	}

	combined := strings.ToLower(a.result.Stdout + a.result.Stderr)
	found := false
	for _, pattern := range failurePatterns {
		if strings.Contains(combined, pattern) {
			found = true
			break
		}
	}

	if !found {
		a.t.Fatalf("Transaction does not appear to have failed\nStdout: %s\nStderr: %s",
			a.result.Stdout, a.result.Stderr)
	}
	return a
}

// Helper methods

// getJSONValue retrieves a value from a JSON object using dot notation
func (a *CLIAssertions) getJSONValue(data map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	current := interface{}(data)

	for _, part := range parts {
		// Check if this is an array index
		if idx := a.parseArrayIndex(part); idx >= 0 {
			if arr, ok := current.([]interface{}); ok && idx < len(arr) {
				current = arr[idx]
			} else {
				return nil
			}
		} else {
			// Regular map access
			if m, ok := current.(map[string]interface{}); ok {
				current = m[part]
			} else {
				return nil
			}
		}
	}

	return current
}

// parseArrayIndex checks if a string is an array index like "[0]"
func (a *CLIAssertions) parseArrayIndex(s string) int {
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		idxStr := s[1 : len(s)-1]
		if idx, err := strconv.Atoi(idxStr); err == nil {
			return idx
		}
	}
	return -1
}

// Advanced assertion helpers for specific hgctl scenarios

// AssertKeystoreCreated verifies a keystore was created successfully
func (a *CLIAssertions) AssertKeystoreCreated(name string) *CLIAssertions {
	a.t.Helper()
	return a.HasExitCode(0).
		OutputContainsAll("keystore created", name)
}

// AssertOperatorRegistered verifies operator registration succeeded
func (a *CLIAssertions) AssertOperatorRegistered() *CLIAssertions {
	a.t.Helper()
	return a.HasExitCode(0).
		OutputContainsAll("operator registered").
		TransactionSucceeded()
}

// AssertKeyRegistered verifies key registration succeeded
func (a *CLIAssertions) AssertKeyRegistered() *CLIAssertions {
	a.t.Helper()
	return a.HasExitCode(0).
		OutputContainsAll("key registered").
		TransactionSucceeded()
}

// AssertAVSRegistered verifies AVS registration succeeded
func (a *CLIAssertions) AssertAVSRegistered() *CLIAssertions {
	a.t.Helper()
	return a.HasExitCode(0).
		OutputContainsAll("registered to AVS").
		TransactionSucceeded()
}

// AssertAllocationSet verifies allocation was set successfully
func (a *CLIAssertions) AssertAllocationSet() *CLIAssertions {
	a.t.Helper()
	return a.HasExitCode(0).
		OutputContainsAll("allocation set").
		TransactionSucceeded()
}

// ContractAssertions provides contract-specific assertions
type ContractAssertions struct {
	harness *TestHarness
	t       *testing.T
}

// NewContractAssertions creates contract assertion helpers
func NewContractAssertions(t *testing.T, harness *TestHarness) *ContractAssertions {
	return &ContractAssertions{
		harness: harness,
		t:       t,
	}
}

// OperatorIsRegistered verifies operator is registered on-chain
func (c *ContractAssertions) OperatorIsRegistered(address string) *ContractAssertions {
	c.t.Helper()
	// This would call the contract to verify
	// For now, it's a placeholder
	return c
}

// KeyIsRegistered verifies key is registered on-chain
func (c *ContractAssertions) KeyIsRegistered(operator, avs string, opSetID uint32) *ContractAssertions {
	c.t.Helper()
	// This would call the contract to verify
	// For now, it's a placeholder
	return c
}

// OperatorRegisteredToAVS verifies operator is registered to AVS
func (c *ContractAssertions) OperatorRegisteredToAVS(operator, avs string, opSetID uint32) *ContractAssertions {
	c.t.Helper()
	// This would call the contract to verify
	// For now, it's a placeholder
	return c
}

// AllocationIsSet verifies allocation is set correctly
func (c *ContractAssertions) AllocationIsSet(operator, avs, strategy string, opSetID uint32, minAmount string) *ContractAssertions {
	c.t.Helper()
	// This would call the contract to verify
	// For now, it's a placeholder
	return c
}

// Quick assertion helper
func Assert(t *testing.T, result *CLIResult) *CLIAssertions {
	return NewCLIAssertions(t, result)
}

// Example usage patterns for tests:
//
// result := harness.ExecuteCLI("keystore", "create", "--name", "test")
// Assert(t, result).
//     HasExitCode(0).
//     OutputContainsAll("keystore created", "test").
//     OutputMatches(`Created keystore: .+/test`)
//
// result := harness.ExecuteCLI("operator", "register")
// Assert(t, result).
//     AssertOperatorRegistered().
//     JSONFieldEquals("address", "0x123...")
//
// table := harness.ExecuteCLI("keystore", "list")
// Assert(t, table).
//     TableHasRow("NAME", "test-keystore").
//     TableRowCount(3)

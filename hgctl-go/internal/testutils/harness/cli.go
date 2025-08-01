package harness

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// CLIExecutor handles execution of the hgctl CLI binary
type CLIExecutor struct {
	binaryPath string
	workDir    string
	Env        []string
}

// CLIResult represents the result of a CLI command execution
type CLIResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Duration time.Duration
}

// NewCLIExecutor creates a new CLI executor
func NewCLIExecutor(binaryPath, workDir string) *CLIExecutor {
	// Set up environment variables
	env := os.Environ()
	
	// Ensure hgctl uses test directory for config
	env = append(env, fmt.Sprintf("HGCTL_CONFIG_DIR=%s/.hgctl", workDir))
	
	return &CLIExecutor{
		binaryPath: binaryPath,
		workDir:    workDir,
		Env:        env,
	}
}

// Execute runs the CLI with the given arguments
func (e *CLIExecutor) Execute(args ...string) (*CLIResult, error) {
	return e.ExecuteWithInput("", args...)
}

// ExecuteWithInput runs the CLI with stdin input
func (e *CLIExecutor) ExecuteWithInput(input string, args ...string) (*CLIResult, error) {
	start := time.Now()

	cmd := exec.Command(e.binaryPath, args...)
	cmd.Dir = e.workDir
	cmd.Env = e.Env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if input != "" {
		cmd.Stdin = strings.NewReader(input)
	}

	err := cmd.Run()
	
	result := &CLIResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: time.Since(start),
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			// Command couldn't be started
			return nil, fmt.Errorf("failed to execute command: %w", err)
		}
	}

	return result, nil
}

// Helper methods for CLIResult

// AssertSuccess verifies the command executed successfully
func (r *CLIResult) AssertSuccess(t *testing.T) {
	t.Helper()
	if r.ExitCode != 0 {
		t.Fatalf("Command failed with exit code %d\nStdout: %s\nStderr: %s", 
			r.ExitCode, r.Stdout, r.Stderr)
	}
}

// AssertFailure verifies the command failed
func (r *CLIResult) AssertFailure(t *testing.T) {
	t.Helper()
	if r.ExitCode == 0 {
		t.Fatalf("Expected command to fail but it succeeded\nStdout: %s", r.Stdout)
	}
}

// AssertExitCode verifies the command exited with a specific code
func (r *CLIResult) AssertExitCode(t *testing.T, expected int) {
	t.Helper()
	if r.ExitCode != expected {
		t.Fatalf("Expected exit code %d, got %d\nStdout: %s\nStderr: %s", 
			expected, r.ExitCode, r.Stdout, r.Stderr)
	}
}

// AssertContains verifies the output contains the expected string
func (r *CLIResult) AssertContains(t *testing.T, expected string) {
	t.Helper()
	combined := r.Stdout + r.Stderr
	if !strings.Contains(combined, expected) {
		t.Fatalf("Output does not contain expected string '%s'\nStdout: %s\nStderr: %s", 
			expected, r.Stdout, r.Stderr)
	}
}

// AssertNotContains verifies the output does not contain the string
func (r *CLIResult) AssertNotContains(t *testing.T, expected string) {
	t.Helper()
	combined := r.Stdout + r.Stderr
	if strings.Contains(combined, expected) {
		t.Fatalf("Output contains unexpected string '%s'\nStdout: %s\nStderr: %s", 
			expected, r.Stdout, r.Stderr)
	}
}

// AssertContainsAll verifies the output contains all expected strings
func (r *CLIResult) AssertContainsAll(t *testing.T, expected ...string) {
	t.Helper()
	combined := r.Stdout + r.Stderr
	for _, exp := range expected {
		if !strings.Contains(combined, exp) {
			t.Fatalf("Output does not contain expected string '%s'\nStdout: %s\nStderr: %s", 
				exp, r.Stdout, r.Stderr)
		}
	}
}

// AssertOutputMatches verifies the output matches a pattern
func (r *CLIResult) AssertOutputMatches(t *testing.T, pattern string) {
	t.Helper()
	// This is a simple substring match, could be extended to support regex
	if !strings.Contains(r.Stdout, pattern) {
		t.Fatalf("Output does not match pattern '%s'\nStdout: %s", pattern, r.Stdout)
	}
}

// ParseJSON unmarshals the stdout as JSON into the provided interface
func (r *CLIResult) ParseJSON(v interface{}) error {
	return json.Unmarshal([]byte(r.Stdout), v)
}

// ParseTable parses tabular output into a slice of maps
// Assumes first line is header, subsequent lines are data
func (r *CLIResult) ParseTable() ([]map[string]string, error) {
	lines := strings.Split(strings.TrimSpace(r.Stdout), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("insufficient lines for table parsing")
	}

	// Parse header
	headers := strings.Fields(lines[0])
	
	// Parse data rows
	var results []map[string]string
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "" {
			continue
		}
		
		fields := strings.Fields(lines[i])
		if len(fields) != len(headers) {
			return nil, fmt.Errorf("row %d has %d fields, expected %d", i, len(fields), len(headers))
		}
		
		row := make(map[string]string)
		for j, header := range headers {
			row[header] = fields[j]
		}
		results = append(results, row)
	}
	
	return results, nil
}

// GetTransactionHash extracts a transaction hash from the output
// Looks for patterns like "Transaction: 0x..." or "tx: 0x..."
func (r *CLIResult) GetTransactionHash() (string, error) {
	lines := strings.Split(r.Stdout, "\n")
	for _, line := range lines {
		line = strings.ToLower(line)
		if strings.Contains(line, "transaction") || strings.Contains(line, "tx") {
			// Look for hex string
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.HasPrefix(part, "0x") && len(part) == 66 {
					return part, nil
				}
			}
		}
	}
	return "", fmt.Errorf("no transaction hash found in output")
}

// SaveOutput saves the command output to files for debugging
func (r *CLIResult) SaveOutput(dir string, prefix string) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Save stdout
	if r.Stdout != "" {
		stdoutPath := filepath.Join(dir, fmt.Sprintf("%s.stdout", prefix))
		if err := os.WriteFile(stdoutPath, []byte(r.Stdout), 0644); err != nil {
			return fmt.Errorf("failed to save stdout: %w", err)
		}
	}

	// Save stderr
	if r.Stderr != "" {
		stderrPath := filepath.Join(dir, fmt.Sprintf("%s.stderr", prefix))
		if err := os.WriteFile(stderrPath, []byte(r.Stderr), 0644); err != nil {
			return fmt.Errorf("failed to save stderr: %w", err)
		}
	}

	// Save metadata
	metadata := map[string]interface{}{
		"exit_code": r.ExitCode,
		"duration":  r.Duration.String(),
		"timestamp": time.Now().Format(time.RFC3339),
	}
	
	metadataPath := filepath.Join(dir, fmt.Sprintf("%s.json", prefix))
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	
	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	return nil
}
package run

import (
	"runtime"
	"testing"
)

func TestTranslateLocalhostForDocker(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		onMacOS  bool
	}{
		{
			name:     "http localhost",
			input:    "http://localhost:8545",
			expected: "http://host.docker.internal:8545",
			onMacOS:  true,
		},
		{
			name:     "https localhost",
			input:    "https://localhost:8545",
			expected: "https://host.docker.internal:8545",
			onMacOS:  true,
		},
		{
			name:     "http 127.0.0.1",
			input:    "http://127.0.0.1:8545",
			expected: "http://host.docker.internal:8545",
			onMacOS:  true,
		},
		{
			name:     "ws localhost",
			input:    "ws://localhost:8545",
			expected: "ws://host.docker.internal:8545",
			onMacOS:  true,
		},
		{
			name:     "wss 127.0.0.1",
			input:    "wss://127.0.0.1:8545",
			expected: "wss://host.docker.internal:8545",
			onMacOS:  true,
		},
		{
			name:     "non-localhost URL",
			input:    "http://example.com:8545",
			expected: "http://example.com:8545",
			onMacOS:  true,
		},
		{
			name:     "multiple localhost occurrences",
			input:    "http://localhost:8545/path?redirect=http://localhost:3000",
			expected: "http://host.docker.internal:8545/path?redirect=http://host.docker.internal:3000",
			onMacOS:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Only run macOS-specific tests on macOS
			if tt.onMacOS && runtime.GOOS != "darwin" {
				// On non-macOS systems, expect no translation
				result := translateLocalhostForDocker(tt.input)
				if result != tt.input {
					t.Errorf("Expected no translation on %s, but got %s -> %s", runtime.GOOS, tt.input, result)
				}
				return
			}

			// On macOS, expect translation
			if runtime.GOOS == "darwin" {
				result := translateLocalhostForDocker(tt.input)
				if result != tt.expected {
					t.Errorf("translateLocalhostForDocker(%s) = %s, want %s", tt.input, result, tt.expected)
				}
			}
		})
	}
}

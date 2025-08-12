package middleware

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadEnvFile(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.env")

	// Write test content
	content := `# This is a comment
OPERATOR_PRIVATE_KEY=0x1234567890abcdef
OPERATOR_KEYSTORE_PASSWORD="my-secret-password"
AVS_ADDRESS='0xABCDEF'

# Another comment
EMPTY_VALUE=
KEY_WITH_SPACES = value with spaces
MALFORMED_LINE_NO_EQUALS
=MALFORMED_LINE_NO_KEY
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	// Load the file
	envVars, err := loadEnvFile(testFile)
	require.NoError(t, err)

	// Verify the loaded values
	assert.Equal(t, "0x1234567890abcdef", envVars["OPERATOR_PRIVATE_KEY"])
	assert.Equal(t, "my-secret-password", envVars["OPERATOR_KEYSTORE_PASSWORD"])
	assert.Equal(t, "0xABCDEF", envVars["AVS_ADDRESS"])
	assert.Equal(t, "", envVars["EMPTY_VALUE"])
	assert.Equal(t, "value with spaces", envVars["KEY_WITH_SPACES"])

	// Malformed lines should be ignored
	_, hasKey := envVars["MALFORMED_LINE_NO_EQUALS"]
	assert.False(t, hasKey)
	_, hasKey = envVars[""]
	assert.False(t, hasKey)
}

func TestExpandPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "home directory expansion",
			input:    "~/test/file.env",
			expected: filepath.Join(os.Getenv("HOME"), "test", "file.env"),
		},
		{
			name:     "absolute path unchanged",
			input:    "/absolute/path/file.env",
			expected: "/absolute/path/file.env",
		},
		{
			name:     "relative path unchanged",
			input:    "relative/path/file.env",
			expected: "relative/path/file.env",
		},
		{
			name:     "current directory unchanged",
			input:    "./file.env",
			expected: "./file.env",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandPath(tt.input)
			if tt.input == "~/test/file.env" {
				// For home directory test, just verify it starts with home
				homeDir, _ := os.UserHomeDir()
				assert.True(t, strings.HasPrefix(result, homeDir))
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestLoadEnvFileNonExistent(t *testing.T) {
	// Try to load a non-existent file
	_, err := loadEnvFile("/non/existent/file.env")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open file")
}

func TestLoadEnvFileEmpty(t *testing.T) {
	// Create an empty file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "empty.env")
	err := os.WriteFile(testFile, []byte(""), 0644)
	require.NoError(t, err)

	// Load the empty file
	envVars, err := loadEnvFile(testFile)
	require.NoError(t, err)
	assert.Empty(t, envVars)
}

func TestLoadEnvFileCommentsOnly(t *testing.T) {
	// Create a file with only comments
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "comments.env")
	content := `# Comment 1
# Comment 2
  # Comment with spaces
#Another comment`
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	// Load the file
	envVars, err := loadEnvFile(testFile)
	require.NoError(t, err)
	assert.Empty(t, envVars)
}

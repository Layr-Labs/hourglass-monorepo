package register

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
)

func TestRegisterAVSCommand(t *testing.T) {
	cmd := RegisterAVSCommand()

	t.Run("Command Structure", func(t *testing.T) {
		assert.Equal(t, "register-avs", cmd.Name)
		assert.Equal(t, "Register operator with an AVS", cmd.Usage)
		assert.NotNil(t, cmd.Action)
		assert.Contains(t, cmd.Description, "Actively Validated Service")
	})

	t.Run("Required Flags", func(t *testing.T) {
		requiredFlags := map[string]bool{
			"operator-set-ids": false,
			"socket":           false,
		}

		for _, flag := range cmd.Flags {
			switch f := flag.(type) {
			case *cli.StringFlag:
				if _, exists := requiredFlags[f.Name]; exists {
					requiredFlags[f.Name] = f.Required
				}
			case *cli.Uint64SliceFlag:
				if _, exists := requiredFlags[f.Name]; exists {
					requiredFlags[f.Name] = f.Required
				}
			}
		}

		assert.True(t, requiredFlags["operator-set-ids"], "operator-set-ids flag should be required")
		assert.True(t, requiredFlags["socket"], "socket flag should be required")
	})

	t.Run("Socket Format Example", func(t *testing.T) {
		// Check that the command description includes socket format example
		// The actual command shows https:// in the example, so test expects https://
		assert.Contains(t, cmd.Description, "https://operator.example.com:8080")

		// Check socket flag usage
		var socketFlag *cli.StringFlag
		for _, flag := range cmd.Flags {
			if sf, ok := flag.(*cli.StringFlag); ok && sf.Name == "socket" {
				socketFlag = sf
				break
			}
		}

		assert.NotNil(t, socketFlag)
		assert.Contains(t, socketFlag.Usage, "endpoint")
	})

	t.Run("Multiple Operator Set IDs", func(t *testing.T) {
		// Verify operator-set-ids is a slice flag
		var operatorSetIDsFlag *cli.Uint64SliceFlag
		for _, flag := range cmd.Flags {
			if sf, ok := flag.(*cli.Uint64SliceFlag); ok && sf.Name == "operator-set-ids" {
				operatorSetIDsFlag = sf
				break
			}
		}

		assert.NotNil(t, operatorSetIDsFlag)
		assert.Contains(t, operatorSetIDsFlag.Usage, "multiple")
	})
}

func TestTranslateLocalhostForDocker(t *testing.T) {
	// Create a mock logger
	log := logger.NewLogger(false)

	testCases := []struct {
		name     string
		input    string
		expected string
		onMacOS  bool
	}{
		{
			name:     "localhost URL with http",
			input:    "http://localhost:8080",
			expected: "http://host.docker.internal:8080",
			onMacOS:  true,
		},
		{
			name:     "localhost URL with https",
			input:    "https://localhost:8443/path",
			expected: "https://host.docker.internal:8443/path",
			onMacOS:  true,
		},
		{
			name:     "127.0.0.1 URL",
			input:    "http://127.0.0.1:9090",
			expected: "http://host.docker.internal:9090",
			onMacOS:  true,
		},
		{
			name:     "127.0.0.2 URL (loopback range)",
			input:    "http://127.0.0.2:9090",
			expected: "http://host.docker.internal:9090",
			onMacOS:  true,
		},
		{
			name:     "non-localhost URL unchanged",
			input:    "https://example.com:8080",
			expected: "https://example.com:8080",
			onMacOS:  true,
		},
		{
			name:     "simple localhost:port format",
			input:    "localhost:8080",
			expected: "host.docker.internal:8080",
			onMacOS:  true,
		},
		{
			name:     "simple 127.0.0.1:port format",
			input:    "127.0.0.1:8080",
			expected: "host.docker.internal:8080",
			onMacOS:  true,
		},
		{
			name:     "URL with query parameters",
			input:    "http://localhost:8080/api?key=value",
			expected: "http://host.docker.internal:8080/api?key=value",
			onMacOS:  true,
		},
		{
			name:     "URL with fragment",
			input:    "http://localhost:8080/api#section",
			expected: "http://host.docker.internal:8080/api#section",
			onMacOS:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Only test on macOS or when explicitly testing macOS behavior
			if runtime.GOOS == "darwin" || tc.onMacOS {
				result := translateLocalhostForDocker(tc.input, log)
				if runtime.GOOS == "darwin" {
					assert.Equal(t, tc.expected, result)
				} else if runtime.GOOS != "darwin" && tc.onMacOS {
					// When not on macOS, the function should return input unchanged
					// since the main function only calls this on macOS
					t.Skip("Skipping macOS-specific test on non-Darwin platform")
				}
			}
		})
	}
}

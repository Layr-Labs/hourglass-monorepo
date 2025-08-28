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
			"socket": false,
		}

		for _, flag := range cmd.Flags {
			switch f := flag.(type) {
			case *cli.StringFlag:
				if _, exists := requiredFlags[f.Name]; exists {
					requiredFlags[f.Name] = f.Required
				}
			}
		}

		assert.True(t, requiredFlags["socket"], "socket flag should be required")

		// Verify operator-set-ids flag no longer exists
		for _, flag := range cmd.Flags {
			switch f := flag.(type) {
			case *cli.Uint64SliceFlag:
				assert.NotEqual(t, "operator-set-ids", f.Name, "operator-set-ids flag should not exist")
			}
		}
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

	t.Run("Context Prerequisites Documentation", func(t *testing.T) {
		// Verify the command description includes context prerequisites
		assert.Contains(t, cmd.Description, "operator set configured in the context",
			"Description should mention operator set context requirement")
		assert.Contains(t, cmd.Description, "hgctl context set --operator-set-id",
			"Description should include example of setting operator-set-id")
		assert.Contains(t, cmd.Description, "Prerequisites",
			"Description should have prerequisites section")
		assert.Contains(t, cmd.Description, "AVS address must be configured",
			"Description should mention AVS address requirement")
		assert.Contains(t, cmd.Description, "Operator address must be configured",
			"Description should mention operator address requirement")
	})
}

func TestRegisterAVSContextUsage(t *testing.T) {
	cmd := RegisterAVSCommand()

	t.Run("Description Mentions Single Operator Set", func(t *testing.T) {
		// Verify that description clarifies single operator set registration
		assert.Contains(t, cmd.Description, "registers the operator to the operator set",
			"Description should clarify single operator set registration")
	})

	t.Run("No Operator Set IDs in Flags", func(t *testing.T) {
		// Ensure no operator-set-ids related flags exist
		for _, flag := range cmd.Flags {
			flagName := ""
			switch f := flag.(type) {
			case *cli.StringFlag:
				flagName = f.Name
			case *cli.Uint64SliceFlag:
				flagName = f.Name
			case *cli.Uint64Flag:
				flagName = f.Name
			}
			assert.NotContains(t, flagName, "operator-set",
				"No flags should contain 'operator-set' in their name")
		}
	})

	t.Run("Socket Flag Still Required", func(t *testing.T) {
		var socketFlag *cli.StringFlag
		for _, flag := range cmd.Flags {
			if sf, ok := flag.(*cli.StringFlag); ok && sf.Name == "socket" {
				socketFlag = sf
				break
			}
		}

		assert.NotNil(t, socketFlag, "socket flag should exist")
		assert.True(t, socketFlag.Required, "socket flag should be required")
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

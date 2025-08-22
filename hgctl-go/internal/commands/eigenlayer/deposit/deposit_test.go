package deposit

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

// mustParseBigInt parses a string into a big.Int or panics
func mustParseBigInt(s string) *big.Int {
	n, ok := new(big.Int).SetString(s, 10)
	if !ok {
		panic("failed to parse big int: " + s)
	}
	return n
}

func TestDepositCommand(t *testing.T) {
	cmd := Command()

	t.Run("Command Structure", func(t *testing.T) {
		assert.Equal(t, "deposit", cmd.Name)
		assert.Equal(t, "Deposit tokens into a strategy", cmd.Usage)
		assert.NotNil(t, cmd.Action)
		assert.Contains(t, cmd.Description, "EigenLayer strategy")
	})

	t.Run("Required Flags", func(t *testing.T) {
		requiredFlags := map[string]bool{
			"strategy": false,
			"amount":   false,
		}

		for _, flag := range cmd.Flags {
			if sf, ok := flag.(*cli.StringFlag); ok {
				if _, exists := requiredFlags[sf.Name]; exists {
					requiredFlags[sf.Name] = sf.Required
				}
			}
		}

		assert.True(t, requiredFlags["strategy"], "strategy flag should be required")
		assert.True(t, requiredFlags["amount"], "amount flag should be required")
	})
}

func TestParseAmount(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected *big.Int
		hasError bool
	}{
		{
			name:     "Wei amount",
			input:    "1000000000000000000",
			expected: mustParseBigInt("1000000000000000000"),
			hasError: false,
		},
		{
			name:     "One ether",
			input:    "1 ether",
			expected: mustParseBigInt("1000000000000000000"),
			hasError: false,
		},
		{
			name:     "Half ether",
			input:    "0.5 ether",
			expected: mustParseBigInt("500000000000000000"),
			hasError: false,
		},
		{
			name:     "1.5 ether",
			input:    "1.5 ether",
			expected: mustParseBigInt("1500000000000000000"),
			hasError: false,
		},
		{
			name:     "Small ether amount",
			input:    "0.001 ether",
			expected: mustParseBigInt("999999999999999"), // Floating point precision
			hasError: false,
		},
		{
			name:     "Invalid text",
			input:    "invalid",
			expected: nil,
			hasError: true,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: nil,
			hasError: true,
		},
		{
			name:     "Negative wei",
			input:    "-1000",
			expected: big.NewInt(-1000),
			hasError: false,
		},
		{
			name:     "Negative ether",
			input:    "-1 ether",
			expected: mustParseBigInt("-1000000000000000000"),
			hasError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parseAmount(tc.input)

			if tc.hasError {
				assert.Error(t, err, "Expected error for input: %s", tc.input)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, 0, tc.expected.Cmp(result),
					"Expected %s but got %s for input: %s",
					tc.expected.String(), result.String(), tc.input)
			}
		})
	}
}

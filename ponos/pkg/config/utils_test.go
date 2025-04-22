package config

import (
	"testing"
)

func TestUtilityFunctions(t *testing.T) {
	t.Run("KebabToSnakeCase", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected string
		}{
			{"hello-world", "hello_world"},
			{"hello", "hello"},
			{"hello-world-example", "hello_world_example"},
			{"", ""},
			{"-", "_"},
			{"hello-", "hello_"},
			{"-hello", "_hello"},
			{"hello--world", "hello__world"},
		}

		for _, tc := range testCases {
			t.Run(tc.input, func(t *testing.T) {
				result := KebabToSnakeCase(tc.input)
				if result != tc.expected {
					t.Errorf("KebabToSnakeCase(%q) = %q, expected %q", tc.input, result, tc.expected)
				}
			})
		}
	})

	t.Run("NormalizeFlagName", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected string
		}{
			{"hello-world", "hello_world"},
			{"hello", "hello"},
			{"hello-world-example", "hello_world_example"},
			{"", ""},
			{"-", "_"},
			{"hello-", "hello_"},
			{"-hello", "_hello"},
			{"hello--world", "hello__world"},
		}

		for _, tc := range testCases {
			t.Run(tc.input, func(t *testing.T) {
				result := NormalizeFlagName(tc.input)
				if result != tc.expected {
					t.Errorf("NormalizeFlagName(%q) = %q, expected %q", tc.input, result, tc.expected)
				}
			})
		}
	})

	t.Run("DefaultInt", func(t *testing.T) {
		testCases := []struct {
			name     string
			value    int
			default_ int
			expected int
		}{
			{"Zero value returns default", 0, 42, 42},
			{"Non-zero value returns itself", 10, 42, 10},
			{"Negative value returns itself", -5, 42, -5},
			{"Zero default and zero value", 0, 0, 0},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := DefaultInt(tc.value, tc.default_)
				if result != tc.expected {
					t.Errorf("DefaultInt(%d, %d) = %d, expected %d", tc.value, tc.default_, result, tc.expected)
				}
			})
		}
	})
}

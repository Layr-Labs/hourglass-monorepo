package harness

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLIExecutor(t *testing.T) {
	t.Run("Execute Simple Command", func(t *testing.T) {
		executor := NewCLIExecutor("echo", "/tmp")
		result, err := executor.Execute("hello", "world")
		
		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.Equal(t, "hello world\n", result.Stdout)
		assert.Empty(t, result.Stderr)
	})

	t.Run("Execute With Input", func(t *testing.T) {
		executor := NewCLIExecutor("cat", "/tmp")
		result, err := executor.ExecuteWithInput("test input\n", "-")
		
		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.Equal(t, "test input\n", result.Stdout)
	})

	t.Run("Handle Command Failure", func(t *testing.T) {
		executor := NewCLIExecutor("ls", "/tmp")
		result, err := executor.Execute("/nonexistent/path")
		
		require.NoError(t, err) // Should not error, just have non-zero exit
		assert.NotEqual(t, 0, result.ExitCode)
		assert.Contains(t, result.Stderr, "No such file")
	})

	t.Run("Parse Table Output", func(t *testing.T) {
		result := &CLIResult{
			Stdout: "NAME    TYPE    AGE\ntest1   ecdsa   1d\ntest2   bn254   2h\n",
		}
		
		table, err := result.ParseTable()
		require.NoError(t, err)
		assert.Len(t, table, 2)
		
		assert.Equal(t, "test1", table[0]["NAME"])
		assert.Equal(t, "ecdsa", table[0]["TYPE"])
		assert.Equal(t, "1d", table[0]["AGE"])
		
		assert.Equal(t, "test2", table[1]["NAME"])
		assert.Equal(t, "bn254", table[1]["TYPE"])
		assert.Equal(t, "2h", table[1]["AGE"])
	})

	t.Run("CLI Result Assertions", func(t *testing.T) {
		result := &CLIResult{
			Stdout:   "Command executed successfully\nOperation: create keystore\nResult: success",
			Stderr:   "",
			ExitCode: 0,
		}
		
		// Test various assertion methods
		result.AssertSuccess(t)
		result.AssertContains(t, "success")
		result.AssertContainsAll(t, "Command", "keystore", "success")
		result.AssertNotContains(t, "error")
		
		// Test output matching
		lines := strings.Split(result.Stdout, "\n")
		assert.Len(t, lines, 3)
	})
}
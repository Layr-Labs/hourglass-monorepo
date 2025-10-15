package integration

import (
	"strings"
	"testing"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/testutils/harness"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppointeeLifecycle(t *testing.T) {
	skipIfShort(t)

	h := harness.NewTestHarness(t)
	require.NoError(t, h.Setup())
	defer h.Teardown()

	testContextName := "appointee-lifecycle-context"
	result, err := h.ExecuteCLI("context", "copy", "--copy-name", testContextName, "--use", "test")
	require.NoError(t, err, "Failed to copy test context")
	require.Equal(t, 0, result.ExitCode, "Context copy should succeed")

	defer func() {
		_, _ = h.ExecuteCLI("context", "use", "test")

		result, err := h.ExecuteCLI("context", "delete", testContextName)
		if err != nil || result.ExitCode != 0 {
			t.Logf("Warning: failed to delete test context %s: %v", testContextName, err)
		}
	}()

	accountAddress := h.ChainConfig.UnregisteredOperator1AccountAddress
	appointee1Address := h.ChainConfig.UnregisteredOperator2AccountAddress
	appointee2Address := h.ChainConfig.ExecOperator2AccountAddress
	targetAddress := h.ChainConfig.PermissionControllerAddress

	selector1 := "0x12345678"
	selector2 := "0x87654321"

	t.Logf("Account: %s", accountAddress)
	t.Logf("Appointee 1: %s", appointee1Address)
	t.Logf("Appointee 2: %s", appointee2Address)
	t.Logf("Target: %s", targetAddress)
	t.Logf("Selector 1: %s", selector1)
	t.Logf("Selector 2: %s", selector2)

	contextResult, err := h.ExecuteCLI("context", "set", "--operator-address", accountAddress)
	require.NoError(t, err)
	require.Equal(t, 0, contextResult.ExitCode)

	t.Run("Add_First_Appointee_For_Selector_1", func(t *testing.T) {
		result, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
			"eigenlayer", "user", "appointee", "set",
			"--account-address", accountAddress,
			"--appointee-address", appointee1Address,
			"--contract-address", targetAddress,
			"--selector", selector1)

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		t.Logf("Set appointee 1 for selector 1: %s", result.Stdout)
	})

	t.Run("Add_Second_Appointee_For_Selector_1", func(t *testing.T) {
		result, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
			"eigenlayer", "user", "appointee", "set",
			"--account-address", accountAddress,
			"--appointee-address", appointee2Address,
			"--contract-address", targetAddress,
			"--selector", selector1)

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		t.Logf("Set appointee 2 for selector 1: %s", result.Stdout)
	})

	t.Run("Add_First_Appointee_For_Selector_2", func(t *testing.T) {
		result, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
			"eigenlayer", "user", "appointee", "set",
			"--account-address", accountAddress,
			"--appointee-address", appointee1Address,
			"--contract-address", targetAddress,
			"--selector", selector2)

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		t.Logf("Set appointee 1 for selector 2: %s", result.Stdout)
	})

	t.Run("Add_Second_Appointee_For_Selector_2", func(t *testing.T) {
		result, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
			"eigenlayer", "user", "appointee", "set",
			"--account-address", accountAddress,
			"--appointee-address", appointee2Address,
			"--contract-address", targetAddress,
			"--selector", selector2)

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		t.Logf("Set appointee 2 for selector 2: %s", result.Stdout)
	})

	t.Run("Verify_Multiple_Appointees_For_Selector_1", func(t *testing.T) {
		result, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
			"eigenlayer", "user", "appointee", "list",
			"--account-address", accountAddress,
			"--contract-address", targetAddress,
			"--selector", selector1)

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		t.Logf("List for selector 1: stdout='%s', stderr='%s'", result.Stdout, result.Stderr)
		assert.Contains(t, strings.ToLower(result.Stdout), strings.ToLower(appointee1Address), "Appointee 1 should be listed")
		assert.Contains(t, strings.ToLower(result.Stdout), strings.ToLower(appointee2Address), "Appointee 2 should be listed")
	})

	t.Run("Verify_Multiple_Appointees_For_Selector_2", func(t *testing.T) {
		result, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
			"eigenlayer", "user", "appointee", "list",
			"--account-address", accountAddress,
			"--contract-address", targetAddress,
			"--selector", selector2)

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.Contains(t, strings.ToLower(result.Stdout), strings.ToLower(appointee1Address), "Appointee 1 should be listed")
		assert.Contains(t, strings.ToLower(result.Stdout), strings.ToLower(appointee2Address), "Appointee 2 should be listed")
		t.Logf("List for selector 2: %s", result.Stdout)
	})

	t.Run("Verify_Appointee1_Listed_Permissions", func(t *testing.T) {
		result, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
			"eigenlayer", "user", "appointee", "list-permissions",
			"--account-address", accountAddress,
			"--appointee-address", appointee1Address)

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)

		selector1NoPrefix := strings.TrimPrefix(selector1, "0x")
		selector2NoPrefix := strings.TrimPrefix(selector2, "0x")

		assert.Contains(t, result.Stdout, appointee1Address, "Should show appointee address")
		assert.Contains(t, result.Stdout, accountAddress, "Should show account address")
		assert.Contains(t, result.Stdout, targetAddress, "Should show target address")
		assert.Contains(t, strings.ToLower(result.Stdout), strings.ToLower(selector1NoPrefix), "Should show selector 1")
		assert.Contains(t, strings.ToLower(result.Stdout), strings.ToLower(selector2NoPrefix), "Should show selector 2")
		t.Logf("Appointee 1 permissions: %s", result.Stdout)
	})

	t.Run("Verify_Appointee2_Listed_Permissions", func(t *testing.T) {
		result, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
			"eigenlayer", "user", "appointee", "list-permissions",
			"--account-address", accountAddress,
			"--appointee-address", appointee2Address)

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)

		selector1NoPrefix := strings.TrimPrefix(selector1, "0x")
		selector2NoPrefix := strings.TrimPrefix(selector2, "0x")

		assert.Contains(t, result.Stdout, appointee2Address, "Should show appointee address")
		assert.Contains(t, result.Stdout, accountAddress, "Should show account address")
		assert.Contains(t, result.Stdout, targetAddress, "Should show target address")
		assert.Contains(t, strings.ToLower(result.Stdout), strings.ToLower(selector1NoPrefix), "Should show selector 1")
		assert.Contains(t, strings.ToLower(result.Stdout), strings.ToLower(selector2NoPrefix), "Should show selector 2")
		t.Logf("Appointee 2 permissions: %s", result.Stdout)
	})

	t.Run("Test_CanCall_Permissions_Are_True", func(t *testing.T) {
		testCases := []struct {
			appointee string
			selector  string
		}{
			{appointee1Address, selector1},
			{appointee1Address, selector2},
			{appointee2Address, selector1},
			{appointee2Address, selector2},
		}

		for _, tc := range testCases {
			result, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
				"eigenlayer", "user", "appointee", "can-call",
				"--appointee-address", tc.appointee,
				"--contract-address", targetAddress,
				"--selector", tc.selector)

			require.NoError(t, err)
			assert.Equal(t, 0, result.ExitCode)
			assert.Contains(t, result.Stdout, "true", "CanCall should be true for %s with selector %s", tc.appointee, tc.selector)
			t.Logf("CanCall %s for selector %s: %s", tc.appointee, tc.selector, result.Stdout)
		}
	})

	t.Run("Remove_First_Appointee_For_Selector_1", func(t *testing.T) {
		result, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
			"eigenlayer", "user", "appointee", "remove",
			"--account-address", accountAddress,
			"--appointee-address", appointee1Address,
			"--contract-address", targetAddress,
			"--selector", selector1)

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		t.Logf("Removed appointee 1 for selector 1: %s", result.Stdout)
	})

	t.Run("Verify_Only_Second_Appointee_Remains_For_Selector_1", func(t *testing.T) {
		result, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
			"eigenlayer", "user", "appointee", "list",
			"--account-address", accountAddress,
			"--contract-address", targetAddress,
			"--selector", selector1)

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.NotContains(t, strings.ToLower(result.Stdout), strings.ToLower(appointee1Address), "Appointee 1 should not be listed")
		assert.Contains(t, strings.ToLower(result.Stdout), strings.ToLower(appointee2Address), "Appointee 2 should still be listed")
		t.Logf("List after first removal: %s", result.Stdout)
	})

	t.Run("Remove_Second_Appointee_For_Selector_1", func(t *testing.T) {
		result, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
			"eigenlayer", "user", "appointee", "remove",
			"--account-address", accountAddress,
			"--appointee-address", appointee2Address,
			"--contract-address", targetAddress,
			"--selector", selector1)

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		t.Logf("Removed appointee 2 for selector 1: %s", result.Stdout)
	})

	t.Run("Verify_List_Is_Empty_After_Selector_1_Removals", func(t *testing.T) {
		result, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
			"eigenlayer", "user", "appointee", "list",
			"--account-address", accountAddress,
			"--contract-address", targetAddress,
			"--selector", selector1)

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.NotContains(t, strings.ToLower(result.Stdout), strings.ToLower(appointee1Address), "Appointee 1 should not be listed")
		assert.NotContains(t, strings.ToLower(result.Stdout), strings.ToLower(appointee2Address), "Appointee 2 should not be listed")
		t.Logf("List after all selector 1 removals: %s", result.Stdout)
	})

	t.Run("Remove_First_Appointee_For_Selector_2", func(t *testing.T) {
		result, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
			"eigenlayer", "user", "appointee", "remove",
			"--account-address", accountAddress,
			"--appointee-address", appointee1Address,
			"--contract-address", targetAddress,
			"--selector", selector2)

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		t.Logf("Removed appointee 1 for selector 2: %s", result.Stdout)
	})

	t.Run("Verify_Only_Second_Appointee_Remains_For_Selector_2", func(t *testing.T) {
		result, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
			"eigenlayer", "user", "appointee", "list",
			"--account-address", accountAddress,
			"--contract-address", targetAddress,
			"--selector", selector2)

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.NotContains(t, strings.ToLower(result.Stdout), strings.ToLower(appointee1Address), "Appointee 1 should not be listed")
		assert.Contains(t, strings.ToLower(result.Stdout), strings.ToLower(appointee2Address), "Appointee 2 should still be listed")
		t.Logf("List after first selector 2 removal: %s", result.Stdout)
	})

	t.Run("Remove_Second_Appointee_For_Selector_2", func(t *testing.T) {
		result, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
			"eigenlayer", "user", "appointee", "remove",
			"--account-address", accountAddress,
			"--appointee-address", appointee2Address,
			"--contract-address", targetAddress,
			"--selector", selector2)

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		t.Logf("Removed appointee 2 for selector 2: %s", result.Stdout)
	})

	t.Run("Verify_List_Is_Empty_After_Selector_2_Removals", func(t *testing.T) {
		result, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
			"eigenlayer", "user", "appointee", "list",
			"--account-address", accountAddress,
			"--contract-address", targetAddress,
			"--selector", selector2)

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.NotContains(t, strings.ToLower(result.Stdout), strings.ToLower(appointee1Address), "Appointee 1 should not be listed")
		assert.NotContains(t, strings.ToLower(result.Stdout), strings.ToLower(appointee2Address), "Appointee 2 should not be listed")
		t.Logf("List after all selector 2 removals: %s", result.Stdout)
	})

	t.Run("Test_CanCall_Permissions_Are_False", func(t *testing.T) {
		testCases := []struct {
			appointee string
			selector  string
		}{
			{appointee1Address, selector1},
			{appointee2Address, selector2},
		}

		for _, tc := range testCases {
			result, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
				"eigenlayer", "user", "appointee", "can-call",
				"--appointee-address", tc.appointee,
				"--contract-address", targetAddress,
				"--selector", tc.selector)

			require.NoError(t, err)
			assert.Equal(t, 0, result.ExitCode)
			assert.Contains(t, result.Stdout, "false", "CanCall should be false for %s with selector %s after removal", tc.appointee, tc.selector)
			t.Logf("CanCall %s for selector %s after removal: %s", tc.appointee, tc.selector, result.Stdout)
		}
	})
}

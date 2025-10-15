package integration

import (
	"strings"
	"testing"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/testutils/harness"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdminLifecycle(t *testing.T) {
	skipIfShort(t)

	h := harness.NewTestHarness(t)
	require.NoError(t, h.Setup())
	defer h.Teardown()

	testContextName := "admin-lifecycle-context"
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
	firstAdminAddress := h.ChainConfig.UnregisteredOperator2AccountAddress
	secondAdminAddress := h.ChainConfig.ExecOperator2AccountAddress

	t.Logf("Test account: %s", accountAddress)
	t.Logf("First admin: %s", firstAdminAddress)
	t.Logf("Second admin: %s", secondAdminAddress)

	t.Run("Setup_SelfAppoint_Account", func(t *testing.T) {
		contextResult, err := h.ExecuteCLI("context", "set", "--operator-address", accountAddress)
		require.NoError(t, err)
		assert.Equal(t, 0, contextResult.ExitCode)

		// Account is not an admin - need to self-appoint
		t.Logf("Account %s is not an admin. Self-appointing...", accountAddress)

		// Add account as pending admin for itself
		addResult, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
			"eigenlayer", "user", "admin", "add",
			"--admin-address", accountAddress)

		require.NoError(t, err)
		require.Equal(t, 0, addResult.ExitCode, "Failed to add account as pending admin")
		assert.Contains(t, addResult.Stdout, "Self-appointment detected")
		t.Logf("Add pending output: %s", addResult.Stdout)

		// Accept as admin (self-appointment)
		acceptResult, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
			"eigenlayer", "user", "admin", "accept",
			"--account-address", accountAddress)

		require.NoError(t, err)
		require.Equal(t, 0, acceptResult.ExitCode, "Failed to accept admin role")
		t.Logf("Accept output: %s", acceptResult.Stdout)

		// Verify account is now an admin
		verifyResult, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
			"eigenlayer", "user", "admin", "is-admin",
			"--admin-address", accountAddress)

		require.NoError(t, err)
		require.Equal(t, 0, verifyResult.ExitCode, "Account should be an admin after acceptance")
		t.Logf("Is-admin check after acceptance: %s", verifyResult.Stdout)
	})

	t.Run("Add_First_Admin_As_Pending", func(t *testing.T) {
		result, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
			"eigenlayer", "user", "admin", "add",
			"--admin-address", firstAdminAddress)

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.Contains(t, result.Stdout, "Successfully added pending admin")
		t.Logf("Add first admin output: %s", result.Stdout)
	})

	t.Run("Verify_First_Admin_Is_Pending", func(t *testing.T) {
		result, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
			"eigenlayer", "user", "admin", "is-pending",
			"--admin-address", firstAdminAddress)

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		t.Logf("Is-pending check: %s", result.Stdout)
	})

	t.Run("Remove_First_Admin_As_Pending", func(t *testing.T) {
		// Remove first admin from pending
		result, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
			"eigenlayer", "user", "admin", "remove-pending",
			"--admin-address", firstAdminAddress)

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.Contains(t, result.Stdout, "Successfully removed pending admin")
		t.Logf("Remove pending output: %s", result.Stdout)
	})

	t.Run("Re_Add_First_Admin_After_Removal", func(t *testing.T) {
		// Add first admin again
		result, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
			"eigenlayer", "user", "admin", "add",
			"--admin-address", firstAdminAddress)

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.Contains(t, result.Stdout, "Successfully added pending admin")
		t.Logf("Re-add first admin output: %s", result.Stdout)
	})

	t.Run("Accept_First_Admin", func(t *testing.T) {
		// Set context to use first admin's address
		contextResult, err := h.ExecuteCLI("context", "set", "--operator-address", firstAdminAddress)
		require.NoError(t, err)
		assert.Equal(t, 0, contextResult.ExitCode)

		// First admin accepts
		result, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered2ECDSA,
			"eigenlayer", "user", "admin", "accept",
			"--account-address", accountAddress)

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.Contains(t, result.Stdout, "Successfully accepted admin")
		t.Logf("Accept first admin output: %s", result.Stdout)
	})

	t.Run("Verify_First_Admin_Is_Admin", func(t *testing.T) {
		result, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
			"eigenlayer", "user", "admin", "is-admin",
			"--account-address", accountAddress,
			"--admin-address", firstAdminAddress)

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		t.Logf("Is-admin check for first admin: %s", result.Stdout)
	})

	t.Run("First_Admin_Adds_Second_Admin", func(t *testing.T) {
		// Set context to first admin
		contextResult, err := h.ExecuteCLI("context", "set", "--operator-address", firstAdminAddress)
		require.NoError(t, err)
		assert.Equal(t, 0, contextResult.ExitCode)

		// First admin adds second admin as pending
		result, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered2ECDSA,
			"eigenlayer", "user", "admin", "add",
			"--account-address", accountAddress,
			"--admin-address", secondAdminAddress)

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.Contains(t, result.Stdout, "Successfully added pending admin")
		t.Logf("First admin adds second admin: %s", result.Stdout)
	})

	t.Run("Verify_Second_Admin_In_Pending_List", func(t *testing.T) {
		result, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
			"eigenlayer", "user", "admin", "list-pending",
			"--account-address", accountAddress)

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.Contains(t, strings.ToLower(result.Stdout), strings.ToLower(secondAdminAddress))
		assert.NotContains(t, result.Stdout, firstAdminAddress)
		t.Logf("Pending admin list: %s", result.Stdout)
	})

	t.Run("Verify_Second_Admin_Is_Pending", func(t *testing.T) {
		result, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
			"eigenlayer", "user", "admin", "is-pending",
			"--account-address", accountAddress,
			"--admin-address", secondAdminAddress)

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		t.Logf("Is-pending check for second admin: %s", result.Stdout)
	})

	t.Run("Accept_Second_Admin", func(t *testing.T) {
		// Set context to second admin's address
		contextResult, err := h.ExecuteCLI("context", "set", "--operator-address", secondAdminAddress)
		require.NoError(t, err)
		assert.Equal(t, 0, contextResult.ExitCode)

		// Second admin accepts
		result, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreExecutor2ECDSA,
			"eigenlayer", "user", "admin", "accept",
			"--account-address", accountAddress)

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.Contains(t, result.Stdout, "Successfully accepted admin")
		t.Logf("Accept second admin output: %s", result.Stdout)
	})

	t.Run("Verify_All_Three_Admins_Listed", func(t *testing.T) {
		result, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
			"eigenlayer", "user", "admin", "list-admins",
			"--account-address", accountAddress)

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.Contains(t, strings.ToLower(result.Stdout), strings.ToLower(accountAddress), "Account should be in admin list")
		assert.Contains(t, strings.ToLower(result.Stdout), strings.ToLower(firstAdminAddress), "First admin should be in admin list")
		assert.Contains(t, strings.ToLower(result.Stdout), strings.ToLower(secondAdminAddress), "Second admin should be in admin list")
		t.Logf("All admins list: %s", result.Stdout)
	})

	t.Run("Remove_Second_Admin", func(t *testing.T) {
		// Set context to first admin (who can remove)
		contextResult, err := h.ExecuteCLI("context", "set", "--operator-address", firstAdminAddress)
		require.NoError(t, err)
		assert.Equal(t, 0, contextResult.ExitCode)

		// First admin removes second admin
		result, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered2ECDSA,
			"eigenlayer", "user", "admin", "remove-admin",
			"--account-address", accountAddress,
			"--admin-address", secondAdminAddress)

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.Contains(t, result.Stdout, "Successfully removed admin")
		t.Logf("Remove second admin output: %s", result.Stdout)
	})

	t.Run("Verify_Second_Admin_Removed", func(t *testing.T) {
		result, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
			"eigenlayer", "user", "admin", "list-admins",
			"--account-address", accountAddress)

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.Contains(t, strings.ToLower(result.Stdout), strings.ToLower(accountAddress))
		assert.Contains(t, strings.ToLower(result.Stdout), strings.ToLower(firstAdminAddress))
		assert.NotContains(t, strings.ToLower(result.Stdout), strings.ToLower(secondAdminAddress), "Second admin should be removed")
		t.Logf("Admin list after removing second: %s", result.Stdout)
	})

	t.Run("Remove_First_Admin", func(t *testing.T) {
		// Set context back to account
		contextResult, err := h.ExecuteCLI("context", "set", "--operator-address", accountAddress)
		require.NoError(t, err)
		assert.Equal(t, 0, contextResult.ExitCode)

		// Account removes first admin
		result, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
			"eigenlayer", "user", "admin", "remove-admin",
			"--admin-address", firstAdminAddress)

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.Contains(t, result.Stdout, "Successfully removed admin")
		t.Logf("Remove first admin output: %s", result.Stdout)
	})

	t.Run("Verify_Only_Account_Remains", func(t *testing.T) {
		result, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
			"eigenlayer", "user", "admin", "list-admins")

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
		assert.Contains(t, strings.ToLower(result.Stdout), strings.ToLower(accountAddress), "Account should still be an admin")
		assert.NotContains(t, strings.ToLower(result.Stdout), strings.ToLower(firstAdminAddress), "First admin should be removed")
		assert.NotContains(t, strings.ToLower(result.Stdout), strings.ToLower(secondAdminAddress), "Second admin should be removed")
		t.Logf("Final admin list: %s", result.Stdout)
	})
}

// TestAdminErrorCases tests error scenarios for admin management
func TestAdminErrorCases(t *testing.T) {
	skipIfShort(t)

	h := harness.NewTestHarness(t)
	require.NoError(t, h.Setup())
	defer h.Teardown()

	// Create a copy of the test context for this test suite
	testContextName := "admin-error-context"
	result, err := h.ExecuteCLI("context", "copy", "--copy-name", testContextName, "--use", "test")
	require.NoError(t, err)
	require.Equal(t, 0, result.ExitCode)

	defer func() {
		_, _ = h.ExecuteCLI("context", "use", "test")
		result, err := h.ExecuteCLI("context", "delete", testContextName)
		if err != nil || result.ExitCode != 0 {
			t.Logf("Warning: failed to delete test context %s: %v", testContextName, err)
		}
	}()

	accountAddress := h.ChainConfig.UnregisteredOperator1AccountAddress
	firstAdminAddress := h.ChainConfig.UnregisteredOperator2AccountAddress

	// Setup: Ensure account is self-appointed
	t.Run("Setup_SelfAppoint", func(t *testing.T) {
		contextResult, err := h.ExecuteCLI("context", "set", "--operator-address", accountAddress)
		require.NoError(t, err)
		require.Equal(t, 0, contextResult.ExitCode)

		listResult, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
			"eigenlayer", "user", "admin", "list-admins")
		require.NoError(t, err)
		require.Equal(t, 0, listResult.ExitCode)

		if !containsAddress(listResult.Stdout, accountAddress) {
			addResult, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
				"eigenlayer", "user", "admin", "add",
				"--admin-address", accountAddress)
			require.NoError(t, err)
			require.Equal(t, 0, addResult.ExitCode)

			acceptResult, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
				"eigenlayer", "user", "admin", "accept",
				"--account-address", accountAddress)
			require.NoError(t, err)
			require.Equal(t, 0, acceptResult.ExitCode)
		}
	})

	t.Run("Cannot_Add_Duplicate_Pending_Admin", func(t *testing.T) {
		// Add first admin as pending
		addResult, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
			"eigenlayer", "user", "admin", "add",
			"--admin-address", firstAdminAddress)
		require.NoError(t, err)
		require.Equal(t, 0, addResult.ExitCode)

		// Try to add same admin again (should fail)
		duplicateResult, _ := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
			"eigenlayer", "user", "admin", "add",
			"--admin-address", firstAdminAddress)

		// Should fail or show error
		assert.NotEqual(t, 0, duplicateResult.ExitCode, "Adding duplicate pending admin should fail")
		t.Logf("Duplicate add attempt output: %s", duplicateResult.Stdout)

		// Cleanup: Remove pending admin for next test
		removeResult, err := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered1ECDSA,
			"eigenlayer", "user", "admin", "remove-pending",
			"--admin-address", firstAdminAddress)
		require.NoError(t, err)
		require.Equal(t, 0, removeResult.ExitCode)
	})

	t.Run("Cannot_Accept_If_Not_Pending", func(t *testing.T) {
		// Set context to first admin
		contextResult, err := h.ExecuteCLI("context", "set", "--operator-address", firstAdminAddress)
		require.NoError(t, err)
		require.Equal(t, 0, contextResult.ExitCode)

		// Try to accept without being added as pending
		acceptResult, _ := h.ExecuteCLIWithOperatorKeystore(harness.KeystoreUnregistered2ECDSA,
			"eigenlayer", "user", "admin", "accept",
			"--account-address", accountAddress)

		// Should fail
		assert.NotEqual(t, 0, acceptResult.ExitCode, "Accepting without being pending should fail")
		t.Logf("Invalid accept attempt output: %s", acceptResult.Stdout)
	})
}

// Helper function to check if an address appears in the output
func containsAddress(output, address string) bool {
	// Normalize addresses for comparison (case-insensitive)
	output = strings.ToLower(output)
	address = strings.ToLower(address)
	return strings.Contains(output, address)
}

package integration

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/signer"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/testutils/client"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/testutils/harness"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestAggregatorDeployment(t *testing.T) {
	skipIfShort(t)

	h := harness.NewTestHarness(t)
	require.NoError(t, h.Setup())
	defer h.Teardown()

	aggregatorContext := "aggregator-deployment-context"
	executorContext := "executor-deployment-context"

	defer func() {
		_, _ = h.ExecuteCLI("context", "use", h.ContextName)
		_, _ = h.ExecuteCLI("context", "delete", aggregatorContext)
		_, _ = h.ExecuteCLI("context", "delete", executorContext)
	}()

	t.Run("Create Aggregator Context", func(t *testing.T) {
		result, err := h.ExecuteCLI("context", "copy", "--copy-name", aggregatorContext, "--use", h.ContextName)
		require.NoError(t, err, "Failed to copy context for aggregator")
		require.Equal(t, 0, result.ExitCode, "Context copy should succeed")

		showResult, err := h.ExecuteCLI("context", "show")
		require.NoError(t, err)
		t.Logf("Current context after copy: %s", showResult.Stdout)

		result, err = h.ExecuteCLI(
			"context",
			"set",
			"--operator-address", h.ChainConfig.OperatorAccountAddress,
			"--operator-set-id", "0",
		)
		require.NoError(t, err)
		require.Equal(t, 0, result.ExitCode)

		err = h.ConfigureSystemKey(harness.KeystoreAggregatorSystem)
		require.NoError(t, err, "Failed to configure aggregator system key")

		signerResult, err := h.ExecuteCLI("signer", "operator", "keystore",
			"--name", harness.KeystoreAggregatorECDSA)
		require.NoError(t, err, "Failed to configure operator signer")
		require.Equal(t, 0, signerResult.ExitCode)
	})

	t.Run("Deploy Aggregator", func(t *testing.T) {
		result, err := h.ExecuteCLI("context", "use", aggregatorContext)
		require.NoError(t, err)
		require.Equal(t, 0, result.ExitCode)

		t.Logf("Running aggregator with operator-set-id 0 (auto-discovery)...")
		result, err = h.ExecuteCLIWithOperatorKeystore(harness.KeystoreAggregatorECDSA, "run")

		if err != nil || result.ExitCode != 0 {
			t.Logf("Run command failed with error: %v", err)
			t.Logf("Exit code: %d", result.ExitCode)
			t.Logf("Stdout: %s", result.Stdout)
			t.Logf("Stderr: %s", result.Stderr)
		}

		require.NoError(t, err, "Run command should not return an error")
		require.Equal(t, 0, result.ExitCode, "Run should succeed")
		time.Sleep(10 * time.Second)
	})

	t.Run("Create Executor Context", func(t *testing.T) {
		result, err := h.ExecuteCLI("context", "use", h.ContextName)
		require.NoError(t, err)
		require.Equal(t, 0, result.ExitCode)

		result, err = h.ExecuteCLI("context", "copy", "--copy-name", executorContext, "--use", h.ContextName)
		require.NoError(t, err, "Failed to copy context for executor")
		require.Equal(t, 0, result.ExitCode, "Context copy should succeed")

		showResult, err := h.ExecuteCLI("context", "show")
		require.NoError(t, err)
		t.Logf("Current context after copy: %s", showResult.Stdout)

		result, err = h.ExecuteCLI(
			"context",
			"set",
			"--operator-address", h.ChainConfig.ExecOperatorAccountAddress,
			"--operator-set-id", "1",
		)
		require.NoError(t, err)
		require.Equal(t, 0, result.ExitCode)

		err = h.ConfigureSystemKey(harness.KeystoreExecutorSystem)
		require.NoError(t, err, "Failed to configure executor system key")

		signerResult, err := h.ExecuteCLI("signer", "operator", "keystore",
			"--name", harness.KeystoreExecutorECDSA)
		require.NoError(t, err, "Failed to configure operator signer")
		require.Equal(t, 0, signerResult.ExitCode)
	})

	t.Run("Deploy Executor", func(t *testing.T) {
		result, err := h.ExecuteCLI("context", "use", executorContext)
		require.NoError(t, err)
		require.Equal(t, 0, result.ExitCode)

		t.Logf("Running executor with operator-set-id 1 (auto-discovery)...")
		result, err = h.ExecuteCLIWithOperatorKeystore(harness.KeystoreExecutorECDSA, "run")

		if err != nil || result.ExitCode != 0 {
			t.Logf("Run command failed with error: %v", err)
			t.Logf("Exit code: %d", result.ExitCode)
			t.Logf("Stdout: %s", result.Stdout)
			t.Logf("Stderr: %s", result.Stderr)
		}

		require.NoError(t, err, "Run command should not return an error")
		require.Equal(t, 0, result.ExitCode, "Run should succeed")
		time.Sleep(10 * time.Second)
	})

	t.Run("Submit Task to Mailbox", func(t *testing.T) {
		s, err := signer.NewPrivateKeySigner(
			h.ChainConfig.AppAccountPk,
			h.L2Client,
			zap.NewNop(),
		)
		require.NoError(t, err)

		chainCaller, err := client.NewChainCaller(
			h.L2Client,
			s,
			*h.ChainConfig,
			h.Logger,
		)
		require.NoError(t, err)

		payload := []byte("test task payload for integration test")

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		t.Logf("Submitting task to mailbox...")
		receipt, err := chainCaller.PublishMessageToInbox(
			ctx,
			h.ChainConfig.AVSAccountAddress,
			1,
			payload,
		)

		if err != nil {
			t.Logf("Failed to submit task: %v", err)
		}

		require.NoError(t, err)
		require.NotNil(t, receipt)
		assert.Equal(t, uint64(1), receipt.Status, "Transaction should succeed")

		t.Logf("Successfully submitted task. Transaction hash: %s", receipt.TxHash.Hex())

		t.Logf("Waiting for task to be processed...")

		select {
		case <-time.After(10 * time.Second):
			t.Logf("Task processing timeout reached")
			t.FailNow()
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.Canceled) {
				t.Logf("Test completed")
				t.FailNow()
			} else {
				t.Logf("Context done with error: %v", ctx.Err())
			}
		}

		t.Logf("Task submission and processing test completed")
	})
}

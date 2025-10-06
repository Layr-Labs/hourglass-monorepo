package integration

import (
	"context"
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

	var aggregatorContext string
	var executorContext string

	t.Run("Create Aggregator Context", func(t *testing.T) {
		// Copy default context to aggregator-context
		aggregatorContext = "aggregator-context"
		result, err := h.ExecuteCLI(
			"context",
			"copy",
			h.ContextName,
			aggregatorContext,
		)
		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)

		// Switch to aggregator context
		result, err = h.ExecuteCLI(
			"context",
			"use",
			aggregatorContext,
		)
		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)

		// Set aggregator-specific configuration
		result, err = h.ExecuteCLI(
			"context",
			"set",
			"--operator-address", h.ChainConfig.OperatorAccountAddress,
			"--operator-set-id", "0",
		)
		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)

		// Configure operator keystore
		err = h.ConfigureSystemKey(harness.KeystoreAggregatorSystem)
		require.NoError(t, err)

		// Set operator keystore
		result, err = h.ExecuteCLIWithOperatorKeystore(harness.KeystoreAggregatorECDSA)
		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
	})

	t.Run("Create Executor Context", func(t *testing.T) {
		// Switch back to default context first
		result, err := h.ExecuteCLI(
			"context",
			"use",
			h.ContextName,
		)
		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)

		// Copy default context to executor-context
		executorContext = "executor-context"
		result, err = h.ExecuteCLI(
			"context",
			"copy",
			h.ContextName,
			executorContext,
		)
		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)

		// Switch to executor context
		result, err = h.ExecuteCLI(
			"context",
			"use",
			executorContext,
		)
		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)

		// Set executor-specific configuration
		result, err = h.ExecuteCLI(
			"context",
			"set",
			"--operator-address", h.ChainConfig.ExecOperatorAccountAddress,
			"--operator-set-id", "1",
		)
		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)

		// Configure system keystore
		err = h.ConfigureSystemKey(harness.KeystoreExecutorSystem)
		require.NoError(t, err)

		// Set operator keystore
		result, err = h.ExecuteCLIWithOperatorKeystore(harness.KeystoreExecutorECDSA)
		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
	})

	t.Run("Deploy Aggregator", func(t *testing.T) {
		// Switch to aggregator context
		result, err := h.ExecuteCLI(
			"context",
			"use",
			aggregatorContext,
		)
		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)

		// Deploy aggregator
		result, err = h.ExecuteCLIWithOperatorKeystore(harness.KeystoreAggregatorECDSA,
			"deploy",
			"aggregator",
			h.ChainConfig.AVSAccountAddress,
		)

		if err != nil || result.ExitCode != 0 {
			t.Logf("Deploy aggregator failed with error: %v", err)
			t.Logf("Exit code: %d", result.ExitCode)
			t.Logf("Stdout: %s", result.Stdout)
			t.Logf("Stderr: %s", result.Stderr)
		}

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)

		// Wait for aggregator to start
		time.Sleep(5 * time.Second)
	})

	t.Run("Deploy Executor", func(t *testing.T) {
		// Switch to executor context
		result, err := h.ExecuteCLI(
			"context",
			"use",
			executorContext,
		)
		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)

		// Deploy executor
		result, err = h.ExecuteCLIWithOperatorKeystore(harness.KeystoreExecutorECDSA,
			"deploy",
			"executor",
			h.ChainConfig.AVSAccountAddress,
		)

		if err != nil || result.ExitCode != 0 {
			t.Logf("Deploy executor failed with error: %v", err)
			t.Logf("Exit code: %d", result.ExitCode)
			t.Logf("Stdout: %s", result.Stdout)
			t.Logf("Stderr: %s", result.Stderr)
		}

		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)

		// Wait for executor to start
		time.Sleep(5 * time.Second)
	})

	t.Run("Submit Task to Mailbox", func(t *testing.T) {
		// Create transaction signer for task submission
		signer, err := signer.NewPrivateKeySigner(
			h.ChainConfig.AppAccountPk,
			h.L2Client,
			zap.NewNop(),
		)
		require.NoError(t, err)

		// Create chain caller
		chainCaller, err := client.NewChainCaller(
			h.L2Client,
			signer,
			*h.ChainConfig,
			h.Logger,
		)
		require.NoError(t, err)

		// Create task payload (simple test payload)
		payload := []byte("test task payload for integration test")

		// Submit task with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 260*time.Second)
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

		// Wait for task to be processed by aggregator and executor
		t.Logf("Waiting for task to be processed...")

		select {
		case <-time.After(240 * time.Second):
			t.Logf("Task processing timeout reached")
		case <-ctx.Done():
			if ctx.Err() == context.Canceled {
				t.Logf("Test completed")
			} else {
				t.Logf("Context done with error: %v", ctx.Err())
			}
		}

		// TODO: Add validation of task completion by checking logs or state
		t.Logf("Task submission and processing test completed")
	})
}

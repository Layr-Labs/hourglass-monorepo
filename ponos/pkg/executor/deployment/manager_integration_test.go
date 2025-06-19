package deployment

import (
	"context"
	"testing"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/containerManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer/serverPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering/localPeeringDataFetcher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestDeploymentManagerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	// Create a real container manager (Docker)
	containerMgr, err := containerManager.NewDockerContainerManager(nil, logger)
	require.NoError(t, err)

	// Ensure cleanup
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := containerMgr.Shutdown(shutdownCtx); err != nil {
			t.Logf("Warning: failed to shutdown container manager: %v", err)
		}
	}()

	// Create peering data fetcher (required by serverPerformer)
	pdf := localPeeringDataFetcher.NewLocalPeeringDataFetcher(&localPeeringDataFetcher.LocalPeeringDataFetcherConfig{
		AggregatorPeers: nil, // Empty for tests
	}, logger)

	t.Run("successful deployment with real container", func(t *testing.T) {
		// Create deployment manager
		deploymentMgr := NewManager(logger)

		// Create performer config
		performerConfig := &avsPerformer.AvsPerformerConfig{
			AvsAddress:           "0x1234567890abcdef",
			ProcessType:          "server",
			SigningCurve:         "bn254",
			PerformerNetworkName: "",
			WorkerCount:          1,
		}

		// Create a real serverPerformer with short health check interval for faster tests
		performer := serverPerformer.NewAvsPerformerServerWithHealthCheckInterval(
			performerConfig,
			pdf,
			logger,
			containerMgr,
			1*time.Second, // Short health check interval for tests
		)

		// Initialize the performer
		err := performer.Initialize(ctx)
		require.NoError(t, err)

		// Deploy with a test container image
		config := DeploymentConfig{
			AvsAddress: performerConfig.AvsAddress,
			Image: avsPerformer.PerformerImage{
				Repository: "hello-performer",
				Tag:        "latest",
			},
			Timeout: 30 * time.Second,
		}

		// Execute deployment
		result, err := deploymentMgr.Deploy(ctx, config, performer)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify deployment result
		assert.Equal(t, DeploymentStatusCompleted, result.Status)
		assert.NotEmpty(t, result.DeploymentID)
		assert.NotEmpty(t, result.PerformerID)
		assert.Contains(t, result.Message, "successfully")
		assert.True(t, result.EndTime.After(result.StartTime))

		// Verify deployment is no longer active
		_, exists := deploymentMgr.GetActiveDeployment(config.AvsAddress)
		assert.False(t, exists, "Deployment should not be active after completion")

		// Verify deployment is in history
		historicalResult, exists := deploymentMgr.GetDeploymentResult(result.DeploymentID)
		assert.True(t, exists, "Deployment should be in history")
		assert.Equal(t, result, historicalResult)

		// Cleanup: shutdown the performer
		err = performer.Shutdown()
		assert.NoError(t, err)
	})

	t.Run("deployment already in progress", func(t *testing.T) {
		// Create deployment manager
		deploymentMgr := NewManager(logger)

		// Create performer config
		performerConfig := &avsPerformer.AvsPerformerConfig{
			AvsAddress:           "0x2234567890abcdef",
			ProcessType:          "server",
			SigningCurve:         "bn254",
			PerformerNetworkName: "",
			WorkerCount:          1,
		}

		// Create serverPerformer with long health check interval to ensure deployment stays in progress
		performer := serverPerformer.NewAvsPerformerServerWithHealthCheckInterval(
			performerConfig,
			pdf,
			logger,
			containerMgr,
			60*time.Second, // Long interval to keep deployment in progress
		)

		err := performer.Initialize(ctx)
		require.NoError(t, err)

		// Start first deployment with long timeout
		config := DeploymentConfig{
			AvsAddress: performerConfig.AvsAddress,
			Image: avsPerformer.PerformerImage{
				Repository: "hello-performer",
				Tag:        "latest",
			},
			Timeout: 2 * time.Minute,
		}

		// Start deployment in background
		deploymentDone := make(chan struct{})
		go func() {
			_, _ = deploymentMgr.Deploy(ctx, config, performer)
			close(deploymentDone)
		}()

		// Wait for deployment to start
		time.Sleep(2 * time.Second)

		// Try to start another deployment for same AVS
		result, err := deploymentMgr.Deploy(ctx, config, performer)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrDeploymentInProgress)

		// Cancel the active deployment
		err = deploymentMgr.CancelDeployment(config.AvsAddress)
		assert.NoError(t, err)

		// Wait for deployment to finish
		select {
		case <-deploymentDone:
			// Good
		case <-time.After(10 * time.Second):
			t.Fatal("Deployment did not finish after cancellation")
		}

		// Cleanup
		err = performer.Shutdown()
		assert.NoError(t, err)
	})

	t.Run("deployment timeout", func(t *testing.T) {
		// Create deployment manager
		deploymentMgr := NewManager(logger)

		// Create performer config
		performerConfig := &avsPerformer.AvsPerformerConfig{
			AvsAddress:           "0x3334567890abcdef",
			ProcessType:          "server",
			SigningCurve:         "bn254",
			PerformerNetworkName: "",
			WorkerCount:          1,
		}

		// Create serverPerformer with very long health check interval to force timeout
		performer := serverPerformer.NewAvsPerformerServerWithHealthCheckInterval(
			performerConfig,
			pdf,
			logger,
			containerMgr,
			5*time.Minute, // Very long interval to ensure timeout
		)

		err := performer.Initialize(ctx)
		require.NoError(t, err)

		// Deploy with short timeout
		config := DeploymentConfig{
			AvsAddress: performerConfig.AvsAddress,
			Image: avsPerformer.PerformerImage{
				Repository: "hello-performer",
				Tag:        "latest",
			},
			Timeout: 3 * time.Second, // Short timeout to trigger failure
		}

		// Execute deployment
		result, err := deploymentMgr.Deploy(ctx, config, performer)
		assert.Error(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, DeploymentStatusFailed, result.Status)
		assert.ErrorIs(t, err, ErrDeploymentTimeout)
		assert.Contains(t, result.Message, "Deployment timed out")

		// Cleanup
		err = performer.Shutdown()
		assert.NoError(t, err)
	})

	t.Run("deployment cancelled", func(t *testing.T) {
		// Create a cancellable context
		deployCtx, cancel := context.WithCancel(ctx)

		// Create deployment manager
		deploymentMgr := NewManager(logger)

		// Create performer config
		performerConfig := &avsPerformer.AvsPerformerConfig{
			AvsAddress:           "0x4434567890abcdef",
			ProcessType:          "server",
			SigningCurve:         "bn254",
			PerformerNetworkName: "",
			WorkerCount:          1,
		}

		// Create serverPerformer
		performer := serverPerformer.NewAvsPerformerServerWithHealthCheckInterval(
			performerConfig,
			pdf,
			logger,
			containerMgr,
			30*time.Second, // Moderate interval
		)

		err := performer.Initialize(ctx)
		require.NoError(t, err)

		// Deploy configuration
		config := DeploymentConfig{
			AvsAddress: performerConfig.AvsAddress,
			Image: avsPerformer.PerformerImage{
				Repository: "hello-performer",
				Tag:        "latest",
			},
			Timeout: 1 * time.Minute,
		}

		// Start deployment in background
		resultChan := make(chan *DeploymentResult)
		errChan := make(chan error)
		go func() {
			result, err := deploymentMgr.Deploy(deployCtx, config, performer)
			resultChan <- result
			errChan <- err
		}()

		// Wait for deployment to start, then cancel
		time.Sleep(2 * time.Second)
		cancel()

		// Get result
		result := <-resultChan
		err = <-errChan

		assert.Error(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, DeploymentStatusCancelled, result.Status)
		assert.ErrorIs(t, err, ErrDeploymentCancelled)

		// Cleanup
		cleanupErr := performer.Shutdown()
		assert.NoError(t, cleanupErr)
	})

	// Note: Promotion failure test would require manipulating the container health
	// or mocking the promotion mechanism, which is harder in an integration test
}

// TestDeploymentManagerMultipleDeployments tests deploying to multiple AVS addresses
func TestDeploymentManagerMultipleDeployments(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	// Create container manager
	containerMgr, err := containerManager.NewDockerContainerManager(nil, logger)
	require.NoError(t, err)
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = containerMgr.Shutdown(shutdownCtx)
	}()

	// Create deployment manager
	deploymentMgr := NewManager(logger)

	// Create peering data fetcher
	pdf := localPeeringDataFetcher.NewLocalPeeringDataFetcher(&localPeeringDataFetcher.LocalPeeringDataFetcherConfig{
		AggregatorPeers: nil,
	}, logger)

	// Deploy to multiple AVS addresses concurrently
	avsAddresses := []string{
		"0xaaa4567890abcdef",
		"0xbbb4567890abcdef",
		"0xccc4567890abcdef",
	}

	type deploymentResult struct {
		avsAddress string
		result     *DeploymentResult
		err        error
	}

	resultChan := make(chan deploymentResult, len(avsAddresses))

	for _, avsAddr := range avsAddresses {
		avsAddr := avsAddr // Capture for goroutine
		go func() {
			// Create performer for this AVS
			performerConfig := &avsPerformer.AvsPerformerConfig{
				AvsAddress:           avsAddr,
				ProcessType:          "server",
				SigningCurve:         "bn254",
				PerformerNetworkName: "",
				WorkerCount:          1,
			}

			performer := serverPerformer.NewAvsPerformerServerWithHealthCheckInterval(
				performerConfig,
				pdf,
				logger,
				containerMgr,
				1*time.Second,
			)

			if err := performer.Initialize(ctx); err != nil {
				resultChan <- deploymentResult{avsAddress: avsAddr, err: err}
				return
			}

			// Deploy
			config := DeploymentConfig{
				AvsAddress: avsAddr,
				Image: avsPerformer.PerformerImage{
					Repository: "hello-performer",
					Tag:        "latest",
				},
				Timeout: 30 * time.Second,
			}

			result, err := deploymentMgr.Deploy(ctx, config, performer)
			resultChan <- deploymentResult{
				avsAddress: avsAddr,
				result:     result,
				err:        err,
			}

			// Note: In real usage, we wouldn't shutdown immediately after deployment
			// This is just for test cleanup
			_ = performer.Shutdown()
		}()
	}

	// Collect results
	successCount := 0
	for i := 0; i < len(avsAddresses); i++ {
		res := <-resultChan
		if res.err == nil {
			successCount++
			assert.Equal(t, DeploymentStatusCompleted, res.result.Status)
			t.Logf("Deployment to %s completed successfully", res.avsAddress)
		} else {
			t.Logf("Deployment to %s failed: %v", res.avsAddress, res.err)
		}
	}

	// All deployments should succeed
	assert.Equal(t, len(avsAddresses), successCount, "All deployments should succeed")
}

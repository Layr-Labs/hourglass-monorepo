package avsContainerPerformer

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/containerManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering/localPeeringDataFetcher"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/performerTask"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// testContainerManager wraps a real ContainerManager to inject TASK_DELAY_MS environment variable
type testContainerManager struct {
	containerManager.ContainerManager
	taskDelayMs int
}

// Create overrides the Create method to inject environment variables
func (tcm *testContainerManager) Create(ctx context.Context, config *containerManager.ContainerConfig) (*containerManager.ContainerInfo, error) {
	// Clone the config to avoid modifying the original
	modifiedConfig := *config

	// Add the task delay environment variable
	if tcm.taskDelayMs > 0 {
		envVar := fmt.Sprintf("TASK_DELAY_MS=%d", tcm.taskDelayMs)
		modifiedConfig.Env = append(modifiedConfig.Env, envVar)
		// Log what we're setting for debugging
		fmt.Printf("Setting environment variable: %s\n", envVar)
		fmt.Printf("TFull environment: %v\n", modifiedConfig.Env)
	}

	// Call the underlying container manager with the modified config
	return tcm.ContainerManager.Create(ctx, &modifiedConfig)
}

func TestPerformerDrainingIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	// Create base container manager
	baseContainerMgr, err := containerManager.NewDockerContainerManager(containerManager.DefaultContainerManagerConfig(), logger)
	require.NoError(t, err)

	// Create peering data fetcher
	pdf := localPeeringDataFetcher.NewLocalPeeringDataFetcher(&localPeeringDataFetcher.LocalPeeringDataFetcherConfig{
		AggregatorPeers: nil,
	}, logger)

	t.Run("draining during promotion and routing to new performer", func(t *testing.T) {
		// Create performer config
		performerConfig := &avsPerformer.AvsPerformerConfig{
			AvsAddress:                     "0x3334567890abcdef",
			ProcessType:                    "server",
			SigningCurve:                   "bn254",
			PerformerNetworkName:           "",
			WorkerCount:                    1,
			ApplicationHealthCheckInterval: 1 * time.Second,
		}

		// Wrap container manager to inject 2-second task delay
		containerMgr := &testContainerManager{
			ContainerManager: baseContainerMgr,
			taskDelayMs:      5000,
		}

		// Create serverPerformer with wrapped container manager
		server := NewAvsContainerPerformerWithContainerManager(
			performerConfig,
			pdf,
			logger,
			containerMgr,
		)
		defer func() {
			assert.NoError(t, server.Shutdown())
		}()

		// Initialize without initial container
		err = server.Initialize(ctx)
		require.NoError(t, err)

		// Deploy the first performer
		firstDeployment, err := server.Deploy(ctx, avsPerformer.PerformerImage{
			Repository: "sleepy-hello-performer",
			Tag:        "latest",
		})
		require.NoError(t, err)
		require.Equal(t, avsPerformer.DeploymentStatusCompleted, firstDeployment.Status)

		// Start a long-running task on current performer
		// Channel to signal when the task has started
		numTasks := 1
		taskStarted := make(chan struct{}, numTasks)
		var wg sync.WaitGroup
		wg.Add(numTasks)

		go func() {
			defer wg.Done()
			task := &performerTask.PerformerTask{
				TaskID:            "long-task",
				AggregatorAddress: "0xaggregator",
				Payload:           []byte("0x0A"),
				Signature:         []byte{},
			}

			// Signal that we're about to start the task
			taskStarted <- struct{}{}
			_, _ = server.RunTask(ctx, task)
		}()

		<-taskStarted
		time.Sleep(100 * time.Millisecond)

		secondDeployment, err := server.Deploy(ctx, avsPerformer.PerformerImage{
			Repository: "sleepy-hello-performer",
			Tag:        "latest",
		})
		require.NoError(t, err)
		require.Equal(t, avsPerformer.DeploymentStatusCompleted, secondDeployment.Status)

		// Verify old performer is in draining state
		server.drainingPerformersMu.Lock()
		_, isDraining := server.drainingPerformers[firstDeployment.PerformerID]
		server.drainingPerformersMu.Unlock()
		assert.True(t, isDraining, "Old performer should be draining")

		// Wait for draining to complete (container to be removed)
		// Poll for completion instead of fixed sleep
		for i := 0; i < 50; i++ {
			time.Sleep(100 * time.Millisecond)
			server.drainingPerformersMu.Lock()
			_, stillDraining := server.drainingPerformers[firstDeployment.PerformerID]
			server.drainingPerformersMu.Unlock()
			if !stillDraining {
				break
			}
		}

		wg.Wait()

		// Verify old performer is no longer draining
		server.drainingPerformersMu.Lock()
		_, isDraining = server.drainingPerformers[firstDeployment.PerformerID]
		server.drainingPerformersMu.Unlock()
		assert.False(t, isDraining, "Old performer should no longer be draining")

		// Verify only new performer remains
		performers := server.ListPerformers()
		require.Len(t, performers, 1)
		assert.Equal(t, secondDeployment.PerformerID, performers[0].PerformerID)
	})

	t.Run("immediate drain when no tasks", func(t *testing.T) {
		// Create performer config
		performerConfig := &avsPerformer.AvsPerformerConfig{
			AvsAddress:                     "0x4434567890abcdef",
			ProcessType:                    "server",
			SigningCurve:                   "bn254",
			PerformerNetworkName:           "",
			WorkerCount:                    1,
			ApplicationHealthCheckInterval: 1 * time.Second,
		}

		// Create serverPerformer
		server := NewAvsContainerPerformerWithContainerManager(
			performerConfig,
			pdf,
			logger,
			baseContainerMgr,
		)
		defer func() {
			assert.NoError(t, server.Shutdown())
		}()

		// Initialize
		err = server.Initialize(ctx)
		require.NoError(t, err)

		// Deploy both performers
		// Deploy first performer
		deployResult1, err := server.Deploy(ctx, avsPerformer.PerformerImage{
			Repository: "sleepy-hello-performer",
			Tag:        "latest",
		})
		require.NoError(t, err)
		require.Equal(t, avsPerformer.DeploymentStatusCompleted, deployResult1.Status)

		// Deploy second performer (becomes next)
		deployResult2, err := server.Deploy(ctx, avsPerformer.PerformerImage{
			Repository: "sleepy-hello-performer",
			Tag:        "latest",
		})
		require.NoError(t, err)
		require.Equal(t, avsPerformer.DeploymentStatusCompleted, deployResult2.Status)

		// Verify only new performer remains
		performers := server.ListPerformers()
		require.Len(t, performers, 1)
		assert.Equal(t, deployResult2.PerformerID, performers[0].PerformerID)
	})
}

func TestDeploymentIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	// Create container manager
	containerMgr, err := containerManager.NewDockerContainerManager(containerManager.DefaultContainerManagerConfig(), logger)
	require.NoError(t, err)

	// Create peering data fetcher
	pdf := localPeeringDataFetcher.NewLocalPeeringDataFetcher(&localPeeringDataFetcher.LocalPeeringDataFetcherConfig{
		AggregatorPeers: nil,
	}, logger)

	t.Run("successful deployment with real container", func(t *testing.T) {
		// Create performer config
		performerConfig := &avsPerformer.AvsPerformerConfig{
			AvsAddress:                     "0x8834567890abcdef",
			ProcessType:                    "server",
			SigningCurve:                   "bn254",
			PerformerNetworkName:           "",
			WorkerCount:                    1,
			ApplicationHealthCheckInterval: 1 * time.Second,
		}

		// Create container performer
		server := NewAvsContainerPerformerWithContainerManager(
			performerConfig,
			pdf,
			logger,
			containerMgr,
		)
		defer func() {
			assert.NoError(t, server.Shutdown())
		}()

		// Initialize the performer
		err = server.Initialize(ctx)
		require.NoError(t, err)

		// Deploy with a test container image
		result, err := server.Deploy(ctx, avsPerformer.PerformerImage{
			Repository: "sleepy-hello-performer",
			Tag:        "latest",
		})
		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify deployment result
		assert.Equal(t, avsPerformer.DeploymentStatusCompleted, result.Status)
		assert.NotEmpty(t, result.ID)
		assert.NotEmpty(t, result.PerformerID)
		assert.Contains(t, result.Message, "successfully")
		assert.True(t, result.EndTime.After(result.StartTime))

		// Verify performer is in service
		performers := server.ListPerformers()
		require.Len(t, performers, 1)
		assert.Equal(t, result.PerformerID, performers[0].PerformerID)
		assert.Equal(t, avsPerformer.PerformerResourceStatusInService, performers[0].Status)
	})

	t.Run("deployment timeout", func(t *testing.T) {
		// Create performer config with very short timeout
		performerConfig := &avsPerformer.AvsPerformerConfig{
			AvsAddress:                     "0x9934567890abcdef",
			ProcessType:                    "server",
			SigningCurve:                   "bn254",
			PerformerNetworkName:           "",
			WorkerCount:                    1,
			ApplicationHealthCheckInterval: 5 * time.Minute, // Very long interval to force timeout
		}

		// Create a custom container manager that delays container creation
		slowContainerMgr := &testContainerManager{
			ContainerManager: containerMgr,
			taskDelayMs:      0,
		}

		// Create container performer
		server := NewAvsContainerPerformerWithContainerManager(
			performerConfig,
			pdf,
			logger,
			slowContainerMgr,
		)
		defer func() {
			assert.NoError(t, server.Shutdown())
		}()

		// Initialize
		err = server.Initialize(ctx)
		require.NoError(t, err)

		// Create a context with very short timeout
		deployCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()

		// Execute deployment
		result, err := server.Deploy(deployCtx, avsPerformer.PerformerImage{
			Repository: "sleepy-hello-performer",
			Tag:        "latest",
		})

		// Should get a timeout error
		assert.Error(t, err)
		if result != nil {
			assert.Equal(t, avsPerformer.DeploymentStatusFailed, result.Status)
			assert.Contains(t, result.Message, "timeout")
		}
	})

	t.Run("deployment cancelled", func(t *testing.T) {
		// Create performer config
		performerConfig := &avsPerformer.AvsPerformerConfig{
			AvsAddress:                     "0xaa34567890abcdef",
			ProcessType:                    "server",
			SigningCurve:                   "bn254",
			PerformerNetworkName:           "",
			WorkerCount:                    1,
			ApplicationHealthCheckInterval: 30 * time.Second,
		}

		// Create container performer
		server := NewAvsContainerPerformerWithContainerManager(
			performerConfig,
			pdf,
			logger,
			containerMgr,
		)
		defer func() {
			assert.NoError(t, server.Shutdown())
		}()

		// Initialize
		err = server.Initialize(ctx)
		require.NoError(t, err)

		// Create a cancellable context
		deployCtx, cancel := context.WithCancel(ctx)

		// Start deployment in background
		resultChan := make(chan *avsPerformer.DeploymentResult)
		errChan := make(chan error)
		go func() {
			result, err := server.Deploy(deployCtx, avsPerformer.PerformerImage{
				Repository: "sleepy-hello-performer",
				Tag:        "latest",
			})
			resultChan <- result
			errChan <- err
		}()

		// Wait for deployment to start, then cancel
		time.Sleep(500 * time.Millisecond)
		cancel()

		// Get result
		result := <-resultChan
		err = <-errChan

		assert.Error(t, err)
		if result != nil {
			// Context cancellation should result in failed status
			assert.Equal(t, avsPerformer.DeploymentStatusFailed, result.Status)
		}
	})

	t.Run("concurrent deployment protection", func(t *testing.T) {
		// Create performer config
		performerConfig := &avsPerformer.AvsPerformerConfig{
			AvsAddress:                     "0xbb34567890abcdef",
			ProcessType:                    "server",
			SigningCurve:                   "bn254",
			PerformerNetworkName:           "",
			WorkerCount:                    1,
			ApplicationHealthCheckInterval: 1 * time.Second,
		}

		// Create container performer
		server := NewAvsContainerPerformerWithContainerManager(
			performerConfig,
			pdf,
			logger,
			containerMgr,
		)
		defer func() {
			assert.NoError(t, server.Shutdown())
		}()

		// Initialize
		err = server.Initialize(ctx)
		require.NoError(t, err)

		// Start first deployment in a goroutine
		var wg sync.WaitGroup
		wg.Add(2)

		deploymentResults := make(chan *avsPerformer.DeploymentResult, 2)
		deploymentErrors := make(chan error, 2)

		// Launch two concurrent deployments
		for i := 0; i < 2; i++ {
			go func() {
				defer wg.Done()
				result, err := server.Deploy(ctx, avsPerformer.PerformerImage{
					Repository: "sleepy-hello-performer",
					Tag:        "latest",
				})
				deploymentResults <- result
				deploymentErrors <- err
			}()
		}

		wg.Wait()
		close(deploymentResults)
		close(deploymentErrors)

		// Collect results
		var results []*avsPerformer.DeploymentResult
		var errors []error
		for result := range deploymentResults {
			results = append(results, result)
		}
		for err := range deploymentErrors {
			errors = append(errors, err)
		}

		// One should succeed, one should fail
		successCount := 0
		failureCount := 0
		for i, err := range errors {
			if err == nil {
				successCount++
				assert.NotNil(t, results[i])
				assert.Equal(t, avsPerformer.DeploymentStatusCompleted, results[i].Status)
			} else {
				failureCount++
				assert.Contains(t, err.Error(), "deployment in progress")
			}
		}

		assert.Equal(t, 1, successCount, "Exactly one deployment should succeed")
		assert.Equal(t, 1, failureCount, "Exactly one deployment should fail")

		// Verify we have exactly one performer
		performers := server.ListPerformers()
		assert.Len(t, performers, 1)
		assert.Equal(t, avsPerformer.PerformerResourceStatusInService, performers[0].Status)
	})
}

func TestListPerformersIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	// Create container manager
	containerMgr, err := containerManager.NewDockerContainerManager(containerManager.DefaultContainerManagerConfig(), logger)
	require.NoError(t, err)

	// Create peering data fetcher
	pdf := localPeeringDataFetcher.NewLocalPeeringDataFetcher(&localPeeringDataFetcher.LocalPeeringDataFetcherConfig{
		AggregatorPeers: nil,
	}, logger)

	t.Run("list performers empty state", func(t *testing.T) {
		// Create performer config
		performerConfig := &avsPerformer.AvsPerformerConfig{
			AvsAddress:                     "0xcc34567890abcdef",
			ProcessType:                    "server",
			SigningCurve:                   "bn254",
			PerformerNetworkName:           "",
			WorkerCount:                    1,
			ApplicationHealthCheckInterval: 1 * time.Second,
		}

		// Create container performer
		server := NewAvsContainerPerformerWithContainerManager(
			performerConfig,
			pdf,
			logger,
			containerMgr,
		)
		defer func() {
			assert.NoError(t, server.Shutdown())
		}()

		// Initialize without any deployment
		err = server.Initialize(ctx)
		require.NoError(t, err)

		// List performers - should be empty
		performers := server.ListPerformers()
		assert.Empty(t, performers, "Should have no performers initially")
	})

	t.Run("list performers with single deployment", func(t *testing.T) {
		// Create performer config
		performerConfig := &avsPerformer.AvsPerformerConfig{
			AvsAddress:                     "0xdd34567890abcdef",
			ProcessType:                    "server",
			SigningCurve:                   "bn254",
			PerformerNetworkName:           "",
			WorkerCount:                    1,
			ApplicationHealthCheckInterval: 1 * time.Second,
		}

		// Create container performer
		server := NewAvsContainerPerformerWithContainerManager(
			performerConfig,
			pdf,
			logger,
			containerMgr,
		)
		defer func() {
			assert.NoError(t, server.Shutdown())
		}()

		// Initialize
		err = server.Initialize(ctx)
		require.NoError(t, err)

		// Deploy a container
		result, err := server.Deploy(ctx, avsPerformer.PerformerImage{
			Repository: "sleepy-hello-performer",
			Tag:        "latest",
		})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, avsPerformer.DeploymentStatusCompleted, result.Status)

		// List performers - should have one
		performers := server.ListPerformers()
		require.Len(t, performers, 1)

		// Verify performer details
		p := performers[0]
		assert.Equal(t, result.PerformerID, p.PerformerID)
		assert.Equal(t, performerConfig.AvsAddress, p.AvsAddress)
		assert.Equal(t, avsPerformer.PerformerResourceStatusInService, p.Status)
		assert.Equal(t, "sleepy-hello-performer", p.ArtifactRegistry)
		assert.Equal(t, "latest", p.ArtifactDigest)
		assert.True(t, p.ContainerHealthy)
		assert.True(t, p.ApplicationHealthy)
		assert.NotEmpty(t, p.ResourceID)
		assert.False(t, p.LastHealthCheck.IsZero())
	})

	t.Run("list performers after removal", func(t *testing.T) {
		// Create performer config
		performerConfig := &avsPerformer.AvsPerformerConfig{
			AvsAddress:                     "0xee34567890abcdef",
			ProcessType:                    "server",
			SigningCurve:                   "bn254",
			PerformerNetworkName:           "",
			WorkerCount:                    1,
			ApplicationHealthCheckInterval: 1 * time.Second,
		}

		// Create container performer
		server := NewAvsContainerPerformerWithContainerManager(
			performerConfig,
			pdf,
			logger,
			containerMgr,
		)
		defer func() {
			assert.NoError(t, server.Shutdown())
		}()

		// Initialize
		err = server.Initialize(ctx)
		require.NoError(t, err)

		// Deploy first performer
		result1, err := server.Deploy(ctx, avsPerformer.PerformerImage{
			Repository: "sleepy-hello-performer",
			Tag:        "latest",
		})
		require.NoError(t, err)
		require.Equal(t, avsPerformer.DeploymentStatusCompleted, result1.Status)

		// Use CreatePerformer to create a staged performer (won't auto-promote)
		createResult, err := server.CreatePerformer(ctx, avsPerformer.PerformerImage{
			Repository: "sleepy-hello-performer",
			Tag:        "latest",
		})
		require.NoError(t, err)
		require.NotNil(t, createResult)

		// Verify we have two performers
		performers := server.ListPerformers()
		assert.Len(t, performers, 2)

		// Find the current and staged performers
		var currentPerformerID, stagedPerformerID string
		for _, p := range performers {
			if p.Status == avsPerformer.PerformerResourceStatusInService {
				currentPerformerID = p.PerformerID
			} else if p.Status == avsPerformer.PerformerResourceStatusStaged {
				stagedPerformerID = p.PerformerID
			}
		}
		assert.NotEmpty(t, currentPerformerID)
		assert.NotEmpty(t, stagedPerformerID)

		// Remove the staged performer
		err = server.RemovePerformer(ctx, stagedPerformerID)
		require.NoError(t, err)

		// List again - should only have current
		performers = server.ListPerformers()
		require.Len(t, performers, 1)
		assert.Equal(t, currentPerformerID, performers[0].PerformerID)
		assert.Equal(t, avsPerformer.PerformerResourceStatusInService, performers[0].Status)

		// Remove the current performer
		err = server.RemovePerformer(ctx, currentPerformerID)
		require.NoError(t, err)

		// List again - should be empty
		performers = server.ListPerformers()
		assert.Empty(t, performers)
	})
}

package taskSession

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	"github.com/Layr-Labs/crypto-libs/pkg/ecdsa"
	executorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/operatorManager"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
)

// Mock structures for testing

type MockOperatorPeerInfo struct {
	mock.Mock
	OperatorAddress string
	OperatorSets    []*peering.OperatorSet
}

func (m *MockOperatorPeerInfo) GetOperatorSet(operatorSetId uint32) (*peering.OperatorSet, error) {
	for _, os := range m.OperatorSets {
		if os.OperatorSetID == operatorSetId {
			return os, nil
		}
	}
	return nil, fmt.Errorf("operator set with ID %d not found", operatorSetId)
}

func (m *MockOperatorPeerInfo) GetSocketForOperatorSet(operatorSetId uint32) (string, error) {
	os, err := m.GetOperatorSet(operatorSetId)
	if err != nil {
		return "", err
	}
	return os.NetworkAddress, nil
}

func (m *MockOperatorPeerInfo) IncludesOperatorSetId(operatorSetId uint32) bool {
	for _, os := range m.OperatorSets {
		if os.OperatorSetID == operatorSetId {
			return true
		}
	}
	return false
}

type MockExecutorServiceClient struct {
	mock.Mock
	shouldFail      bool
	response        *executorV1.TaskResult
	responseDelay   time.Duration
	responseSize    int
	taskSubmissions []*executorV1.TaskSubmission
}

func (m *MockExecutorServiceClient) SubmitTask(ctx context.Context, taskSubmission *executorV1.TaskSubmission, opts ...grpc.CallOption) (*executorV1.TaskResult, error) {
	m.taskSubmissions = append(m.taskSubmissions, taskSubmission)

	if m.responseDelay > 0 {
		select {
		case <-time.After(m.responseDelay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if m.shouldFail {
		return nil, errors.New("mock executor client error")
	}

	return m.response, nil
}

// Note: Only implementing the methods we need for testing
// Additional gRPC methods would be implemented if needed

// Test helper functions

func createBN254TestOperators(count int) ([]*peering.OperatorPeerInfo, []*bn254.PrivateKey, error) {
	operators := make([]*peering.OperatorPeerInfo, count)
	privateKeys := make([]*bn254.PrivateKey, count)

	for i := 0; i < count; i++ {
		privKey, pubKey, err := bn254.GenerateKeyPair()
		if err != nil {
			return nil, nil, err
		}

		privateKeys[i] = privKey
		operators[i] = &peering.OperatorPeerInfo{
			OperatorAddress: fmt.Sprintf("0x%040x", i+1),
			OperatorSets: []*peering.OperatorSet{
				{
					OperatorSetID: 1,
					WrappedPublicKey: peering.WrappedPublicKey{
						PublicKey: pubKey,
					},
					NetworkAddress: fmt.Sprintf("localhost:900%d", i),
				},
			},
		}
	}

	return operators, privateKeys, nil
}

func createECDSATestOperators(count int) ([]*peering.OperatorPeerInfo, []*ecdsa.PrivateKey, error) {
	operators := make([]*peering.OperatorPeerInfo, count)
	privateKeys := make([]*ecdsa.PrivateKey, count)

	for i := 0; i < count; i++ {
		privKey, _, err := ecdsa.GenerateKeyPair()
		if err != nil {
			return nil, nil, err
		}

		ecdsaAddr, err := privKey.DeriveAddress()
		if err != nil {
			return nil, nil, err
		}

		privateKeys[i] = privKey
		operators[i] = &peering.OperatorPeerInfo{
			OperatorAddress: ecdsaAddr.String(),
			OperatorSets: []*peering.OperatorSet{
				{
					OperatorSetID: 1,
					WrappedPublicKey: peering.WrappedPublicKey{
						ECDSAAddress: ecdsaAddr,
					},
					NetworkAddress: fmt.Sprintf("localhost:900%d", i),
				},
			},
		}
	}

	return operators, privateKeys, nil
}

func createTestTask() *types.Task {
	deadline := time.Now().Add(10 * time.Minute)
	return &types.Task{
		TaskId:              "0x1234567890abcdef",
		AVSAddress:          "0xavsaddress",
		OperatorSetId:       1,
		ThresholdBips:       7500, // 75%
		Payload:             []byte("test-payload"),
		DeadlineUnixSeconds: &deadline,
	}
}

// Test cases

func TestNewBN254TaskSession(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		operators, _, err := createBN254TestOperators(3)
		require.NoError(t, err)

		task := createTestTask()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		operatorPeersWeight := &operatorManager.PeerWeight{
			Operators: operators,
		}

		logger := zaptest.NewLogger(t)

		session, err := NewBN254TaskSession(
			ctx, cancel, task, "0xaggregator", []byte("signature"),
			operatorPeersWeight, logger,
		)

		require.NoError(t, err)
		require.NotNil(t, session)
		assert.Equal(t, task, session.Task)
		assert.Equal(t, "0xaggregator", session.aggregatorAddress)
		assert.Equal(t, []byte("signature"), session.aggregatorSignature)
		assert.Equal(t, operatorPeersWeight, session.operatorPeersWeight)
		assert.Equal(t, uint32(0), session.resultsCount.Load())
		assert.False(t, session.thresholdMet.Load())
	})

	t.Run("invalid operator set", func(t *testing.T) {
		operators, _, err := createBN254TestOperators(1)
		require.NoError(t, err)

		task := createTestTask()
		task.OperatorSetId = 999 // Non-existent operator set

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		operatorPeersWeight := &operatorManager.PeerWeight{
			Operators: operators,
		}

		logger := zaptest.NewLogger(t)

		session, err := NewBN254TaskSession(
			ctx, cancel, task, "0xaggregator", []byte("signature"),
			operatorPeersWeight, logger,
		)

		require.Error(t, err)
		require.Nil(t, session)
		assert.Contains(t, err.Error(), "failed to get operator set")
	})
}

func TestNewECDSATaskSession(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		operators, _, err := createECDSATestOperators(3)
		require.NoError(t, err)

		task := createTestTask()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		operatorPeersWeight := &operatorManager.PeerWeight{
			Operators: operators,
		}

		logger := zaptest.NewLogger(t)

		session, err := NewECDSATaskSession(
			ctx, cancel, task, "0xaggregator", []byte("signature"),
			operatorPeersWeight, logger,
		)

		require.NoError(t, err)
		require.NotNil(t, session)
		assert.Equal(t, task, session.Task)
		assert.Equal(t, "0xaggregator", session.aggregatorAddress)
		assert.Equal(t, []byte("signature"), session.aggregatorSignature)
		assert.Equal(t, operatorPeersWeight, session.operatorPeersWeight)
	})
}

func TestTaskSession_Process_ContextHandling(t *testing.T) {
	t.Run("context timeout during process", func(t *testing.T) {
		operators, _, err := createBN254TestOperators(1)
		require.NoError(t, err)

		task := createTestTask()
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		operatorPeersWeight := &operatorManager.PeerWeight{
			Operators: operators,
		}

		logger := zaptest.NewLogger(t)

		session, err := NewBN254TaskSession(
			ctx, cancel, task, "0xaggregator", []byte("signature"),
			operatorPeersWeight, logger,
		)
		require.NoError(t, err)

		// The process will timeout because the mock gRPC client will fail to connect
		cert, err := session.Process()
		require.Error(t, err)
		require.Nil(t, cert)
		assert.Contains(t, err.Error(), "deadline exceeded")
	})

	t.Run("context cancellation", func(t *testing.T) {
		operators, _, err := createBN254TestOperators(1)
		require.NoError(t, err)

		task := createTestTask()
		ctx, cancel := context.WithCancel(context.Background())

		operatorPeersWeight := &operatorManager.PeerWeight{
			Operators: operators,
		}

		logger := zaptest.NewLogger(t)

		session, err := NewBN254TaskSession(
			ctx, cancel, task, "0xaggregator", []byte("signature"),
			operatorPeersWeight, logger,
		)
		require.NoError(t, err)

		// Cancel context immediately
		cancel()

		cert, err := session.Process()
		require.Error(t, err)
		require.Nil(t, cert)
		assert.Contains(t, err.Error(), "context done")
	})
}

func TestTaskSession_BasicFunctionality(t *testing.T) {
	t.Run("BN254 - verify task session state", func(t *testing.T) {
		operators, _, err := createBN254TestOperators(4)
		require.NoError(t, err)

		task := createTestTask()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		operatorPeersWeight := &operatorManager.PeerWeight{
			Operators:              operators,
			RootReferenceTimestamp: 123456789,
		}

		logger := zaptest.NewLogger(t)

		session, err := NewBN254TaskSession(
			ctx, cancel, task, "0xaggregator", []byte("signature"),
			operatorPeersWeight, logger,
		)
		require.NoError(t, err)

		// Verify initial state
		assert.Equal(t, uint32(0), session.resultsCount.Load())
		assert.False(t, session.thresholdMet.Load())
		assert.NotNil(t, session.taskAggregator)
		assert.Equal(t, task.TaskId, session.Task.TaskId)

		// Verify operator peer weight configuration
		assert.Equal(t, operatorPeersWeight, session.operatorPeersWeight)
		assert.Len(t, operatorPeersWeight.Operators, 4)
	})

	t.Run("ECDSA - verify task session state", func(t *testing.T) {
		operators, _, err := createECDSATestOperators(3)
		require.NoError(t, err)

		task := createTestTask()
		task.ThresholdBips = 6667 // ~67% (2 out of 3)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		operatorPeersWeight := &operatorManager.PeerWeight{
			Operators:              operators,
			RootReferenceTimestamp: 123456789,
		}

		logger := zaptest.NewLogger(t)

		session, err := NewECDSATaskSession(
			ctx, cancel, task, "0xaggregator", []byte("signature"),
			operatorPeersWeight, logger,
		)
		require.NoError(t, err)

		// Verify initial state
		assert.Equal(t, uint32(0), session.resultsCount.Load())
		assert.False(t, session.thresholdMet.Load())
		assert.NotNil(t, session.taskAggregator)
		assert.Equal(t, task.TaskId, session.Task.TaskId)

		// Verify ECDSA-specific configuration
		assert.Equal(t, operatorPeersWeight, session.operatorPeersWeight)
		assert.Len(t, operatorPeersWeight.Operators, 3)
	})
}

func TestTaskSession_DeadlockScenario(t *testing.T) {
	t.Run("deadlock when insufficient operators successfully submit tasks", func(t *testing.T) {
		operators, _, err := createBN254TestOperators(4)
		require.NoError(t, err)

		task := createTestTask()
		task.ThresholdBips = 7500 // 75% threshold (need 3 out of 4 operators)

		// Set a short timeout to simulate deadlock scenario where
		// operators don't respond in time
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		operatorPeersWeight := &operatorManager.PeerWeight{
			Operators:              operators,
			RootReferenceTimestamp: 123456789,
		}

		logger := zaptest.NewLogger(t)

		session, err := NewBN254TaskSession(
			ctx, cancel, task, "0xaggregator", []byte("signature"),
			operatorPeersWeight, logger,
		)
		require.NoError(t, err)

		// This test simulates a deadlock scenario where:
		// 1. The task session is configured with a 75% threshold (3 out of 4 operators needed)
		// 2. The context has a very short timeout (500ms)
		// 3. All gRPC calls will fail due to no running servers at the test addresses
		// 4. The session will timeout before reaching the required threshold
		//
		// This represents the deadlock condition described in the user request:
		// "Insufficient Operators Successfully Submit Tasks" - where the system
		// waits for enough operators to respond but never receives sufficient
		// responses to meet the signing threshold.

		cert, err := session.Process()

		// Verify that we have a deadlock scenario
		require.Error(t, err)
		require.Nil(t, cert)
		assert.Contains(t, err.Error(), "deadline exceeded")

		// Verify that no threshold was met (core deadlock condition)
		assert.False(t, session.thresholdMet.Load(), "Threshold should not be met in deadlock scenario")

		// Verify that the aggregator never reached the signing threshold
		// This is the core of the deadlock: we need responses but don't get enough
		assert.False(t, session.taskAggregator.SigningThresholdMet(),
			"Task aggregator should not have met signing threshold with insufficient responses")

		// Verify initial state remained unchanged due to no successful responses
		assert.Equal(t, uint32(0), session.resultsCount.Load(),
			"Should have zero successful responses in complete deadlock")
	})

	t.Run("deadlock scenario with high threshold requirement", func(t *testing.T) {
		operators, _, err := createBN254TestOperators(5)
		require.NoError(t, err)

		task := createTestTask()
		task.ThresholdBips = 9000 // 90% threshold (need 5 out of 5 operators - very strict)

		ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		defer cancel()

		operatorPeersWeight := &operatorManager.PeerWeight{
			Operators:              operators,
			RootReferenceTimestamp: 123456789,
		}

		logger := zaptest.NewLogger(t)

		session, err := NewBN254TaskSession(
			ctx, cancel, task, "0xaggregator", []byte("signature"),
			operatorPeersWeight, logger,
		)
		require.NoError(t, err)

		// This test simulates a more realistic deadlock scenario:
		// 1. High threshold (90%) requiring almost all operators to respond
		// 2. Short timeout simulating network issues or slow responses
		// 3. This represents scenarios where even if some operators respond,
		//    the threshold is so high that it's difficult to achieve consensus

		cert, err := session.Process()

		require.Error(t, err)
		require.Nil(t, cert)
		assert.Contains(t, err.Error(), "deadline exceeded")

		// Verify deadlock conditions
		assert.False(t, session.thresholdMet.Load(),
			"High threshold should not be met in deadlock scenario")
		assert.False(t, session.taskAggregator.SigningThresholdMet(),
			"90% threshold should be impossible to meet with failed connections")
	})

	t.Run("verify task session internal state during deadlock", func(t *testing.T) {
		operators, _, err := createBN254TestOperators(3)
		require.NoError(t, err)

		task := createTestTask()
		task.ThresholdBips = 6667 // 67% threshold (need 2 out of 3)

		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		operatorPeersWeight := &operatorManager.PeerWeight{
			Operators: operators,
		}

		logger := zaptest.NewLogger(t)

		session, err := NewBN254TaskSession(
			ctx, cancel, task, "0xaggregator", []byte("signature"),
			operatorPeersWeight, logger,
		)
		require.NoError(t, err)

		// Store initial state
		initialResultsCount := session.resultsCount.Load()
		initialThresholdMet := session.thresholdMet.Load()

		// Attempt process - should deadlock
		cert, err := session.Process()

		// Verify deadlock occurred
		require.Error(t, err)
		require.Nil(t, cert)

		// Verify internal state consistency during deadlock
		assert.Equal(t, initialResultsCount, session.resultsCount.Load(),
			"Results count should remain unchanged in complete deadlock")
		assert.Equal(t, initialThresholdMet, session.thresholdMet.Load(),
			"Threshold met flag should remain unchanged in deadlock")

		// Verify the task session maintains proper state even in failure scenarios
		assert.NotNil(t, session.Task, "Task should remain accessible")
		assert.NotNil(t, session.taskAggregator, "Task aggregator should remain accessible")
		assert.Equal(t, "0xaggregator", session.aggregatorAddress,
			"Aggregator address should remain unchanged")
	})
}

func TestTaskSession_ResponseSizeLimit(t *testing.T) {
	t.Run("verify maximum task response size constant", func(t *testing.T) {
		// Test that the constant is set to expected value
		expectedSize := 1.5 * 1024 * 1024 // 1.5MB
		assert.Equal(t, expectedSize, float64(maximumTaskResponseSize))
	})
}

func TestTaskSession_EdgeCases(t *testing.T) {
	t.Run("empty operator list", func(t *testing.T) {
		task := createTestTask()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Empty operator list
		operatorPeersWeight := &operatorManager.PeerWeight{
			Operators: []*peering.OperatorPeerInfo{},
		}

		logger := zaptest.NewLogger(t)

		session, err := NewBN254TaskSession(
			ctx, cancel, task, "0xaggregator", []byte("signature"),
			operatorPeersWeight, logger,
		)

		// Should return error due to no operators
		require.Error(t, err)
		require.Nil(t, session)
	})

	t.Run("zero threshold", func(t *testing.T) {
		operators, _, err := createBN254TestOperators(3)
		require.NoError(t, err)

		task := createTestTask()
		task.ThresholdBips = 0 // Invalid threshold
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		operatorPeersWeight := &operatorManager.PeerWeight{
			Operators: operators,
		}

		logger := zaptest.NewLogger(t)

		session, err := NewBN254TaskSession(
			ctx, cancel, task, "0xaggregator", []byte("signature"),
			operatorPeersWeight, logger,
		)

		// Should return error due to invalid threshold
		require.Error(t, err)
		require.Nil(t, session)
	})

	t.Run("maximum threshold", func(t *testing.T) {
		operators, _, err := createBN254TestOperators(2)
		require.NoError(t, err)

		task := createTestTask()
		task.ThresholdBips = 10000 // 100% threshold
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		operatorPeersWeight := &operatorManager.PeerWeight{
			Operators: operators,
		}

		logger := zaptest.NewLogger(t)

		session, err := NewBN254TaskSession(
			ctx, cancel, task, "0xaggregator", []byte("signature"),
			operatorPeersWeight, logger,
		)

		// Should successfully create session with 100% threshold
		require.NoError(t, err)
		require.NotNil(t, session)
		assert.Equal(t, task.ThresholdBips, session.Task.ThresholdBips)
	})
}

func TestTaskSession_TaskValidation(t *testing.T) {
	t.Run("valid task properties", func(t *testing.T) {
		operators, _, err := createBN254TestOperators(3)
		require.NoError(t, err)

		task := createTestTask()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		operatorPeersWeight := &operatorManager.PeerWeight{
			Operators: operators,
		}

		logger := zaptest.NewLogger(t)

		session, err := NewBN254TaskSession(
			ctx, cancel, task, "0xaggregator", []byte("signature"),
			operatorPeersWeight, logger,
		)
		require.NoError(t, err)

		// Verify task properties are correctly set
		assert.Equal(t, "0x1234567890abcdef", session.Task.TaskId)
		assert.Equal(t, "0xavsaddress", session.Task.AVSAddress)
		assert.Equal(t, uint32(1), session.Task.OperatorSetId)
		assert.Equal(t, uint16(7500), session.Task.ThresholdBips)
		assert.Equal(t, []byte("test-payload"), session.Task.Payload)
		assert.NotNil(t, session.Task.DeadlineUnixSeconds)
	})

	t.Run("task with different signature schemes", func(t *testing.T) {
		// Test BN254
		operators, _, err := createBN254TestOperators(2)
		require.NoError(t, err)

		task := createTestTask()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		operatorPeersWeight := &operatorManager.PeerWeight{
			Operators: operators,
		}

		logger := zaptest.NewLogger(t)

		bn254Session, err := NewBN254TaskSession(
			ctx, cancel, task, "0xaggregator", []byte("signature"),
			operatorPeersWeight, logger,
		)
		require.NoError(t, err)
		require.NotNil(t, bn254Session)

		// Test ECDSA
		ecdsaOperators, _, err := createECDSATestOperators(2)
		require.NoError(t, err)

		ecdsaOperatorPeersWeight := &operatorManager.PeerWeight{
			Operators: ecdsaOperators,
		}

		ecdsaSession, err := NewECDSATaskSession(
			ctx, cancel, task, "0xaggregator", []byte("signature"),
			ecdsaOperatorPeersWeight, logger,
		)
		require.NoError(t, err)
		require.NotNil(t, ecdsaSession)

		// Both should have the same task configuration
		assert.Equal(t, bn254Session.Task.TaskId, ecdsaSession.Task.TaskId)
		assert.Equal(t, bn254Session.Task.ThresholdBips, ecdsaSession.Task.ThresholdBips)
	})
}

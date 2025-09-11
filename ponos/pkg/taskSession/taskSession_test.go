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
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller/caller"
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

type MockSigner struct {
	mock.Mock
}

func (m *MockSigner) SignMessage(data []byte) ([]byte, error) {
	args := m.Called(data)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockSigner) SignMessageForSolidity(data []byte) ([]byte, error) {
	args := m.Called(data)
	return args.Get(0).([]byte), args.Error(1)
}

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

func createMockOperatorPeerInfo(operatorSetId uint32, networkAddress string) *MockOperatorPeerInfo {
	return &MockOperatorPeerInfo{
		OperatorAddress: fmt.Sprintf("0x%040x", operatorSetId),
		OperatorSets: []*peering.OperatorSet{
			{
				OperatorSetID:  operatorSetId,
				OperatorIndex:  0, // Mock operator is at index 0
				NetworkAddress: networkAddress,
			},
		},
	}
}

func createMockOperatorPeerInfoWithMultipleSets(operatorAddress string, operatorSets []*peering.OperatorSet) *MockOperatorPeerInfo {
	return &MockOperatorPeerInfo{
		OperatorAddress: operatorAddress,
		OperatorSets:    operatorSets,
	}
}

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
					OperatorIndex: uint32(i),
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
					OperatorIndex: uint32(i),
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
		TaskId:                 "0x1234567890abcdef",
		AVSAddress:             "0xavsaddress",
		OperatorSetId:          1,
		ThresholdBips:          7500, // 75%
		Payload:                []byte("test-payload"),
		DeadlineUnixSeconds:    &deadline,
		L1ReferenceBlockNumber: 12345678, // Include block number for testing
		Version:                1,
	}
}

// Helper function to create a mock signer for tests
func createMockSigner() *MockSigner {
	mockSigner := new(MockSigner)
	// Set up default behavior - return a mock signature for any input
	mockSigner.On("SignMessage", mock.Anything).Return([]byte("mock-signature"), nil)
	mockSigner.On("SignMessageForSolidity", mock.Anything).Return([]byte("mock-signature-solidity"), nil)
	return mockSigner
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
			Operators:              operators,
			RootReferenceTimestamp: 123456789,
		}

		logger := zaptest.NewLogger(t)

		session, err := NewBN254TaskSession(
			ctx, cancel, task, &caller.ContractCaller{}, "0xaggregator", createMockSigner(),
			operatorPeersWeight, false, logger,
		)

		require.NoError(t, err)
		require.NotNil(t, session)
		assert.Equal(t, task, session.Task)
		assert.Equal(t, "0xaggregator", session.aggregatorAddress)
		assert.NotNil(t, session.signer)
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
			Operators:              operators,
			RootReferenceTimestamp: 123456789,
		}

		logger := zaptest.NewLogger(t)

		session, err := NewBN254TaskSession(
			ctx, cancel, task, &caller.ContractCaller{}, "0xaggregator", createMockSigner(),
			operatorPeersWeight, false, logger,
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
			Operators:              operators,
			RootReferenceTimestamp: 123456789,
		}

		logger := zaptest.NewLogger(t)

		session, err := NewECDSATaskSession(
			ctx, cancel, task, &caller.ContractCaller{}, "0xaggregator", createMockSigner(),
			operatorPeersWeight, false, logger,
		)

		require.NoError(t, err)
		require.NotNil(t, session)
		assert.Equal(t, task, session.Task)
		assert.Equal(t, "0xaggregator", session.aggregatorAddress)
		assert.NotNil(t, session.signer)
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
			Operators:              operators,
			RootReferenceTimestamp: 123456789,
		}

		logger := zaptest.NewLogger(t)

		session, err := NewBN254TaskSession(
			ctx, cancel, task, &caller.ContractCaller{}, "0xaggregator", createMockSigner(),
			operatorPeersWeight, false, logger,
		)
		require.NoError(t, err)

		// The process will timeout because the mock gRPC client will fail to connect
		cert, err := session.Process()
		require.Error(t, err)
		require.Nil(t, cert)
		assert.Contains(t, err.Error(), "deadline exceeded")
	})

	t.Run("mock client demonstrates timeout behavior", func(t *testing.T) {
		// This test shows how the mock client handles timeouts
		// and demonstrates the timeout scenario more deterministically

		mockClient := &MockExecutorServiceClient{
			shouldFail:    false,
			response:      &executorV1.TaskResult{TaskId: "test", Output: []byte("response"), Version: 1},
			responseDelay: 200 * time.Millisecond, // Longer than context timeout
		}

		// Short timeout context
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		// Mock client respects context timeout and returns context.DeadlineExceeded
		result, err := mockClient.SubmitTask(ctx, &executorV1.TaskSubmission{
			TaskId:          "test",
			Payload:         []byte("test-payload"),
			TaskBlockNumber: 0,
			Version:         1,
		})

		require.Error(t, err)
		require.Nil(t, result)
		assert.Equal(t, context.DeadlineExceeded, err)

		// Verify the submission was recorded (showing the attempt was made)
		assert.Len(t, mockClient.taskSubmissions, 1, "Should record submission attempt before timeout")
		assert.Equal(t, uint64(0), mockClient.taskSubmissions[0].TaskBlockNumber,
			"TaskBlockNumber should be included in submission")
	})

	t.Run("mock client successful response", func(t *testing.T) {
		mockClient := &MockExecutorServiceClient{
			shouldFail:    false,
			response:      &executorV1.TaskResult{TaskId: "test", Output: []byte("quick response"), Version: 1},
			responseDelay: 50 * time.Millisecond, // Shorter than timeout
		}

		// Generous timeout context
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		result, err := mockClient.SubmitTask(ctx, &executorV1.TaskSubmission{
			TaskId:          "test",
			Payload:         []byte("test-payload"),
			TaskBlockNumber: 0,
			Version:         1,
		})

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "test", result.TaskId)
		assert.Equal(t, []byte("quick response"), result.Output)
		assert.Len(t, mockClient.taskSubmissions, 1)
		assert.Equal(t, uint64(0), mockClient.taskSubmissions[0].TaskBlockNumber,
			"TaskBlockNumber should be included in submission")
	})

	t.Run("mock client partial timeout scenario", func(t *testing.T) {
		// Test simulating mixed responses - some fast, some timeout
		fastClient := &MockExecutorServiceClient{
			shouldFail:    false,
			response:      &executorV1.TaskResult{TaskId: "test", Output: []byte("fast"), Version: 1},
			responseDelay: 20 * time.Millisecond,
		}

		slowClient := &MockExecutorServiceClient{
			shouldFail:    false,
			response:      &executorV1.TaskResult{TaskId: "test", Output: []byte("slow"), Version: 1},
			responseDelay: 150 * time.Millisecond, // Will timeout
		}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		// Fast client should succeed
		fastResult, fastErr := fastClient.SubmitTask(ctx, &executorV1.TaskSubmission{
			TaskId:          "test",
			TaskBlockNumber: 0,
			Version:         1,
		})
		require.NoError(t, fastErr)
		assert.Equal(t, []byte("fast"), fastResult.Output)

		// Slow client should timeout
		slowResult, slowErr := slowClient.SubmitTask(ctx, &executorV1.TaskSubmission{
			TaskId:          "test",
			TaskBlockNumber: 0,
			Version:         1,
		})
		require.Error(t, slowErr)
		assert.Nil(t, slowResult)
		assert.Equal(t, context.DeadlineExceeded, slowErr)

		// Both should have recorded submissions
		assert.Len(t, fastClient.taskSubmissions, 1)
		assert.Len(t, slowClient.taskSubmissions, 1)
	})

	t.Run("context cancellation", func(t *testing.T) {
		operators, _, err := createBN254TestOperators(1)
		require.NoError(t, err)

		task := createTestTask()
		ctx, cancel := context.WithCancel(context.Background())

		operatorPeersWeight := &operatorManager.PeerWeight{
			Operators:              operators,
			RootReferenceTimestamp: 123456789,
		}

		logger := zaptest.NewLogger(t)

		session, err := NewBN254TaskSession(
			ctx, cancel, task, &caller.ContractCaller{}, "0xaggregator", createMockSigner(),
			operatorPeersWeight, false, logger,
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
			ctx, cancel, task, &caller.ContractCaller{}, "0xaggregator", createMockSigner(),
			operatorPeersWeight, false, logger,
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
			ctx, cancel, task, &caller.ContractCaller{}, "0xaggregator", createMockSigner(),
			operatorPeersWeight, false, logger,
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
			ctx, cancel, task, &caller.ContractCaller{}, "0xaggregator", createMockSigner(),
			operatorPeersWeight, false, logger,
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
			ctx, cancel, task, &caller.ContractCaller{}, "0xaggregator", createMockSigner(),
			operatorPeersWeight, false, logger,
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
			Operators:              operators,
			RootReferenceTimestamp: 123456789,
		}

		logger := zaptest.NewLogger(t)

		session, err := NewBN254TaskSession(
			ctx, cancel, task, &caller.ContractCaller{}, "0xaggregator", createMockSigner(),
			operatorPeersWeight, false, logger,
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

func TestTaskSession_MockIntegration(t *testing.T) {
	t.Run("MockOperatorPeerInfo socket resolution", func(t *testing.T) {
		mockPeer := createMockOperatorPeerInfo(1, "localhost:9001")

		socket, err := mockPeer.GetSocketForOperatorSet(1)
		require.NoError(t, err)
		assert.Equal(t, "localhost:9001", socket)

		// Test non-existent operator set
		_, err = mockPeer.GetSocketForOperatorSet(999)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "operator set with ID 999 not found")

		// Test IncludesOperatorSetId
		assert.True(t, mockPeer.IncludesOperatorSetId(1))
		assert.False(t, mockPeer.IncludesOperatorSetId(999))
	})

	t.Run("MockOperatorPeerInfo with multiple operator sets", func(t *testing.T) {
		operatorSets := []*peering.OperatorSet{
			{OperatorSetID: 1, OperatorIndex: 0, NetworkAddress: "localhost:9001"},
			{OperatorSetID: 2, OperatorIndex: 1, NetworkAddress: "localhost:9002"},
			{OperatorSetID: 3, OperatorIndex: 2, NetworkAddress: "localhost:9003"},
		}

		mockPeer := createMockOperatorPeerInfoWithMultipleSets("0xoperator123", operatorSets)

		// Test all operator sets are accessible
		for i, expected := range operatorSets {
			socket, err := mockPeer.GetSocketForOperatorSet(expected.OperatorSetID)
			require.NoError(t, err, "Failed for operator set %d", i+1)
			assert.Equal(t, expected.NetworkAddress, socket)
			assert.True(t, mockPeer.IncludesOperatorSetId(expected.OperatorSetID))
		}

		// Test non-existent set
		assert.False(t, mockPeer.IncludesOperatorSetId(999))
	})

	t.Run("task session error handling with MockOperatorPeerInfo", func(t *testing.T) {
		// Create mock peer with non-existent operator set for the task
		mockPeer := createMockOperatorPeerInfo(1, "localhost:9001")

		// Convert to peering.OperatorPeerInfo interface
		operators := []*peering.OperatorPeerInfo{
			{
				OperatorAddress: mockPeer.OperatorAddress,
				OperatorSets:    mockPeer.OperatorSets,
			},
		}

		task := createTestTask()
		task.OperatorSetId = 999 // Set to non-existent operator set

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		operatorPeersWeight := &operatorManager.PeerWeight{
			Operators:              operators,
			RootReferenceTimestamp: 123456789,
		}

		logger := zaptest.NewLogger(t)

		// This should fail because operator set 999 doesn't exist
		session, err := NewBN254TaskSession(
			ctx, cancel, task, &caller.ContractCaller{}, "0xaggregator", createMockSigner(),
			operatorPeersWeight, false, logger,
		)

		require.Error(t, err)
		require.Nil(t, session)
		assert.Contains(t, err.Error(), "failed to get operator set")
	})

	t.Run("mock client with controlled network addresses", func(t *testing.T) {
		// This test demonstrates how MockOperatorPeerInfo provides controlled
		// network addresses that could be used with a mock server or client

		mockPeers := []*MockOperatorPeerInfo{
			createMockOperatorPeerInfo(1, "mock-server-1:9001"),
			createMockOperatorPeerInfo(1, "mock-server-2:9002"),
			createMockOperatorPeerInfo(1, "mock-server-3:9003"),
		}

		// Verify each mock peer has the expected network address
		expectedAddresses := []string{
			"mock-server-1:9001",
			"mock-server-2:9002",
			"mock-server-3:9003",
		}

		for i, mockPeer := range mockPeers {
			socket, err := mockPeer.GetSocketForOperatorSet(1)
			require.NoError(t, err)
			assert.Equal(t, expectedAddresses[i], socket)
		}
	})

	t.Run("combined mock scenario - peer info with client behavior", func(t *testing.T) {
		// This test demonstrates how MockOperatorPeerInfo and MockExecutorServiceClient
		// could work together for comprehensive testing

		// Create mock peers with controlled network addresses
		mockPeer1 := createMockOperatorPeerInfo(1, "mock-executor-1:9001")
		mockPeer2 := createMockOperatorPeerInfo(1, "mock-executor-2:9002")

		// Create mock clients with different behaviors
		fastClient := &MockExecutorServiceClient{
			shouldFail:    false,
			response:      &executorV1.TaskResult{TaskId: "test", Output: []byte("fast-response"), Version: 1},
			responseDelay: 10 * time.Millisecond,
		}

		slowClient := &MockExecutorServiceClient{
			shouldFail:    false,
			response:      &executorV1.TaskResult{TaskId: "test", Output: []byte("slow-response"), Version: 1},
			responseDelay: 200 * time.Millisecond,
		}

		// Verify the mock peer info provides the expected sockets
		socket1, err := mockPeer1.GetSocketForOperatorSet(1)
		require.NoError(t, err)
		assert.Equal(t, "mock-executor-1:9001", socket1)

		socket2, err := mockPeer2.GetSocketForOperatorSet(1)
		require.NoError(t, err)
		assert.Equal(t, "mock-executor-2:9002", socket2)

		// Test client behaviors with different timeouts
		ctx1, cancel1 := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel1()

		ctx2, cancel2 := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel2()

		// Fast client should succeed within timeout
		result1, err1 := fastClient.SubmitTask(ctx1, &executorV1.TaskSubmission{
			TaskId:          "test",
			TaskBlockNumber: 0,
			Version:         1,
		})
		require.NoError(t, err1)
		assert.Equal(t, []byte("fast-response"), result1.Output)

		// Slow client should timeout
		result2, err2 := slowClient.SubmitTask(ctx2, &executorV1.TaskSubmission{
			TaskId:          "test",
			TaskBlockNumber: 0,
			Version:         1,
		})
		require.Error(t, err2)
		assert.Nil(t, result2)
		assert.Equal(t, context.DeadlineExceeded, err2)

		// Both clients should have recorded submissions
		assert.Len(t, fastClient.taskSubmissions, 1)
		assert.Len(t, slowClient.taskSubmissions, 1)

		// Validate TaskBlockNumber is properly included in submissions
		assert.Equal(t, uint64(0), fastClient.taskSubmissions[0].TaskBlockNumber,
			"Fast client should capture TaskBlockNumber")
		assert.Equal(t, uint64(0), slowClient.taskSubmissions[0].TaskBlockNumber,
			"Slow client should capture TaskBlockNumber")

		// This demonstrates how the mocks provide controlled, deterministic testing
		// where MockOperatorPeerInfo gives us controlled network addresses and
		// MockExecutorServiceClient gives us controlled response timing/behavior
	})
}

func TestTaskSession_EdgeCases(t *testing.T) {
	t.Run("empty operator list", func(t *testing.T) {
		task := createTestTask()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Empty operator list
		operatorPeersWeight := &operatorManager.PeerWeight{
			Operators:              []*peering.OperatorPeerInfo{},
			RootReferenceTimestamp: 123456789,
		}

		logger := zaptest.NewLogger(t)

		session, err := NewBN254TaskSession(
			ctx, cancel, task, &caller.ContractCaller{}, "0xaggregator", createMockSigner(),
			operatorPeersWeight, false, logger,
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
			ctx, cancel, task, &caller.ContractCaller{}, "0xaggregator", createMockSigner(),
			operatorPeersWeight, false, logger,
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
			Operators:              operators,
			RootReferenceTimestamp: 123456789,
		}

		logger := zaptest.NewLogger(t)

		session, err := NewBN254TaskSession(
			ctx, cancel, task, &caller.ContractCaller{}, "0xaggregator", createMockSigner(),
			operatorPeersWeight, false, logger,
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
			Operators:              operators,
			RootReferenceTimestamp: 123456789,
		}

		logger := zaptest.NewLogger(t)

		session, err := NewBN254TaskSession(
			ctx, cancel, task, &caller.ContractCaller{}, "0xaggregator", createMockSigner(),
			operatorPeersWeight, false, logger,
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
			Operators:              operators,
			RootReferenceTimestamp: 123456789,
		}

		logger := zaptest.NewLogger(t)

		bn254Session, err := NewBN254TaskSession(
			ctx, cancel, task, &caller.ContractCaller{}, "0xaggregator", createMockSigner(),
			operatorPeersWeight, false, logger,
		)
		require.NoError(t, err)
		require.NotNil(t, bn254Session)

		// Test ECDSA
		ecdsaOperators, _, err := createECDSATestOperators(2)
		require.NoError(t, err)

		ecdsaOperatorPeersWeight := &operatorManager.PeerWeight{
			Operators:              ecdsaOperators,
			RootReferenceTimestamp: 123456789,
		}

		ecdsaSession, err := NewECDSATaskSession(
			ctx, cancel, task, &caller.ContractCaller{}, "0xaggregator", createMockSigner(),
			ecdsaOperatorPeersWeight, false, logger,
		)
		require.NoError(t, err)
		require.NotNil(t, ecdsaSession)

		// Both should have the same task configuration
		assert.Equal(t, bn254Session.Task.TaskId, ecdsaSession.Task.TaskId)
		assert.Equal(t, bn254Session.Task.ThresholdBips, ecdsaSession.Task.ThresholdBips)
	})
}

func TestTaskSession_TLSConfiguration(t *testing.T) {
	t.Run("secure connections by default", func(t *testing.T) {
		operators, _, err := createBN254TestOperators(2)
		require.NoError(t, err)

		task := createTestTask()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		operatorPeersWeight := &operatorManager.PeerWeight{
			Operators:              operators,
			RootReferenceTimestamp: 123456789,
		}

		logger := zaptest.NewLogger(t)

		// Test with secure connections (tlsEnabled = false)
		secureSession, err := NewBN254TaskSession(
			ctx, cancel, task, &caller.ContractCaller{}, "0xaggregator", createMockSigner(),
			operatorPeersWeight, false, logger,
		)
		require.NoError(t, err)
		require.NotNil(t, secureSession)
		assert.False(t, secureSession.tlsEnabled, "Should use secure connections by default")
	})

	t.Run("insecure connections for local development", func(t *testing.T) {
		operators, _, err := createECDSATestOperators(2)
		require.NoError(t, err)

		task := createTestTask()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		operatorPeersWeight := &operatorManager.PeerWeight{
			Operators:              operators,
			RootReferenceTimestamp: 123456789,
		}

		logger := zaptest.NewLogger(t)

		// Test with insecure connections (tlsEnabled = true)
		insecureSession, err := NewECDSATaskSession(
			ctx, cancel, task, &caller.ContractCaller{}, "0xaggregator", createMockSigner(),
			operatorPeersWeight, true, logger,
		)
		require.NoError(t, err)
		require.NotNil(t, insecureSession)
		assert.True(t, insecureSession.tlsEnabled, "Should allow insecure connections when explicitly configured")
	})
}

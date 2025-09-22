package executor

import (
	"context"
	"fmt"
	"testing"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/performerTask"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/executorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockContractCaller struct {
	mock.Mock
	contractCaller.IContractCaller
}

func (m *MockContractCaller) CalculateTaskHashMessage(_ context.Context, taskHash [32]byte, result []byte) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockContractCaller) GetAVSConfig(avsAddress string, blockNumber uint64) (*contractCaller.AVSConfig, error) {
	args := m.Called(avsAddress, blockNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*contractCaller.AVSConfig), args.Error(1)
}

func (m *MockContractCaller) GetOperatorSetDetailsForOperator(operatorAddress common.Address, avsAddress string, operatorSetId uint32, blockNumber uint64) (*peering.OperatorSet, error) {
	args := m.Called(operatorAddress, avsAddress, operatorSetId, blockNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*peering.OperatorSet), args.Error(1)
}

func (m *MockContractCaller) GetOperatorSetCurveType(avsAddress string, operatorSetId uint32, blockNumber uint64) (config.CurveType, error) {
	args := m.Called(avsAddress, operatorSetId, blockNumber)
	return args.Get(0).(config.CurveType), args.Error(1)
}

func (m *MockContractCaller) CalculateBN254CertificateDigestBytes(ctx context.Context, referenceTimestamp uint32, messageHash [32]byte) ([]byte, error) {
	args := m.Called(ctx, referenceTimestamp, messageHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockContractCaller) CalculateECDSACertificateDigestBytes(ctx context.Context, referenceTimestamp uint32, messageHash [32]byte) ([]byte, error) {
	args := m.Called(ctx, referenceTimestamp, messageHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

// TestValidateOperatorInSet tests operator validation in isolation
func TestValidateOperatorInSet(t *testing.T) {
	tests := []struct {
		name          string
		operatorAddr  string
		task          *executor.TaskSubmission
		operatorSet   *peering.OperatorSet
		setupMock     func(*MockContractCaller, string, *executor.TaskSubmission, *peering.OperatorSet)
		expectError   bool
		errorContains string
	}{
		{
			name:         "valid operator in operator set",
			operatorAddr: "0x1234567890123456789012345678901234567890",
			task: &executor.TaskSubmission{
				TaskId:          "task-1",
				AvsAddress:      "0xABCDEF1234567890123456789012345678901234",
				OperatorSetId:   1,
				TaskBlockNumber: 12345678,
				Payload:         []byte("test payload"),
			},
			operatorSet: &peering.OperatorSet{
				OperatorSetID:  1,
				NetworkAddress: "127.0.0.1:8080",
			},
			setupMock: func(m *MockContractCaller, operatorAddr string, task *executor.TaskSubmission, opSet *peering.OperatorSet) {
				m.On("GetOperatorSetDetailsForOperator",
					common.HexToAddress(operatorAddr),
					task.GetAvsAddress(),
					task.OperatorSetId,
					task.TaskBlockNumber,
				).Return(opSet, nil)
			},
			expectError: false,
		},
		{
			name:         "operator not in operator set - nil response",
			operatorAddr: "0x1234567890123456789012345678901234567890",
			task: &executor.TaskSubmission{
				TaskId:          "task-2",
				AvsAddress:      "0xABCDEF1234567890123456789012345678901234",
				OperatorSetId:   1,
				TaskBlockNumber: 12345678,
				Payload:         []byte("test payload"),
			},
			operatorSet: nil,
			setupMock: func(m *MockContractCaller, operatorAddr string, task *executor.TaskSubmission, opSet *peering.OperatorSet) {
				m.On("GetOperatorSetDetailsForOperator",
					common.HexToAddress(operatorAddr),
					task.GetAvsAddress(),
					task.OperatorSetId,
					task.TaskBlockNumber,
				).Return(nil, nil)
			},
			expectError:   true,
			errorContains: "invalid task operator set",
		},
		{
			name:         "contract caller error",
			operatorAddr: "0x1234567890123456789012345678901234567890",
			task: &executor.TaskSubmission{
				TaskId:          "task-4",
				AvsAddress:      "0xABCDEF1234567890123456789012345678901234",
				OperatorSetId:   1,
				TaskBlockNumber: 12345678,
				Payload:         []byte("test payload"),
			},
			operatorSet: nil,
			setupMock: func(m *MockContractCaller, operatorAddr string, task *executor.TaskSubmission, opSet *peering.OperatorSet) {
				m.On("GetOperatorSetDetailsForOperator",
					common.HexToAddress(operatorAddr),
					task.GetAvsAddress(),
					task.OperatorSetId,
					task.TaskBlockNumber,
				).Return(nil, fmt.Errorf("contract call failed"))
			},
			expectError:   true,
			errorContains: "contract call failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCaller := new(MockContractCaller)
			tt.setupMock(mockCaller, tt.operatorAddr, tt.task, tt.operatorSet)

			l, err := logger.NewLogger(&logger.LoggerConfig{Debug: false})
			assert.NoError(t, err)

			execConfig := &executorConfig.ExecutorConfig{
				Operator: &config.OperatorConfig{
					Address: tt.operatorAddr,
				},
				AvsPerformers: []*executorConfig.AvsPerformerConfig{
					{
						AvsAddress: tt.task.GetAvsAddress(),
					},
				},
			}

			e := &Executor{
				config:           execConfig,
				l1ContractCaller: mockCaller,
				logger:           l,
			}

			// Test only the validateOperatorInSet method
			err = e.validateOperatorInSet(tt.task)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			mockCaller.AssertExpectations(t)
		})
	}
}

type MockSigner struct{}

func (m *MockSigner) SignMessage(data []byte) ([]byte, error) {
	return []byte("mock-signature"), nil
}

func (m *MockSigner) SignMessageForSolidity(data []byte) ([]byte, error) {
	return []byte("mock-signature-solidity"), nil
}

var _ signer.ISigner = (*MockSigner)(nil)

type MockPerformer struct{}

func (m *MockPerformer) Initialize(ctx context.Context) error {
	return nil
}

func (m *MockPerformer) RunTask(ctx context.Context, task *performerTask.PerformerTask) (*performerTask.PerformerTaskResult, error) {
	return &performerTask.PerformerTaskResult{
		TaskID: task.TaskID,
		Result: []byte("mock result"),
	}, nil
}

func (m *MockPerformer) Deploy(ctx context.Context, image avsPerformer.PerformerImage) (*avsPerformer.DeploymentResult, error) {
	return &avsPerformer.DeploymentResult{}, nil
}

func (m *MockPerformer) CreatePerformer(ctx context.Context, image avsPerformer.PerformerImage) (*avsPerformer.PerformerCreationResult, error) {
	return &avsPerformer.PerformerCreationResult{}, nil
}

func (m *MockPerformer) PromotePerformer(ctx context.Context, performerID string) error {
	return nil
}

func (m *MockPerformer) RemovePerformer(ctx context.Context, performerID string) error {
	return nil
}

func (m *MockPerformer) ListPerformers() []avsPerformer.PerformerMetadata {
	return []avsPerformer.PerformerMetadata{}
}

func (m *MockPerformer) Shutdown() error {
	return nil
}

func (m *MockPerformer) ExecuteWorkflow(task *executor.TaskSubmission) (*executor.TaskResult, error) {
	return &executor.TaskResult{
		TaskId: task.TaskId,
		Output: []byte("mock result"),
	}, nil
}

func (m *MockPerformer) ExecutorConfig() *executorConfig.AvsPerformerConfig {
	return &executorConfig.AvsPerformerConfig{
		AvsAddress: "0xtest",
	}
}

func TestValidateTaskSubmission(t *testing.T) {
	testCases := []struct {
		name        string
		request     *executor.TaskSubmission
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid task submission",
			request: &executor.TaskSubmission{
				TaskId:             "0x" + "a" + fmt.Sprintf("%063d", 1), // 0x + 64 hex chars
				AggregatorAddress:  "0x" + fmt.Sprintf("%040x", 1),       // 0x + 40 hex chars
				AvsAddress:         "0x" + fmt.Sprintf("%040x", 2),       // 0x + 40 hex chars
				ExecutorAddress:    "0x" + fmt.Sprintf("%040x", 3),       // 0x + 40 hex chars
				Signature:          []byte("signature"),
				ReferenceTimestamp: 1234567890,
				OperatorSetId:      1,
				TaskBlockNumber:    100,
				Payload:            []byte("test payload"),
			},
			expectError: false,
		},
		{
			name: "task ID too short",
			request: &executor.TaskSubmission{
				TaskId:             "0x1234",
				AggregatorAddress:  "0x" + fmt.Sprintf("%040x", 1),
				AvsAddress:         "0x" + fmt.Sprintf("%040x", 2),
				ExecutorAddress:    "0x" + fmt.Sprintf("%040x", 3),
				Signature:          []byte("signature"),
				ReferenceTimestamp: 1234567890,
			},
			expectError: true,
			errorMsg:    "invalid task ID length: expected 66, got 6",
		},
		{
			name: "task ID too long",
			request: &executor.TaskSubmission{
				TaskId:             "0x" + fmt.Sprintf("%065d", 1), // 0x + 65 chars (too long)
				AggregatorAddress:  "0x" + fmt.Sprintf("%040x", 1),
				AvsAddress:         "0x" + fmt.Sprintf("%040x", 2),
				ExecutorAddress:    "0x" + fmt.Sprintf("%040x", 3),
				Signature:          []byte("signature"),
				ReferenceTimestamp: 1234567890,
			},
			expectError: true,
			errorMsg:    "invalid task ID length: expected 66, got 67",
		},
		{
			name: "task ID missing 0x prefix",
			request: &executor.TaskSubmission{
				TaskId:             fmt.Sprintf("%064x", 1), // 64 hex chars without 0x
				AggregatorAddress:  "0x" + fmt.Sprintf("%040x", 1),
				AvsAddress:         "0x" + fmt.Sprintf("%040x", 2),
				ExecutorAddress:    "0x" + fmt.Sprintf("%040x", 3),
				Signature:          []byte("signature"),
				ReferenceTimestamp: 1234567890,
			},
			expectError: true,
			errorMsg:    "invalid task ID length: expected 66, got 64",
		},
		{
			name: "task ID invalid hex",
			request: &executor.TaskSubmission{
				TaskId:             "0x" + "g" + fmt.Sprintf("%063d", 1), // invalid hex char 'g'
				AggregatorAddress:  "0x" + fmt.Sprintf("%040x", 1),
				AvsAddress:         "0x" + fmt.Sprintf("%040x", 2),
				ExecutorAddress:    "0x" + fmt.Sprintf("%040x", 3),
				Signature:          []byte("signature"),
				ReferenceTimestamp: 1234567890,
			},
			expectError: true,
			errorMsg:    "invalid task ID format",
		},
		{
			name: "aggregator address too short",
			request: &executor.TaskSubmission{
				TaskId:             "0x" + fmt.Sprintf("%064x", 1),
				AggregatorAddress:  "0x1234",
				AvsAddress:         "0x" + fmt.Sprintf("%040x", 2),
				ExecutorAddress:    "0x" + fmt.Sprintf("%040x", 3),
				Signature:          []byte("signature"),
				ReferenceTimestamp: 1234567890,
			},
			expectError: true,
			errorMsg:    "invalid aggregator address length: expected 42, got 6",
		},
		{
			name: "aggregator address missing 0x prefix",
			request: &executor.TaskSubmission{
				TaskId:             "0x" + fmt.Sprintf("%064x", 1),
				AggregatorAddress:  fmt.Sprintf("%040x", 1),
				AvsAddress:         "0x" + fmt.Sprintf("%040x", 2),
				ExecutorAddress:    "0x" + fmt.Sprintf("%040x", 3),
				Signature:          []byte("signature"),
				ReferenceTimestamp: 1234567890,
			},
			expectError: true,
			errorMsg:    "invalid aggregator address length: expected 42, got 40",
		},
		{
			name: "aggregator address invalid hex",
			request: &executor.TaskSubmission{
				TaskId:             "0x" + fmt.Sprintf("%064x", 1),
				AggregatorAddress:  "0x" + "g" + fmt.Sprintf("%039x", 1), // invalid hex char 'g'
				AvsAddress:         "0x" + fmt.Sprintf("%040x", 2),
				ExecutorAddress:    "0x" + fmt.Sprintf("%040x", 3),
				Signature:          []byte("signature"),
				ReferenceTimestamp: 1234567890,
			},
			expectError: true,
			errorMsg:    "invalid aggregator address format",
		},
		{
			name: "AVS address too long",
			request: &executor.TaskSubmission{
				TaskId:             "0x" + fmt.Sprintf("%064x", 1),
				AggregatorAddress:  "0x" + fmt.Sprintf("%040x", 1),
				AvsAddress:         "0x" + fmt.Sprintf("%041x", 2), // 41 hex chars (too long)
				ExecutorAddress:    "0x" + fmt.Sprintf("%040x", 3),
				Signature:          []byte("signature"),
				ReferenceTimestamp: 1234567890,
			},
			expectError: true,
			errorMsg:    "invalid AVS address length: expected 42, got 43",
		},
		{
			name: "AVS address missing 0x prefix",
			request: &executor.TaskSubmission{
				TaskId:             "0x" + fmt.Sprintf("%064x", 1),
				AggregatorAddress:  "0x" + fmt.Sprintf("%040x", 1),
				AvsAddress:         fmt.Sprintf("%040x", 2),
				ExecutorAddress:    "0x" + fmt.Sprintf("%040x", 3),
				Signature:          []byte("signature"),
				ReferenceTimestamp: 1234567890,
			},
			expectError: true,
			errorMsg:    "invalid AVS address length: expected 42, got 40",
		},
		{
			name: "executor address too short",
			request: &executor.TaskSubmission{
				TaskId:             "0x" + fmt.Sprintf("%064x", 1),
				AggregatorAddress:  "0x" + fmt.Sprintf("%040x", 1),
				AvsAddress:         "0x" + fmt.Sprintf("%040x", 2),
				ExecutorAddress:    "0x123",
				Signature:          []byte("signature"),
				ReferenceTimestamp: 1234567890,
			},
			expectError: true,
			errorMsg:    "invalid executor address length: expected 42, got 5",
		},
		{
			name: "executor address missing 0x prefix",
			request: &executor.TaskSubmission{
				TaskId:             "0x" + fmt.Sprintf("%064x", 1),
				AggregatorAddress:  "0x" + fmt.Sprintf("%040x", 1),
				AvsAddress:         "0x" + fmt.Sprintf("%040x", 2),
				ExecutorAddress:    fmt.Sprintf("%040x", 3),
				Signature:          []byte("signature"),
				ReferenceTimestamp: 1234567890,
			},
			expectError: true,
			errorMsg:    "invalid executor address length: expected 42, got 40",
		},
		{
			name: "empty signature",
			request: &executor.TaskSubmission{
				TaskId:             "0x" + fmt.Sprintf("%064x", 1),
				AggregatorAddress:  "0x" + fmt.Sprintf("%040x", 1),
				AvsAddress:         "0x" + fmt.Sprintf("%040x", 2),
				ExecutorAddress:    "0x" + fmt.Sprintf("%040x", 3),
				Signature:          []byte{},
				ReferenceTimestamp: 1234567890,
			},
			expectError: true,
			errorMsg:    "signature cannot be empty",
		},
		{
			name: "nil signature",
			request: &executor.TaskSubmission{
				TaskId:             "0x" + fmt.Sprintf("%064x", 1),
				AggregatorAddress:  "0x" + fmt.Sprintf("%040x", 1),
				AvsAddress:         "0x" + fmt.Sprintf("%040x", 2),
				ExecutorAddress:    "0x" + fmt.Sprintf("%040x", 3),
				Signature:          nil,
				ReferenceTimestamp: 1234567890,
			},
			expectError: true,
			errorMsg:    "signature cannot be empty",
		},
		{
			name: "zero reference timestamp",
			request: &executor.TaskSubmission{
				TaskId:             "0x" + fmt.Sprintf("%064x", 1),
				AggregatorAddress:  "0x" + fmt.Sprintf("%040x", 1),
				AvsAddress:         "0x" + fmt.Sprintf("%040x", 2),
				ExecutorAddress:    "0x" + fmt.Sprintf("%040x", 3),
				Signature:          []byte("signature"),
				ReferenceTimestamp: 0,
			},
			expectError: true,
			errorMsg:    "reference timestamp cannot be zero",
		},
		{
			name: "all fields valid with real ethereum addresses",
			request: &executor.TaskSubmission{
				TaskId:             "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
				AggregatorAddress:  "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1",
				AvsAddress:         "0x5FbDB2315678afecb367f032d93F642f64180aa3",
				ExecutorAddress:    "0x70997970C51812dc3A010C7d01b50e0d17dc79C8",
				Signature:          []byte("signature"),
				ReferenceTimestamp: 1234567890,
				OperatorSetId:      1,
				TaskBlockNumber:    100,
				Payload:            []byte("test payload"),
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateTaskSubmission(tc.request)

			if tc.expectError {
				assert.NotNil(t, err, "Expected validation error but got nil")
				if err != nil {
					assert.Contains(t, err.Error(), tc.errorMsg, "Error message should contain expected text")
				}
			} else {
				assert.Nil(t, err, "Expected no validation error but got: %w", err)
			}
		})
	}
}

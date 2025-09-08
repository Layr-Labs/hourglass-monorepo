package executor

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/performerTask"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer"
	"testing"

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

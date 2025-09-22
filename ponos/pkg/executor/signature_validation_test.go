package executor

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"sync"
	"testing"

	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	ecdsacrypto "github.com/Layr-Labs/crypto-libs/pkg/ecdsa"
	executorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/executorConfig"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/storage/memory"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/performerTask"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestValidateTaskSignature_ValidECDSA tests validation of a valid ECDSA signature
func TestValidateTaskSignature_ValidECDSA(t *testing.T) {
	// Setup
	aggregatorPrivKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	aggregatorAddress := crypto.PubkeyToAddress(aggregatorPrivKey.PublicKey).Hex()
	executorAddress := "0x1234567890123456789012345678901234567890"
	avsAddress := "0xabcdef1234567890abcdef1234567890abcdef12"

	// Create mock contract caller
	mockCaller := NewEnhancedMockContractCaller()
	mockCaller.SetupValidAggregator(avsAddress, aggregatorAddress, config.CurveTypeECDSA)

	// Create signed task
	task := CreateSignedTaskSubmission(t, aggregatorPrivKey, executorAddress, avsAddress)

	// Create executor
	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	require.NoError(t, err)

	e := &Executor{
		config: &executorConfig.ExecutorConfig{
			Operator: &config.OperatorConfig{
				Address: executorAddress,
			},
		},
		l1ContractCaller: mockCaller,
		logger:           l,
	}

	// Test
	err = e.validateTaskSignature(task)

	// Assert
	assert.NoError(t, err, "Valid ECDSA signature should pass validation")
	mockCaller.AssertExpectations(t)
}

// TestValidateTaskSignature_ValidBN254 tests validation of a valid BN254 signature
func TestValidateTaskSignature_ValidBN254(t *testing.T) {
	// Setup
	aggregatorPrivKey, aggregatorPubKey, err := bn254.GenerateKeyPair()
	require.NoError(t, err)
	aggregatorAddress := "0xaggregator00000000000000000000000000000"
	executorAddress := "0x1234567890123456789012345678901234567890"
	avsAddress := "0xabcdef1234567890abcdef1234567890abcdef12"

	// Create mock contract caller with manual setup for BN254
	mockCaller := NewEnhancedMockContractCaller()
	mockCaller.SetupValidAggregator(avsAddress, aggregatorAddress, config.CurveTypeBN254)

	// Manually set up the BN254 public key
	key := fmt.Sprintf("%s-%s-%d", aggregatorAddress, avsAddress, 0)
	mockCaller.operatorSets[key].WrappedPublicKey.PublicKey = aggregatorPubKey

	// Create signed task
	task := CreateSignedTaskSubmission(t, aggregatorPrivKey, executorAddress, avsAddress)
	task.AggregatorAddress = aggregatorAddress // Override since BN254 doesn't derive address

	// Create executor
	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	require.NoError(t, err)

	e := &Executor{
		config: &executorConfig.ExecutorConfig{
			Operator: &config.OperatorConfig{
				Address: executorAddress,
			},
		},
		l1ContractCaller: mockCaller,
		logger:           l,
	}

	// Test
	err = e.validateTaskSignature(task)

	// Assert
	assert.NoError(t, err, "Valid BN254 signature should pass validation")
	mockCaller.AssertExpectations(t)
}

// TestValidateTaskSignature_InvalidSignature tests rejection of invalid signatures
func TestValidateTaskSignature_InvalidSignature(t *testing.T) {
	// Setup
	aggregatorPrivKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	aggregatorAddress := crypto.PubkeyToAddress(aggregatorPrivKey.PublicKey).Hex()
	executorAddress := "0x1234567890123456789012345678901234567890"
	avsAddress := "0xabcdef1234567890abcdef1234567890abcdef12"

	// Create mock contract caller
	mockCaller := NewEnhancedMockContractCaller()
	mockCaller.SetupValidAggregator(avsAddress, aggregatorAddress, config.CurveTypeECDSA)

	// Create task with invalid signature
	task := CreateValidTaskSubmission(t, aggregatorPrivKey, executorAddress, avsAddress)
	task.Signature = []byte("invalid signature data that is long enough but wrong")

	// Create executor
	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	require.NoError(t, err)

	e := &Executor{
		config: &executorConfig.ExecutorConfig{
			Operator: &config.OperatorConfig{
				Address: executorAddress,
			},
		},
		l1ContractCaller: mockCaller,
		logger:           l,
	}

	// Test
	err = e.validateTaskSignature(task)

	// Assert
	assert.Error(t, err, "Invalid signature should fail validation")
	assert.Contains(t, err.Error(), "signature", "Error should mention signature")
}

// TestValidateTaskSignature_WrongExecutor tests rejection when signature is for different executor
func TestValidateTaskSignature_WrongExecutor(t *testing.T) {
	// Setup
	aggregatorPrivKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	aggregatorAddress := crypto.PubkeyToAddress(aggregatorPrivKey.PublicKey).Hex()
	executorAddress := "0x1234567890123456789012345678901234567890"
	wrongExecutorAddress := "0x9999999999999999999999999999999999999999"
	avsAddress := "0xabcdef1234567890abcdef1234567890abcdef12"

	// Create mock contract caller
	mockCaller := NewEnhancedMockContractCaller()
	mockCaller.SetupValidAggregator(avsAddress, aggregatorAddress, config.CurveTypeECDSA)

	// Create task signed for wrong executor
	task := CreateSignedTaskSubmission(t, aggregatorPrivKey, wrongExecutorAddress, avsAddress)

	// Create executor with different address
	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	require.NoError(t, err)

	e := &Executor{
		config: &executorConfig.ExecutorConfig{
			Operator: &config.OperatorConfig{
				Address: executorAddress, // Different from what was signed
			},
		},
		l1ContractCaller: mockCaller,
		logger:           l,
	}

	// Test
	err = e.validateTaskSignature(task)

	// Assert
	assert.Error(t, err, "Signature for wrong executor should fail validation")
	assert.Contains(t, err.Error(), "signature verification failed", "Error should mention signature verification")
}

// TestValidateTaskSignature_TamperedPayload tests rejection when payload is tampered
func TestValidateTaskSignature_TamperedPayload(t *testing.T) {
	// Setup
	aggregatorPrivKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	aggregatorAddress := crypto.PubkeyToAddress(aggregatorPrivKey.PublicKey).Hex()
	executorAddress := "0x1234567890123456789012345678901234567890"
	avsAddress := "0xabcdef1234567890abcdef1234567890abcdef12"

	// Create mock contract caller
	mockCaller := NewEnhancedMockContractCaller()
	mockCaller.SetupValidAggregator(avsAddress, aggregatorAddress, config.CurveTypeECDSA)

	// Create signed task then tamper with payload
	task := CreateSignedTaskSubmission(t, aggregatorPrivKey, executorAddress, avsAddress)
	task.Payload = []byte("tampered payload data")

	// Create executor
	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	require.NoError(t, err)

	e := &Executor{
		config: &executorConfig.ExecutorConfig{
			Operator: &config.OperatorConfig{
				Address: executorAddress,
			},
		},
		l1ContractCaller: mockCaller,
		logger:           l,
	}

	// Test
	err = e.validateTaskSignature(task)

	// Assert
	assert.Error(t, err, "Tampered payload should fail validation")
	assert.Contains(t, err.Error(), "signature verification failed", "Error should mention signature verification")
}

// TestValidateTaskSignature_WrongAggregator tests rejection when signed by non-aggregator
func TestValidateTaskSignature_WrongAggregator(t *testing.T) {
	// Setup
	aggregatorPrivKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	aggregatorAddress := crypto.PubkeyToAddress(aggregatorPrivKey.PublicKey).Hex()

	wrongPrivKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	executorAddress := "0x1234567890123456789012345678901234567890"
	avsAddress := "0xabcdef1234567890abcdef1234567890abcdef12"

	// Create mock contract caller - setup with correct aggregator
	mockCaller := NewEnhancedMockContractCaller()
	mockCaller.SetupValidAggregator(avsAddress, aggregatorAddress, config.CurveTypeECDSA)

	// Create task signed by wrong key
	task := CreateSignedTaskSubmission(t, wrongPrivKey, executorAddress, avsAddress)
	task.AggregatorAddress = aggregatorAddress // Claim to be the aggregator

	// Create executor
	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	require.NoError(t, err)

	e := &Executor{
		config: &executorConfig.ExecutorConfig{
			Operator: &config.OperatorConfig{
				Address: executorAddress,
			},
		},
		l1ContractCaller: mockCaller,
		logger:           l,
	}

	// Test
	err = e.validateTaskSignature(task)

	// Assert
	assert.Error(t, err, "Signature from wrong aggregator should fail validation")
	assert.Contains(t, err.Error(), "signature verification failed", "Error should mention signature verification")
}

// TestValidateTaskSignature_MissingAVSConfig tests handling of missing AVS config
func TestValidateTaskSignature_MissingAVSConfig(t *testing.T) {
	// Setup
	aggregatorPrivKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	executorAddress := "0x1234567890123456789012345678901234567890"
	avsAddress := "0xabcdef1234567890abcdef1234567890abcdef12"

	// Create mock contract caller that returns error for AVS config
	mockCaller := NewEnhancedMockContractCaller()
	mockCaller.On("GetAVSConfig", avsAddress, mock.AnythingOfType("uint64")).Return(nil, assert.AnError)

	// Create signed task
	task := CreateSignedTaskSubmission(t, aggregatorPrivKey, executorAddress, avsAddress)

	// Create executor
	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	require.NoError(t, err)

	e := &Executor{
		config: &executorConfig.ExecutorConfig{
			Operator: &config.OperatorConfig{
				Address: executorAddress,
			},
		},
		l1ContractCaller: mockCaller,
		logger:           l,
	}

	// Test
	err = e.validateTaskSignature(task)

	// Assert
	assert.Error(t, err, "Missing AVS config should fail validation")
	assert.Contains(t, err.Error(), "AVS config", "Error should mention AVS config")
	mockCaller.AssertExpectations(t)
}

// TestHandleReceivedTask_ValidTaskAccepted tests that valid tasks are accepted and passed to performer
func TestHandleReceivedTask_ValidTaskAccepted(t *testing.T) {
	// Setup
	aggregatorPrivKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	aggregatorAddress := crypto.PubkeyToAddress(aggregatorPrivKey.PublicKey).Hex()
	executorAddress := "0x1234567890123456789012345678901234567890"
	avsAddress := "0xabcdef1234567890abcdef1234567890abcdef12"

	// Create mock contract caller
	mockCaller := NewEnhancedMockContractCaller()
	mockCaller.SetupValidAggregator(avsAddress, aggregatorAddress, config.CurveTypeECDSA)
	mockCaller.SetupValidOperatorSet(executorAddress, avsAddress, 1)
	mockCaller.SetupOperatorSetCurveType(avsAddress, 1, config.CurveTypeECDSA)

	// Setup certificate digest calculation
	mockCaller.On("CalculateECDSACertificateDigestBytes", mock.Anything, mock.AnythingOfType("uint32"), mock.AnythingOfType("[32]uint8")).
		Return([]byte("certificate_digest"), nil)

	// Create mock performer
	mockPerformer := NewConfigurableMockPerformer()
	mockPerformer.SetCustomResponse([]byte("task completed successfully"))

	// Create signed task
	task := CreateSignedTaskSubmission(t, aggregatorPrivKey, executorAddress, avsAddress)

	// Create executor with signers
	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	require.NoError(t, err)

	executorPrivKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	e := &Executor{
		config: &executorConfig.ExecutorConfig{
			Operator: &config.OperatorConfig{
				Address: executorAddress,
			},
		},
		l1ContractCaller: mockCaller,
		logger:           l,
		avsPerformers:    &sync.Map{},
		store:            memory.NewInMemoryExecutorStore(),
		inflightTasks:    &sync.Map{},
		ecdsaSigner:      NewECDSATestSigner(executorPrivKey),
	}

	// Store performer
	e.avsPerformers.Store(avsAddress, mockPerformer)

	// Test
	result, err := e.handleReceivedTask(context.Background(), task)

	// Assert
	assert.NoError(t, err, "Valid task should be accepted")
	assert.NotNil(t, result, "Result should not be nil")
	assert.Equal(t, task.TaskId, result.TaskId, "Result should have correct task ID")
	assert.Equal(t, []byte("task completed successfully"), result.Output, "Result should have performer output")
	mockCaller.AssertExpectations(t)
}

// TestHandleReceivedTask_InvalidTaskRejected tests that invalid tasks are rejected
func TestHandleReceivedTask_InvalidTaskRejected(t *testing.T) {
	// Setup
	aggregatorPrivKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	aggregatorAddress := crypto.PubkeyToAddress(aggregatorPrivKey.PublicKey).Hex()
	executorAddress := "0x1234567890123456789012345678901234567890"
	wrongExecutorAddress := "0x9999999999999999999999999999999999999999"
	avsAddress := "0xabcdef1234567890abcdef1234567890abcdef12"

	// Create mock contract caller
	mockCaller := NewEnhancedMockContractCaller()
	mockCaller.SetupValidAggregator(avsAddress, aggregatorAddress, config.CurveTypeECDSA)
	mockCaller.SetupValidOperatorSet(executorAddress, avsAddress, 1)

	// Create mock performer
	mockPerformer := NewConfigurableMockPerformer()

	// Create task signed for wrong executor (signature won't match)
	task := CreateSignedTaskSubmission(t, aggregatorPrivKey, wrongExecutorAddress, avsAddress)
	// Set the executor address to match what the executor expects (to pass initial validation)
	task.ExecutorAddress = executorAddress

	// Create executor with signers
	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	require.NoError(t, err)

	executorPrivKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	e := &Executor{
		config: &executorConfig.ExecutorConfig{
			Operator: &config.OperatorConfig{
				Address: executorAddress,
			},
		},
		l1ContractCaller: mockCaller,
		logger:           l,
		avsPerformers:    &sync.Map{},
		store:            memory.NewInMemoryExecutorStore(),
		inflightTasks:    &sync.Map{},
		ecdsaSigner:      NewECDSATestSigner(executorPrivKey),
	}

	// Store performer
	e.avsPerformers.Store(avsAddress, mockPerformer)

	// Test
	result, err := e.handleReceivedTask(context.Background(), task)

	// Assert
	assert.Error(t, err, "Invalid task should be rejected")
	assert.Nil(t, result, "Result should be nil for rejected task")
	assert.Contains(t, err.Error(), "signature verification failed", "Error should mention signature verification")

	// Verify performer was never called
	mockPerformer.AssertNotCalled(t, "RunTask")
	mockCaller.AssertExpectations(t)
}

// TestSignResult_ECDSA tests ECDSA signature generation for task results
func TestSignResult_ECDSA(t *testing.T) {
	// Setup
	executorPrivKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	executorAddress := crypto.PubkeyToAddress(executorPrivKey.PublicKey).Hex()
	avsAddress := "0xabcdef1234567890abcdef1234567890abcdef12"

	// Create mock contract caller
	mockCaller := NewEnhancedMockContractCaller()
	mockCaller.SetupOperatorSetCurveType(avsAddress, 1, config.CurveTypeECDSA)

	// Setup certificate digest calculation expectation
	// The mock should return some deterministic bytes when asked to calculate the certificate digest
	expectedDigestBytes := []byte("mocked_certificate_digest_for_ecdsa")
	mockCaller.On("CalculateECDSACertificateDigestBytes", mock.Anything, mock.Anything, mock.Anything).Return(expectedDigestBytes, nil).Maybe()

	// Create executor with ECDSA signer
	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	require.NoError(t, err)

	e := &Executor{
		config: &executorConfig.ExecutorConfig{
			Operator: &config.OperatorConfig{
				Address: executorAddress,
			},
		},
		l1ContractCaller: mockCaller,
		logger:           l,
		ecdsaSigner:      NewECDSATestSigner(executorPrivKey),
	}

	// Create task and result
	task := &performerTask.PerformerTask{
		TaskID:        "0x0000000000000000000000000000000000000000000000000000000000000001",
		Avs:           avsAddress,
		OperatorSetId: 1,
		Payload:       []byte("test payload"),
	}

	result := &performerTask.PerformerTaskResult{
		TaskID: task.TaskID,
		Result: []byte("task output data"),
	}

	// Test
	resultSig, authSig, err := e.signResult(context.Background(), task, result)

	// Assert
	assert.NoError(t, err, "ECDSA signing should succeed")
	assert.NotEmpty(t, resultSig, "Result signature should not be empty")
	assert.NotEmpty(t, authSig, "Auth signature should not be empty")
	// ECDSA signatures are typically 65 bytes (r: 32, s: 32, v: 1)
	assert.Len(t, resultSig, 65, "Result signature should be 65 bytes for ECDSA")
	assert.Len(t, authSig, 65, "Auth signature should be 65 bytes for ECDSA")

	mockCaller.AssertExpectations(t)
}

// TestSignResult_BN254 tests BN254 signature generation for task results
func TestSignResult_BN254(t *testing.T) {
	// Setup
	resultPayload := []byte("test resultPayload")
	executorPrivKey, _, err := bn254.GenerateKeyPair()
	require.NoError(t, err)
	executorAddress := "0x1234567890123456789012345678901234567890"
	avsAddress := "0xabcdef1234567890abcdef1234567890abcdef12"

	// Create mock contract caller
	mockCaller := NewEnhancedMockContractCaller()
	mockCaller.SetupOperatorSetCurveType(avsAddress, 1, config.CurveTypeBN254)

	// Setup certificate digest calculation expectation
	// The mock should return some deterministic bytes when asked to calculate the certificate digest
	expectedDigestBytes := []byte("mocked_certificate_digest_for_bn254")
	mockCaller.On("CalculateBN254CertificateDigestBytes", mock.Anything, mock.Anything, mock.Anything).Return(expectedDigestBytes, nil).Maybe()

	// Create executor with BN254 signer
	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	require.NoError(t, err)

	e := &Executor{
		config: &executorConfig.ExecutorConfig{
			Operator: &config.OperatorConfig{
				Address: executorAddress,
			},
		},
		l1ContractCaller: mockCaller,
		logger:           l,
		bn254Signer:      NewBN254TestSigner(executorPrivKey),
	}

	// Create task and result
	task := &performerTask.PerformerTask{
		TaskID:        "0x0000000000000000000000000000000000000000000000000000000000000001",
		Avs:           avsAddress,
		OperatorSetId: 1,
		Payload:       []byte("test resultPayload"),
	}

	result := &performerTask.PerformerTaskResult{
		TaskID: task.TaskID,
		Result: resultPayload,
	}

	// Test
	resultSig, authSig, err := e.signResult(context.Background(), task, result)

	// Assert
	assert.NoError(t, err, "BN254 signing should succeed")
	assert.NotEmpty(t, resultSig, "Result signature should not be empty")
	assert.NotEmpty(t, authSig, "Auth signature should not be empty")
	// BN254 signatures are 64 bytes (two field elements of 32 bytes each)
	assert.Len(t, resultSig, 64, "Result signature should be 64 bytes for BN254")
	assert.Len(t, authSig, 64, "Auth signature should be 64 bytes for BN254")

	mockCaller.AssertExpectations(t)
}

// TestSignResult_BindsToExecutor tests that result signatures are bound to specific executor
func TestSignResult_BindsToExecutor(t *testing.T) {
	// Setup two different executors
	executor1PrivKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	executor1Address := crypto.PubkeyToAddress(executor1PrivKey.PublicKey).Hex()

	executor2PrivKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	executor2Address := crypto.PubkeyToAddress(executor2PrivKey.PublicKey).Hex()

	avsAddress := "0xabcdef1234567890abcdef1234567890abcdef12"

	// Create mock contract caller
	mockCaller := NewEnhancedMockContractCaller()
	mockCaller.SetupOperatorSetCurveType(avsAddress, 1, config.CurveTypeECDSA)

	// Setup certificate digest calculation expectation
	mockCaller.On("CalculateECDSACertificateDigestBytes", mock.Anything, mock.Anything, mock.Anything).Return([]byte("mocked_certificate_digest"), nil).Maybe()

	// Create logger
	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	require.NoError(t, err)

	// Create first executor
	e1 := &Executor{
		config: &executorConfig.ExecutorConfig{
			Operator: &config.OperatorConfig{
				Address: executor1Address,
			},
		},
		l1ContractCaller: mockCaller,
		logger:           l,
		ecdsaSigner:      NewECDSATestSigner(executor1PrivKey),
	}

	// Create second executor
	e2 := &Executor{
		config: &executorConfig.ExecutorConfig{
			Operator: &config.OperatorConfig{
				Address: executor2Address,
			},
		},
		l1ContractCaller: mockCaller,
		logger:           l,
		ecdsaSigner:      NewECDSATestSigner(executor2PrivKey),
	}

	// Create identical task and result for both
	task := &performerTask.PerformerTask{
		TaskID:        "0x0000000000000000000000000000000000000000000000000000000000000001",
		Avs:           avsAddress,
		OperatorSetId: 1,
		Payload:       []byte("test payload"),
	}

	result := &performerTask.PerformerTaskResult{
		TaskID: task.TaskID,
		Result: []byte("task output data"),
	}

	// Sign with both executors
	resultSig1, authSig1, err := e1.signResult(context.Background(), task, result)
	require.NoError(t, err)

	resultSig2, authSig2, err := e2.signResult(context.Background(), task, result)
	require.NoError(t, err)

	// Assert
	// Result signatures will be different because different private keys are used
	// But they're signing the same data (certificate digest)
	assert.NotEqual(t, resultSig1, resultSig2, "Different private keys produce different signatures")
	// Auth signatures should also be different (different operator addresses + different keys)
	assert.NotEqual(t, authSig1, authSig2, "Different executors should produce different auth signatures")
}

// TestSignResult_IncludesAllFields tests that result signatures include all required fields
func TestSignResult_IncludesAllFields(t *testing.T) {
	// Setup
	resultPayload := []byte("complex output with unicode 你好世界 and emojis 🚀")
	executorPrivKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	executorAddress := crypto.PubkeyToAddress(executorPrivKey.PublicKey).Hex()
	avsAddress := "0xabcdef1234567890abcdef1234567890abcdef12"

	// Create mock contract caller
	mockCaller := NewEnhancedMockContractCaller()
	mockCaller.SetupOperatorSetCurveType(avsAddress, 2, config.CurveTypeECDSA)

	// Setup certificate digest calculation expectation
	mockCaller.On("CalculateECDSACertificateDigestBytes", mock.Anything, mock.Anything, mock.Anything).Return([]byte("mocked_certificate_digest"), nil).Maybe()

	// Create executor
	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	require.NoError(t, err)

	e := &Executor{
		config: &executorConfig.ExecutorConfig{
			Operator: &config.OperatorConfig{
				Address: executorAddress,
			},
		},
		l1ContractCaller: mockCaller,
		logger:           l,
		ecdsaSigner:      NewECDSATestSigner(executorPrivKey),
	}

	// Create task with all fields populated
	task := &performerTask.PerformerTask{
		TaskID:        "0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
		Avs:           avsAddress,
		OperatorSetId: 2,
		Payload:       []byte("complex payload with special chars !@#$%^&*()"),
	}

	result := &performerTask.PerformerTaskResult{
		TaskID: task.TaskID,
		Result: resultPayload,
	}

	// Test
	resultSig, authSig, err := e.signResult(context.Background(), task, result)

	// Assert
	assert.NoError(t, err, "Signing should succeed")
	assert.NotEmpty(t, resultSig, "Result signature should not be empty")
	assert.NotEmpty(t, authSig, "Auth signature should not be empty")
	mockCaller.AssertExpectations(t)
}

// TestSignResult_MissingSignerError tests error handling when signer is not configured
func TestSignResult_MissingSignerError(t *testing.T) {
	// Setup
	executorAddress := "0x1234567890123456789012345678901234567890"
	avsAddress := "0xabcdef1234567890abcdef1234567890abcdef12"

	// Create mock contract caller
	mockCaller := NewEnhancedMockContractCaller()
	mockCaller.SetupOperatorSetCurveType(avsAddress, 1, config.CurveTypeECDSA)

	// Create executor WITHOUT signer
	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	require.NoError(t, err)

	e := &Executor{
		config: &executorConfig.ExecutorConfig{
			Operator: &config.OperatorConfig{
				Address: executorAddress,
			},
		},
		l1ContractCaller: mockCaller,
		logger:           l,
		ecdsaSigner:      nil, // No signer configured
	}

	// Create task and result
	task := &performerTask.PerformerTask{
		TaskID:        "0x0000000000000000000000000000000000000000000000000000000000000001",
		Avs:           avsAddress,
		OperatorSetId: 1,
		Payload:       []byte("test payload"),
	}

	result := &performerTask.PerformerTaskResult{
		TaskID: task.TaskID,
		Result: []byte("task output data"),
	}

	// Test
	resultSig, authSig, err := e.signResult(context.Background(), task, result)

	// Assert
	assert.Error(t, err, "Should fail when signer is not configured")
	assert.Contains(t, err.Error(), "signer is not initialized", "Error should mention missing signer")
	assert.Nil(t, resultSig, "Result signature should be nil on error")
	assert.Nil(t, authSig, "Auth signature should be nil on error")

	mockCaller.AssertExpectations(t)
}

// ============================================================================
// Test Helper Types and Functions
// ============================================================================

// EnhancedMockContractCaller provides a more configurable mock for testing
type EnhancedMockContractCaller struct {
	mock.Mock
	contractCaller.IContractCaller

	// Configurable responses
	avsConfigs   map[string]*contractCaller.AVSConfig
	operatorSets map[string]*peering.OperatorSet
	curveTypes   map[string]config.CurveType
}

func (m *EnhancedMockContractCaller) CalculateTaskHashMessage(_ context.Context, taskHash [32]byte, result []byte) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

// NewEnhancedMockContractCaller creates a new enhanced mock contract caller
func NewEnhancedMockContractCaller() *EnhancedMockContractCaller {
	return &EnhancedMockContractCaller{
		avsConfigs:   make(map[string]*contractCaller.AVSConfig),
		operatorSets: make(map[string]*peering.OperatorSet),
		curveTypes:   make(map[string]config.CurveType),
	}
}

// SetupValidAggregator configures a valid aggregator for testing
func (m *EnhancedMockContractCaller) SetupValidAggregator(avsAddress, aggregatorAddress string, curveType config.CurveType) {
	// Setup AVS config
	m.avsConfigs[avsAddress] = &contractCaller.AVSConfig{
		AggregatorOperatorSetId: 0,
		ExecutorOperatorSetIds:  []uint32{1},
	}

	// Setup aggregator operator set
	key := fmt.Sprintf("%s-%s-%d", aggregatorAddress, avsAddress, 0)
	opSet := &peering.OperatorSet{
		OperatorSetID: 0,
		CurveType:     curveType,
	}

	// Set public key based on curve type
	if curveType == config.CurveTypeECDSA {
		opSet.WrappedPublicKey = peering.WrappedPublicKey{
			ECDSAAddress: common.HexToAddress(aggregatorAddress),
		}
	}

	m.operatorSets[key] = opSet

	// Setup GetAVSConfig mock - now expects blockNumber parameter
	m.On("GetAVSConfig", avsAddress, mock.AnythingOfType("uint64")).Return(m.avsConfigs[avsAddress], nil).Maybe()

	// Setup GetOperatorSetDetailsForOperator mock for aggregator
	m.On("GetOperatorSetDetailsForOperator",
		common.HexToAddress(aggregatorAddress),
		avsAddress,
		uint32(0),
		mock.AnythingOfType("uint64"),
	).Return(m.operatorSets[key], nil).Maybe()
}

// SetupValidOperatorSet configures a valid operator set for testing
func (m *EnhancedMockContractCaller) SetupValidOperatorSet(operatorAddress, avsAddress string, operatorSetId uint32) {
	key := fmt.Sprintf("%s-%s-%d", operatorAddress, avsAddress, operatorSetId)
	m.operatorSets[key] = &peering.OperatorSet{
		OperatorSetID: operatorSetId,
	}

	m.On("GetOperatorSetDetailsForOperator",
		common.HexToAddress(operatorAddress),
		avsAddress,
		operatorSetId,
		mock.AnythingOfType("uint64"),
	).Return(m.operatorSets[key], nil).Maybe()
}

// SetupOperatorSetCurveType configures the curve type for an operator set
func (m *EnhancedMockContractCaller) SetupOperatorSetCurveType(avsAddress string, operatorSetId uint32, curveType config.CurveType) {
	key := fmt.Sprintf("%s-%d", avsAddress, operatorSetId)
	m.curveTypes[key] = curveType

	m.On("GetOperatorSetCurveType", avsAddress, operatorSetId, mock.AnythingOfType("uint64")).Return(curveType, nil).Maybe()
}

// GetAVSConfig implementation
func (m *EnhancedMockContractCaller) GetAVSConfig(avsAddress string, blockNumber uint64) (*contractCaller.AVSConfig, error) {
	args := m.Called(avsAddress, blockNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*contractCaller.AVSConfig), args.Error(1)
}

// GetOperatorSetDetailsForOperator implementation
func (m *EnhancedMockContractCaller) GetOperatorSetDetailsForOperator(operatorAddress common.Address, avsAddress string, operatorSetId uint32, blockNumber uint64) (*peering.OperatorSet, error) {
	args := m.Called(operatorAddress, avsAddress, operatorSetId, blockNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*peering.OperatorSet), args.Error(1)
}

// GetOperatorSetCurveType implementation
func (m *EnhancedMockContractCaller) GetOperatorSetCurveType(avsAddress string, operatorSetId uint32, blockNumber uint64) (config.CurveType, error) {
	args := m.Called(avsAddress, operatorSetId, blockNumber)
	return args.Get(0).(config.CurveType), args.Error(1)
}

// CalculateBN254CertificateDigestBytes implementation
func (m *EnhancedMockContractCaller) CalculateBN254CertificateDigestBytes(ctx context.Context, referenceTimestamp uint32, messageHash [32]byte) ([]byte, error) {
	args := m.Called(ctx, referenceTimestamp, messageHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

// CalculateECDSACertificateDigestBytes implementation
func (m *EnhancedMockContractCaller) CalculateECDSACertificateDigestBytes(ctx context.Context, referenceTimestamp uint32, messageHash [32]byte) ([]byte, error) {
	args := m.Called(ctx, referenceTimestamp, messageHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

// ConfigurableMockPerformer provides controllable behavior for testing
type ConfigurableMockPerformer struct {
	mock.Mock
	avsPerformer.IAvsPerformer

	// Control behavior
	shouldFail     bool
	customResponse []byte
	errorMessage   string
}

// NewConfigurableMockPerformer creates a new configurable mock performer
func NewConfigurableMockPerformer() *ConfigurableMockPerformer {
	return &ConfigurableMockPerformer{
		customResponse: []byte("mock result"),
	}
}

// SetFailure configures the performer to fail
func (p *ConfigurableMockPerformer) SetFailure(shouldFail bool, errorMessage string) {
	p.shouldFail = shouldFail
	p.errorMessage = errorMessage
}

// SetCustomResponse sets a custom response for the performer
func (p *ConfigurableMockPerformer) SetCustomResponse(response []byte) {
	p.customResponse = response
}

// RunTask implementation
func (p *ConfigurableMockPerformer) RunTask(ctx context.Context, task *performerTask.PerformerTask) (*performerTask.PerformerTaskResult, error) {
	// Only call if expectations were set
	if len(p.ExpectedCalls) > 0 {
		p.Called(ctx, task)
	}

	// Check if we should use configured behavior
	if p.shouldFail {
		return nil, fmt.Errorf("%s", p.errorMessage)
	}

	// Return configured response
	return &performerTask.PerformerTaskResult{
		TaskID: task.TaskID,
		Result: p.customResponse,
	}, nil
}

// ExecuteWorkflow implementation
func (p *ConfigurableMockPerformer) ExecuteWorkflow(task *executorV1.TaskSubmission) (*executorV1.TaskResult, error) {
	if p.shouldFail {
		return nil, fmt.Errorf("%s", p.errorMessage)
	}

	return &executorV1.TaskResult{
		TaskId:  task.TaskId,
		Output:  p.customResponse,
		Version: 1,
	}, nil
}

// TestSigner provides signing capabilities for tests
type TestSigner struct {
	privateKey interface{} // *ecdsa.PrivateKey or *bn254.PrivateKey
	curveType  config.CurveType
}

// NewECDSATestSigner creates a test signer with ECDSA key
func NewECDSATestSigner(privateKey *ecdsa.PrivateKey) *TestSigner {
	return &TestSigner{
		privateKey: privateKey,
		curveType:  config.CurveTypeECDSA,
	}
}

// NewBN254TestSigner creates a test signer with BN254 key
func NewBN254TestSigner(privateKey *bn254.PrivateKey) *TestSigner {
	return &TestSigner{
		privateKey: privateKey,
		curveType:  config.CurveTypeBN254,
	}
}

// SignMessage signs a message
func (s *TestSigner) SignMessage(data []byte) ([]byte, error) {
	switch s.curveType {
	case config.CurveTypeECDSA:
		key := s.privateKey.(*ecdsa.PrivateKey)
		hash := crypto.Keccak256Hash(data)
		return crypto.Sign(hash.Bytes(), key)
	case config.CurveTypeBN254:
		key := s.privateKey.(*bn254.PrivateKey)
		sig, err := key.Sign(data)
		if err != nil {
			return nil, err
		}
		return sig.Bytes(), nil
	default:
		return nil, fmt.Errorf("unsupported curve type: %v", s.curveType)
	}
}

// SignMessageForSolidity signs a message for Solidity verification
func (s *TestSigner) SignMessageForSolidity(data []byte) ([]byte, error) {
	return s.SignMessage(data)
}

var _ signer.ISigner = (*TestSigner)(nil)

// CreateValidTaskSubmission creates a valid task submission for testing
func CreateValidTaskSubmission(t *testing.T, aggregatorKey interface{}, executorAddress, avsAddress string) *executorV1.TaskSubmission {
	taskID := "0x0000000000000000000000000000000000000000000000000000000000000001"
	operatorSetID := uint32(1)
	payload := []byte("test payload data")

	// Determine aggregator address based on key type
	var aggregatorAddress string
	switch key := aggregatorKey.(type) {
	case *ecdsa.PrivateKey:
		aggregatorAddress = crypto.PubkeyToAddress(key.PublicKey).Hex()
	case *bn254.PrivateKey:
		// For BN254, use a dummy address
		aggregatorAddress = "0xaggregator00000000000000000000000000000"
	default:
		t.Fatalf("unsupported key type: %T", aggregatorKey)
	}

	return &executorV1.TaskSubmission{
		TaskId:             taskID,
		AvsAddress:         avsAddress,
		ExecutorAddress:    executorAddress,
		AggregatorAddress:  aggregatorAddress,
		OperatorSetId:      operatorSetID,
		TaskBlockNumber:    12345678,
		Payload:            payload,
		ReferenceTimestamp: 42,
		Version:            1,
	}
}

// SignTaskSubmission signs a task submission with the given key
func SignTaskSubmission(t *testing.T, task *executorV1.TaskSubmission, signerKey interface{}, executorAddress string) *executorV1.TaskSubmission {
	// Create the message to sign
	message, _ := util.EncodeTaskSubmissionMessageVersioned(
		task.TaskId,
		task.AvsAddress,
		executorAddress,
		task.ReferenceTimestamp,
		task.TaskBlockNumber,
		task.OperatorSetId,
		task.Payload,
		task.Version,
	)

	// Sign based on key type
	switch key := signerKey.(type) {
	case *ecdsa.PrivateKey:
		messageHash := crypto.Keccak256Hash(message)
		signature, err := crypto.Sign(messageHash.Bytes(), key)
		require.NoError(t, err)

		// Convert to ECDSA signature format
		r := new(big.Int).SetBytes(signature[:32])
		s := new(big.Int).SetBytes(signature[32:64])
		ecdsaSig := &ecdsacrypto.Signature{
			V: signature[64] + 27,
			R: r,
			S: s,
		}
		task.Signature = ecdsaSig.Bytes()

	case *bn254.PrivateKey:
		messageHash := crypto.Keccak256Hash(message)
		sig, err := key.Sign(messageHash.Bytes())
		require.NoError(t, err)
		task.Signature = sig.Bytes()

	default:
		t.Fatalf("unsupported key type: %T", signerKey)
	}

	return task
}

// CreateSignedTaskSubmission creates and signs a task submission in one step
func CreateSignedTaskSubmission(t *testing.T, aggregatorKey interface{}, executorAddress, avsAddress string) *executorV1.TaskSubmission {
	task := CreateValidTaskSubmission(t, aggregatorKey, executorAddress, avsAddress)
	return SignTaskSubmission(t, task, aggregatorKey, executorAddress)
}

package types

import (
	"bytes"
	"math/big"
	"testing"
	"time"

	"github.com/iden3/go-iden3-crypto/keccak256"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionLogParser/log"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskSignatureData_DeterministicEncoding(t *testing.T) {
	// Test that the same data produces the same bytes
	var resultDigest [32]byte
	copy(resultDigest[:], keccak256.Hash([]byte("test output")))
	sigData1 := &AuthSignatureData{
		TaskId:          "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		AvsAddress:      "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1",
		OperatorAddress: "0x5B38Da6a701c568545dCfcB03FcB875f56beddC4",
		OperatorSetId:   42,
		ResultSigDigest: resultDigest,
	}

	sigData2 := &AuthSignatureData{
		TaskId:          "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		AvsAddress:      "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1",
		OperatorAddress: "0x5B38Da6a701c568545dCfcB03FcB875f56beddC4",
		OperatorSetId:   42,
		ResultSigDigest: resultDigest,
	}

	bytes1 := sigData1.ToSigningBytes()
	bytes2 := sigData2.ToSigningBytes()

	// Should be identical for same input
	assert.Equal(t, bytes1, bytes2, "Same input should produce identical bytes")
	assert.Equal(t, 160, len(bytes1), "Output should be exactly 160 bytes")

	// Change one field and verify different output
	sigData2.OperatorSetId = 43
	bytes3 := sigData2.ToSigningBytes()
	assert.NotEqual(t, bytes1, bytes3, "Different OperatorSetId should produce different bytes")

	// Change another field
	sigData2.OperatorSetId = 42 // Reset
	var newDigest [32]byte
	copy(newDigest[:], keccak256.Hash([]byte("different output")))
	sigData2.ResultSigDigest = newDigest
	bytes4 := sigData2.ToSigningBytes()
	assert.NotEqual(t, bytes1, bytes4, "Different Output should produce different bytes")

	// Test that addresses are normalized (case insensitive)
	sigData3 := &AuthSignatureData{
		TaskId:          "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		AvsAddress:      "0x742D35CC6634C0532925A3B844BC9E7595F0BEB1", // uppercase
		OperatorAddress: "0x5b38da6a701c568545dcfcb03fcb875f56beddc4", // lowercase
		OperatorSetId:   42,
		ResultSigDigest: sigData1.ResultSigDigest,
	}
	bytes5 := sigData3.ToSigningBytes()
	assert.Equal(t, bytes1, bytes5, "Address case should not affect output")
}

func TestTaskSignatureData_BytesStructure(t *testing.T) {
	var resultDigest [32]byte
	copy(resultDigest[:], keccak256.Hash([]byte("test output")))
	sigData := &AuthSignatureData{
		TaskId:          "0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
		AvsAddress:      "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1",
		OperatorAddress: "0x5B38Da6a701c568545dCfcB03FcB875f56beddC4",
		OperatorSetId:   1,
		ResultSigDigest: resultDigest,
	}

	signedBytes := sigData.ToSigningBytes()

	// Verify total length
	assert.Equal(t, 160, len(signedBytes), "Total length should be 160 bytes")

	// Verify structure:
	// Bytes 0-31: TaskId (32 bytes)
	// Bytes 32-63: AvsAddress (20 bytes padded to 32)
	// Bytes 64-95: OperatorAddress (20 bytes padded to 32)
	// Bytes 96-127: OperatorSetId (uint32 padded to 32)
	// Bytes 128-159: Output hash (32 bytes)

	// Check that addresses are properly padded (first 12 bytes should be zeros)
	avsAddressSection := signedBytes[32:64]
	operatorAddressSection := signedBytes[64:96]

	// First 12 bytes should be padding (zeros)
	assert.True(t, bytes.Equal(avsAddressSection[:12], make([]byte, 12)), "AvsAddress should be left-padded with zeros")
	assert.True(t, bytes.Equal(operatorAddressSection[:12], make([]byte, 12)), "OperatorAddress should be left-padded with zeros")

	// Check OperatorSetId is properly padded (first 28 bytes should be zeros, last 4 bytes contain the uint32)
	operatorSetIdSection := signedBytes[96:128]
	assert.True(t, bytes.Equal(operatorSetIdSection[:28], make([]byte, 28)), "OperatorSetId should be left-padded with zeros")
	assert.Equal(t, byte(0), operatorSetIdSection[28])
	assert.Equal(t, byte(0), operatorSetIdSection[29])
	assert.Equal(t, byte(0), operatorSetIdSection[30])
	assert.Equal(t, byte(1), operatorSetIdSection[31]) // OperatorSetId = 1
}

func TestEndToEndSignatureVerification(t *testing.T) {
	// Generate a private key for testing
	privKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	operatorAddr := crypto.PubkeyToAddress(privKey.PublicKey)
	hash := keccak256.Hash([]byte("result data"))
	var resultDigest [32]byte
	copy(resultDigest[:], keccak256.Hash(hash))

	verifyData := &AuthSignatureData{
		TaskId:          "0xdeadbeef00000000000000000000000000000000000000000000000000000000",
		AvsAddress:      "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1",
		OperatorAddress: operatorAddr.Hex(),
		OperatorSetId:   1,
		ResultSigDigest: resultDigest,
	}

	// Sign the data
	signedBytes := verifyData.ToSigningBytes()
	digest := util.GetKeccak256Digest(signedBytes)
	sig, err := crypto.Sign(digest[:], privKey)
	require.NoError(t, err)

	// Create task result
	taskResult := &TaskResult{
		TaskId:          verifyData.TaskId,
		AvsAddress:      verifyData.AvsAddress,
		OperatorAddress: verifyData.OperatorAddress,
		OperatorSetId:   verifyData.OperatorSetId,
		Output:          []byte("result data"),
		ResultSignature: sig,
	}

	verifyBytes := verifyData.ToSigningBytes()
	verifyDigest := util.GetKeccak256Digest(verifyBytes)

	// Should match original digest
	assert.Equal(t, digest, verifyDigest, "Reconstructed digest should match original")

	// Verify signature is valid
	sigPublicKey, err := crypto.SigToPub(verifyDigest[:], taskResult.ResultSignature)
	require.NoError(t, err)

	recoveredAddr := crypto.PubkeyToAddress(*sigPublicKey)
	assert.Equal(t, operatorAddr, recoveredAddr, "Recovered address should match operator address")
}

func TestTaskSignatureData_SecurityProperties(t *testing.T) {
	// Test that changing any field produces a different signature
	var resultDigest [32]byte
	copy(resultDigest[:], keccak256.Hash([]byte("test output")))

	baseData := &AuthSignatureData{
		TaskId:          "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		AvsAddress:      "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1",
		OperatorAddress: "0x5B38Da6a701c568545dCfcB03FcB875f56beddC4",
		OperatorSetId:   42,
		ResultSigDigest: resultDigest,
	}

	baseBytes := baseData.ToSigningBytes()
	baseDigest := util.GetKeccak256Digest(baseBytes)

	// Test TaskId change
	{
		modified := *baseData
		modified.TaskId = "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
		modBytes := modified.ToSigningBytes()
		modDigest := util.GetKeccak256Digest(modBytes)
		assert.NotEqual(t, baseDigest, modDigest, "Different TaskId should produce different digest")
	}

	// Test AvsAddress change
	{
		modified := *baseData
		modified.AvsAddress = "0x0000000000000000000000000000000000000001"
		modBytes := modified.ToSigningBytes()
		modDigest := util.GetKeccak256Digest(modBytes)
		assert.NotEqual(t, baseDigest, modDigest, "Different AvsAddress should produce different digest")
	}

	// Test OperatorAddress change
	{
		modified := *baseData
		modified.OperatorAddress = "0x0000000000000000000000000000000000000002"
		modBytes := modified.ToSigningBytes()
		modDigest := util.GetKeccak256Digest(modBytes)
		assert.NotEqual(t, baseDigest, modDigest, "Different OperatorAddress should produce different digest")
	}

	// Test OperatorSetId change
	{
		modified := *baseData
		modified.OperatorSetId = 999
		modBytes := modified.ToSigningBytes()
		modDigest := util.GetKeccak256Digest(modBytes)
		assert.NotEqual(t, baseDigest, modDigest, "Different OperatorSetId should produce different digest")
	}

	// Test Output change
	{
		modified := *baseData
		var nextDigest [32]byte
		copy(nextDigest[:], keccak256.Hash([]byte("completely different output")))
		modified.ResultSigDigest = nextDigest
		modBytes := modified.ToSigningBytes()
		modDigest := util.GetKeccak256Digest(modBytes)
		assert.NotEqual(t, baseDigest, modDigest, "Different Output should produce different digest")
	}
}

func TestTaskSignatureData_PreventReplayAttacks(t *testing.T) {
	// Generate keys for two different operators
	privKey1, err := crypto.GenerateKey()
	require.NoError(t, err)
	operatorAddr1 := crypto.PubkeyToAddress(privKey1.PublicKey)

	privKey2, err := crypto.GenerateKey()
	require.NoError(t, err)
	operatorAddr2 := crypto.PubkeyToAddress(privKey2.PublicKey)

	var resultDigest [32]byte
	copy(resultDigest[:], keccak256.Hash([]byte("test output")))
	// Original task for operator 1
	originalData := &AuthSignatureData{
		TaskId:          "0xdeadbeef00000000000000000000000000000000000000000000000000000000",
		AvsAddress:      "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1",
		OperatorAddress: operatorAddr1.Hex(),
		OperatorSetId:   1,
		ResultSigDigest: resultDigest,
	}

	// Sign with operator 1's key
	signedBytes := originalData.ToSigningBytes()
	digest := util.GetKeccak256Digest(signedBytes)
	sig1, err := crypto.Sign(digest[:], privKey1)
	require.NoError(t, err)

	// Try to replay the same task to a different operator set
	replayData := &AuthSignatureData{
		TaskId:          originalData.TaskId,
		AvsAddress:      originalData.AvsAddress,
		OperatorAddress: operatorAddr2.Hex(), // Different operator
		OperatorSetId:   2,                   // Different operator set
		ResultSigDigest: originalData.ResultSigDigest,
	}

	replayBytes := replayData.ToSigningBytes()
	replayDigest := util.GetKeccak256Digest(replayBytes)

	// The digest should be different, preventing replay
	assert.NotEqual(t, digest, replayDigest, "Different operator/set should produce different digest")

	// Verify that the original signature won't validate for the replay attempt
	sigPublicKey, err := crypto.SigToPub(replayDigest[:], sig1)
	if err == nil {
		recoveredAddr := crypto.PubkeyToAddress(*sigPublicKey)
		assert.NotEqual(t, operatorAddr2, recoveredAddr, "Signature should not validate for different operator")
	}
}

func TestNewTaskFromLog_DeadlineConversion(t *testing.T) {
	tests := []struct {
		name                string
		taskDeadlineSeconds int64
		wantDeadlineTime    time.Time
	}{
		{
			name:                "unix epoch",
			taskDeadlineSeconds: 0,
			wantDeadlineTime:    time.Unix(0, 0),
		},
		{
			name:                "recent timestamp",
			taskDeadlineSeconds: 1704067200, // 2024-01-01 00:00:00 UTC
			wantDeadlineTime:    time.Unix(1704067200, 0),
		},
		{
			name:                "future timestamp",
			taskDeadlineSeconds: 2000000000, // 2033-05-18 03:33:20 UTC
			wantDeadlineTime:    time.Unix(2000000000, 0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock decoded log
			avsAddr := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1")
			taskId := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
			inboxAddress := "0x5B38Da6a701c568545dCfcB03FcB875f56beddC4"

			decodedLog := &log.DecodedLog{
				LogIndex:  1,
				Address:   inboxAddress,
				EventName: "TaskCreated",
				Arguments: []log.Argument{
					{Name: "creatorAddress", Value: "0x123", Indexed: true},
					{Name: "taskId", Value: taskId, Indexed: true},
					{Name: "avsAddress", Value: avsAddr, Indexed: true},
				},
				OutputData: map[string]interface{}{
					"ExecutorOperatorSetId":           uint32(1),
					"OperatorTableReferenceTimestamp": uint32(1704067200),
					"TaskDeadline":                    big.NewInt(tt.taskDeadlineSeconds),
					"Payload":                         []byte("test payload"),
				},
			}

			// Create mock block
			block := &ethereum.EthereumBlock{
				Hash:      ethereum.EthereumHexString("0xblockhash"),
				Number:    ethereum.EthereumQuantity(12345),
				ChainId:   config.ChainId(1),
				Timestamp: ethereum.EthereumQuantity(1704067200),
			}

			// Call NewTaskFromLog
			task, err := NewTaskFromLog(decodedLog, block, inboxAddress)
			require.NoError(t, err)
			require.NotNil(t, task)

			// Verify deadline is set correctly
			require.NotNil(t, task.DeadlineUnixSeconds, "DeadlineUnixSeconds should not be nil")
			assert.Equal(t, tt.wantDeadlineTime.Unix(), task.DeadlineUnixSeconds.Unix(),
				"DeadlineUnixSeconds should match expected timestamp")

			// Verify other fields are set correctly
			assert.Equal(t, taskId, task.TaskId)
			assert.Equal(t, "0x742d35cc6634c0532925a3b844bc9e7595f0beb1", task.AVSAddress) // lowercased
			assert.Equal(t, uint32(1), task.OperatorSetId)
			assert.Equal(t, inboxAddress, task.CallbackAddr)
			assert.Equal(t, []byte("test payload"), task.Payload)
			assert.Equal(t, config.ChainId(1), task.ChainId)
			assert.Equal(t, uint64(12345), task.SourceBlockNumber)
			assert.Equal(t, uint32(1704067200), task.ReferenceTimestamp)
		})
	}
}

func TestNewTaskFromLog_InvalidDeadline(t *testing.T) {
	// Test that a deadline larger than MaxInt64 returns an error
	avsAddr := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1")
	taskId := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	inboxAddress := "0x5B38Da6a701c568545dCfcB03FcB875f56beddC4"

	// Create a TaskDeadline that's too large (> MaxInt64)
	veryLargeDeadline := new(big.Int)
	veryLargeDeadline.SetString("9999999999999999999", 10) // Much larger than MaxInt64

	decodedLog := &log.DecodedLog{
		LogIndex:  1,
		Address:   inboxAddress,
		EventName: "TaskCreated",
		Arguments: []log.Argument{
			{Name: "creatorAddress", Value: "0x123", Indexed: true},
			{Name: "taskId", Value: taskId, Indexed: true},
			{Name: "avsAddress", Value: avsAddr, Indexed: true},
		},
		OutputData: map[string]interface{}{
			"ExecutorOperatorSetId":           uint32(1),
			"OperatorTableReferenceTimestamp": uint32(1704067200),
			"TaskDeadline":                    veryLargeDeadline,
			"Payload":                         []byte("test payload"),
		},
	}

	block := &ethereum.EthereumBlock{
		Hash:      ethereum.EthereumHexString("0xblockhash"),
		Number:    ethereum.EthereumQuantity(12345),
		ChainId:   config.ChainId(1),
		Timestamp: ethereum.EthereumQuantity(1704067200),
	}

	// Call NewTaskFromLog - should return error
	task, err := NewTaskFromLog(decodedLog, block, inboxAddress)
	require.Error(t, err)
	assert.Nil(t, task)
	assert.Contains(t, err.Error(), "task deadline too large")
}

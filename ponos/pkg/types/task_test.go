package types

import (
	"bytes"
	"testing"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskSignatureData_DeterministicEncoding(t *testing.T) {
	// Test that the same data produces the same bytes
	sigData1 := &TaskSignatureData{
		TaskId:          "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		AvsAddress:      "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1",
		OperatorAddress: "0x5B38Da6a701c568545dCfcB03FcB875f56beddC4",
		OperatorSetId:   42,
		Output:          []byte("test output"),
	}

	sigData2 := &TaskSignatureData{
		TaskId:          "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		AvsAddress:      "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1",
		OperatorAddress: "0x5B38Da6a701c568545dCfcB03FcB875f56beddC4",
		OperatorSetId:   42,
		Output:          []byte("test output"),
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
	sigData2.Output = []byte("different output")
	bytes4 := sigData2.ToSigningBytes()
	assert.NotEqual(t, bytes1, bytes4, "Different Output should produce different bytes")

	// Test that addresses are normalized (case insensitive)
	sigData3 := &TaskSignatureData{
		TaskId:          "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		AvsAddress:      "0x742D35CC6634C0532925A3B844BC9E7595F0BEB1", // uppercase
		OperatorAddress: "0x5b38da6a701c568545dcfcb03fcb875f56beddc4", // lowercase
		OperatorSetId:   42,
		Output:          []byte("test output"),
	}
	bytes5 := sigData3.ToSigningBytes()
	assert.Equal(t, bytes1, bytes5, "Address case should not affect output")
}

func TestTaskSignatureData_BytesStructure(t *testing.T) {
	sigData := &TaskSignatureData{
		TaskId:          "0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
		AvsAddress:      "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1",
		OperatorAddress: "0x5B38Da6a701c568545dCfcB03FcB875f56beddC4",
		OperatorSetId:   1,
		Output:          []byte("result data"),
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

	sigData := &TaskSignatureData{
		TaskId:          "0xdeadbeef00000000000000000000000000000000000000000000000000000000",
		AvsAddress:      "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1",
		OperatorAddress: operatorAddr.Hex(),
		OperatorSetId:   1,
		Output:          []byte("result data"),
	}

	// Sign the data
	signedBytes := sigData.ToSigningBytes()
	digest := util.GetKeccak256Digest(signedBytes)
	sig, err := crypto.Sign(digest[:], privKey)
	require.NoError(t, err)

	// Create task result
	taskResult := &TaskResult{
		TaskId:          sigData.TaskId,
		AvsAddress:      sigData.AvsAddress,
		OperatorAddress: sigData.OperatorAddress,
		OperatorSetId:   sigData.OperatorSetId,
		Output:          sigData.Output,
		Signature:       sig,
	}

	// Verify - reconstruct signing data and check
	verifyData := &TaskSignatureData{
		TaskId:          taskResult.TaskId,
		AvsAddress:      taskResult.AvsAddress,
		OperatorAddress: taskResult.OperatorAddress,
		OperatorSetId:   taskResult.OperatorSetId,
		Output:          taskResult.Output,
	}

	verifyBytes := verifyData.ToSigningBytes()
	verifyDigest := util.GetKeccak256Digest(verifyBytes)

	// Should match original digest
	assert.Equal(t, digest, verifyDigest, "Reconstructed digest should match original")

	// Verify signature is valid
	sigPublicKey, err := crypto.SigToPub(verifyDigest[:], taskResult.Signature)
	require.NoError(t, err)

	recoveredAddr := crypto.PubkeyToAddress(*sigPublicKey)
	assert.Equal(t, operatorAddr, recoveredAddr, "Recovered address should match operator address")
}

func TestTaskSignatureData_SecurityProperties(t *testing.T) {
	// Test that changing any field produces a different signature
	baseData := &TaskSignatureData{
		TaskId:          "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		AvsAddress:      "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1",
		OperatorAddress: "0x5B38Da6a701c568545dCfcB03FcB875f56beddC4",
		OperatorSetId:   42,
		Output:          []byte("test output"),
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
		modified.Output = []byte("completely different output")
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

	// Original task for operator 1
	originalData := &TaskSignatureData{
		TaskId:          "0xdeadbeef00000000000000000000000000000000000000000000000000000000",
		AvsAddress:      "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1",
		OperatorAddress: operatorAddr1.Hex(),
		OperatorSetId:   1,
		Output:          []byte("result data"),
	}

	// Sign with operator 1's key
	signedBytes := originalData.ToSigningBytes()
	digest := util.GetKeccak256Digest(signedBytes)
	sig1, err := crypto.Sign(digest[:], privKey1)
	require.NoError(t, err)

	// Try to replay the same task to a different operator set
	replayData := &TaskSignatureData{
		TaskId:          originalData.TaskId,
		AvsAddress:      originalData.AvsAddress,
		OperatorAddress: operatorAddr2.Hex(), // Different operator
		OperatorSetId:   2,                   // Different operator set
		Output:          originalData.Output,
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

package util

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeTaskSubmissionMessage(t *testing.T) {
	tests := []struct {
		name             string
		taskId           string
		avsAddress       string
		executorAddress  string
		operatorSetId    uint32
		blockNumber      uint64
		timestamp        uint32
		payload          []byte
		expectedLength   int
		checkDeterminism bool
		version          uint32
	}{
		{
			name:             "basic encoding",
			taskId:           "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			avsAddress:       "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb2",
			executorAddress:  "0x5aAeb6053f3E94C9b9A09f33669435E7Ef1BeAed",
			operatorSetId:    42,
			blockNumber:      12345678,
			timestamp:        1234567890,
			payload:          []byte("test payload"),
			expectedLength:   224,
			checkDeterminism: true,
			version:          1,
		},
		{
			name:             "empty payload",
			taskId:           "0xdeadbeef",
			avsAddress:       "0x0000000000000000000000000000000000000001",
			executorAddress:  "0x0000000000000000000000000000000000000002",
			operatorSetId:    0,
			blockNumber:      1,
			timestamp:        0,
			payload:          []byte{},
			expectedLength:   224,
			checkDeterminism: true,
			version:          1,
		},
		{
			name:             "max operator set id and timestamp",
			taskId:           "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			avsAddress:       "0xffffffffffffffffffffffffffffffffffffffff",
			executorAddress:  "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
			operatorSetId:    ^uint32(0), // max uint32
			blockNumber:      999999999,
			timestamp:        ^uint32(0), // max uint32
			payload:          []byte("another test"),
			expectedLength:   224,
			checkDeterminism: true,
			version:          1,
		},
		{
			name:             "addresses with mixed case",
			taskId:           "0xABCDEF1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF1234567890",
			avsAddress:       "0xAbCdEf1234567890aBcDeF1234567890aBcDeF12",
			executorAddress:  "0x1234567890AbCdEf1234567890AbCdEf12345678",
			operatorSetId:    100,
			blockNumber:      54321,
			timestamp:        987654321,
			payload:          []byte("mixed case test"),
			expectedLength:   224,
			checkDeterminism: true,
			version:          1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode the message
			encoded, _ := EncodeTaskSubmissionMessageVersioned(
				tt.taskId,
				tt.avsAddress,
				tt.executorAddress,
				tt.timestamp,
				tt.blockNumber,
				tt.operatorSetId,
				tt.payload,
				tt.version,
			)

			// Check length
			assert.Equal(t, tt.expectedLength, len(encoded), "encoded message should be exactly 224 bytes")

			// Check determinism - encoding same data should produce same result
			if tt.checkDeterminism {
				encoded2, _ := EncodeTaskSubmissionMessageVersioned(
					tt.taskId,
					tt.avsAddress,
					tt.executorAddress,
					tt.timestamp,
					tt.blockNumber,
					tt.operatorSetId,
					tt.payload,
					tt.version,
				)
				assert.True(t, bytes.Equal(encoded, encoded2), "encoding should be deterministic")
			}

			// Verify structure of encoded data
			// Bytes 0-31: taskId
			taskIdBytes := common.HexToHash(tt.taskId).Bytes()
			assert.Equal(t, taskIdBytes, encoded[0:32], "first 32 bytes should be taskId")

			// Bytes 32-63: avsAddress (padded)
			avsAddr := common.HexToAddress(tt.avsAddress).Bytes()
			expectedAvsBytes := common.LeftPadBytes(avsAddr, 32)
			assert.Equal(t, expectedAvsBytes, encoded[32:64], "bytes 32-63 should be padded AVS address")

			// Bytes 64-95: executorAddress (padded)
			execAddr := common.HexToAddress(tt.executorAddress).Bytes()
			expectedExecBytes := common.LeftPadBytes(execAddr, 32)
			assert.Equal(t, expectedExecBytes, encoded[64:96], "bytes 64-95 should be padded executor address")

			// Bytes 96-127: operatorSetId (padded)
			// Check that the uint32 is in the last 4 bytes of the 32-byte segment
			assert.Equal(t, byte(tt.operatorSetId>>24), encoded[124])
			assert.Equal(t, byte(tt.operatorSetId>>16), encoded[125])
			assert.Equal(t, byte(tt.operatorSetId>>8), encoded[126])
			assert.Equal(t, byte(tt.operatorSetId), encoded[127])

			// Bytes 128-159: timestamp (padded)
			// Check that the uint32 is in the last 4 bytes of the 32-byte segment
			assert.Equal(t, byte(tt.timestamp>>24), encoded[156])
			assert.Equal(t, byte(tt.timestamp>>16), encoded[157])
			assert.Equal(t, byte(tt.timestamp>>8), encoded[158])
			assert.Equal(t, byte(tt.timestamp), encoded[159])

			// Bytes 160-191: blockNumber (padded)
			// Check that the uint64 is in the last 8 bytes of the 32-byte segment
			assert.Equal(t, byte(tt.blockNumber>>56), encoded[184])
			assert.Equal(t, byte(tt.blockNumber>>48), encoded[185])
			assert.Equal(t, byte(tt.blockNumber>>40), encoded[186])
			assert.Equal(t, byte(tt.blockNumber>>32), encoded[187])
			assert.Equal(t, byte(tt.blockNumber>>24), encoded[188])
			assert.Equal(t, byte(tt.blockNumber>>16), encoded[189])
			assert.Equal(t, byte(tt.blockNumber>>8), encoded[190])
			assert.Equal(t, byte(tt.blockNumber), encoded[191])

			// Bytes 192-223: payload digest
			expectedDigest := crypto.Keccak256Hash(tt.payload)
			assert.Equal(t, expectedDigest[:], encoded[192:224], "last 32 bytes should be payload digest")
		})
	}
}

func TestEncodeTaskResultMessage(t *testing.T) {
	tests := []struct {
		name            string
		taskId          string
		avsAddress      string
		operatorAddress string
		operatorSetId   uint32
		output          []byte
		expectedLength  int
	}{
		{
			name:            "basic result encoding",
			taskId:          "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			avsAddress:      "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb2",
			operatorAddress: "0x5aAeb6053f3E94C9b9A09f33669435E7Ef1BeAed",
			operatorSetId:   42,
			output:          []byte("task result"),
			expectedLength:  160,
		},
		{
			name:            "empty output",
			taskId:          "0xdeadbeef",
			avsAddress:      "0x0000000000000000000000000000000000000001",
			operatorAddress: "0x0000000000000000000000000000000000000002",
			operatorSetId:   0,
			output:          []byte{},
			expectedLength:  160,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode the message
			encoded, _ := EncodeTaskResultMessageVersioned(
				tt.taskId,
				tt.avsAddress,
				tt.operatorAddress,
				tt.operatorSetId,
				tt.output,
				1,
			)

			// Check length
			assert.Equal(t, tt.expectedLength, len(encoded), "encoded message should be exactly 160 bytes")

			// Check determinism
			encoded2, _ := EncodeTaskResultMessageVersioned(
				tt.taskId,
				tt.avsAddress,
				tt.operatorAddress,
				tt.operatorSetId,
				tt.output,
				1,
			)
			assert.True(t, bytes.Equal(encoded, encoded2), "encoding should be deterministic")

			// Verify structure matches TaskSubmission format (for consistency)
			// The format should be identical, just with output instead of payload
			outputDigest := crypto.Keccak256Hash(tt.output)
			assert.Equal(t, outputDigest[:], encoded[128:160], "last 32 bytes should be output digest")
		})
	}
}

func TestSignatureDataStructures(t *testing.T) {
	t.Run("TaskSubmissionSignatureData", func(t *testing.T) {
		payload := []byte("test payload data")
		payloadDigest := crypto.Keccak256Hash(payload)

		sigData := &TaskSubmissionSignatureData{
			TaskId:          "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			AvsAddress:      "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb2",
			ExecutorAddress: "0x5aAeb6053f3E94C9b9A09f33669435E7Ef1BeAed",
			OperatorSetId:   42,
			BlockNumber:     12345678,
			Timestamp:       1234567890,
			PayloadDigest:   [32]byte(payloadDigest),
		}

		bytes := sigData.ToSigningBytes()
		assert.Equal(t, 224, len(bytes), "signing bytes should be 224 bytes")

		// Verify the same result as EncodeTaskSubmissionMessage
		encodedDirect, _ := EncodeTaskSubmissionMessageVersioned(
			sigData.TaskId,
			sigData.AvsAddress,
			sigData.ExecutorAddress,
			sigData.Timestamp,
			sigData.BlockNumber,
			sigData.OperatorSetId,
			payload,
			1,
		)
		assert.Equal(t, encodedDirect, bytes, "ToSigningBytes should match EncodeTaskSubmissionMessage")
	})

	t.Run("TaskResultSignatureData", func(t *testing.T) {
		output := []byte("task execution result")
		outputDigest := crypto.Keccak256Hash(output)

		sigData := &TaskResultSignatureData{
			TaskId:          "0xdeadbeef",
			AvsAddress:      "0x0000000000000000000000000000000000000001",
			OperatorAddress: "0x0000000000000000000000000000000000000002",
			OperatorSetId:   100,
			OutputDigest:    [32]byte(outputDigest),
		}

		bytes := sigData.ToSigningBytes()
		assert.Equal(t, 160, len(bytes), "signing bytes should be 160 bytes")

		// Verify the same result as EncodeTaskResultMessage
		encodedDirect, _ := EncodeTaskResultMessageVersioned(
			sigData.TaskId,
			sigData.AvsAddress,
			sigData.OperatorAddress,
			sigData.OperatorSetId,
			output,
			1,
		)
		assert.Equal(t, encodedDirect, bytes, "ToSigningBytes should match EncodeTaskResultMessage")
	})
}

func TestAbiEncodingHelpers(t *testing.T) {
	t.Run("AbiEncodeUint32", func(t *testing.T) {
		tests := []struct {
			value    uint32
			expected string // hex representation
		}{
			{0, "0000000000000000000000000000000000000000000000000000000000000000"},
			{1, "0000000000000000000000000000000000000000000000000000000000000001"},
			{255, "00000000000000000000000000000000000000000000000000000000000000ff"},
			{256, "0000000000000000000000000000000000000000000000000000000000000100"},
			{^uint32(0), "00000000000000000000000000000000000000000000000000000000ffffffff"},
		}

		for _, tt := range tests {
			encoded := AbiEncodeUint32(tt.value)
			assert.Equal(t, 32, len(encoded), "encoded uint32 should be 32 bytes")
			assert.Equal(t, tt.expected, hex.EncodeToString(encoded))
		}
	})

	t.Run("AbiEncodeAddress", func(t *testing.T) {
		tests := []struct {
			address  string
			expected string // hex representation (lowercase)
		}{
			{
				"0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb2",
				"000000000000000000000000742d35cc6634c0532925a3b844bc9e7595f0beb2",
			},
			{
				"0x0000000000000000000000000000000000000001",
				"0000000000000000000000000000000000000000000000000000000000000001",
			},
			{
				"0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
				"000000000000000000000000ffffffffffffffffffffffffffffffffffffffff",
			},
		}

		for _, tt := range tests {
			encoded := AbiEncodeAddress(tt.address)
			assert.Equal(t, 32, len(encoded), "encoded address should be 32 bytes")
			assert.Equal(t, tt.expected, hex.EncodeToString(encoded))
		}
	})

	t.Run("AbiEncodeBytes32", func(t *testing.T) {
		tests := []struct {
			value    string
			expected string // hex representation
		}{
			{
				"0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
				"1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			},
			{
				"0xdeadbeef",
				"00000000000000000000000000000000000000000000000000000000deadbeef",
			},
			{
				"0x00",
				"0000000000000000000000000000000000000000000000000000000000000000",
			},
		}

		for _, tt := range tests {
			encoded := AbiEncodeBytes32(tt.value)
			assert.Equal(t, 32, len(encoded), "encoded bytes32 should be 32 bytes")
			assert.Equal(t, tt.expected, hex.EncodeToString(encoded))
		}
	})
}

func TestTimestampFieldInEncoding(t *testing.T) {
	// Test that different timestamps produce different encoded messages
	taskId := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	avsAddress := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb2"
	executorAddress := "0x5aAeb6053f3E94C9b9A09f33669435E7Ef1BeAed"
	operatorSetId := uint32(42)
	blockNumber := uint64(12345678)
	payload := []byte("test payload")

	// Create messages with different timestamps
	timestamp1 := uint32(1000000000)
	timestamp2 := uint32(2000000000)

	message1, _ := EncodeTaskSubmissionMessageVersioned(taskId, avsAddress, executorAddress, timestamp1, blockNumber, operatorSetId, payload, 1)
	message2, _ := EncodeTaskSubmissionMessageVersioned(taskId, avsAddress, executorAddress, timestamp2, blockNumber, operatorSetId, payload, 1)

	// Messages should be different due to different timestamps
	assert.False(t, bytes.Equal(message1, message2),
		"messages with different timestamps should be different")

	// Verify that only the timestamp bytes (128-159) are different
	assert.Equal(t, message1[0:128], message2[0:128],
		"bytes before timestamp should be the same")
	assert.NotEqual(t, message1[128:160], message2[128:160],
		"timestamp bytes should be different")
	assert.Equal(t, message1[160:], message2[160:],
		"bytes after timestamp should be the same")

	// Test specific timestamp values encoding
	timestampZero := uint32(0)
	messageZero, _ := EncodeTaskSubmissionMessageVersioned(taskId, avsAddress, executorAddress, timestampZero, blockNumber, operatorSetId, payload, 1)
	// Check that timestamp 0 is properly encoded (all zeros in timestamp position)
	for i := 128; i < 156; i++ {
		assert.Equal(t, byte(0), messageZero[i], "timestamp 0 should be all zeros")
	}

	timestampMax := ^uint32(0) // max uint32
	messageMax, _ := EncodeTaskSubmissionMessageVersioned(taskId, avsAddress, executorAddress, timestampMax, blockNumber, operatorSetId, payload, 1)
	// Check that max timestamp is properly encoded
	assert.Equal(t, byte(0xff), messageMax[156])
	assert.Equal(t, byte(0xff), messageMax[157])
	assert.Equal(t, byte(0xff), messageMax[158])
	assert.Equal(t, byte(0xff), messageMax[159])
}

func TestSignatureBindingPreventsReplay(t *testing.T) {
	// This test verifies that signatures are bound to specific executors
	// and cannot be reused for different executors

	taskId := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	avsAddress := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb2"
	executor1 := "0x5aAeb6053f3E94C9b9A09f33669435E7Ef1BeAed"
	executor2 := "0x1111111111111111111111111111111111111111"
	operatorSetId := uint32(42)
	payload := []byte("important task data")

	// Create messages for two different executors
	blockNumber := uint64(12345)
	timestamp := uint32(1234567890)
	message1, _ := EncodeTaskSubmissionMessageVersioned(taskId, avsAddress, executor1, timestamp, blockNumber, operatorSetId, payload, 1)
	message2, _ := EncodeTaskSubmissionMessageVersioned(taskId, avsAddress, executor2, timestamp, blockNumber, operatorSetId, payload, 1)

	// Messages should be different even though all other parameters are the same
	assert.False(t, bytes.Equal(message1, message2),
		"messages for different executors should be different")

	// Verify that the only difference is in the executor address bytes (64-95)
	assert.Equal(t, message1[0:64], message2[0:64],
		"taskId and avsAddress should be the same")
	assert.NotEqual(t, message1[64:96], message2[64:96],
		"executor address bytes should be different")
	assert.Equal(t, message1[96:], message2[96:],
		"operatorSetId and payload digest should be the same")
}

func TestCompatibilityWithSolidityAbiEncoding(t *testing.T) {
	// This test ensures our encoding matches Solidity's abi.encode() behavior
	// for the message format:
	// abi.encode(taskId, avsAddress, executorAddress, operatorSetId, payloadDigest)

	taskId := "0x0000000000000000000000000000000000000000000000000000000000000001"
	avsAddress := "0x0000000000000000000000000000000000000002"
	executorAddress := "0x0000000000000000000000000000000000000003"
	operatorSetId := uint32(1)
	payload := []byte{} // empty payload for simplicity

	blockNumber := uint64(999)
	timestamp := uint32(42)
	encoded, _ := EncodeTaskSubmissionMessageVersioned(taskId, avsAddress, executorAddress, timestamp, blockNumber, operatorSetId, payload, 1)

	// Expected encoding:
	// taskId:          0000000000000000000000000000000000000000000000000000000000000001
	// avsAddress:      0000000000000000000000000000000000000000000000000000000000000002
	// executorAddress: 0000000000000000000000000000000000000000000000000000000000000003
	// operatorSetId:   0000000000000000000000000000000000000000000000000000000000000001
	// timestamp:       000000000000000000000000000000000000000000000000000000000000002a (42 in hex)
	// blockNumber:     00000000000000000000000000000000000000000000000000000000000003e7 (999 in hex)
	// payloadDigest:   c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470 (keccak256 of empty)

	expectedTaskId := "0000000000000000000000000000000000000000000000000000000000000001"
	expectedAvsAddr := "0000000000000000000000000000000000000000000000000000000000000002"
	expectedExecAddr := "0000000000000000000000000000000000000000000000000000000000000003"
	expectedOpSetId := "0000000000000000000000000000000000000000000000000000000000000001"
	expectedTimestamp := "000000000000000000000000000000000000000000000000000000000000002a"
	expectedBlockNum := "00000000000000000000000000000000000000000000000000000000000003e7"
	expectedPayloadDigest := "c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"

	require.Equal(t, expectedTaskId, hex.EncodeToString(encoded[0:32]))
	require.Equal(t, expectedAvsAddr, hex.EncodeToString(encoded[32:64]))
	require.Equal(t, expectedExecAddr, hex.EncodeToString(encoded[64:96]))
	require.Equal(t, expectedOpSetId, hex.EncodeToString(encoded[96:128]))
	require.Equal(t, expectedTimestamp, hex.EncodeToString(encoded[128:160]))
	require.Equal(t, expectedBlockNum, hex.EncodeToString(encoded[160:192]))
	require.Equal(t, expectedPayloadDigest, hex.EncodeToString(encoded[192:224]))
}

func BenchmarkEncodeTaskSubmissionMessage(b *testing.B) {
	taskId := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	avsAddress := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb2"
	executorAddress := "0x5aAeb6053f3E94C9b9A09f33669435E7Ef1BeAed"
	operatorSetId := uint32(42)
	blockNumber := uint64(12345678)
	timestamp := uint32(1234567890)
	payload := []byte("benchmark payload data that could be quite large in production")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = EncodeTaskSubmissionMessageVersioned(taskId, avsAddress, executorAddress, timestamp, blockNumber, operatorSetId, payload, 1)
	}
}

func BenchmarkEncodeTaskResultMessage(b *testing.B) {
	taskId := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	avsAddress := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb2"
	operatorAddress := "0x5aAeb6053f3E94C9b9A09f33669435E7Ef1BeAed"
	operatorSetId := uint32(42)
	output := []byte("benchmark output data that represents task execution results")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = EncodeTaskResultMessageVersioned(taskId, avsAddress, operatorAddress, operatorSetId, output, 1)
	}
}

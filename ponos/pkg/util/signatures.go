package util

import (
	"encoding/binary"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// TaskSubmissionSignatureData represents data signed by aggregator for a specific executor
type TaskSubmissionSignatureData struct {
	TaskId          string   // 32 bytes when encoded
	AvsAddress      string   // 20 bytes padded to 32
	ExecutorAddress string   // 20 bytes padded to 32
	OperatorSetId   uint32   // 4 bytes padded to 32
	BlockNumber     uint64   // 8 bytes padded to 32
	PayloadDigest   [32]byte // 32 bytes (keccak256 of payload)
}

// ToSigningBytes creates deterministic ABI-encoded bytes for signing
// Format: taskId(bytes32) || avsAddress(address) || executorAddress(address) ||
//
//	operatorSetId(uint32) || blockNumber(uint64) || payloadDigest(bytes32)
//
// Total: 192 bytes
func (tsd *TaskSubmissionSignatureData) ToSigningBytes() []byte {
	result := make([]byte, 0, 192)

	// TaskId as 32 bytes
	taskIdBytes := common.HexToHash(tsd.TaskId).Bytes()
	result = append(result, taskIdBytes...)

	// AVS address padded to 32 bytes
	avsAddr := common.HexToAddress(tsd.AvsAddress).Bytes()
	result = append(result, common.LeftPadBytes(avsAddr, 32)...)

	// Executor address padded to 32 bytes
	execAddr := common.HexToAddress(tsd.ExecutorAddress).Bytes()
	result = append(result, common.LeftPadBytes(execAddr, 32)...)

	// OperatorSetId as uint32 padded to 32 bytes
	operSetId := make([]byte, 32)
	binary.BigEndian.PutUint32(operSetId[28:], tsd.OperatorSetId)
	result = append(result, operSetId...)

	// BlockNumber as uint64 padded to 32 bytes
	blockNum := make([]byte, 32)
	binary.BigEndian.PutUint64(blockNum[24:], tsd.BlockNumber)
	result = append(result, blockNum...)

	// Payload digest (already 32 bytes)
	result = append(result, tsd.PayloadDigest[:]...)

	return result
}

// TaskResultSignatureData represents data signed by executor when returning results
type TaskResultSignatureData struct {
	TaskId          string   // 32 bytes when encoded
	AvsAddress      string   // 20 bytes padded to 32
	OperatorAddress string   // 20 bytes padded to 32 (executor's own address)
	OperatorSetId   uint32   // 4 bytes padded to 32
	OutputDigest    [32]byte // 32 bytes (keccak256 of output)
}

// ToSigningBytes creates deterministic ABI-encoded bytes for signing
// Same format as TaskSubmissionSignatureData for consistency
func (trd *TaskResultSignatureData) ToSigningBytes() []byte {
	result := make([]byte, 0, 160)

	// TaskId as 32 bytes
	taskIdBytes := common.HexToHash(trd.TaskId).Bytes()
	result = append(result, taskIdBytes...)

	// AVS address padded to 32 bytes
	avsAddr := common.HexToAddress(trd.AvsAddress).Bytes()
	result = append(result, common.LeftPadBytes(avsAddr, 32)...)

	// Operator address padded to 32 bytes
	operAddr := common.HexToAddress(trd.OperatorAddress).Bytes()
	result = append(result, common.LeftPadBytes(operAddr, 32)...)

	// OperatorSetId as uint32 padded to 32 bytes
	operSetId := make([]byte, 32)
	binary.BigEndian.PutUint32(operSetId[28:], trd.OperatorSetId)
	result = append(result, operSetId...)

	// Output digest (already 32 bytes)
	result = append(result, trd.OutputDigest[:]...)

	return result
}

// EncodeTaskSubmissionMessage creates the message to be signed by aggregator for a specific executor
func EncodeTaskSubmissionMessage(
	taskId string,
	avsAddress string,
	executorAddress string,
	operatorSetId uint32,
	blockNumber uint64,
	payload []byte,
) []byte {

	payloadDigest := crypto.Keccak256Hash(payload)

	sigData := &TaskSubmissionSignatureData{
		TaskId:          taskId,
		AvsAddress:      avsAddress,
		ExecutorAddress: executorAddress,
		OperatorSetId:   operatorSetId,
		BlockNumber:     blockNumber,
		PayloadDigest:   [32]byte(payloadDigest),
	}

	return sigData.ToSigningBytes()
}

// EncodeTaskResultMessage creates the message to be signed by executor when returning results
func EncodeTaskResultMessage(
	taskId string,
	avsAddress string,
	operatorAddress string,
	operatorSetId uint32,
	output []byte,
) []byte {

	outputDigest := crypto.Keccak256Hash(output)

	sigData := &TaskResultSignatureData{
		TaskId:          taskId,
		AvsAddress:      avsAddress,
		OperatorAddress: operatorAddress,
		OperatorSetId:   operatorSetId,
		OutputDigest:    [32]byte(outputDigest),
	}

	return sigData.ToSigningBytes()
}

// AbiEncodeUint32 encodes a uint32 as 32 bytes (ABI standard)
func AbiEncodeUint32(value uint32) []byte {
	result := make([]byte, 32)
	binary.BigEndian.PutUint32(result[28:32], value)
	return result
}

// AbiEncodeAddress encodes an address as 32 bytes (ABI standard)
func AbiEncodeAddress(addr string) []byte {
	address := common.HexToAddress(addr).Bytes()
	return common.LeftPadBytes(address, 32)
}

// AbiEncodeBytes32 encodes a hash/bytes32 value
func AbiEncodeBytes32(value string) []byte {
	return common.HexToHash(value).Bytes()
}

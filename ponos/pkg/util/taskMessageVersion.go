package util

import (
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
)

// Version constants for message encoding
const (
	TaskMessageVersionV1 uint8 = 1
)

// EncodeTaskSubmissionMessageVersioned creates a versioned task submission message
func EncodeTaskSubmissionMessageVersioned(
	taskId string,
	avsAddress string,
	executorAddress string,
	referenceTimestamp uint32,
	blockNumber uint64,
	operatorSetId uint32,
	payload []byte,
	version uint8,
) ([]byte, error) {
	switch version {
	case TaskMessageVersionV1:
		return EncodeTaskSubmissionMessage(taskId, avsAddress, executorAddress, referenceTimestamp, blockNumber, operatorSetId, payload), nil
	default:
		return nil, fmt.Errorf("unsupported task submission message version: %d", version)
	}
}

// EncodeTaskResultMessageVersioned creates a versioned task result message
func EncodeTaskResultMessageVersioned(
	taskId string,
	avsAddress string,
	operatorAddress string,
	operatorSetId uint32,
	output []byte,
	version uint8,
) ([]byte, error) {
	switch version {
	case TaskMessageVersionV1:
		return EncodeTaskResultMessage(taskId, avsAddress, operatorAddress, operatorSetId, output), nil
	default:
		return nil, fmt.Errorf("unsupported task result message version: %d", version)
	}
}

// EncodeTaskSubmissionMessage creates the message to be signed by aggregator for a specific executor
func EncodeTaskSubmissionMessage(
	taskId string,
	avsAddress string,
	executorAddress string,
	referenceTimestamp uint32,
	blockNumber uint64,
	operatorSetId uint32,
	payload []byte,
) []byte {

	payloadDigest := crypto.Keccak256Hash(payload)

	sigData := &TaskSubmissionSignatureData{
		TaskId:          taskId,
		AvsAddress:      avsAddress,
		ExecutorAddress: executorAddress,
		OperatorSetId:   operatorSetId,
		BlockNumber:     blockNumber,
		Timestamp:       referenceTimestamp,
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

package aggregation

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/Layr-Labs/crypto-libs/pkg/ecdsa"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type AggregatedECDSACertificate struct {
	// the unique identifier for the task
	TaskId []byte

	// the output of the task
	TaskResponse []byte

	// keccak256 hash of the task response
	TaskResponseDigest [32]byte

	// signatures of the signers
	// operatorAddress --> signature
	SignersSignatures map[common.Address][]byte

	// the time the certificate was signed
	SignedAt *time.Time
}

func (cert *AggregatedECDSACertificate) GetFinalSignature() ([]byte, error) {
	if len(cert.SignersSignatures) == 0 {
		return nil, fmt.Errorf("no signatures found in certificate")
	}

	// Collect addresses
	addresses := make([]common.Address, 0, len(cert.SignersSignatures))
	for addr := range cert.SignersSignatures {
		addresses = append(addresses, addr)
	}

	// Sort by raw bytes using slices.Compare
	slices.SortFunc(addresses, func(a, b common.Address) int {
		return slices.Compare(a[:], b[:])
	})

	// Concatenate signatures in sorted order
	var finalSignature []byte
	for _, addr := range addresses {
		sig := cert.SignersSignatures[addr]
		if len(sig) != 65 {
			return nil, fmt.Errorf("signature for address %s has invalid length: expected 65, got %d",
				addr.Hex(), len(sig))
		}
		finalSignature = append(finalSignature, sig...)
	}

	return finalSignature, nil
}

func (cert *AggregatedECDSACertificate) GetTaskMessageHash() [32]byte {
	return util.GetKeccak256Digest(cert.TaskResponse)
}

type ReceivedECDSAResponseWithDigest struct {
	// TaskId is the unique identifier for the task
	TaskId string

	// The full task result from the operator
	TaskResult *types.TaskResult

	// signature is the result signature from the operator
	Signature *ecdsa.Signature

	// OutputDigest is a keccak256 hash of the output bytes (for consensus tracking)
	OutputDigest [32]byte
}

// signerInfo holds information about an individual signer
type ecdsaSignerInfo struct {
	publicKey common.Address
	signature []byte
	operator  *Operator[common.Address]
}

// digestGroup tracks all signers for a specific output digest
type ecdsaDigestGroup struct {
	// Operators who signed this specific digest
	signers map[string]*ecdsaSignerInfo // operator address -> info

	// Representative response for this digest
	response *ReceivedECDSAResponseWithDigest

	// Count of signatures for this digest
	count int
}

type aggregatedECDSAOperators struct {
	// Group signatures by the digest they signed
	digestGroups map[[32]byte]*ecdsaDigestGroup

	// Track the most common digest
	mostCommonDigest [32]byte
	mostCommonCount  int
}

type ECDSATaskResultAggregator struct {
	mu                 sync.Mutex
	TaskId             string
	TaskCreatedBlock   uint64
	OperatorSetId      uint32
	ThresholdBips      uint16
	TaskData           []byte
	TaskExpirationTime *time.Time
	Operators          []*Operator[common.Address]
	ReceivedSignatures map[string]*ReceivedECDSAResponseWithDigest // operator address -> signature

	AggregatePublicKeys []common.Address

	aggregatedOperators *aggregatedECDSAOperators
	// Add more fields as needed for aggregation
}

func NewECDSATaskResultAggregator(
	_ context.Context,
	taskId string,
	taskCreatedBlock uint64,
	operatorSetId uint32,
	thresholdBips uint16,
	taskData []byte,
	taskExpirationTime *time.Time,
	operators []*Operator[common.Address],
) (*ECDSATaskResultAggregator, error) {
	if len(taskId) == 0 {
		return nil, ErrInvalidTaskId
	}
	if len(operators) == 0 {
		return nil, ErrNoOperatorAddresses
	}
	if thresholdBips == 0 || thresholdBips > 10_000 {
		return nil, ErrInvalidThreshold
	}

	aggPub := util.Map(operators, func(o *Operator[common.Address], i uint64) common.Address {
		return o.PublicKey
	})

	cert := &ECDSATaskResultAggregator{
		TaskId:              taskId,
		TaskCreatedBlock:    taskCreatedBlock,
		OperatorSetId:       operatorSetId,
		ThresholdBips:       thresholdBips,
		TaskData:            taskData,
		TaskExpirationTime:  taskExpirationTime,
		Operators:           operators,
		AggregatePublicKeys: aggPub,
	}
	return cert, nil
}

func (tra *ECDSATaskResultAggregator) ProcessNewSignature(
	_ context.Context,
	taskResponse *types.TaskResult,
) error {
	tra.mu.Lock()
	defer tra.mu.Unlock()

	// Validate task ID matches the expected task ID for this aggregator
	if tra.TaskId != taskResponse.TaskId {
		return fmt.Errorf("task ID mismatch: expected %s, got %s", tra.TaskId, taskResponse.TaskId)
	}

	// Validate OperatorSetId matches
	if taskResponse.OperatorSetId != tra.OperatorSetId {
		return fmt.Errorf("operator set ID mismatch: expected %d, got %d",
			tra.OperatorSetId, taskResponse.OperatorSetId)
	}

	// Validate operator is in the allowed set
	operator := util.Find(tra.Operators, func(op *Operator[common.Address]) bool {
		return strings.EqualFold(op.Address, taskResponse.OperatorAddress)
	})
	if operator == nil {
		return fmt.Errorf("operator %s is not in the allowed set", taskResponse.OperatorAddress)
	}

	if len(taskResponse.ResultSignature) == 0 {
		return fmt.Errorf("result signature is empty")
	}
	if len(taskResponse.AuthSignature) == 0 {
		return fmt.Errorf("auth signature is empty")
	}

	// Initialize map if nil
	if tra.ReceivedSignatures == nil {
		tra.ReceivedSignatures = make(map[string]*ReceivedECDSAResponseWithDigest)
	}

	// check to see if the operator has already submitted a signature
	if _, ok := tra.ReceivedSignatures[taskResponse.OperatorAddress]; ok {
		return fmt.Errorf("operator %s has already submitted a signature", taskResponse.OperatorAddress)
	}

	// Calculate output digest for consensus tracking
	outputDigest := util.GetKeccak256Digest(taskResponse.Output)

	// Verify both signatures
	sig, err := tra.VerifyResponseSignature(taskResponse, operator, outputDigest)
	if err != nil {
		return fmt.Errorf("failed to verify signatures: %w", err)
	}

	rr := &ReceivedECDSAResponseWithDigest{
		TaskId:       tra.TaskId,
		TaskResult:   taskResponse,
		Signature:    sig,
		OutputDigest: outputDigest,
	}

	tra.ReceivedSignatures[taskResponse.OperatorAddress] = rr

	// Aggregate when generating the final certificate
	if tra.aggregatedOperators == nil {
		tra.aggregatedOperators = &aggregatedECDSAOperators{
			digestGroups: make(map[[32]byte]*ecdsaDigestGroup),
		}
	}

	// Get or create the digest group for this output
	group, exists := tra.aggregatedOperators.digestGroups[outputDigest]
	if !exists {
		group = &ecdsaDigestGroup{
			signers:  make(map[string]*ecdsaSignerInfo),
			response: rr,
			count:    0,
		}
		tra.aggregatedOperators.digestGroups[outputDigest] = group
	}

	// Store signer info for later aggregation
	group.signers[taskResponse.OperatorAddress] = &ecdsaSignerInfo{
		publicKey: operator.PublicKey,
		signature: taskResponse.ResultSignature,
		operator:  operator,
	}
	group.count++

	// Update most common tracking
	if group.count > tra.aggregatedOperators.mostCommonCount {
		tra.aggregatedOperators.mostCommonCount = group.count
		tra.aggregatedOperators.mostCommonDigest = outputDigest
	}

	return nil
}

func (tra *ECDSATaskResultAggregator) SigningThresholdMet() bool {
	// Check if threshold is met (by count)
	required := int((float64(tra.ThresholdBips) / 10_000.0) * float64(len(tra.Operators)))
	if required == 0 {
		required = 1 // Always require at least one
	}
	if tra.aggregatedOperators == nil {
		return false
	}
	// Check if the most common digest has enough signatures
	return tra.aggregatedOperators.mostCommonCount >= required
}

// VerifyResponseSignature verifies both result and auth signatures
func (tra *ECDSATaskResultAggregator) VerifyResponseSignature(
	taskResponse *types.TaskResult,
	operator *Operator[common.Address],
	outputDigest [32]byte,
) (*ecdsa.Signature, error) {
	// Step 1: Verify the result signature (for storage)
	resultSig, err := ecdsa.NewSignatureFromBytes(taskResponse.ResultSignature)
	if err != nil {
		return nil, fmt.Errorf("failed to parse result signature: %w", err)
	}

	// Verify result signature matches output digest
	if verified, err := resultSig.VerifyWithAddress(outputDigest[:], operator.PublicKey); err != nil {
		return nil, fmt.Errorf("result signature verification failed: %w", err)
	} else if !verified {
		return nil, fmt.Errorf("result signature verification failed: signature does not match operator public key")
	}

	// Step 2: Verify the auth signature (for identity)
	authSig, err := ecdsa.NewSignatureFromBytes(taskResponse.AuthSignature)
	if err != nil {
		return nil, fmt.Errorf("failed to parse auth signature: %w", err)
	}

	// Create the auth data that should have been signed
	resultSigDigest := util.GetKeccak256Digest(taskResponse.ResultSignature)
	authData := &types.AuthSignatureData{
		TaskId:          taskResponse.TaskId,
		AvsAddress:      taskResponse.AvsAddress,
		OperatorAddress: taskResponse.OperatorAddress,
		OperatorSetId:   taskResponse.OperatorSetId,
		ResultSigDigest: resultSigDigest,
	}
	authBytes := authData.ToSigningBytes()
	authDigest := util.GetKeccak256Digest(authBytes)

	// Verify auth signature
	if verified, err := authSig.VerifyWithAddress(authDigest[:], operator.PublicKey); err != nil {
		return nil, fmt.Errorf("auth signature verification failed: %w", err)
	} else if !verified {
		return nil, fmt.Errorf("auth signature verification failed: signature does not match operator public key")
	}

	// Additional validation: ensure claimed operator matches expected
	if !strings.EqualFold(taskResponse.OperatorAddress, operator.Address) {
		return nil, fmt.Errorf("operator address mismatch: expected %s, got %s",
			operator.Address, taskResponse.OperatorAddress)
	}

	return resultSig, nil
}

func (tra *ECDSATaskResultAggregator) GenerateFinalCertificate() (*AggregatedECDSACertificate, error) {
	if tra.aggregatedOperators == nil || len(tra.aggregatedOperators.digestGroups) == 0 {
		return nil, fmt.Errorf("no signatures collected")
	}

	// Find the winning digest group
	winningGroup := tra.aggregatedOperators.digestGroups[tra.aggregatedOperators.mostCommonDigest]
	if winningGroup == nil || winningGroup.count == 0 {
		return nil, fmt.Errorf("no signatures for winning digest")
	}

	// Collect only the signatures that signed the winning message
	signersSignatures := make(map[common.Address][]byte)
	for _, signer := range winningGroup.signers {
		signersSignatures[signer.operator.GetAddress()] = signer.signature
	}

	taskIdBytes, err := hexutil.Decode(tra.TaskId)
	if err != nil {
		return nil, fmt.Errorf("failed to decode taskId: %w", err)
	}

	return &AggregatedECDSACertificate{
		TaskId:             taskIdBytes,
		TaskResponse:       winningGroup.response.TaskResult.Output,
		TaskResponseDigest: winningGroup.response.OutputDigest,
		SignersSignatures:  signersSignatures,
		SignedAt:           new(time.Time),
	}, nil
}

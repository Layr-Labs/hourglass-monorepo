package aggregation

import (
	"context"
	"fmt"
	"slices"
	"sort"
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

	// public keys for all operators that did not sign the task
	NonSignersPubKeys []common.Address

	// public keys for all operators that were selected to participate in the task
	AllOperatorsPubKeys []common.Address

	// aggregated signature of the signers
	// operatorAddress --> signature
	SignersSignatures map[common.Address][]byte

	// aggregated public key of the signers
	SignersPublicKeys []common.Address

	// the time the certificate was signed
	SignedAt *time.Time
}

func (cert *AggregatedECDSACertificate) GetFinalSignature() ([]byte, error) {
	// Extract and sort addresses
	addresses := make([]common.Address, 0, len(cert.SignersSignatures))
	for addr := range cert.SignersSignatures {
		addresses = append(addresses, addr)
	}
	sort.Slice(addresses, func(i, j int) bool {
		return addresses[i].Hex() < addresses[j].Hex()
	})

	// Concatenate in sorted order
	var finalSignature []byte
	for _, addr := range addresses {
		finalSignature = append(finalSignature,
			cert.SignersSignatures[addr]...)
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

type aggregatedECDSAOperators struct {
	// aggregated public keys of signers
	signersPublicKeys []common.Address

	// aggregated signatures of signers (not really aggregated, but a map)
	// operatorAddress --> signature
	signersSignatures map[common.Address][]byte

	// operators that have signed (operatorAddress --> true)
	// signersOperatorSet map[string]bool

	// simple count of signers. eventually this could represent stake weight or something
	totalSigners int

	// Track digest counts to find most common result
	// digest -> count
	digestCounts map[[32]byte]int

	// Track which response to use for each digest
	// digest -> response
	digestResponses map[[32]byte]*ReceivedECDSAResponseWithDigest

	// The response with the most common digest (updated as we aggregate)
	mostCommonResponse *ReceivedECDSAResponseWithDigest

	// The response with the most common digest (updated as we aggregate)
	mostCommonCount int
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
	ctx context.Context,
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
	ctx context.Context,
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

	// Begin aggregating signatures and public keys.
	// Track digest counts to select the most common result for the final certificate
	if tra.aggregatedOperators == nil {
		// no signers yet, initialize the aggregated operators
		tra.aggregatedOperators = &aggregatedECDSAOperators{
			signersPublicKeys: []common.Address{operator.PublicKey},

			signersSignatures: map[common.Address][]byte{
				operator.GetAddress(): taskResponse.ResultSignature,
			},

			// initialize the map of signers (operatorAddress --> true) to track who actually signed
			// signersOperatorSet: map[string]bool{taskResponse.OperatorAddress: true},

			// initialize the count of signers (could eventually be weight or something else)
			totalSigners: 1,

			// Initialize digest tracking maps
			digestCounts:       make(map[[32]byte]int),
			digestResponses:    make(map[[32]byte]*ReceivedECDSAResponseWithDigest),
			mostCommonResponse: rr,
			mostCommonCount:    1,
		}
		// Track this digest for the first operator
		tra.aggregatedOperators.digestCounts[outputDigest] = 1
		tra.aggregatedOperators.digestResponses[outputDigest] = rr
		tra.aggregatedOperators.mostCommonResponse = rr
	} else {
		tra.aggregatedOperators.signersSignatures[operator.GetAddress()] = taskResponse.ResultSignature
		tra.aggregatedOperators.signersPublicKeys = append(tra.aggregatedOperators.signersPublicKeys, operator.PublicKey)
		tra.aggregatedOperators.totalSigners++

		// Update digest count using output digest for consensus tracking
		newCount := tra.aggregatedOperators.digestCounts[outputDigest] + 1
		tra.aggregatedOperators.digestCounts[outputDigest] = newCount

		// Store the first response for this digest (if not already stored)
		if _, exists := tra.aggregatedOperators.digestResponses[outputDigest]; !exists {
			tra.aggregatedOperators.digestResponses[outputDigest] = rr
		}

		// Check if this digest is now the most common
		if newCount > tra.aggregatedOperators.mostCommonCount {
			tra.aggregatedOperators.mostCommonCount = newCount
			tra.aggregatedOperators.mostCommonResponse = tra.aggregatedOperators.digestResponses[outputDigest]
		}
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
	return tra.aggregatedOperators.totalSigners >= required
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
	operatorAddresses := make([]common.Address, 0)
	for addr := range tra.aggregatedOperators.signersSignatures {
		operatorAddresses = append(operatorAddresses, addr)
	}
	// sort the operator addresses to ensure deterministic order for the certificate verifier
	sort.Slice(operatorAddresses, func(i, j int) bool {
		return operatorAddresses[i].Hex() < operatorAddresses[j].Hex()
	})

	nonSignerOperators := make([]*Operator[common.Address], 0)
	for _, operator := range tra.Operators {
		if !slices.Contains(operatorAddresses, operator.GetAddress()) {
			// if the operator is not in the signers set, add them to the non-signer list
			nonSignerOperators = append(nonSignerOperators, operator)
			continue
		}
	}

	nonSignerPublicKeys := util.Map(nonSignerOperators, func(op *Operator[common.Address], i uint64) common.Address {
		return op.PublicKey
	})

	allOperatorAddresses := util.Map(tra.Operators, func(op *Operator[common.Address], i uint64) common.Address {
		return op.GetAddress()
	})
	sort.Slice(allOperatorAddresses, func(i, j int) bool {
		return allOperatorAddresses[i].Hex() < allOperatorAddresses[j].Hex()
	})

	allPublicKeys := []common.Address{}
	for _, opAddr := range allOperatorAddresses {
		op := util.Find(tra.Operators, func(o *Operator[common.Address]) bool {
			return strings.EqualFold(o.Address, opAddr.String())
		})
		if op != nil {
			allPublicKeys = append(allPublicKeys, op.PublicKey)
		}
	}

	taskIdBytes, err := hexutil.Decode(tra.TaskId)
	if err != nil {
		return nil, fmt.Errorf("failed to decode taskId: %w", err)
	}

	// Use the most common response for the certificate
	if tra.aggregatedOperators.mostCommonResponse == nil {
		return nil, fmt.Errorf("no common response found")
	}

	return &AggregatedECDSACertificate{
		TaskId:              taskIdBytes,
		TaskResponse:        tra.aggregatedOperators.mostCommonResponse.TaskResult.Output,
		TaskResponseDigest:  tra.aggregatedOperators.mostCommonResponse.OutputDigest,
		NonSignersPubKeys:   nonSignerPublicKeys,
		AllOperatorsPubKeys: allPublicKeys,
		SignersPublicKeys:   tra.aggregatedOperators.signersPublicKeys,
		SignersSignatures:   tra.aggregatedOperators.signersSignatures,
		SignedAt:            new(time.Time),
	}, nil
}

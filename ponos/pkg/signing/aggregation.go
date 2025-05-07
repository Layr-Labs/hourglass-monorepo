package signing

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/bn254"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"math/big"
	"slices"
	"strings"
	"sync"
	"time"
)

type Operator struct {
	Address   string
	PublicKey *bn254.PublicKey
}

// InitializeNewTaskWithWindow initializes a new aggregation certificate for a task window.
// All required data must be provided as arguments; no network or chain calls are performed.
func InitializeNewTaskWithWindow(
	ctx context.Context,
	taskId []byte,
	taskCreatedBlock uint32,
	operatorSetId uint32,
	thresholdPercentage uint8,
	taskData []byte,
	timeToExpiry time.Duration,
	operators []*Operator,
) (*AggregationCertificate, error) {
	if len(taskId) == 0 {
		return nil, ErrInvalidTaskId
	}
	if len(operators) == 0 {
		return nil, ErrNoOperatorAddresses
	}
	if thresholdPercentage == 0 || thresholdPercentage > 100 {
		return nil, ErrInvalidThreshold
	}
	aggPub, err := AggregatePublicKeys(util.Map(operators, func(o *Operator, i uint64) *bn254.PublicKey {
		return o.PublicKey
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate public keys: %w", err)
	}

	cert := &AggregationCertificate{
		TaskId:              taskId,
		TaskCreatedBlock:    taskCreatedBlock,
		OperatorSetId:       operatorSetId,
		ThresholdPercentage: thresholdPercentage,
		TaskData:            taskData,
		TimeToExpiry:        timeToExpiry,
		Operators:           operators,
		AggregatePublicKey:  aggPub,
	}
	return cert, nil
}

// Error variables for input validation
var (
	ErrInvalidTaskId       = fmt.Errorf("taskId must not be empty")
	ErrNoOperatorAddresses = fmt.Errorf("operatorAddresses must not be empty")
	ErrInvalidThreshold    = fmt.Errorf("thresholdPercentage must be between 1 and 100")
)

// AggregationCertificate represents the data needed to initialize a new aggregation task window.
type AggregationCertificate struct {
	mu                  sync.Mutex
	TaskId              []byte
	TaskCreatedBlock    uint32
	OperatorSetId       uint32
	ThresholdPercentage uint8
	TaskData            []byte
	TimeToExpiry        time.Duration
	Operators           []*Operator
	ReceivedSignatures  map[string]*ReceivedResponse // operator address -> signature
	AggregatePublicKey  *bn254.PublicKey

	aggregatedOperators *aggregatedOperators
	// Add more fields as needed for aggregation
}

type ReceivedResponse struct {
	TaskId     []byte
	TaskResult *types.TaskResult
	Signature  *bn254.Signature
}

type aggregatedOperators struct {
	// aggregated public keys of signers
	signersG2 *bn254.G1Point

	// aggregated signatures of signers
	signersAggSig *bn254.Signature

	// operators that have signed (operatorAddress --> true)
	signersOperatorSet map[string]bool

	// simple count of signers. eventually this could represent stake weight or something
	totalSigners int
}

func (ac *AggregationCertificate) SigningThresholdMet() bool {
	// Check if threshold is met (by count)
	required := int((float64(ac.ThresholdPercentage) / 100.0) * float64(len(ac.Operators)))
	if required == 0 {
		required = 1 // Always require at least one
	}
	return len(ac.ReceivedSignatures) >= required
}

// ProcessNewSignature processes a new signature submission from an operator.
// Returns true if the threshold is met after this submission, false otherwise.
func (ac *AggregationCertificate) ProcessNewSignature(
	ctx context.Context,
	taskId []byte,
	taskResponse *types.TaskResult,
) error {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	// Validate operator is in the allowed set
	foundOp := util.Find(ac.Operators, func(op *Operator) bool {
		return op.Address == taskResponse.OperatorAddress
	})
	if foundOp == nil {
		return fmt.Errorf("operator %s is not in the allowed set", taskResponse.OperatorAddress)
	}

	if len(taskResponse.Signature) == 0 {
		return fmt.Errorf("signature is empty")
	}

	// Initialize map if nil
	if ac.ReceivedSignatures == nil {
		ac.ReceivedSignatures = make(map[string]*ReceivedResponse)
	}

	// check to see if the operator has already submitted a signature
	if _, ok := ac.ReceivedSignatures[taskResponse.OperatorAddress]; ok {
		return fmt.Errorf("operator %s has already submitted a signature", taskResponse.OperatorAddress)
	}

	// verify the signature
	sig, err := ac.VerifyResponseSignature(taskResponse)
	if err != nil {
		return fmt.Errorf("failed to verify signature: %w", err)
	}

	rr := &ReceivedResponse{
		TaskId:     taskId,
		TaskResult: taskResponse,
		Signature:  sig,
	}

	// Store the signature
	ac.ReceivedSignatures[taskResponse.OperatorAddress] = rr

	// verify signature
	if ac.aggregatedOperators == nil {
		// no signers yet, initialize the aggregated operators
		ac.aggregatedOperators = &aggregatedOperators{
			// operator's public key
			signersG2: bn254.NewZeroG1Point().Add(bn254.NewG1Point(sig.Sig.X.BigInt(&big.Int{}), sig.Sig.Y.BigInt(&big.Int{}))),

			signersAggSig: sig,

			signersOperatorSet: map[string]bool{taskResponse.OperatorAddress: true},

			totalSigners: 1,
		}
	} else {
		ac.aggregatedOperators.signersG2.Add(bn254.NewG1Point(sig.Sig.X.BigInt(&big.Int{}), sig.Sig.Y.BigInt(&big.Int{})))
	}

	return nil
}

// VerifyResponseSignature verifies that the signature of the response is valid against
// the operators public key.
func (ac *AggregationCertificate) VerifyResponseSignature(taskResponse *types.TaskResult) (*bn254.Signature, error) {
	digestBytes := util.GetKeccak256Digest(taskResponse.Output)
	sig, err := bn254.NewSignatureFromBytes(taskResponse.Signature)
	if err != nil {
		return nil, fmt.Errorf("failed to create signature from bytes: %w", err)
	}

	operator := util.Find(ac.Operators, func(op *Operator) bool {
		return strings.EqualFold(op.Address, taskResponse.OperatorAddress)
	})
	if operator == nil {
		return nil, fmt.Errorf("operator %s not found in allowed set", taskResponse.OperatorAddress)
	}
	if verified, err := sig.Verify(operator.PublicKey, digestBytes[:]); err != nil {
		return nil, fmt.Errorf("signature verification failed: %w", err)
	} else if !verified {
		return nil, fmt.Errorf("signature verification failed: signature does not match operator public key")
	}
	return sig, nil
}

func AggregatePublicKeys(pubKeys []*bn254.PublicKey) (*bn254.PublicKey, error) {
	return bn254.AggregatePublicKeys(pubKeys)
}

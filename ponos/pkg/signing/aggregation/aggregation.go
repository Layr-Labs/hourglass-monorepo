package aggregation

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/bn254"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
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
) (*TaskResultAggregator, error) {
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

	cert := &TaskResultAggregator{
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

type AggregatedCertificate struct {
	// the unique identifier for the task
	TaskId []byte

	// the output of the task
	TaskResponse []byte

	// keccak256 hash of the task response
	TaskResponseDigest []byte

	// public keys for all operators that did not sign the task
	NonSignersPubKeys []*bn254.PublicKey

	// public keys for all operators that were selected to participate in the task
	AllOperatorsPubKeys []*bn254.PublicKey

	// aggregated signature of the signers
	SignersSignature *bn254.Signature

	// aggregated public key of the signers
	SignersPublicKey *bn254.G2Point
}

// GenerateFinalCertificate generates the final aggregated certificate for the task.
func (ac *TaskResultAggregator) GenerateFinalCertificate() (*AggregatedCertificate, error) {
	// TODO(seanmcgary): nonSignerOperatorIds should be a list of operatorIds which is the hash of their public key
	nonSignerOperatorIds := make([]*Operator, 0)
	for _, operator := range ac.Operators {
		if _, ok := ac.aggregatedOperators.signersOperatorSet[operator.Address]; !ok {
			nonSignerOperatorIds = append(nonSignerOperatorIds, operator)
		}
	}

	// TODO: add this based on the avs registry
	// the contract requires a sorted nonSignersOperatorIds
	// sort.SliceStable(nonSignerOperatorIds, func(i, j int) bool {
	// 	iOprInt := new(big.Int).SetBytes(nonSignerOperatorIds[i][:])
	// 	jOprInt := new(big.Int).SetBytes(nonSignerOperatorIds[j][:])
	// 	return iOprInt.Cmp(jOprInt) == -1
	// })

	nonSignerPublicKeys := make([]*bn254.PublicKey, 0)
	for _, operatorId := range nonSignerOperatorIds {
		operator := util.Find(ac.Operators, func(op *Operator) bool {
			return strings.EqualFold(op.Address, operatorId.Address)
		})
		nonSignerPublicKeys = append(nonSignerPublicKeys, operator.PublicKey)
	}

	allPublicKeys := util.Map(ac.Operators, func(o *Operator, i uint64) *bn254.PublicKey {
		return o.PublicKey
	})

	return &AggregatedCertificate{
		TaskId:              ac.TaskId,
		TaskResponse:        ac.aggregatedOperators.lastReceivedResponse.TaskResult.Output,
		TaskResponseDigest:  ac.aggregatedOperators.lastReceivedResponse.Digest,
		NonSignersPubKeys:   nonSignerPublicKeys,
		AllOperatorsPubKeys: allPublicKeys,
		SignersPublicKey:    ac.aggregatedOperators.signersG2,
		SignersSignature:    ac.aggregatedOperators.signersAggSig,
	}, nil
}

// TaskResultAggregator represents the data needed to initialize a new aggregation task window.
type TaskResultAggregator struct {
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
	// TaskId is the unique identifier for the task
	TaskId []byte

	// The full task result from the operator
	TaskResult *types.TaskResult

	// signature is the signature of the task result from the operator signed with their bls key
	Signature *bn254.Signature

	// digest is a keccak256 hash of the task result
	Digest []byte
}

type aggregatedOperators struct {
	// aggregated public keys of signers
	signersG2 *bn254.G2Point

	// aggregated signatures of signers
	signersAggSig *bn254.Signature

	// operators that have signed (operatorAddress --> true)
	signersOperatorSet map[string]bool

	// simple count of signers. eventually this could represent stake weight or something
	totalSigners int

	lastReceivedResponse *ReceivedResponse
}

func (ac *TaskResultAggregator) SigningThresholdMet() bool {
	// Check if threshold is met (by count)
	required := int((float64(ac.ThresholdPercentage) / 100.0) * float64(len(ac.Operators)))
	if required == 0 {
		required = 1 // Always require at least one
	}
	return ac.aggregatedOperators.totalSigners >= required
}

// ProcessNewSignature processes a new signature submission from an operator.
// Returns true if the threshold is met after this submission, false otherwise.
func (ac *TaskResultAggregator) ProcessNewSignature(
	ctx context.Context,
	taskId []byte,
	taskResponse *types.TaskResult,
) error {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	// Validate operator is in the allowed set
	operator := util.Find(ac.Operators, func(op *Operator) bool {
		return op.Address == taskResponse.OperatorAddress
	})
	if operator == nil {
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
	sig, digest, err := ac.VerifyResponseSignature(taskResponse, operator)
	if err != nil {
		return fmt.Errorf("failed to verify signature: %w", err)
	}

	rr := &ReceivedResponse{
		TaskId:     taskId,
		TaskResult: taskResponse,
		Signature:  sig,
		Digest:     digest,
	}

	// Store the signature
	ac.ReceivedSignatures[taskResponse.OperatorAddress] = rr

	// verify signature
	if ac.aggregatedOperators == nil {
		// no signers yet, initialize the aggregated operators
		ac.aggregatedOperators = &aggregatedOperators{
			// operator's public key
			signersG2: bn254.NewZeroG2Point().AddPublicKey(operator.PublicKey),

			signersAggSig: sig,

			signersOperatorSet: map[string]bool{taskResponse.OperatorAddress: true},

			totalSigners: 1,

			lastReceivedResponse: rr,
		}
	} else {
		ac.aggregatedOperators.signersG2.AddPublicKey(operator.PublicKey)
		ac.aggregatedOperators.signersAggSig.Add(sig)
		ac.aggregatedOperators.signersOperatorSet[taskResponse.OperatorAddress] = true
		ac.aggregatedOperators.totalSigners++
		ac.aggregatedOperators.lastReceivedResponse = rr
	}

	return nil
}

// VerifyResponseSignature verifies that the signature of the response is valid against
// the operators public key.
func (ac *TaskResultAggregator) VerifyResponseSignature(taskResponse *types.TaskResult, operator *Operator) (*bn254.Signature, []byte, error) {
	digestBytes := util.GetKeccak256Digest(taskResponse.Output)
	sig, err := bn254.NewSignatureFromBytes(taskResponse.Signature)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create signature from bytes: %w", err)
	}

	if verified, err := sig.Verify(operator.PublicKey, digestBytes[:]); err != nil {
		return nil, nil, fmt.Errorf("signature verification failed: %w", err)
	} else if !verified {
		return nil, nil, fmt.Errorf("signature verification failed: signature does not match operator public key")
	}
	return sig, digestBytes[:], nil
}

// AggregatePublicKeys aggregates a list of public keys into a single public key.
func AggregatePublicKeys(pubKeys []*bn254.PublicKey) (*bn254.PublicKey, error) {
	return bn254.AggregatePublicKeys(pubKeys)
}

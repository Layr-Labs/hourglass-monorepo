package aggregation

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
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

	// the time the certificate was signed
	SignedAt *time.Time
}

// TaskResultAggregator represents the data needed to initialize a new aggregation task window.
type TaskResultAggregator struct {
	mu                  sync.Mutex
	TaskId              string
	TaskCreatedBlock    uint64
	OperatorSetId       uint32
	ThresholdPercentage uint8
	TaskData            []byte
	TaskExpirationTime  *time.Time
	Operators           []*Operator
	ReceivedSignatures  map[string]*ReceivedResponseWithDigest // operator address -> signature
	AggregatePublicKey  *bn254.PublicKey

	aggregatedOperators *aggregatedOperators
	// Add more fields as needed for aggregation
}

// NewTaskResultAggregator initializes a new aggregation certificate for a task window.
// All required data must be provided as arguments; no network or chain calls are performed.
func NewTaskResultAggregator(
	ctx context.Context,
	taskId string,
	taskCreatedBlock uint64,
	operatorSetId uint32,
	thresholdPercentage uint8,
	taskData []byte,
	taskExpirationTime *time.Time,
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
		TaskExpirationTime:  taskExpirationTime,
		Operators:           operators,
		AggregatePublicKey:  aggPub,
	}
	return cert, nil
}

type ReceivedResponseWithDigest struct {
	// TaskId is the unique identifier for the task
	TaskId string

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

	lastReceivedResponse *ReceivedResponseWithDigest
}

func (tra *TaskResultAggregator) SigningThresholdMet() bool {
	// Check if threshold is met (by count)
	required := int((float64(tra.ThresholdPercentage) / 100.0) * float64(len(tra.Operators)))
	if required == 0 {
		required = 1 // Always require at least one
	}
	return tra.aggregatedOperators.totalSigners >= required
}

// ProcessNewSignature processes a new signature submission from an operator.
// Returns true if the threshold is met after this submission, false otherwise.
func (tra *TaskResultAggregator) ProcessNewSignature(
	ctx context.Context,
	taskId string,
	taskResponse *types.TaskResult,
) error {
	tra.mu.Lock()
	defer tra.mu.Unlock()

	// Validate operator is in the allowed set
	operator := util.Find(tra.Operators, func(op *Operator) bool {
		return op.Address == taskResponse.OperatorAddress
	})
	if operator == nil {
		return fmt.Errorf("operator %s is not in the allowed set", taskResponse.OperatorAddress)
	}

	if len(taskResponse.Signature) == 0 {
		return fmt.Errorf("signature is empty")
	}

	// Initialize map if nil
	if tra.ReceivedSignatures == nil {
		tra.ReceivedSignatures = make(map[string]*ReceivedResponseWithDigest)
	}

	// check to see if the operator has already submitted a signature
	if _, ok := tra.ReceivedSignatures[taskResponse.OperatorAddress]; ok {
		return fmt.Errorf("operator %s has already submitted a signature", taskResponse.OperatorAddress)
	}

	// verify the signature
	sig, digest, err := tra.VerifyResponseSignature(taskResponse, operator)
	if err != nil {
		return fmt.Errorf("failed to verify signature: %w", err)
	}

	rr := &ReceivedResponseWithDigest{
		TaskId:     taskId,
		TaskResult: taskResponse,
		Signature:  sig,
		Digest:     digest,
	}

	tra.ReceivedSignatures[taskResponse.OperatorAddress] = rr

	// Begin aggregating signatures and public keys.
	// The lastReceivedResponse will end up being the value used to for the final certificate.
	//
	// TODO: probably need some kind of comparison on results, otherwise the last operator in
	// will always be the one that is used for the final certificate and could potentially be
	// wrong or malicious.
	if tra.aggregatedOperators == nil {
		// no signers yet, initialize the aggregated operators
		tra.aggregatedOperators = &aggregatedOperators{
			// operator's public key to start an aggregated public key
			signersG2: bn254.NewZeroG2Point().AddPublicKey(operator.PublicKey),

			// signature of the task result payload
			signersAggSig: sig,

			// initialize the map of signers (operatorAddress --> true) to track who actually signed
			signersOperatorSet: map[string]bool{taskResponse.OperatorAddress: true},

			// initialize the count of signers (could eventually be weight or something else)
			totalSigners: 1,

			// store the last received response
			lastReceivedResponse: rr,
		}
	} else {
		tra.aggregatedOperators.signersG2.AddPublicKey(operator.PublicKey)
		tra.aggregatedOperators.signersAggSig.Add(sig)
		tra.aggregatedOperators.signersOperatorSet[taskResponse.OperatorAddress] = true
		tra.aggregatedOperators.totalSigners++
		tra.aggregatedOperators.lastReceivedResponse = rr
	}

	return nil
}

// VerifyResponseSignature verifies that the signature of the response is valid against
// the operators public key.
func (tra *TaskResultAggregator) VerifyResponseSignature(taskResponse *types.TaskResult, operator *Operator) (*bn254.Signature, []byte, error) {
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

// GenerateFinalCertificate generates the final aggregated certificate for the task.
func (tra *TaskResultAggregator) GenerateFinalCertificate() (*AggregatedCertificate, error) {
	// TODO(seanmcgary): nonSignerOperatorIds should be a list of operatorIds which is the hash of their public key
	nonSignerOperatorIds := make([]*Operator, 0)
	for _, operator := range tra.Operators {
		if _, ok := tra.aggregatedOperators.signersOperatorSet[operator.Address]; !ok {
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
		operator := util.Find(tra.Operators, func(op *Operator) bool {
			return strings.EqualFold(op.Address, operatorId.Address)
		})
		nonSignerPublicKeys = append(nonSignerPublicKeys, operator.PublicKey)
	}

	allPublicKeys := util.Map(tra.Operators, func(o *Operator, i uint64) *bn254.PublicKey {
		return o.PublicKey
	})

	taskIdBytes, err := hexutil.Decode(tra.TaskId)
	if err != nil {
		return nil, fmt.Errorf("failed to decode taskId: %w", err)
	}

	return &AggregatedCertificate{
		TaskId:              taskIdBytes,
		TaskResponse:        tra.aggregatedOperators.lastReceivedResponse.TaskResult.Output,
		TaskResponseDigest:  tra.aggregatedOperators.lastReceivedResponse.Digest,
		NonSignersPubKeys:   nonSignerPublicKeys,
		AllOperatorsPubKeys: allPublicKeys,
		SignersPublicKey:    tra.aggregatedOperators.signersG2,
		SignersSignature:    tra.aggregatedOperators.signersAggSig,
		SignedAt:            new(time.Time),
	}, nil
}

// AggregatePublicKeys aggregates a list of public keys into a single public key.
func AggregatePublicKeys(pubKeys []*bn254.PublicKey) (*bn254.PublicKey, error) {
	return bn254.AggregatePublicKeys(pubKeys)
}

package aggregation

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	"github.com/Layr-Labs/crypto-libs/pkg/signing"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"strings"
	"sync"
	"time"
)

type AggregatedBN254Certificate struct {
	// the unique identifier for the task
	TaskId []byte

	// the output of the task
	TaskResponse []byte

	// keccak256 hash of the task response
	TaskResponseDigest []byte

	// public keys for all operators that did not sign the task
	NonSignersPubKeys []signing.PublicKey

	// public keys for all operators that were selected to participate in the task
	AllOperatorsPubKeys []signing.PublicKey

	// aggregated signature of the signers
	SignersSignature *bn254.Signature

	// aggregated public key of the signers
	SignersPublicKey *bn254.G2Point

	// the time the certificate was signed
	SignedAt *time.Time
}

// BN254TaskResultAggregator represents the data needed to initialize a new aggregation task window.
type BN254TaskResultAggregator struct {
	mu                 sync.Mutex
	TaskId             string
	TaskCreatedBlock   uint64
	OperatorSetId      uint32
	ThresholdBips      uint16
	TaskData           []byte
	TaskExpirationTime *time.Time
	Operators          []*Operator[signing.PublicKey]
	ReceivedSignatures map[string]*ReceivedBN254ResponseWithDigest // operator address -> signature
	AggregatePublicKey signing.PublicKey

	aggregatedOperators *aggregatedBN254Operators
	// Add more fields as needed for aggregation
}

// NewBN254TaskResultAggregator initializes a new aggregation certificate for a task window.
// All required data must be provided as arguments; no network or chain calls are performed.
func NewBN254TaskResultAggregator(
	ctx context.Context,
	taskId string,
	taskCreatedBlock uint64,
	operatorSetId uint32,
	thresholdBips uint16,
	taskData []byte,
	taskExpirationTime *time.Time,
	operators []*Operator[signing.PublicKey],
) (*BN254TaskResultAggregator, error) {
	if len(taskId) == 0 {
		return nil, ErrInvalidTaskId
	}
	if len(operators) == 0 {
		return nil, ErrNoOperatorAddresses
	}
	if thresholdBips == 0 || thresholdBips > 10_000 {
		return nil, ErrInvalidThreshold
	}

	aggPub, err := AggregatePublicKeys(util.Map(operators, func(o *Operator[signing.PublicKey], i uint64) signing.PublicKey {
		return o.PublicKey
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate public keys: %w", err)
	}

	cert := &BN254TaskResultAggregator{
		TaskId:             taskId,
		TaskCreatedBlock:   taskCreatedBlock,
		OperatorSetId:      operatorSetId,
		ThresholdBips:      thresholdBips,
		TaskData:           taskData,
		TaskExpirationTime: taskExpirationTime,
		Operators:          operators,
		AggregatePublicKey: aggPub,
	}
	return cert, nil
}

type ReceivedBN254ResponseWithDigest struct {
	// TaskId is the unique identifier for the task
	TaskId string

	// The full task result from the operator
	TaskResult *types.TaskResult

	// signature is the signature of the task result from the operator signed with their bls key
	Signature *bn254.Signature

	// digest is a keccak256 hash of the task result
	Digest []byte
}

type aggregatedBN254Operators struct {
	// aggregated public keys of signers
	signersG2 *bn254.G2Point

	// aggregated signatures of signers
	signersAggSig *bn254.Signature

	// operators that have signed (operatorAddress --> true)
	signersOperatorSet map[string]bool

	// simple count of signers. eventually this could represent stake weight or something
	totalSigners int

	// Track digest counts to find most common result
	// digest -> count
	digestCounts map[[32]byte]int

	// Track which response to use for each digest
	// digest -> response
	digestResponses map[[32]byte]*ReceivedBN254ResponseWithDigest

	// The response with the most common digest (updated as we aggregate)
	mostCommonResponse *ReceivedBN254ResponseWithDigest

	// The response with the most common digest (updated as we aggregate)
	mostCommonCount int
}

func (tra *BN254TaskResultAggregator) SigningThresholdMet() bool {
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

// ProcessNewSignature processes a new signature submission from an operator.
// Returns true if the threshold is met after this submission, false otherwise.
func (tra *BN254TaskResultAggregator) ProcessNewSignature(
	ctx context.Context,
	taskId string,
	taskResponse *types.TaskResult,
) error {
	tra.mu.Lock()
	defer tra.mu.Unlock()

	// Validate operator is in the allowed set
	operator := util.Find(tra.Operators, func(op *Operator[signing.PublicKey]) bool {
		return strings.EqualFold(op.Address, taskResponse.OperatorAddress)
	})
	if operator == nil {
		return fmt.Errorf("operator %s is not in the allowed set", taskResponse.OperatorAddress)
	}

	if len(taskResponse.Signature) == 0 {
		return fmt.Errorf("signature is empty")
	}

	// Initialize map if nil
	if tra.ReceivedSignatures == nil {
		tra.ReceivedSignatures = make(map[string]*ReceivedBN254ResponseWithDigest)
	}

	// check to see if the operator has already submitted a signature
	if _, ok := tra.ReceivedSignatures[taskResponse.OperatorAddress]; ok {
		return fmt.Errorf("operator %s has already submitted a signature", taskResponse.OperatorAddress)
	}

	// verify the signature
	sig, err := tra.VerifyResponseSignature(taskResponse, operator)
	if err != nil {
		return fmt.Errorf("failed to verify signature: %w", err)
	}

	var digest [32]byte
	copy(digest[:], taskResponse.OutputDigest)

	rr := &ReceivedBN254ResponseWithDigest{
		TaskId:     taskId,
		TaskResult: taskResponse,
		Signature:  sig,
		Digest:     taskResponse.OutputDigest,
	}

	tra.ReceivedSignatures[taskResponse.OperatorAddress] = rr

	bn254PubKey, err := bn254.NewPublicKeyFromBytes(operator.PublicKey.Bytes())
	if err != nil {
		return fmt.Errorf("failed to create public key from bytes: %w", err)
	}

	// Begin aggregating signatures and public keys.
	// Track digest counts to select the most common result for the final certificate
	if tra.aggregatedOperators == nil {
		// no signers yet, initialize the aggregated operators
		tra.aggregatedOperators = &aggregatedBN254Operators{
			// operator's public key to start an aggregated public key
			signersG2: bn254.NewZeroG2Point().AddPublicKey(bn254PubKey),

			// signature of the task result payload
			signersAggSig: sig,

			// initialize the map of signers (operatorAddress --> true) to track who actually signed
			signersOperatorSet: map[string]bool{taskResponse.OperatorAddress: true},

			// initialize the count of signers (could eventually be weight or something else)
			totalSigners: 1,

			// Initialize digest tracking maps
			digestCounts:       make(map[[32]byte]int),
			digestResponses:    make(map[[32]byte]*ReceivedBN254ResponseWithDigest),
			mostCommonResponse: rr,
			mostCommonCount:    1,
		}
		// Track this digest for the first operator
		tra.aggregatedOperators.digestCounts[digest] = 1
		tra.aggregatedOperators.digestResponses[digest] = rr
		tra.aggregatedOperators.mostCommonResponse = rr
	} else {
		tra.aggregatedOperators.signersG2.AddPublicKey(bn254PubKey)
		tra.aggregatedOperators.signersAggSig.Add(sig)
		tra.aggregatedOperators.signersOperatorSet[taskResponse.OperatorAddress] = true
		tra.aggregatedOperators.totalSigners++

		// Update digest count
		newCount := tra.aggregatedOperators.digestCounts[digest] + 1
		tra.aggregatedOperators.digestCounts[digest] = newCount

		// Store the first response for this digest (if not already stored)
		if _, exists := tra.aggregatedOperators.digestResponses[digest]; !exists {
			tra.aggregatedOperators.digestResponses[digest] = rr
		}

		// Check if this digest is now the most common
		if newCount > tra.aggregatedOperators.mostCommonCount {
			tra.aggregatedOperators.mostCommonCount = newCount
			tra.aggregatedOperators.mostCommonResponse = tra.aggregatedOperators.digestResponses[digest]
		}
	}

	return nil
}

// VerifyResponseSignature verifies that the signature of the response is valid against
// the operators public key.
func (tra *BN254TaskResultAggregator) VerifyResponseSignature(taskResponse *types.TaskResult, operator *Operator[signing.PublicKey]) (*bn254.Signature, error) {
	sig, err := bn254.NewSignatureFromBytes(taskResponse.Signature)
	if err != nil {
		return nil, fmt.Errorf("failed to create signature from bytes: %w", err)
	}

	var digest [32]byte
	copy(digest[:], taskResponse.OutputDigest)

	if verified, err := sig.VerifySolidityCompatible(operator.PublicKey.(*bn254.PublicKey), digest); err != nil {
		return nil, fmt.Errorf("signature verification failed: %w", err)
	} else if !verified {
		return nil, fmt.Errorf("signature verification failed: signature does not match operator public key")
	}
	return sig, nil
}

// GenerateFinalCertificate generates the final aggregated certificate for the task.
func (tra *BN254TaskResultAggregator) GenerateFinalCertificate() (*AggregatedBN254Certificate, error) {
	// TODO(seanmcgary): nonSignerOperatorIds should be a list of operatorIds which is the hash of their public key
	nonSignerOperatorIds := make([]*Operator[signing.PublicKey], 0)
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

	nonSignerPublicKeys := make([]signing.PublicKey, 0)
	for _, operatorId := range nonSignerOperatorIds {
		operator := util.Find(tra.Operators, func(op *Operator[signing.PublicKey]) bool {
			return strings.EqualFold(op.Address, operatorId.Address)
		})
		nonSignerPublicKeys = append(nonSignerPublicKeys, operator.PublicKey)
	}

	allPublicKeys := util.Map(tra.Operators, func(o *Operator[signing.PublicKey], i uint64) signing.PublicKey {
		return o.PublicKey
	})

	taskIdBytes, err := hexutil.Decode(tra.TaskId)
	if err != nil {
		return nil, fmt.Errorf("failed to decode taskId: %w", err)
	}

	// Use the most common response for the certificate
	if tra.aggregatedOperators.mostCommonResponse == nil {
		return nil, fmt.Errorf("no common response found")
	}

	return &AggregatedBN254Certificate{
		TaskId:              taskIdBytes,
		TaskResponse:        tra.aggregatedOperators.mostCommonResponse.TaskResult.Output,
		TaskResponseDigest:  tra.aggregatedOperators.mostCommonResponse.Digest,
		NonSignersPubKeys:   nonSignerPublicKeys,
		AllOperatorsPubKeys: allPublicKeys,
		SignersPublicKey:    tra.aggregatedOperators.signersG2,
		SignersSignature:    tra.aggregatedOperators.signersAggSig,
		SignedAt:            new(time.Time),
	}, nil
}

// AggregatePublicKeys aggregates a list of public keys into a single public key.
func AggregatePublicKeys(pubKeys []signing.PublicKey) (signing.PublicKey, error) {
	bn254Keys := make([]*bn254.PublicKey, len(pubKeys))
	for i, pk := range pubKeys {
		if pk == nil {
			return nil, fmt.Errorf("public key at index %d is nil", i)
		}
		bn254Pk, err := bn254.NewPublicKeyFromBytes(pk.Bytes())
		if err != nil {
			return nil, fmt.Errorf("public key at index %d is not a bn254 public key", i)
		}
		bn254Keys[i] = bn254Pk
	}
	aggregatedKey, err := bn254.AggregatePublicKeys(bn254Keys)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate public keys: %w", err)
	}

	return aggregatedKey, err
}

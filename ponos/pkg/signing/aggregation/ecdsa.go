package aggregation

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/crypto-libs/pkg/ecdsa"
	"github.com/Layr-Labs/crypto-libs/pkg/signing"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"strings"
	"sync"
	"time"
)

type AggregatedECDSACertificate struct {
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
	SignersSignatures [][]byte

	// aggregated public key of the signers
	SignersPublicKeys []signing.PublicKey

	// the time the certificate was signed
	SignedAt *time.Time
}

type ReceivedECDSAResponseWithDigest struct {
	// TaskId is the unique identifier for the task
	TaskId string

	// The full task result from the operator
	TaskResult *types.TaskResult

	// signature is the signature of the task result from the operator signed with their bls key
	Signature *ecdsa.Signature

	// digest is a keccak256 hash of the task result
	Digest []byte
}

type aggregatedECDSAOperators struct {
	// aggregated public keys of signers
	signersPublicKeys []signing.PublicKey

	// aggregated signatures of signers
	signersSignatures [][]byte

	// operators that have signed (operatorAddress --> true)
	signersOperatorSet map[string]bool

	// simple count of signers. eventually this could represent stake weight or something
	totalSigners int

	lastReceivedResponse *ReceivedECDSAResponseWithDigest
}

type ECDSATaskResultAggregator struct {
	mu                  sync.Mutex
	TaskId              string
	TaskCreatedBlock    uint64
	OperatorSetId       uint32
	ThresholdPercentage uint8
	TaskData            []byte
	TaskExpirationTime  *time.Time
	Operators           []*Operator[signing.PublicKey]
	ReceivedSignatures  map[string]*ReceivedECDSAResponseWithDigest // operator address -> signature

	AggregatePublicKeys []signing.PublicKey

	aggregatedOperators *aggregatedECDSAOperators
	// Add more fields as needed for aggregation
}

func NewECDSATaskResultAggregator(
	ctx context.Context,
	taskId string,
	taskCreatedBlock uint64,
	operatorSetId uint32,
	thresholdPercentage uint8,
	taskData []byte,
	taskExpirationTime *time.Time,
	operators []*Operator[signing.PublicKey],
) (*ECDSATaskResultAggregator, error) {
	if len(taskId) == 0 {
		return nil, ErrInvalidTaskId
	}
	if len(operators) == 0 {
		return nil, ErrNoOperatorAddresses
	}
	if thresholdPercentage == 0 || thresholdPercentage > 100 {
		return nil, ErrInvalidThreshold
	}

	aggPub := util.Map(operators, func(o *Operator[signing.PublicKey], i uint64) signing.PublicKey {
		return o.PublicKey
	})

	cert := &ECDSATaskResultAggregator{
		TaskId:              taskId,
		TaskCreatedBlock:    taskCreatedBlock,
		OperatorSetId:       operatorSetId,
		ThresholdPercentage: thresholdPercentage,
		TaskData:            taskData,
		TaskExpirationTime:  taskExpirationTime,
		Operators:           operators,
		AggregatePublicKeys: aggPub,
	}
	return cert, nil
}

func (tra *ECDSATaskResultAggregator) ProcessNewSignature(
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
		tra.ReceivedSignatures = make(map[string]*ReceivedECDSAResponseWithDigest)
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

	rr := &ReceivedECDSAResponseWithDigest{
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
		tra.aggregatedOperators = &aggregatedECDSAOperators{
			signersPublicKeys: []signing.PublicKey{operator.PublicKey},

			signersSignatures: [][]byte{taskResponse.Signature},

			// initialize the map of signers (operatorAddress --> true) to track who actually signed
			signersOperatorSet: map[string]bool{taskResponse.OperatorAddress: true},

			// initialize the count of signers (could eventually be weight or something else)
			totalSigners: 1,

			// store the last received response
			lastReceivedResponse: rr,
		}
	} else {
		tra.aggregatedOperators.signersSignatures = append(tra.aggregatedOperators.signersSignatures, taskResponse.Signature)
		tra.aggregatedOperators.signersPublicKeys = append(tra.aggregatedOperators.signersPublicKeys, operator.PublicKey)
		tra.aggregatedOperators.signersOperatorSet[taskResponse.OperatorAddress] = true
		tra.aggregatedOperators.totalSigners++
		tra.aggregatedOperators.lastReceivedResponse = rr
	}

	return nil
}

func (tra *ECDSATaskResultAggregator) SigningThresholdMet() bool {
	// Check if threshold is met (by count)
	required := int((float64(tra.ThresholdPercentage) / 100.0) * float64(len(tra.Operators)))
	if required == 0 {
		required = 1 // Always require at least one
	}
	if tra.aggregatedOperators == nil {
		return false
	}
	return tra.aggregatedOperators.totalSigners >= required
}

// TODO(seanmcgary): update this
func (tra *ECDSATaskResultAggregator) VerifyResponseSignature(taskResponse *types.TaskResult, operator *Operator[signing.PublicKey]) (*ecdsa.Signature, []byte, error) {
	/*
		digestBytes := util.GetKeccak256Digest(taskResponse.Output)
		sig, err := bn254.NewSignatureFromBytes(taskResponse.Signature)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create signature from bytes: %w", err)
		}

		if verified, err := sig.VerifySolidityCompatible(operator.PublicKey, digestBytes); err != nil {
			return nil, nil, fmt.Errorf("signature verification failed: %w", err)
		} else if !verified {
			return nil, nil, fmt.Errorf("signature verification failed: signature does not match operator public key")
		}
		return sig, digestBytes[:], nil*/
	return nil, nil, fmt.Errorf("ECDSA signature verification not implemented yet")
}

func (tra *ECDSATaskResultAggregator) GenerateFinalCertificate() (*AggregatedECDSACertificate, error) {
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

	return &AggregatedECDSACertificate{
		TaskId:              taskIdBytes,
		TaskResponse:        tra.aggregatedOperators.lastReceivedResponse.TaskResult.Output,
		TaskResponseDigest:  tra.aggregatedOperators.lastReceivedResponse.Digest,
		NonSignersPubKeys:   nonSignerPublicKeys,
		AllOperatorsPubKeys: allPublicKeys,
		SignersPublicKeys:   tra.aggregatedOperators.signersPublicKeys,
		SignersSignatures:   tra.aggregatedOperators.signersSignatures,
		SignedAt:            new(time.Time),
	}, nil
}

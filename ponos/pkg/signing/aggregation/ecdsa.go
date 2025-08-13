package aggregation

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/crypto-libs/pkg/ecdsa"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"slices"
	"sort"
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

	// signature is the signature of the task result from the operator signed with their bls key
	Signature *ecdsa.Signature

	// digest is a keccak256 hash of the task result
	Digest [32]byte
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

	lastReceivedResponse *ReceivedECDSAResponseWithDigest
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
	taskId string,
	taskResponse *types.TaskResult,
) error {
	if taskId != taskResponse.TaskId {
		return fmt.Errorf("task ID mismatch: expected %s, got %s", taskId, taskResponse.TaskId)
	}

	tra.mu.Lock()
	defer tra.mu.Unlock()

	// Validate operator is in the allowed set
	operator := util.Find(tra.Operators, func(op *Operator[common.Address]) bool {
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
	sig, err := tra.VerifyResponseSignature(taskResponse, operator)
	if err != nil {
		return fmt.Errorf("failed to verify signature: %w", err)
	}

	var digest [32]byte
	copy(digest[:], taskResponse.OutputDigest)

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
			signersPublicKeys: []common.Address{operator.PublicKey},

			signersSignatures: map[common.Address][]byte{
				operator.GetAddress(): taskResponse.Signature,
			},

			// initialize the map of signers (operatorAddress --> true) to track who actually signed
			// signersOperatorSet: map[string]bool{taskResponse.OperatorAddress: true},

			// initialize the count of signers (could eventually be weight or something else)
			totalSigners: 1,

			// store the last received response
			lastReceivedResponse: rr,
		}
	} else {
		tra.aggregatedOperators.signersSignatures[operator.GetAddress()] = taskResponse.Signature
		tra.aggregatedOperators.signersPublicKeys = append(tra.aggregatedOperators.signersPublicKeys, operator.PublicKey)
		//tra.aggregatedOperators.signersOperatorSet[taskResponse.OperatorAddress] = true
		tra.aggregatedOperators.totalSigners++
		tra.aggregatedOperators.lastReceivedResponse = rr
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

func (tra *ECDSATaskResultAggregator) VerifyResponseSignature(taskResponse *types.TaskResult, operator *Operator[common.Address]) (*ecdsa.Signature, error) {
	sig, err := ecdsa.NewSignatureFromBytes(taskResponse.Signature)
	if err != nil {
		return nil, fmt.Errorf("failed to create signature from bytes: %w", err)
	}

	if verified, err := sig.VerifyWithAddress(taskResponse.OutputDigest, operator.PublicKey); err != nil {
		return nil, fmt.Errorf("signature verification failed: %w", err)
	} else if !verified {
		return nil, fmt.Errorf("signature verification failed: signature does not match operator public key")
	}
	return sig, nil
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

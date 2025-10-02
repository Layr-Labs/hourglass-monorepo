package aggregation

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	"github.com/Layr-Labs/crypto-libs/pkg/ecdsa"
	"github.com/Layr-Labs/crypto-libs/pkg/signing"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/ethereum/go-ethereum/common"
)

type ITaskResultAggregator[SigT, CertT, PubKeyT any] interface {
	SigningThresholdMet() bool

	ProcessNewSignature(
		ctx context.Context,
		taskResponse *types.TaskResult,
	) error

	VerifyResponseSignature(taskResponse *types.TaskResult, operator *Operator[PubKeyT], outputDigest [32]byte) (*SigT, error)

	GenerateFinalCertificate() (*CertT, error)
}

type Operator[PubKeyT any] struct {
	Address       string
	PublicKey     PubKeyT
	OperatorIndex uint32
	Weights       []*big.Int
}

func (o *Operator[PubKeyT]) GetAddress() common.Address {
	return common.HexToAddress(o.Address)
}

var (
	ErrInvalidTaskId             = fmt.Errorf("taskId must not be empty")
	ErrNoOperatorAddresses       = fmt.Errorf("operatorAddresses must not be empty")
	ErrInvalidThreshold          = fmt.Errorf("thresholdPercentage must be between 1 and 100")
	ErrInvalidReferenceTimestamp = fmt.Errorf("referenceTimestamp must be positive")
)

type BN254TaskResultAggregator struct {
	ctx                context.Context
	mu                 sync.Mutex
	TaskId             string
	ReferenceTimestamp uint32
	OperatorSetId      uint32
	ThresholdBips      uint16
	l1ContractCaller   contractCaller.IContractCaller
	TaskData           []byte
	TaskExpirationTime *time.Time
	Operators          []*Operator[signing.PublicKey]
	ReceivedSignatures map[string]*ReceivedBN254ResponseWithDigest // operator address -> signature
	AggregatePublicKey signing.PublicKey

	aggregatedOperators *aggregatedBN254Operators
}

type ReceivedBN254ResponseWithDigest struct {
	// TaskId is the unique identifier for the task
	TaskId string

	// The full task result from the operator
	TaskResult *types.TaskResult

	// signature is the signature of the task result from the operator signed with their bls key
	Signature *bn254.Signature

	// OutputDigest is a keccak256 hash of the output bytes (for consensus tracking)
	OutputDigest [32]byte
}

// signerInfo holds information about an individual signer
type signerInfo struct {
	publicKey *bn254.PublicKey
	signature *bn254.Signature
	operator  *Operator[signing.PublicKey]
}

// digestGroup tracks all signers for a specific output digest
type digestGroup struct {
	// Operators who signed this specific digest
	signers map[string]*signerInfo // operator address -> info

	// Representative response for this digest
	response *ReceivedBN254ResponseWithDigest

	// Total stake weight of all signers for this digest
	currentWeight *big.Int
}

type aggregatedBN254Operators struct {
	// Group signatures by the digest they signed
	digestGroups map[[32]byte]*digestGroup

	// Track the digest with the highest stake weight
	winningDigest [32]byte
	winningWeight *big.Int
}

type AggregatedBN254Certificate struct {
	// the unique identifier for the task
	TaskId []byte

	// the output of the task
	TaskResponse []byte

	// keccak256 hash of the task response
	TaskResponseDigest [32]byte

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

	// Non-signer operators sorted by OperatorIndex (for contract submission)
	NonSignerOperators []*Operator[signing.PublicKey]

	// All operators sorted by OperatorIndex (for merkle proof generation)
	SortedOperatorsByIndex []*Operator[signing.PublicKey]
}

func (cert *AggregatedBN254Certificate) ToSubmitParams() *contractCaller.BN254TaskResultParams {
	params := &contractCaller.BN254TaskResultParams{
		TaskId:             cert.TaskId,
		TaskResponse:       cert.TaskResponse,
		TaskResponseDigest: cert.TaskResponseDigest,
		SignersSignature:   cert.SignersSignature,
		SignersPublicKey:   cert.SignersPublicKey,
	}

	params.NonSignerOperators = make([]contractCaller.BN254NonSignerOperator, len(cert.NonSignerOperators))
	for i, op := range cert.NonSignerOperators {
		params.NonSignerOperators[i] = contractCaller.BN254NonSignerOperator{
			OperatorIndex: op.OperatorIndex,
			PublicKey:     op.PublicKey.Bytes(),
		}
	}

	params.SortedOperatorsByIndex = make([]contractCaller.BN254OperatorWithWeights, len(cert.SortedOperatorsByIndex))
	for i, op := range cert.SortedOperatorsByIndex {
		params.SortedOperatorsByIndex[i] = contractCaller.BN254OperatorWithWeights{
			OperatorIndex: op.OperatorIndex,
			PublicKey:     op.PublicKey.Bytes(),
			Weights:       op.Weights,
		}
	}

	return params
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

	// Total stake weight of all signers for this digest
	currentWeight *big.Int
}

type aggregatedECDSAOperators struct {
	// Group signatures by the digest they signed
	digestGroups map[[32]byte]*ecdsaDigestGroup

	// Track the digest with the highest stake weight
	winningDigest [32]byte
	winningWeight *big.Int
}

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

// ToSubmitParams converts the certificate to contract submission parameters
func (cert *AggregatedECDSACertificate) ToSubmitParams() *contractCaller.ECDSATaskResultParams {
	return &contractCaller.ECDSATaskResultParams{
		TaskId:             cert.TaskId,
		TaskResponse:       cert.TaskResponse,
		TaskResponseDigest: cert.TaskResponseDigest,
		SignersSignatures:  cert.SignersSignatures,
	}
}

package aggregation

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	"github.com/Layr-Labs/crypto-libs/pkg/signing"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

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
}

// ToSubmitParams converts the certificate to contract submission parameters
func (cert *AggregatedBN254Certificate) ToSubmitParams() *contractCaller.BN254TaskResultParams {
	params := &contractCaller.BN254TaskResultParams{
		TaskId:             cert.TaskId,
		TaskResponse:       cert.TaskResponse,
		TaskResponseDigest: cert.TaskResponseDigest,
		SignersSignature:   cert.SignersSignature,
		SignersPublicKey:   cert.SignersPublicKey,
	}

	// Convert NonSignerOperators
	params.NonSignerOperators = make([]contractCaller.BN254NonSignerOperator, len(cert.NonSignerOperators))
	for i, op := range cert.NonSignerOperators {
		params.NonSignerOperators[i] = contractCaller.BN254NonSignerOperator{
			OperatorIndex: op.OperatorIndex,
			PublicKey:     op.PublicKey.Bytes(),
		}
	}

	return params
}

// BN254TaskResultAggregator represents the data needed to initialize a new aggregation task window.
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
	// Add more fields as needed for aggregation
}

// NewBN254TaskResultAggregator initializes a new aggregation certificate for a task window.
// All required data must be provided as arguments; no network or chain calls are performed.
func NewBN254TaskResultAggregator(
	ctx context.Context,
	taskId string,
	referenceTimestamp uint32,
	operatorSetId uint32,
	thresholdBips uint16,
	l1ContractCaller contractCaller.IContractCaller,
	taskData []byte,
	taskExpirationTime *time.Time,
	operators []*Operator[signing.PublicKey],
) (*BN254TaskResultAggregator, error) {
	if len(taskId) == 0 {
		return nil, ErrInvalidTaskId
	}
	if referenceTimestamp == 0 {
		return nil, ErrInvalidReferenceTimestamp
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
		ctx:                ctx,
		TaskId:             taskId,
		ReferenceTimestamp: referenceTimestamp,
		OperatorSetId:      operatorSetId,
		ThresholdBips:      thresholdBips,
		l1ContractCaller:   l1ContractCaller,
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

	// Count of signatures for this digest
	count int
}

type aggregatedBN254Operators struct {
	// Group signatures by the digest they signed
	digestGroups map[[32]byte]*digestGroup

	// Track the most common digest
	mostCommonDigest [32]byte
	mostCommonCount  int

	// Track total count of all signers across all digests
	totalSignerCount int
}

func (tra *BN254TaskResultAggregator) SigningThresholdMet() bool {
	// Check if threshold is met based on total signers (not just most common)
	required := int((float64(tra.ThresholdBips) / 10_000.0) * float64(len(tra.Operators)))
	if required == 0 {
		required = 1 // Always require at least one
	}
	if tra.aggregatedOperators == nil {
		return false
	}
	// Check if we have enough total signatures across all digests
	// The most common among them will be selected as the winner
	return tra.aggregatedOperators.totalSignerCount >= required
}

// ProcessNewSignature processes a new signature submission from an operator.
// Returns true if the threshold is met after this submission, false otherwise.
func (tra *BN254TaskResultAggregator) ProcessNewSignature(
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
	operator := util.Find(tra.Operators, func(op *Operator[signing.PublicKey]) bool {
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
		tra.ReceivedSignatures = make(map[string]*ReceivedBN254ResponseWithDigest)
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

	rr := &ReceivedBN254ResponseWithDigest{
		TaskId:       tra.TaskId,
		TaskResult:   taskResponse,
		Signature:    sig,
		OutputDigest: outputDigest,
	}

	tra.ReceivedSignatures[taskResponse.OperatorAddress] = rr

	bn254PubKey, err := bn254.NewPublicKeyFromBytes(operator.PublicKey.Bytes())
	if err != nil {
		return fmt.Errorf("failed to create public key from bytes: %w", err)
	}

	// Aggregate when generating the final certificate
	if tra.aggregatedOperators == nil {
		tra.aggregatedOperators = &aggregatedBN254Operators{
			digestGroups: make(map[[32]byte]*digestGroup),
		}
	}

	// Get or create the digest group for this output
	group, exists := tra.aggregatedOperators.digestGroups[outputDigest]
	if !exists {
		group = &digestGroup{
			signers:  make(map[string]*signerInfo),
			response: rr,
			count:    0,
		}
		tra.aggregatedOperators.digestGroups[outputDigest] = group
	}

	// Store signer info for later aggregation
	group.signers[taskResponse.OperatorAddress] = &signerInfo{
		publicKey: bn254PubKey,
		signature: sig,
		operator:  operator,
	}
	group.count++

	// Update most common tracking
	if group.count > tra.aggregatedOperators.mostCommonCount {
		tra.aggregatedOperators.mostCommonCount = group.count
		tra.aggregatedOperators.mostCommonDigest = outputDigest
	}

	// Increment total signer count
	tra.aggregatedOperators.totalSignerCount++

	return nil
}

// VerifyResponseSignature verifies both result and auth signatures
func (tra *BN254TaskResultAggregator) VerifyResponseSignature(
	taskResponse *types.TaskResult,
	operator *Operator[signing.PublicKey],
	outputDigest [32]byte,
) (*bn254.Signature, error) {
	if !strings.EqualFold(taskResponse.OperatorAddress, operator.Address) {
		return nil, fmt.Errorf("operator address mismatch: expected %s, got %s",
			operator.Address, taskResponse.OperatorAddress)
	}

	// Step 1: Verify the result signature
	resultSig, err := bn254.NewSignatureFromBytes(taskResponse.ResultSignature)
	if err != nil {
		return nil, fmt.Errorf("failed to parse result signature: %w", err)
	}

	// Verify result signature matches output digest
	bn254PubKey, ok := operator.PublicKey.(*bn254.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to cast public key to bn254.PublicKey")
	}
	signedOverDigest, err := tra.l1ContractCaller.CalculateBN254CertificateDigestBytes(
		tra.ctx,
		tra.ReferenceTimestamp,
		outputDigest,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to calculate signature: %w", err)
	}

	var digestData [32]byte
	copy(digestData[:], signedOverDigest)
	if verified, err := resultSig.VerifySolidityCompatible(bn254PubKey, digestData); err != nil {
		return nil, fmt.Errorf("result signature verification failed: %w", err)
	} else if !verified {
		return nil, fmt.Errorf("result signature verification failed: signature does not match operator public key")
	}

	// Step 2: Verify the auth signature (for identity)
	authSig, err := bn254.NewSignatureFromBytes(taskResponse.AuthSignature)
	if err != nil {
		return nil, fmt.Errorf("failed to parse auth signature: %w", err)
	}

	// TODO: populate and use tra avs address and operator address to properly verify the expected result
	authData := &types.AuthSignatureData{
		TaskId:          tra.TaskId,
		AvsAddress:      taskResponse.AvsAddress,
		OperatorAddress: taskResponse.OperatorAddress,
		OperatorSetId:   tra.OperatorSetId,
		ResultSigDigest: util.GetKeccak256Digest(taskResponse.ResultSignature),
	}

	authBytes := authData.ToSigningBytes()
	authBytesDigest := util.GetKeccak256Digest(authBytes)
	hashCopy := make([]byte, 32)
	copy(hashCopy, authBytesDigest[:])

	if verified, err := authSig.Verify(operator.PublicKey.(*bn254.PublicKey), hashCopy); err != nil {
		return nil, fmt.Errorf("auth signature verification failed: %w", err)
	} else if !verified {
		return nil, fmt.Errorf("auth signature verification failed: signature does not match operator public key")
	}

	return resultSig, nil
}

// GenerateFinalCertificate generates the final aggregated certificate for the task.
func (tra *BN254TaskResultAggregator) GenerateFinalCertificate() (*AggregatedBN254Certificate, error) {
	if tra.aggregatedOperators == nil || len(tra.aggregatedOperators.digestGroups) == 0 {
		return nil, fmt.Errorf("no signatures collected")
	}

	// Find the winning digest group
	winningGroup := tra.aggregatedOperators.digestGroups[tra.aggregatedOperators.mostCommonDigest]
	if winningGroup == nil || winningGroup.count == 0 {
		return nil, fmt.Errorf("no signatures for winning digest")
	}

	// Aggregate only the signatures that signed the winning message
	var aggregatedSig *bn254.Signature
	aggregatedPubKey := bn254.NewZeroG2Point()

	for _, signer := range winningGroup.signers {
		if aggregatedSig == nil {
			aggregatedSig = signer.signature
		} else {
			aggregatedSig.Add(signer.signature)
		}
		aggregatedPubKey.AddPublicKey(signer.publicKey)
	}

	// IMPORTANT: All operators who didn't sign the winning digest are non-signers
	// This includes operators who signed a different digest
	nonSignerOperators := make([]*Operator[signing.PublicKey], 0)
	for _, operator := range tra.Operators {
		_, signedWinning := winningGroup.signers[operator.Address]
		if !signedWinning {
			// Either didn't sign at all, or signed a different digest
			nonSignerOperators = append(nonSignerOperators, operator)
		}
	}

	// Sort non-signers by OperatorIndex as required by the certificate verifier
	sort.SliceStable(nonSignerOperators, func(i, j int) bool {
		return nonSignerOperators[i].OperatorIndex < nonSignerOperators[j].OperatorIndex
	})

	nonSignerPublicKeys := make([]signing.PublicKey, 0)
	for _, operator := range nonSignerOperators {
		nonSignerPublicKeys = append(nonSignerPublicKeys, operator.PublicKey)
	}

	allPublicKeys := util.Map(tra.Operators, func(o *Operator[signing.PublicKey], i uint64) signing.PublicKey {
		return o.PublicKey
	})

	taskIdBytes, err := hexutil.Decode(tra.TaskId)
	if err != nil {
		return nil, fmt.Errorf("failed to decode taskId: %w", err)
	}

	return &AggregatedBN254Certificate{
		TaskId:              taskIdBytes,
		TaskResponse:        winningGroup.response.TaskResult.Output,
		TaskResponseDigest:  winningGroup.response.OutputDigest,
		NonSignersPubKeys:   nonSignerPublicKeys,
		AllOperatorsPubKeys: allPublicKeys,
		SignersPublicKey:    aggregatedPubKey,
		SignersSignature:    aggregatedSig,
		SignedAt:            new(time.Time),
		NonSignerOperators:  nonSignerOperators,
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

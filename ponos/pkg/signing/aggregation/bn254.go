package aggregation

import (
	"context"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"time"

	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	"github.com/Layr-Labs/crypto-libs/pkg/signing"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

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

func (tra *BN254TaskResultAggregator) SigningThresholdMet() bool {
	if tra.aggregatedOperators == nil || len(tra.aggregatedOperators.digestGroups) == 0 {
		return false
	}

	mostCommonGroup := tra.aggregatedOperators.digestGroups[tra.aggregatedOperators.mostCommonDigest]
	if mostCommonGroup == nil {
		return false
	}

	totalStake := big.NewInt(0)
	for _, op := range tra.Operators {
		if len(op.Weights) > 0 {
			totalStake.Add(totalStake, op.Weights[0])
		}
	}

	if totalStake.Sign() == 0 {
		return false
	}

	signersStake := big.NewInt(0)
	for signerAddr := range mostCommonGroup.signers {
		for _, op := range tra.Operators {
			if strings.EqualFold(op.Address, signerAddr) && len(op.Weights) > 0 {
				signersStake.Add(signersStake, op.Weights[0])
				break
			}
		}
	}

	thresholdStake := new(big.Int).Mul(totalStake, big.NewInt(int64(tra.ThresholdBips)))
	thresholdStake.Quo(thresholdStake, big.NewInt(10000))

	return signersStake.Cmp(thresholdStake) >= 0
}

func (tra *BN254TaskResultAggregator) ProcessNewSignature(
	ctx context.Context,
	taskResponse *types.TaskResult,
) error {
	tra.mu.Lock()
	defer tra.mu.Unlock()

	err := tra.validateTaskResponse(taskResponse)
	if err != nil {
		return fmt.Errorf("failed to validate task response: %w", err)
	}

	operator := util.Find(tra.Operators, func(op *Operator[signing.PublicKey]) bool {
		return strings.EqualFold(op.Address, taskResponse.OperatorAddress)
	})
	if operator == nil {
		return fmt.Errorf("operator %s is not in the allowed set", taskResponse.OperatorAddress)
	}

	if tra.ReceivedSignatures == nil {
		tra.ReceivedSignatures = make(map[string]*ReceivedBN254ResponseWithDigest)
	}

	var taskMessageHash [32]byte
	copy(taskMessageHash[:], common.HexToHash(taskResponse.TaskId).Bytes())

	outputTaskMessage, err := tra.l1ContractCaller.CalculateTaskMessageHash(ctx, taskMessageHash, taskResponse.Output)
	if err != nil {
		return fmt.Errorf("failed to calculate task hash: %w", err)
	}

	sig, err := tra.VerifyResponseSignature(taskResponse, operator, outputTaskMessage)
	if err != nil {
		return fmt.Errorf("failed to verify signatures: %w", err)
	}

	rr := &ReceivedBN254ResponseWithDigest{
		TaskId:       tra.TaskId,
		TaskResult:   taskResponse,
		Signature:    sig,
		OutputDigest: outputTaskMessage,
	}

	tra.ReceivedSignatures[taskResponse.OperatorAddress] = rr

	bn254PubKey, err := bn254.NewPublicKeyFromBytes(operator.PublicKey.Bytes())
	if err != nil {
		return fmt.Errorf("failed to create public key from bytes: %w", err)
	}

	if tra.aggregatedOperators == nil {
		tra.aggregatedOperators = &aggregatedBN254Operators{
			digestGroups: make(map[[32]byte]*digestGroup),
		}
	}

	group, exists := tra.aggregatedOperators.digestGroups[outputTaskMessage]
	if !exists {
		group = &digestGroup{
			signers:  make(map[string]*signerInfo),
			response: rr,
			count:    0,
		}
		tra.aggregatedOperators.digestGroups[outputTaskMessage] = group
	}

	group.signers[taskResponse.OperatorAddress] = &signerInfo{
		publicKey: bn254PubKey,
		signature: sig,
		operator:  operator,
	}
	group.count++

	tra.updateMostCommonResponse(group, outputTaskMessage)

	tra.aggregatedOperators.totalSignerCount++

	return nil
}

func (tra *BN254TaskResultAggregator) updateMostCommonResponse(group *digestGroup, outputTaskMessage [32]byte) {

	if group.count > tra.aggregatedOperators.mostCommonCount {
		tra.aggregatedOperators.mostCommonCount = group.count
		tra.aggregatedOperators.mostCommonDigest = outputTaskMessage
	} else if group.count == tra.aggregatedOperators.mostCommonCount && group.count > 0 {

		currentGroupStake := tra.calculateGroupStake(group)

		mostCommonGroup := tra.aggregatedOperators.digestGroups[tra.aggregatedOperators.mostCommonDigest]
		mostCommonGroupStake := tra.calculateGroupStake(mostCommonGroup)

		if currentGroupStake.Cmp(mostCommonGroupStake) > 0 {
			tra.aggregatedOperators.mostCommonCount = group.count
			tra.aggregatedOperators.mostCommonDigest = outputTaskMessage
		}
	}
}

func (tra *BN254TaskResultAggregator) calculateGroupStake(group *digestGroup) *big.Int {
	totalStake := big.NewInt(0)
	if group != nil {
		for _, signer := range group.signers {
			if signer.operator != nil && len(signer.operator.Weights) > 0 {
				totalStake.Add(totalStake, signer.operator.Weights[0])
			}
		}
	}
	return totalStake
}

func (tra *BN254TaskResultAggregator) VerifyResponseSignature(
	taskResponse *types.TaskResult,
	operator *Operator[signing.PublicKey],
	outputDigest [32]byte,
) (*bn254.Signature, error) {
	if !strings.EqualFold(taskResponse.OperatorAddress, operator.Address) {
		return nil, fmt.Errorf("operator address mismatch: expected %s, got %s",
			operator.Address, taskResponse.OperatorAddress)
	}

	resultSig, err := bn254.NewSignatureFromBytes(taskResponse.ResultSignature)
	if err != nil {
		return nil, fmt.Errorf("failed to parse result signature: %w", err)
	}

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

	authSig, err := bn254.NewSignatureFromBytes(taskResponse.AuthSignature)
	if err != nil {
		return nil, fmt.Errorf("failed to parse auth signature: %w", err)
	}

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

func (tra *BN254TaskResultAggregator) GenerateFinalCertificate() (*AggregatedBN254Certificate, error) {
	if tra.aggregatedOperators == nil || len(tra.aggregatedOperators.digestGroups) == 0 {
		return nil, fmt.Errorf("no signatures collected")
	}

	winningGroup := tra.aggregatedOperators.digestGroups[tra.aggregatedOperators.mostCommonDigest]
	if winningGroup == nil || winningGroup.count == 0 {
		return nil, fmt.Errorf("no signatures for winning digest")
	}

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

	// Sort all operators by their operator index
	sortedOperators := make([]*Operator[signing.PublicKey], len(tra.Operators))
	copy(sortedOperators, tra.Operators)
	sort.SliceStable(sortedOperators, func(i, j int) bool {
		return sortedOperators[i].OperatorIndex < sortedOperators[j].OperatorIndex
	})

	allPublicKeys := util.Map(sortedOperators, func(o *Operator[signing.PublicKey], i uint64) signing.PublicKey {
		return o.PublicKey
	})

	taskIdBytes, err := hexutil.Decode(tra.TaskId)
	if err != nil {
		return nil, fmt.Errorf("failed to decode taskId: %w", err)
	}

	return &AggregatedBN254Certificate{
		TaskId:                 taskIdBytes,
		TaskResponse:           winningGroup.response.TaskResult.Output,
		TaskResponseDigest:     winningGroup.response.OutputDigest,
		NonSignersPubKeys:      nonSignerPublicKeys,
		AllOperatorsPubKeys:    allPublicKeys,
		SignersPublicKey:       aggregatedPubKey,
		SignersSignature:       aggregatedSig,
		SignedAt:               new(time.Time),
		NonSignerOperators:     nonSignerOperators,
		SortedOperatorsByIndex: sortedOperators,
	}, nil
}

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

func (tra *BN254TaskResultAggregator) validateTaskResponse(taskResponse *types.TaskResult) error {
	if tra.TaskId != taskResponse.TaskId {
		return fmt.Errorf("task ID mismatch: expected %s, got %s", tra.TaskId, taskResponse.TaskId)
	}

	if len(taskResponse.ResultSignature) == 0 {
		return fmt.Errorf("result signature is empty")
	}
	if len(taskResponse.AuthSignature) == 0 {
		return fmt.Errorf("auth signature is empty")
	}

	if _, ok := tra.ReceivedSignatures[taskResponse.OperatorAddress]; ok {
		return fmt.Errorf("operator %s has already submitted a signature", taskResponse.OperatorAddress)
	}

	if taskResponse.OperatorSetId != tra.OperatorSetId {
		return fmt.Errorf("operator set ID mismatch: expected %d, got %d",
			tra.OperatorSetId, taskResponse.OperatorSetId)
	}

	return nil
}

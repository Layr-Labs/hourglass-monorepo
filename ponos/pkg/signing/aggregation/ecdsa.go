package aggregation

import (
	"context"
	"fmt"
	"math/big"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/Layr-Labs/crypto-libs/pkg/ecdsa"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type ECDSATaskResultAggregator struct {
	mu                  sync.Mutex
	TaskId              string
	ReferenceTimestamp  uint32
	OperatorSetId       uint32
	ThresholdBips       uint16
	TaskData            []byte
	TaskExpirationTime  *time.Time
	Operators           []*Operator[common.Address]
	OperatorSignatures  map[string]*ReceivedECDSAResponseWithDigest
	AggregatePublicKeys []common.Address
	aggregatedOperators *aggregatedECDSAOperators
	L1ContractCaller    contractCaller.IContractCaller
}

func NewECDSATaskResultAggregator(
	_ context.Context,
	taskId string,
	referenceTimestamp uint32,
	operatorSetId uint32,
	thresholdBips uint16,
	l1ContractCaller contractCaller.IContractCaller,
	taskData []byte,
	taskExpirationTime *time.Time,
	operators []*Operator[common.Address],
) (*ECDSATaskResultAggregator, error) {
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

	aggPub := util.Map(operators, func(o *Operator[common.Address], i uint64) common.Address {
		return o.PublicKey
	})

	cert := &ECDSATaskResultAggregator{
		TaskId:              taskId,
		ReferenceTimestamp:  referenceTimestamp,
		OperatorSetId:       operatorSetId,
		ThresholdBips:       thresholdBips,
		TaskData:            taskData,
		TaskExpirationTime:  taskExpirationTime,
		Operators:           operators,
		AggregatePublicKeys: aggPub,
		L1ContractCaller:    l1ContractCaller,
	}
	return cert, nil
}

func (tra *ECDSATaskResultAggregator) SigningThresholdMet() bool {
	if tra.aggregatedOperators == nil || len(tra.aggregatedOperators.digestGroups) == 0 {
		return false
	}

	winningGroup := tra.aggregatedOperators.digestGroups[tra.aggregatedOperators.winningDigest]
	if winningGroup == nil {
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

	signersStake := winningGroup.currentWeight
	if signersStake == nil {
		return false
	}

	thresholdStake := new(big.Int).Mul(totalStake, big.NewInt(int64(tra.ThresholdBips)))
	thresholdStake.Quo(thresholdStake, big.NewInt(10000))

	return signersStake.Cmp(thresholdStake) >= 0
}

func (tra *ECDSATaskResultAggregator) ProcessNewSignature(
	ctx context.Context,
	taskResponse *types.TaskResult,
) error {
	tra.mu.Lock()
	defer tra.mu.Unlock()

	if tra.TaskId != taskResponse.TaskId {
		return fmt.Errorf("task ID mismatch: expected %s, got %s", tra.TaskId, taskResponse.TaskId)
	}

	if taskResponse.OperatorSetId != tra.OperatorSetId {
		return fmt.Errorf("operator set ID mismatch: expected %d, got %d",
			tra.OperatorSetId, taskResponse.OperatorSetId)
	}

	operator := util.Find(tra.Operators, func(op *Operator[common.Address]) bool {
		match := strings.EqualFold(op.Address, taskResponse.OperatorAddress)
		return match
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
	if tra.OperatorSignatures == nil {
		tra.OperatorSignatures = make(map[string]*ReceivedECDSAResponseWithDigest)
	}

	// check to see if the operator has already submitted a signature
	if _, ok := tra.OperatorSignatures[taskResponse.OperatorAddress]; ok {
		return fmt.Errorf("operator %s has already submitted a signature", taskResponse.OperatorAddress)
	}

	var taskMessageHash [32]byte
	copy(taskMessageHash[:], common.HexToHash(taskResponse.TaskId).Bytes())

	outputTaskMessage, err := tra.L1ContractCaller.CalculateTaskMessageHash(ctx, taskMessageHash, taskResponse.Output)
	if err != nil {
		return fmt.Errorf("failed to calculate task message hash: %w", err)
	}

	sig, err := tra.VerifyResponseSignature(taskResponse, operator, outputTaskMessage)
	if err != nil {
		return fmt.Errorf("failed to verify signatures: %w", err)
	}

	rr := &ReceivedECDSAResponseWithDigest{
		TaskId:       tra.TaskId,
		TaskResult:   taskResponse,
		Signature:    sig,
		OutputDigest: outputTaskMessage,
	}

	tra.OperatorSignatures[taskResponse.OperatorAddress] = rr

	if tra.aggregatedOperators == nil {
		tra.aggregatedOperators = &aggregatedECDSAOperators{
			digestGroups: make(map[[32]byte]*ecdsaDigestGroup),
		}
	}

	group, exists := tra.aggregatedOperators.digestGroups[outputTaskMessage]
	if !exists {
		group = &ecdsaDigestGroup{
			signers:       make(map[string]*ecdsaSignerInfo),
			response:      rr,
			count:         0,
			currentWeight: big.NewInt(0),
		}
		tra.aggregatedOperators.digestGroups[outputTaskMessage] = group
	}

	group.signers[taskResponse.OperatorAddress] = &ecdsaSignerInfo{
		publicKey: operator.PublicKey,
		signature: taskResponse.ResultSignature,
		operator:  operator,
	}
	group.count++

	if len(operator.Weights) > 0 {
		group.currentWeight.Add(group.currentWeight, operator.Weights[0])
	}

	tra.updateWinningResponse(group, outputTaskMessage)

	return nil
}

func (tra *ECDSATaskResultAggregator) VerifyResponseSignature(
	taskResponse *types.TaskResult,
	operator *Operator[common.Address],
	outputDigest [32]byte,
) (*ecdsa.Signature, error) {
	if !strings.EqualFold(taskResponse.OperatorAddress, operator.Address) {
		return nil, fmt.Errorf("operator address mismatch: expected %s, got %s",
			operator.Address, taskResponse.OperatorAddress)
	}

	resultSig, err := ecdsa.NewSignatureFromBytes(taskResponse.ResultSignature)
	if err != nil {
		return nil, fmt.Errorf("failed to parse result signature: %w", err)
	}

	signedOverBytes, err := tra.L1ContractCaller.CalculateECDSACertificateDigestBytes(
		context.Background(),
		tra.ReferenceTimestamp,
		outputDigest,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to calculate ECDSA certificate digest: %w", err)
	}

	resultSigHash := util.GetKeccak256Digest(signedOverBytes)
	resultHashCopy := make([]byte, 32)
	copy(resultHashCopy, resultSigHash[:])

	if verified, err := resultSig.VerifyWithAddress(resultHashCopy, operator.PublicKey); err != nil {
		return nil, fmt.Errorf("result signature verification failed: %w", err)
	} else if !verified {
		return nil, fmt.Errorf("result signature verification failed: signature does not match operator public key")
	}

	authSig, err := ecdsa.NewSignatureFromBytes(taskResponse.AuthSignature)
	if err != nil {
		return nil, fmt.Errorf("failed to parse auth signature: %w", err)
	}

	authData := &types.AuthSignatureData{
		TaskId:          taskResponse.TaskId,
		AvsAddress:      taskResponse.AvsAddress,
		OperatorAddress: taskResponse.OperatorAddress,
		OperatorSetId:   taskResponse.OperatorSetId,
		ResultSigDigest: util.GetKeccak256Digest(resultSig.Bytes()),
	}

	authBytes := authData.ToSigningBytes()
	authBytesDigest := util.GetKeccak256Digest(authBytes)
	authHashCopy := make([]byte, 32)
	copy(authHashCopy, authBytesDigest[:])

	if verified, err := authSig.VerifyWithAddress(authHashCopy, operator.PublicKey); err != nil {
		return nil, fmt.Errorf("auth signature verification failed: %w", err)
	} else if !verified {
		return nil, fmt.Errorf("auth signature verification failed: signature does not match operator public key")
	}

	return resultSig, nil
}

func (tra *ECDSATaskResultAggregator) GenerateFinalCertificate() (*AggregatedECDSACertificate, error) {
	if tra.aggregatedOperators == nil || len(tra.aggregatedOperators.digestGroups) == 0 {
		return nil, fmt.Errorf("no signatures collected")
	}

	winningGroup := tra.aggregatedOperators.digestGroups[tra.aggregatedOperators.winningDigest]
	if winningGroup == nil || len(winningGroup.signers) == 0 {
		return nil, fmt.Errorf("no signatures for winning digest")
	}

	signersSignatures := make(map[common.Address][]byte)
	for _, signer := range winningGroup.signers {
		signersSignatures[signer.operator.GetAddress()] = signer.signature
	}

	taskIdBytes, err := hexutil.Decode(tra.TaskId)
	if err != nil {
		return nil, fmt.Errorf("failed to decode taskId: %w", err)
	}

	return &AggregatedECDSACertificate{
		TaskId:             taskIdBytes,
		TaskResponse:       winningGroup.response.TaskResult.Output,
		TaskResponseDigest: winningGroup.response.OutputDigest,
		SignersSignatures:  signersSignatures,
		SignedAt:           new(time.Time),
	}, nil
}

func (cert *AggregatedECDSACertificate) GetFinalSignature() ([]byte, error) {
	if len(cert.SignersSignatures) == 0 {
		return nil, fmt.Errorf("no signatures found in certificate")
	}

	addresses := make([]common.Address, 0, len(cert.SignersSignatures))
	for addr := range cert.SignersSignatures {
		addresses = append(addresses, addr)
	}

	// Sort by address raw bytes
	slices.SortFunc(addresses, func(a, b common.Address) int {
		cmp := slices.Compare(a[:], b[:])
		return cmp
	})

	// Concatenate signatures in sorted order
	var finalSignature []byte
	for _, addr := range addresses {
		sig := cert.SignersSignatures[addr]
		if len(sig) != 65 {
			return nil, fmt.Errorf("signature for address %s has invalid length: expected 65, got %d",
				addr.Hex(), len(sig))
		}
		finalSignature = append(finalSignature, sig...)
	}

	return finalSignature, nil
}

func (tra *ECDSATaskResultAggregator) updateWinningResponse(group *ecdsaDigestGroup, outputTaskMessage [32]byte) {
	if tra.aggregatedOperators.winningWeight == nil {
		tra.aggregatedOperators.winningWeight = new(big.Int).Set(group.currentWeight)
		tra.aggregatedOperators.winningDigest = outputTaskMessage
		return
	}

	cmp := group.currentWeight.Cmp(tra.aggregatedOperators.winningWeight)
	if cmp > 0 {
		tra.aggregatedOperators.winningWeight = new(big.Int).Set(group.currentWeight)
		tra.aggregatedOperators.winningDigest = outputTaskMessage
	}
}

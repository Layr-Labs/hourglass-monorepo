package contractCaller

import (
	"context"
	"fmt"
	"math/big"

	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/middleware-bindings/IBN254TableCalculator"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"github.com/ethereum/go-ethereum/common"
	ethereumTypes "github.com/ethereum/go-ethereum/core/types"
)

type AVSConfig struct {
	AggregatorOperatorSetId uint32
	ExecutorOperatorSetIds  []uint32
}

type OperatorTableData struct {
	OperatorWeights            [][]*big.Int
	Operators                  []common.Address
	LatestReferenceTimestamp   uint32
	LatestReferenceBlockNumber uint32
	TableUpdaterAddresses      map[uint64]common.Address
	OperatorInfoTreeRoot       [32]byte
	OperatorInfos              []BN254OperatorInfo
}

// BN254OperatorInfo contains BN254 operator public key and weights for merkle proof generation
type BN254OperatorInfo struct {
	PubkeyX *big.Int
	PubkeyY *big.Int
	Weights []*big.Int
}

type LatestReferenceTimeAndBlock struct {
	LatestReferenceTimestamp   uint32
	LatestReferenceBlockNumber uint32
}

type TaskMailboxExecutorOperatorSetConfig struct {
	TaskHook     common.Address
	TaskSLA      *big.Int
	FeeToken     common.Address
	CurveType    uint8
	FeeCollector common.Address
	Consensus    struct {
		ConsensusType uint8
		Value         []byte
	}
	TaskMetadata []byte
}

func (tm *TaskMailboxExecutorOperatorSetConfig) GetConsensusValue() (uint16, error) {
	return util.AbiDecodeUint16(tm.Consensus.Value)
}

func (tm *TaskMailboxExecutorOperatorSetConfig) GetCurveType() (config.CurveType, error) {
	return config.ConvertSolidityEnumToCurveType(tm.CurveType)
}

// BN254TaskResultParams contains all fields needed to submit a BN254 task result
type BN254TaskResultParams struct {
	TaskId                 []byte
	TaskResponse           []byte
	TaskResponseDigest     [32]byte
	SignersSignature       *bn254.Signature
	SignersPublicKey       *bn254.G2Point
	NonSignerOperators     []BN254NonSignerOperator
	SortedOperatorsByIndex []BN254OperatorWithWeights // All operators sorted by index with weights
}

// BN254OperatorWithWeights contains operator info including weights for merkle proof generation
type BN254OperatorWithWeights struct {
	OperatorIndex uint32
	PublicKey     []byte     // BN254 public key bytes
	Weights       []*big.Int // Operator stake weights
}

// BN254NonSignerOperator contains operator info for non-signers
type BN254NonSignerOperator struct {
	OperatorIndex uint32
	PublicKey     []byte // BN254 public key bytes
}

// ECDSATaskResultParams contains all fields needed to submit an ECDSA task result
type ECDSATaskResultParams struct {
	TaskId             []byte
	TaskResponse       []byte
	TaskResponseDigest [32]byte
	SignersSignatures  map[common.Address][]byte
}

// ErrOperatorKeyNotRegistered is returned when an operator has not registered a key for the specified operator set
var ErrOperatorKeyNotRegistered = fmt.Errorf("operator key not registered for operator set")

type IContractCaller interface {
	SubmitBN254TaskResult(ctx context.Context, params *BN254TaskResultParams, infos []BN254OperatorInfo, globalTableRootReferenceTimestamp uint32, operatorInfoTreeRoot [32]byte) (*ethereumTypes.Receipt, error)

	SubmitBN254TaskResultRetryable(
		ctx context.Context,
		params *BN254TaskResultParams,
		infos []BN254OperatorInfo,
		globalTableRootReferenceTimestamp uint32,
		operatorInfoTreeRoot [32]byte,
	) (*ethereumTypes.Receipt, error)

	SubmitECDSATaskResult(
		ctx context.Context,
		params *ECDSATaskResultParams,
		globalTableRootReferenceTimestamp uint32,
	) (*ethereumTypes.Receipt, error)

	SubmitECDSATaskResultRetryable(
		ctx context.Context,
		params *ECDSATaskResultParams,
		globalTableRootReferenceTimestamp uint32,
	) (*ethereumTypes.Receipt, error)

	VerifyBN254Certificate(
		ctx context.Context,
		avsAddress common.Address,
		operatorSetId uint32,
		params *BN254TaskResultParams,
		operatorInfos []BN254OperatorInfo,
		globalTableRootReferenceTimestamp uint32,
		operatorInfoTreeRoot [32]byte,
		thresholdPercentage uint16,
	) (bool, error)

	VerifyECDSACertificate(
		messageHash [32]byte,
		signature []byte,
		avsAddress common.Address,
		operatorSetId uint32,
		globalTableRootReferenceTimestamp uint32,
		threshold uint16,
	) (bool, []common.Address, error)

	GetAVSConfig(avsAddress string, blockNumber uint64) (*AVSConfig, error)

	GetOperatorSetCurveType(avsAddress string, operatorSetId uint32, blockNumber uint64) (config.CurveType, error)

	GetOperatorSetMembersWithPeering(avsAddress string, operatorSetId uint32, blockNumber uint64) ([]*peering.OperatorPeerInfo, error)

	GetOperatorSetDetailsForOperator(operatorAddress common.Address, avsAddress string, operatorSetId uint32, blockNumber uint64) (*peering.OperatorSet, error)

	PublishMessageToInbox(ctx context.Context, avsAddress string, operatorSetId uint32, payload []byte) (*ethereumTypes.Receipt, error)

	GetOperatorTableDataForOperatorSet(ctx context.Context, avsAddress common.Address, operatorSetId uint32, chainId config.ChainId, atBlockNumber uint64) (*OperatorTableData, error)

	GetTableUpdaterReferenceTimeAndBlock(
		ctx context.Context,
		tableUpdaterAddr common.Address,
		atBlockNumber uint64,
	) (*LatestReferenceTimeAndBlock, error)

	GetSupportedChainsForMultichain(ctx context.Context, referenceBlockNumber uint64) ([]*big.Int, []common.Address, error)

	CalculateECDSACertificateDigestBytes(ctx context.Context, referenceTimestamp uint32, messageHash [32]byte) ([]byte, error)

	CalculateBN254CertificateDigestBytes(ctx context.Context, referenceTimestamp uint32, messageHash [32]byte) ([]byte, error)

	CalculateTaskMessageHash(ctx context.Context, taskHash [32]byte, result []byte) ([32]byte, error)

	GetExecutorOperatorSetTaskConfig(ctx context.Context, avsAddress common.Address, opsetId uint32, blockNumber uint64) (*TaskMailboxExecutorOperatorSetConfig, error)

	CreateGenerationReservation(
		ctx context.Context,
		avsAddress common.Address,
		operatorSetId uint32,
		operatorTableCalculatorAddress common.Address,
		owner common.Address,
		maxStalenessPeriod uint32,
	) (*ethereumTypes.Receipt, error)

	GetTableCalculatorAddress(curveType config.CurveType) common.Address

	GetOperatorInfos(
		ctx context.Context,
		avsAddress common.Address,
		opSetId uint32,
		referenceBlockNumber uint64,
	) ([]IBN254TableCalculator.IOperatorTableCalculatorTypesBN254OperatorInfo, error)

	GetOperatorInfoTreeRoot(
		ctx context.Context,
		avsAddress common.Address,
		opSetId uint32,
		taskBlockNumber uint64,
		referenceTimestamp uint32,
	) ([32]byte, error)

	// ------------------------------------------------------------------------
	// Helper functions for test setup
	// ------------------------------------------------------------------------

	GetOperatorBN254KeyRegistrationMessageHash(
		ctx context.Context,
		operatorAddress common.Address,
		avsAddress common.Address,
		operatorSetId uint32,
		keyData []byte,
	) ([32]byte, error)

	GetOperatorECDSAKeyRegistrationMessageHash(
		ctx context.Context,
		operatorAddress common.Address,
		avsAddress common.Address,
		operatorSetId uint32,
		signingKeyAddress common.Address,
	) ([32]byte, error)

	ConfigureAVSOperatorSet(ctx context.Context, avsAddress common.Address, operatorSetId uint32, curveType config.CurveType) (*ethereumTypes.Receipt, error)

	RegisterKeyWithKeyRegistrar(
		ctx context.Context,
		operatorAddress common.Address,
		avsAddress common.Address,
		operatorSetId uint32,
		sigBytes []byte,
		keyData []byte,
	) (*ethereumTypes.Receipt, error)

	CreateOperatorAndRegisterWithAvs(
		ctx context.Context,
		avsAddress common.Address,
		operatorAddress common.Address,
		operatorSetIds []uint32,
		socket string,
		allocationDelay uint32,
		metadataUri string,
	) (*ethereumTypes.Receipt, error)

	EncodeBN254KeyData(pubKey *bn254.PublicKey) ([]byte, error)

	SetupTaskMailboxForAvs(
		ctx context.Context,
		avsAddress common.Address,
		taskHookAddress common.Address,
		executorOperatorSetIds []uint32,
		curveTypes []config.CurveType,
	) error
}

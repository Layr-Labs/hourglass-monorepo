package contractCaller

import (
	"context"
	"math/big"

	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
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
	TaskId             []byte
	TaskResponse       []byte
	TaskResponseDigest [32]byte
	SignersSignature   *bn254.Signature
	SignersPublicKey   *bn254.G2Point
	NonSignerOperators []BN254NonSignerOperator
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

type IContractCaller interface {
	SubmitBN254TaskResult(
		ctx context.Context,
		params *BN254TaskResultParams,
		globalTableRootReferenceTimestamp uint32,
	) (*ethereumTypes.Receipt, error)

	SubmitBN254TaskResultRetryable(
		ctx context.Context,
		params *BN254TaskResultParams,
		globalTableRootReferenceTimestamp uint32,
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

	GetAVSConfig(avsAddress string, blockNumber uint64) (*AVSConfig, error)

	GetOperatorSetCurveType(avsAddress string, operatorSetId uint32, blockNumber uint64) (config.CurveType, error)

	GetOperatorSetMembersWithPeering(avsAddress string, operatorSetId uint32, blockNumber uint64) ([]*peering.OperatorPeerInfo, error)

	GetOperatorSetDetailsForOperator(operatorAddress common.Address, avsAddress string, operatorSetId uint32, blockNumber uint64) (*peering.OperatorSet, error)

	PublishMessageToInbox(ctx context.Context, avsAddress string, operatorSetId uint32, payload []byte) (*ethereumTypes.Receipt, error)

	GetOperatorTableDataForOperatorSet(
		ctx context.Context,
		avsAddress common.Address,
		operatorSetId uint32,
		chainId config.ChainId,
		referenceBlocknumber uint64,
	) (*OperatorTableData, error)

	GetTableUpdaterReferenceTimeAndBlock(
		ctx context.Context,
		tableUpdaterAddr common.Address,
		atBlockNumber uint64,
	) (*LatestReferenceTimeAndBlock, error)

	GetSupportedChainsForMultichain(ctx context.Context, referenceBlockNumber uint64) ([]*big.Int, []common.Address, error)

	CalculateECDSACertificateDigestBytes(ctx context.Context, referenceTimestamp uint32, messageHash [32]byte) ([]byte, error)

	CalculateBN254CertificateDigestBytes(ctx context.Context, referenceTimestamp uint32, messageHash [32]byte) ([]byte, error)

	GetExecutorOperatorSetTaskConfig(ctx context.Context, avsAddress common.Address, opsetId uint32, blockNumber uint64) (*TaskMailboxExecutorOperatorSetConfig, error)

	CreateGenerationReservation(
		ctx context.Context,
		avsAddress common.Address,
		operatorSetId uint32,
		operatorTableCalculatorAddress common.Address,
		owner common.Address,
		maxStalenessPeriod uint32,
	) (*ethereumTypes.Receipt, error)

	SetOperatorTableCalculator(
		ctx context.Context,
		avsAddress common.Address,
		operatorSetId uint32,
		operatorTableCalculatorAddress common.Address,
	) (*ethereumTypes.Receipt, error)

	GetTableCalculatorAddress(curveType config.CurveType) common.Address

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

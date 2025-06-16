package contractCaller

import (
	"context"
	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/aggregation"
	"github.com/ethereum/go-ethereum/common"
	ethereumTypes "github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

type AVSConfig struct {
	AggregatorOperatorSetId uint32
	ExecutorOperatorSetIds  []uint32
}

type ExecutorOperatorSetTaskConfig struct {
	CertificateVerifier      string
	TaskHook                 string
	FeeToken                 string
	FeeCollector             string
	TaskSLA                  *big.Int
	StakeProportionThreshold uint16
	TaskMetadata             []byte
}

type CurveType uint8

const (
	CurveTypeUnknown CurveType = 0 // Unknown curve type
	CurveTypeECDSA   CurveType = 1
	CurveTypeBN254   CurveType = 2 // BN254 is the only supported curve type for now
)

type OperatorTableData struct {
	OperatorWeights          [][]*big.Int
	Operators                []common.Address
	LatestReferenceTimestamp uint32
}

type IContractCaller interface {
	SubmitTaskResult(
		ctx context.Context,
		aggCert *aggregation.AggregatedCertificate,
		globalTableRootReferenceTimestamp uint32,
	) (*ethereumTypes.Receipt, error)

	SubmitTaskResultRetryable(
		ctx context.Context,
		aggCert *aggregation.AggregatedCertificate,
		globalTableRootReferenceTimestamp uint32,
	) (*ethereumTypes.Receipt, error)

	GetAVSConfig(avsAddress string) (*AVSConfig, error)

	GetOperatorSetMembersWithPeering(avsAddress string, operatorSetId uint32) ([]*peering.OperatorPeerInfo, error)

	GetOperatorSetDetailsForOperator(operatorAddress common.Address, avsAddress string, operatorSetId uint32) (*peering.OperatorSet, error)

	PublishMessageToInbox(ctx context.Context, avsAddress string, operatorSetId uint32, payload []byte) (*ethereumTypes.Receipt, error)

	GetOperatorRegistrationMessageHash(
		ctx context.Context,
		operatorAddress common.Address,
		avsAddress common.Address,
		operatorSetId uint32,
		keyData []byte,
	) ([32]byte, error)

	ConfigureAVSOperatorSet(ctx context.Context, avsAddress common.Address, operatorSetId uint32, curveType CurveType) (*ethereumTypes.Receipt, error)

	RegisterKeyWithKeyRegistrar(
		ctx context.Context,
		operatorAddress common.Address,
		avsAddress common.Address,
		operatorSetId uint32,
		signature *bn254.Signature,
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

	GetOperatorTableDataForOperatorSet(
		ctx context.Context,
		avsAddress common.Address,
		operatorSetId uint32,
		chainId config.ChainId,
		referenceBlocknumber uint64,
	) (*OperatorTableData, error)
}

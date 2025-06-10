package contractCaller

import (
	"context"
	"github.com/Layr-Labs/hourglass-monorepo/contracts/pkg/bindings/ITaskAVSRegistrar"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/aggregation"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/bn254"
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

type IContractCaller interface {
	SubmitTaskResult(ctx context.Context, task *aggregation.AggregatedCertificate) (*ethereumTypes.Receipt, error)

	SubmitTaskResultRetryable(ctx context.Context, aggCert *aggregation.AggregatedCertificate) (*ethereumTypes.Receipt, error)

	GetAVSConfig(avsAddress string) (*AVSConfig, error)

	GetOperatorSetMembersWithPeering(avsAddress string, operatorSetId uint32) ([]*peering.OperatorPeerInfo, error)

	GetOperatorSetDetailsForOperator(operatorAddress common.Address, avsAddress string, operatorSetId uint32) (*peering.OperatorSet, error)

	PublishMessageToInbox(ctx context.Context, avsAddress string, operatorSetId uint32, payload []byte) (*ethereumTypes.Receipt, error)

	GetOperatorRegistrationMessageHash(ctx context.Context, address common.Address) (ITaskAVSRegistrar.BN254G1Point, error)

	CreateOperatorAndRegisterWithAvs(
		ctx context.Context,
		avsAddress common.Address,
		operatorAddress common.Address,
		operatorSetIds []uint32,
		publicKey *bn254.PublicKey,
		signature *bn254.Signature,
		socket string,
		allocationDelay uint32,
		metadataUri string,
	) (*ethereumTypes.Receipt, error)
}

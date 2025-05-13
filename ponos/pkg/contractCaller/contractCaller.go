package contractCaller

import (
	"context"
	"github.com/Layr-Labs/hourglass-monorepo/contracts/pkg/bindings/ITaskAVSRegistrar"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/aggregation"
	"github.com/ethereum/go-ethereum/common"
	ethereumTypes "github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

type AVSConfig struct {
	ResultSubmitter         string
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
	// TODO: task will need a certificate
	SubmitTaskResult(ctx context.Context, task *aggregation.AggregatedCertificate) (*ethereumTypes.Receipt, error)

	GetAVSConfig(avsAddress string) (*AVSConfig, error)

	GetTaskConfigForExecutorOperatorSet(avsAddress string, operatorSetId uint32) (*ExecutorOperatorSetTaskConfig, error)

	GetOperatorSets(avsAddress string) ([]uint32, error)

	GetOperatorSetMembers(avsAddress string, operatorSetId uint32) ([]string, error)

	GetMembersForAllOperatorSets(avsAddress string) (map[uint32][]string, error)

	GetOperatorSetMembersWithPeering(avsAddress string, operatorSetId uint32) ([]*peering.OperatorPeerInfo, error)

	PublishMessageToInbox(ctx context.Context, avsAddress string, operatorSetId uint32, payload []byte) (*ethereumTypes.Receipt, error)

	GetOperatorRegistrationMessageHash(ctx context.Context, address common.Address) (ITaskAVSRegistrar.BN254G1Point, error)
}

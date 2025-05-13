package contractCaller

import (
	"context"
	"github.com/Layr-Labs/hourglass-monorepo/contracts/pkg/bindings/ITaskAVSRegistrar"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/taskSession"
	"github.com/ethereum/go-ethereum/common"
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
	SubmitTaskResult(ctx context.Context, task *taskSession.TaskSession) error

	GetAVSConfig(avsAddress string) (*AVSConfig, error)

	GetTaskConfigForExecutorOperatorSet(avsAddress string, operatorSetId uint32) (*ExecutorOperatorSetTaskConfig, error)

	GetOperatorSets(avsAddress string) ([]uint32, error)

	GetOperatorSetMembers(avsAddress string, operatorSetId uint32) ([]string, error)

	GetMembersForAllOperatorSets(avsAddress string) (map[uint32][]string, error)

	GetOperatorSetMembersWithPeering(avsAddress string, operatorSetId uint32) ([]*peering.OperatorPeerInfo, error)

	PublishMessageToInbox(ctx context.Context, avsAddress string, operatorSetId uint32, payload []byte) (interface{}, error)

	GetOperatorRegistrationMessageHash(ctx context.Context, address common.Address) (ITaskAVSRegistrar.BN254G1Point, error)
}

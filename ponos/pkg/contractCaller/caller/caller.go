package caller

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/IAllocationManager"
	"github.com/Layr-Labs/hourglass-monorepo/contracts/pkg/bindings/ITaskAVSRegistrar"
	"github.com/Layr-Labs/hourglass-monorepo/contracts/pkg/bindings/ITaskMailbox"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/bn254"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/taskSession"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
	"math/big"
	"slices"
	"strings"
)

type ContractCallerConfig struct {
	PrivateKey          string
	AVSRegistrarAddress string
	TaskMailboxAddress  string
}

type ContractCaller struct {
	avsRegistrarCaller      *ITaskAVSRegistrar.ITaskAVSRegistrarCaller
	taskMailboxCaller       *ITaskMailbox.ITaskMailboxCaller
	taskMailboxTransactor   *ITaskMailbox.ITaskMailboxTransactor
	allocationManagerCaller *IAllocationManager.IAllocationManagerCaller
	ethclient               *ethclient.Client
	config                  *ContractCallerConfig
	logger                  *zap.Logger
}

func NewContractCallerFromEthereumClient(
	config *ContractCallerConfig,
	ethClient *ethereum.Client,
	logger *zap.Logger,
) (*ContractCaller, error) {
	client, err := ethClient.GetEthereumContractCaller()
	if err != nil {
		return nil, err
	}

	return NewContractCaller(config, client, logger)
}

func NewContractCaller(
	cfg *ContractCallerConfig,
	ethclient *ethclient.Client,
	logger *zap.Logger,
) (*ContractCaller, error) {
	avsRegistrarCaller, err := ITaskAVSRegistrar.NewITaskAVSRegistrarCaller(common.HexToAddress(cfg.AVSRegistrarAddress), ethclient)
	if err != nil {
		return nil, fmt.Errorf("failed to create AVSRegistrar caller: %w", err)
	}

	taskMailboxCaller, err := ITaskMailbox.NewITaskMailboxCaller(common.HexToAddress(cfg.TaskMailboxAddress), ethclient)
	if err != nil {
		return nil, fmt.Errorf("failed to create TaskMailbox caller: %w", err)
	}
	taskMailboxTransactor, err := ITaskMailbox.NewITaskMailboxTransactor(common.HexToAddress(cfg.TaskMailboxAddress), ethclient)
	if err != nil {
		return nil, fmt.Errorf("failed to create TaskMailbox transactor: %w", err)
	}

	chainId, err := ethclient.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	coreContracts, err := config.GetCoreContractsForChainId(config.ChainId(chainId.Uint64()))
	if err != nil {
		return nil, fmt.Errorf("failed to get core contracts: %w", err)
	}

	allocationManagerCaller, err := IAllocationManager.NewIAllocationManagerCaller(common.HexToAddress(coreContracts.AllocationManager), ethclient)
	if err != nil {
		return nil, fmt.Errorf("failed to create AllocationManager caller: %w", err)
	}

	return &ContractCaller{
		avsRegistrarCaller:      avsRegistrarCaller,
		taskMailboxCaller:       taskMailboxCaller,
		taskMailboxTransactor:   taskMailboxTransactor,
		allocationManagerCaller: allocationManagerCaller,
		ethclient:               ethclient,
		config:                  cfg,
		logger:                  logger,
	}, nil
}

func (cc *ContractCaller) getECDSAPrivateKeyFromString(privateKey string) (*ecdsa.PrivateKey, error) {
	privateKey = strings.TrimPrefix(privateKey, "0x")

	pk, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}
	return pk, nil
}

func (cc *ContractCaller) buildTxOps(ctx context.Context, pk *ecdsa.PrivateKey) (*bind.TransactOpts, error) {
	chainId, err := cc.ethclient.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	opts, err := bind.NewKeyedTransactorWithChainID(pk, chainId)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %w", err)
	}
	opts.NoSend = true
	return opts, nil
}

func (cc *ContractCaller) SubmitTaskResult(ctx context.Context, ts *taskSession.TaskSession) error {
	privateKey, err := cc.getECDSAPrivateKeyFromString(cc.config.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	noSendTxOpts, err := cc.buildTxOps(ctx, privateKey)
	if err != nil {
		return fmt.Errorf("failed to build transaction options: %w", err)
	}
	taskIdBytes, err := hexutil.Decode(ts.Task.TaskId)
	if err != nil {
		return fmt.Errorf("invalid taskId hex: %w", err)
	}
	if len(taskIdBytes) != 32 {
		return fmt.Errorf("taskId must be 32 bytes, got %d", len(taskIdBytes))
	}
	var taskId [32]byte
	copy(taskId[:], taskIdBytes)
	cc.logger.Sugar().Infow("submitting task result", "taskId", taskId)

	cert := ITaskMailbox.IBN254CertificateVerifierBN254Certificate{
		ReferenceTimestamp: uint32(ts.Task.BlockNumber),
		MessageHash:        [32]byte{},
		Sig: ITaskMailbox.BN254G1Point{
			X: big.NewInt(0),
			Y: big.NewInt(0),
		},
		Apk: ITaskMailbox.BN254G2Point{
			X: [2]*big.Int{big.NewInt(0), big.NewInt(0)},
			Y: [2]*big.Int{big.NewInt(0), big.NewInt(0)},
		},
		NonsignerIndices:   []uint32{},
		NonSignerWitnesses: []ITaskMailbox.IBN254CertificateVerifierBN254OperatorInfoWitness{},
	}

	operatorMap := ts.GetOperatorOutputsMap()
	aggregated, err := encodeOperatorOutputMap(operatorMap)
	if err != nil {
		return fmt.Errorf("failed to encode operator-output map: %w", err)
	}

	tx, err := cc.taskMailboxTransactor.SubmitResult(noSendTxOpts, taskId, cert, aggregated)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	receipt, err := cc.EstimateGasPriceAndLimitAndSendTx(ctx, noSendTxOpts.From, tx, privateKey, "SubmitTaskSession")
	if err != nil {
		return fmt.Errorf("failed to send transaction: %w", err)
	}

	cc.logger.Sugar().Infow("Successfully submitted task session result",
		zap.String("taskId", ts.Task.TaskId),
		zap.String("transactionHash", receipt.TxHash.Hex()),
	)
	return nil
}

func encodeOperatorOutputMap(m map[string][]byte) ([]byte, error) {
	var buf bytes.Buffer

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	for _, op := range keys {
		output := m[op]

		opBytes := []byte(op)
		opLen := uint32(len(opBytes))
		outLen := uint32(len(output))

		if err := binary.Write(&buf, binary.BigEndian, opLen); err != nil {
			return nil, err
		}
		if _, err := buf.Write(opBytes); err != nil {
			return nil, err
		}
		if err := binary.Write(&buf, binary.BigEndian, outLen); err != nil {
			return nil, err
		}
		if _, err := buf.Write(output); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

func (cc *ContractCaller) GetOperatorSets(avsAddress string) ([]uint32, error) {
	avsAddr := common.HexToAddress(avsAddress)
	opSets, err := cc.allocationManagerCaller.GetRegisteredSets(&bind.CallOpts{}, avsAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to get operator sets: %w", err)
	}
	opsetIds := make([]uint32, len(opSets))
	for i, opSet := range opSets {
		opsetIds[i] = opSet.Id
	}
	return opsetIds, nil
}

func (cc *ContractCaller) GetOperatorSetMembers(avsAddress string, operatorSetId uint32) ([]string, error) {
	avsAddr := common.HexToAddress(avsAddress)
	operatorSet, err := cc.allocationManagerCaller.GetMembers(&bind.CallOpts{}, IAllocationManager.OperatorSet{
		Avs: avsAddr,
		Id:  operatorSetId,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get operator set members: %w", err)
	}
	members := make([]string, len(operatorSet))
	for i, member := range operatorSet {
		members[i] = member.String()
	}
	return members, nil
}

func (cc *ContractCaller) GetOperatorSetMembersWithPeering(
	avsAddress string,
	operatorSetId uint32,
) ([]*peering.OperatorPeerInfo, error) {
	members, err := cc.GetOperatorSetMembers(avsAddress, operatorSetId)
	if err != nil {
		return nil, err
	}

	peerMembers, err := cc.avsRegistrarCaller.GetBatchOperatorPubkeyInfoAndSocket(&bind.CallOpts{}, util.Map(members, func(mem string, i uint64) common.Address {
		return common.HexToAddress(mem)
	}))
	if err != nil {
		cc.logger.Sugar().Errorf("failed to get operator set members with peering: %v", err)
		return nil, err
	}

	allMembers := make([]*peering.OperatorPeerInfo, len(peerMembers))
	for i, pm := range peerMembers {
		pubKey, err := bn254.NewPublicKeyFromSolidity(pm.PubkeyInfo.PubkeyG2)
		if err != nil {
			cc.logger.Sugar().Errorf("failed to convert public key: %v", err)
			return nil, err
		}

		allMembers = append(allMembers, &peering.OperatorPeerInfo{
			NetworkAddress:  pm.Socket,
			PublicKey:       hex.EncodeToString(pubKey.Bytes()),
			OperatorAddress: members[i],
			OperatorSetIds:  []uint32{operatorSetId},
		})
	}
	return allMembers, nil
}

func (cc *ContractCaller) GetMembersForAllOperatorSets(avsAddress string) (map[uint32][]string, error) {
	operatorSets, err := cc.GetOperatorSets(avsAddress)
	if err != nil {
		return nil, err
	}

	opsetMembers := make(map[uint32][]string)
	for _, operatorSetId := range operatorSets {
		members, err := cc.GetOperatorSetMembers(avsAddress, operatorSetId)
		if err != nil {
			return nil, err
		}
		opsetMembers[operatorSetId] = members
	}
	return opsetMembers, nil
}

func (cc *ContractCaller) GetAVSConfig(avsAddress string) (*contractCaller.AVSConfig, error) {
	avsAddr := common.HexToAddress(avsAddress)
	avsConfig, err := cc.taskMailboxCaller.GetAvsConfig(&bind.CallOpts{}, avsAddr)
	if err != nil {
		return nil, err
	}

	return &contractCaller.AVSConfig{
		ResultSubmitter:         avsConfig.ResultSubmitter.String(),
		AggregatorOperatorSetId: avsConfig.AggregatorOperatorSetId,
		ExecutorOperatorSetIds:  avsConfig.ExecutorOperatorSetIds,
	}, nil
}

func (cc *ContractCaller) GetTaskConfigForExecutorOperatorSet(avsAddress string, operatorSetId uint32) (*contractCaller.ExecutorOperatorSetTaskConfig, error) {
	avsAddr := common.HexToAddress(avsAddress)
	taskCfg, err := cc.taskMailboxCaller.GetExecutorOperatorSetTaskConfig(&bind.CallOpts{}, ITaskMailbox.OperatorSet{
		Avs: avsAddr,
		Id:  operatorSetId,
	})
	if err != nil {
		return nil, err
	}

	return &contractCaller.ExecutorOperatorSetTaskConfig{
		CertificateVerifier:      taskCfg.CertificateVerifier.String(),
		TaskHook:                 taskCfg.TaskHook.String(),
		FeeToken:                 taskCfg.FeeToken.String(),
		FeeCollector:             taskCfg.FeeCollector.String(),
		TaskSLA:                  taskCfg.TaskSLA,
		StakeProportionThreshold: taskCfg.StakeProportionThreshold,
		TaskMetadata:             taskCfg.TaskMetadata,
	}, nil
}

func (cc *ContractCaller) PublishMessageToInbox(ctx context.Context, avsAddress string, operatorSetId uint32, payload []byte) (interface{}, error) {
	privateKey, err := cc.getECDSAPrivateKeyFromString(cc.config.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to get public key ECDSA")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	noSendTxOpts, err := cc.buildTxOps(ctx, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction options: %w", err)
	}

	tx, err := cc.taskMailboxTransactor.CreateTask(noSendTxOpts, ITaskMailbox.ITaskMailboxTypesTaskParams{
		RefundCollector: address,
		AvsFee:          new(big.Int).SetUint64(0),
		ExecutorOperatorSet: ITaskMailbox.OperatorSet{
			Avs: common.HexToAddress(avsAddress),
			Id:  operatorSetId,
		},
		Payload: payload,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	receipt, err := cc.EstimateGasPriceAndLimitAndSendTx(ctx, noSendTxOpts.From, tx, privateKey, "PublishMessageToInbox")
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}
	cc.logger.Sugar().Infow("Successfully published message to inbox",
		zap.String("transactionHash", receipt.TxHash.Hex()),
	)
	return receipt, nil
}

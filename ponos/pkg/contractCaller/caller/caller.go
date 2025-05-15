package caller

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/binary"
	"fmt"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/IAllocationManager"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/IDelegationManager"
	"github.com/Layr-Labs/hourglass-monorepo/contracts/pkg/bindings/ITaskAVSRegistrar"
	"github.com/Layr-Labs/hourglass-monorepo/contracts/pkg/bindings/ITaskMailbox"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller"
	cryptoUtils "github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/crypto"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/aggregation"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/bn254"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
	"math/big"
	"slices"
)

type ContractCallerConfig struct {
	PrivateKey          string
	AVSRegistrarAddress string
	TaskMailboxAddress  string
}

type ContractCaller struct {
	avsRegistrarCaller          *ITaskAVSRegistrar.ITaskAVSRegistrarCaller
	taskMailboxCaller           *ITaskMailbox.ITaskMailboxCaller
	taskMailboxTransactor       *ITaskMailbox.ITaskMailboxTransactor
	allocationManagerCaller     *IAllocationManager.IAllocationManagerCaller
	allocationManagerTransactor *IAllocationManager.IAllocationManagerTransactor
	delegationManagerTransactor *IDelegationManager.IDelegationManagerTransactor
	ethclient                   *ethclient.Client
	config                      *ContractCallerConfig
	logger                      *zap.Logger
	coreContracts               *config.CoreContractAddresses
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

	allocationManagerTransactor, err := IAllocationManager.NewIAllocationManagerTransactor(common.HexToAddress(coreContracts.AllocationManager), ethclient)
	if err != nil {
		return nil, fmt.Errorf("failed to create AllocationManager transactor: %w", err)
	}

	delegationManagerTransactor, err := IDelegationManager.NewIDelegationManagerTransactor(common.HexToAddress(coreContracts.DelegationManager), ethclient)
	if err != nil {
		return nil, fmt.Errorf("failed to create DelegationManager transactor: %w", err)
	}

	return &ContractCaller{
		avsRegistrarCaller:          avsRegistrarCaller,
		taskMailboxCaller:           taskMailboxCaller,
		taskMailboxTransactor:       taskMailboxTransactor,
		allocationManagerCaller:     allocationManagerCaller,
		allocationManagerTransactor: allocationManagerTransactor,
		delegationManagerTransactor: delegationManagerTransactor,
		ethclient:                   ethclient,
		coreContracts:               coreContracts,
		config:                      cfg,
		logger:                      logger,
	}, nil
}

func (cc *ContractCaller) buildNoSendOptsWithPrivateKey(ctx context.Context) (*bind.TransactOpts, *ecdsa.PrivateKey, error) {
	privateKey, err := cryptoUtils.StringToECDSAPrivateKey(cc.config.PrivateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	noSendTxOpts, err := cc.buildTxOps(ctx, privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build transaction options: %w", err)
	}
	return noSendTxOpts, privateKey, nil
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

func (cc *ContractCaller) SubmitTaskResult(ctx context.Context, aggCert *aggregation.AggregatedCertificate) (*types.Receipt, error) {
	noSendTxOpts, privateKey, err := cc.buildNoSendOptsWithPrivateKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction options: %w", err)
	}

	if len(aggCert.TaskId) != 32 {
		return nil, fmt.Errorf("taskId must be 32 bytes, got %d", len(aggCert.TaskId))
	}
	var taskId [32]byte
	copy(taskId[:], aggCert.TaskId)
	cc.logger.Sugar().Infow("submitting task result", "taskId", taskId)

	// Convert signature to G1 point in precompile format
	g1Point := &bn254.G1Point{
		G1Affine: aggCert.SignersSignature.GetG1Point(),
	}
	g1Bytes, err := g1Point.ToPrecompileFormat()
	if err != nil {
		return nil, fmt.Errorf("signature not in correct subgroup: %w", err)
	}

	// Convert public key to G2 point in precompile format
	g2Bytes, err := aggCert.SignersPublicKey.ToPrecompileFormat()
	if err != nil {
		return nil, fmt.Errorf("public key not in correct subgroup: %w", err)
	}

	var digest [32]byte
	copy(digest[:], aggCert.TaskResponseDigest)

	cert := ITaskMailbox.IBN254CertificateVerifierBN254Certificate{
		ReferenceTimestamp: uint32(aggCert.SignedAt.Unix()),
		MessageHash:        digest,
		Sig: ITaskMailbox.BN254G1Point{
			X: new(big.Int).SetBytes(g1Bytes[0:32]),
			Y: new(big.Int).SetBytes(g1Bytes[32:64]),
		},
		Apk: ITaskMailbox.BN254G2Point{
			X: [2]*big.Int{
				new(big.Int).SetBytes(g2Bytes[0:32]),
				new(big.Int).SetBytes(g2Bytes[32:64]),
			},
			Y: [2]*big.Int{
				new(big.Int).SetBytes(g2Bytes[64:96]),
				new(big.Int).SetBytes(g2Bytes[96:128]),
			},
		},
		// TODO: technically these are all empty since we default to needing all operators to sign.
		NonsignerIndices:   []uint32{},
		NonSignerWitnesses: []ITaskMailbox.IBN254CertificateVerifierBN254OperatorInfoWitness{},
	}
	fmt.Printf("taskId: %v\n", taskId)
	fmt.Printf("Submitting task: %+v\n", cert)

	tx, err := cc.taskMailboxTransactor.SubmitResult(noSendTxOpts, taskId, cert, aggCert.TaskResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	return cc.EstimateGasPriceAndLimitAndSendTx(ctx, noSendTxOpts.From, tx, privateKey, "SubmitTaskSession")
}

//nolint:unused
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

	allMembers := make([]*peering.OperatorPeerInfo, 0)
	for i, pm := range peerMembers {
		pubKey, err := bn254.NewPublicKeyFromSolidity(pm.PubkeyInfo.PubkeyG1, pm.PubkeyInfo.PubkeyG2)
		if err != nil {
			cc.logger.Sugar().Errorf("failed to convert public key: %v", err)
			return nil, err
		}

		allMembers = append(allMembers, &peering.OperatorPeerInfo{
			NetworkAddress:  pm.Socket,
			PublicKey:       pubKey,
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

func (cc *ContractCaller) PublishMessageToInbox(ctx context.Context, avsAddress string, operatorSetId uint32, payload []byte) (*types.Receipt, error) {
	privateKey, err := cryptoUtils.StringToECDSAPrivateKey(cc.config.PrivateKey)
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

func (cc *ContractCaller) GetOperatorRegistrationMessageHash(ctx context.Context, operatorAddress common.Address) (ITaskAVSRegistrar.BN254G1Point, error) {
	return cc.avsRegistrarCaller.PubkeyRegistrationMessageHash(&bind.CallOpts{
		Context: ctx,
	}, operatorAddress)
}

func (cc *ContractCaller) createOperator(ctx context.Context, operatorAddress common.Address, allocationDelay uint32, metadataUri string) (*types.Receipt, error) {
	noSendTxOpts, privateKey, err := cc.buildNoSendOptsWithPrivateKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction options: %w", err)
	}

	tx, err := cc.delegationManagerTransactor.RegisterAsOperator(
		noSendTxOpts,
		operatorAddress,
		allocationDelay,
		metadataUri,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	return cc.EstimateGasPriceAndLimitAndSendTx(ctx, noSendTxOpts.From, tx, privateKey, "RegisterAsOperator")
}

func (cc *ContractCaller) CreateOperatorRegistrationPayload(
	publicKey *bn254.PublicKey,
	signature *bn254.Signature,
	socket string,
) ([]byte, error) {
	// Convert G2 point to precompile format
	g2Point := &bn254.G2Point{
		G2Affine: publicKey.GetG2Point(),
	}
	g2Bytes, err := g2Point.ToPrecompileFormat()
	if err != nil {
		return nil, fmt.Errorf("public key not in correct subgroup: %w", err)
	}

	// Convert G1 point to precompile format
	g1Point := &bn254.G1Point{
		G1Affine: signature.GetG1Point(),
	}
	g1Bytes, err := g1Point.ToPrecompileFormat()
	if err != nil {
		return nil, fmt.Errorf("signature not in correct subgroup: %w", err)
	}

	registrationPayload := ITaskAVSRegistrar.ITaskAVSRegistrarTypesPubkeyRegistrationParams{
		PubkeyRegistrationSignature: ITaskAVSRegistrar.BN254G1Point{
			X: new(big.Int).SetBytes(g1Bytes[0:32]),
			Y: new(big.Int).SetBytes(g1Bytes[32:64]),
		},
		PubkeyG1: ITaskAVSRegistrar.BN254G1Point{
			X: publicKey.GetG1Point().X.BigInt(new(big.Int)),
			Y: publicKey.GetG1Point().Y.BigInt(new(big.Int)),
		},
		PubkeyG2: ITaskAVSRegistrar.BN254G2Point{
			X: [2]*big.Int{
				new(big.Int).SetBytes(g2Bytes[0:32]),
				new(big.Int).SetBytes(g2Bytes[32:64]),
			},
			Y: [2]*big.Int{
				new(big.Int).SetBytes(g2Bytes[64:96]),
				new(big.Int).SetBytes(g2Bytes[96:128]),
			},
		},
	}

	return cc.avsRegistrarCaller.PackRegisterPayload(&bind.CallOpts{}, socket, registrationPayload)
}

func (cc *ContractCaller) registerOperatorWithAvs(
	ctx context.Context,
	operatorAddress common.Address,
	avsAddress common.Address,
	operatorSetIds []uint32,
	publicKey *bn254.PublicKey,
	signature *bn254.Signature,
	socket string,
) (*types.Receipt, error) {
	noSendTxOpts, privateKey, err := cc.buildNoSendOptsWithPrivateKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction options: %w", err)
	}

	packedBytes, err := cc.CreateOperatorRegistrationPayload(publicKey, signature, socket)
	if err != nil {
		return nil, fmt.Errorf("failed to create operator registration payload: %w", err)
	}

	tx, err := cc.allocationManagerTransactor.RegisterForOperatorSets(noSendTxOpts, operatorAddress, IAllocationManager.IAllocationManagerTypesRegisterParams{
		Avs:            avsAddress,
		OperatorSetIds: operatorSetIds,
		Data:           packedBytes,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	return cc.EstimateGasPriceAndLimitAndSendTx(ctx, noSendTxOpts.From, tx, privateKey, "RegisterForOperatorSets")
}

func (cc *ContractCaller) CreateOperatorAndRegisterWithAvs(
	ctx context.Context,
	avsAddress common.Address,
	operatorAddress common.Address,
	operatorSetIds []uint32,
	publicKey *bn254.PublicKey,
	signature *bn254.Signature,
	socket string,
	allocationDelay uint32,
	metadataUri string,
) (*types.Receipt, error) {
	createdOperator, err := cc.createOperator(ctx, operatorAddress, allocationDelay, metadataUri)
	if err != nil {
		return nil, fmt.Errorf("failed to register as operator: %w", err)
	}
	cc.logger.Sugar().Infow("Successfully registered as operator",
		zap.Any("receipt", createdOperator),
	)
	cc.logger.Sugar().Infow("Registering operator with AVS")

	return cc.registerOperatorWithAvs(ctx, operatorAddress, avsAddress, operatorSetIds, publicKey, signature, socket)
}

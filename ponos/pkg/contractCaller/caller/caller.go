package caller

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/IAllocationManager"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/ICrossChainRegistry"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/IDelegationManager"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/IKeyRegistrar"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/IOperatorTableCalculator"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/IOperatorTableUpdater"
	"github.com/Layr-Labs/hourglass-monorepo/contracts/pkg/bindings/ITaskMailbox"
	"github.com/Layr-Labs/hourglass-monorepo/contracts/pkg/bindings/TaskAVSRegistrarBase"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller"
	cryptoUtils "github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/crypto"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/aggregation"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
	"math/big"
	"time"
)

type ContractCallerConfig struct {
	PrivateKey                string
	AVSRegistrarAddress       string
	TaskMailboxAddress        string
	KeyRegistrarAddress       string
	CrossChainRegistryAddress string
}

type ContractCaller struct {
	avsRegistrarCaller *TaskAVSRegistrarBase.TaskAVSRegistrarBaseCaller
	taskMailbox        *ITaskMailbox.ITaskMailbox
	allocationManager  *IAllocationManager.IAllocationManager
	delegationManager  *IDelegationManager.IDelegationManager
	crossChainRegistry *ICrossChainRegistry.ICrossChainRegistry
	keyRegistrar       *IKeyRegistrar.IKeyRegistrar
	ethclient          *ethclient.Client
	config             *ContractCallerConfig
	logger             *zap.Logger
	coreContracts      *config.CoreContractAddresses
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
	logger.Sugar().Infow("Creating contract caller",
		zap.String("AVSRegistrarAddress", cfg.AVSRegistrarAddress),
		zap.String("TaskMailboxAddress", cfg.TaskMailboxAddress),
	)
	avsRegistrarCaller, err := TaskAVSRegistrarBase.NewTaskAVSRegistrarBaseCaller(common.HexToAddress(cfg.AVSRegistrarAddress), ethclient)
	if err != nil {
		return nil, fmt.Errorf("failed to create AVSRegistrar caller: %w", err)
	}

	taskMailbox, err := ITaskMailbox.NewITaskMailbox(common.HexToAddress(cfg.TaskMailboxAddress), ethclient)
	if err != nil {
		return nil, fmt.Errorf("failed to create TaskMailbox: %w", err)
	}

	keyRegistrar, err := IKeyRegistrar.NewIKeyRegistrar(common.HexToAddress(cfg.KeyRegistrarAddress), ethclient)
	if err != nil {
		return nil, fmt.Errorf("failed to create KeyRegistrar: %w", err)
	}

	chainId, err := ethclient.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	coreContracts, err := config.GetCoreContractsForChainId(config.ChainId(chainId.Uint64()))
	if err != nil {
		return nil, fmt.Errorf("failed to get core contracts: %w", err)
	}

	allocationManager, err := IAllocationManager.NewIAllocationManager(common.HexToAddress(coreContracts.AllocationManager), ethclient)
	if err != nil {
		return nil, fmt.Errorf("failed to create AllocationManager: %w", err)
	}

	crossChainRegistry, err := ICrossChainRegistry.NewICrossChainRegistry(common.HexToAddress(cfg.CrossChainRegistryAddress), ethclient)
	if err != nil {
		return nil, fmt.Errorf("failed to create CrossChainRegistry: %w", err)
	}

	delegationManager, err := IDelegationManager.NewIDelegationManager(common.HexToAddress(coreContracts.DelegationManager), ethclient)
	if err != nil {
		return nil, fmt.Errorf("failed to create DelegationManager transactor: %w", err)
	}

	return &ContractCaller{
		avsRegistrarCaller: avsRegistrarCaller,
		taskMailbox:        taskMailbox,
		allocationManager:  allocationManager,
		keyRegistrar:       keyRegistrar,
		delegationManager:  delegationManager,
		crossChainRegistry: crossChainRegistry,
		ethclient:          ethclient,
		coreContracts:      coreContracts,
		config:             cfg,
		logger:             logger,
	}, nil
}

func (cc *ContractCaller) SubmitTaskResultRetryable(
	ctx context.Context,
	aggCert *aggregation.AggregatedCertificate,
	globalTableRootReferenceTimestamp uint32,
) (*types.Receipt, error) {
	backoffs := []int{1, 3, 5, 10, 20}
	for i, backoff := range backoffs {
		res, err := cc.SubmitTaskResult(ctx, aggCert, globalTableRootReferenceTimestamp)
		if err != nil {
			if i == len(backoffs)-1 {
				cc.logger.Sugar().Errorw("failed to submit task result after retries",
					zap.String("taskId", hexutil.Encode(aggCert.TaskId)),
					zap.Error(err),
				)
				return nil, fmt.Errorf("failed to submit task result: %w", err)
			}
			cc.logger.Sugar().Errorw("failed to submit task result, retrying",
				zap.String("taskId", hexutil.Encode(aggCert.TaskId)),
				zap.Int("attempt", i+1),
			)
			time.Sleep(time.Second * time.Duration(backoff))
			continue
		}
		return res, nil
	}
	return nil, fmt.Errorf("failed to submit task result after retries")
}

func (cc *ContractCaller) SubmitTaskResult(
	ctx context.Context,
	aggCert *aggregation.AggregatedCertificate,
	globalTableRootReferenceTimestamp uint32,
) (*types.Receipt, error) {
	noSendTxOpts, privateKey, err := cc.buildNoSendOptsWithPrivateKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction options: %w", err)
	}

	if len(aggCert.TaskId) != 32 {
		return nil, fmt.Errorf("taskId must be 32 bytes, got %d", len(aggCert.TaskId))
	}
	var taskId [32]byte
	copy(taskId[:], aggCert.TaskId)
	cc.logger.Sugar().Infow("submitting task result",
		zap.String("taskId", hexutil.Encode(taskId[:])),
		zap.String("mailboxAddress", cc.config.TaskMailboxAddress),
	)

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

	cert := ITaskMailbox.IBN254CertificateVerifierTypesBN254Certificate{
		ReferenceTimestamp: globalTableRootReferenceTimestamp,
		MessageHash:        digest,
		Signature: ITaskMailbox.BN254G1Point{
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
		NonSignerWitnesses: []ITaskMailbox.IBN254CertificateVerifierTypesBN254OperatorInfoWitness{},
	}

	tx, err := cc.taskMailbox.SubmitResult(noSendTxOpts, taskId, cert, aggCert.TaskResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	return cc.EstimateGasPriceAndLimitAndSendTx(ctx, noSendTxOpts.From, tx, privateKey, "SubmitTaskSession")
}

func (cc *ContractCaller) GetOperatorSetMembersWithPeering(
	avsAddress string,
	operatorSetId uint32,
) ([]*peering.OperatorPeerInfo, error) {
	operatorSetStringAddrs, err := cc.getOperatorSetMembers(avsAddress, operatorSetId)
	if err != nil {
		return nil, err
	}

	operatorSetMemberAddrs := util.Map(operatorSetStringAddrs, func(address string, i uint64) common.Address {
		return common.HexToAddress(address)
	})

	allMembers := make([]*peering.OperatorPeerInfo, 0)
	for i, member := range operatorSetMemberAddrs {
		operatorSetInfo, err := cc.GetOperatorSetDetailsForOperator(member, avsAddress, operatorSetId)
		if err != nil {
			cc.logger.Sugar().Errorf("failed to get operator set details for operator %s: %v", member.Hex(), err)
			return nil, err
		}

		allMembers = append(allMembers, &peering.OperatorPeerInfo{
			OperatorAddress: operatorSetStringAddrs[i],
			OperatorSets:    []*peering.OperatorSet{operatorSetInfo},
		})
	}
	return allMembers, nil
}

func (cc *ContractCaller) GetOperatorSetDetailsForOperator(operatorAddress common.Address, avsAddress string, operatorSetId uint32) (*peering.OperatorSet, error) {
	opset := IKeyRegistrar.OperatorSet{
		Avs: common.HexToAddress(avsAddress),
		Id:  operatorSetId,
	}
	curveType, err := cc.keyRegistrar.GetOperatorSetCurveType(&bind.CallOpts{}, opset)
	if err != nil {
		cc.logger.Sugar().Errorf("failed to get operator set curve type: %v", err)
		return nil, err
	}
	// bn254 curve is the only supported curve type for now
	if curveType != 2 {
		return nil, fmt.Errorf("unsupported curve type %d for operator set %d", curveType, operatorSetId)
	}

	solidityPubKey, err := cc.keyRegistrar.GetBN254Key(&bind.CallOpts{}, opset, operatorAddress)
	if err != nil {
		cc.logger.Sugar().Errorf("failed to get operator set public key: %v", err)
		return nil, err
	}

	pubKey, err := bn254.NewPublicKeyFromSolidity(
		&bn254.SolidityBN254G1Point{
			X: solidityPubKey.G1Point.X,
			Y: solidityPubKey.G1Point.Y,
		},
		&bn254.SolidityBN254G2Point{
			X: [2]*big.Int{
				solidityPubKey.G2Point.X[0],
				solidityPubKey.G2Point.X[1],
			},
			Y: [2]*big.Int{
				solidityPubKey.G2Point.Y[0],
				solidityPubKey.G2Point.Y[1],
			},
		},
	)
	if err != nil {
		cc.logger.Sugar().Errorf("failed to convert public key: %v", err)
		return nil, err
	}

	socket, err := cc.avsRegistrarCaller.GetOperatorSocket(&bind.CallOpts{}, operatorAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get operator socket: %w", err)
	}

	return &peering.OperatorSet{
		OperatorSetID:  operatorSetId,
		PublicKey:      pubKey,
		NetworkAddress: socket,
	}, nil
}

func (cc *ContractCaller) GetAVSConfig(avsAddress string) (*contractCaller.AVSConfig, error) {
	avsAddr := common.HexToAddress(avsAddress)
	avsConfig, err := cc.taskMailbox.GetAvsConfig(&bind.CallOpts{}, avsAddr)
	if err != nil {
		return nil, err
	}

	return &contractCaller.AVSConfig{
		AggregatorOperatorSetId: avsConfig.AggregatorOperatorSetId,
		ExecutorOperatorSetIds:  avsConfig.ExecutorOperatorSetIds,
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

	tx, err := cc.taskMailbox.CreateTask(noSendTxOpts, ITaskMailbox.ITaskMailboxTypesTaskParams{
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

func (cc *ContractCaller) GetOperatorRegistrationMessageHash(
	ctx context.Context,
	operatorAddress common.Address,
	avsAddress common.Address,
	operatorSetId uint32,
	keyData []byte,
) ([32]byte, error) {
	return cc.keyRegistrar.GetBN254KeyRegistrationMessageHash(&bind.CallOpts{Context: ctx}, operatorAddress, IKeyRegistrar.OperatorSet{
		Avs: avsAddress,
		Id:  operatorSetId,
	}, keyData)
}

func (cc *ContractCaller) EncodeBN254KeyData(pubKey *bn254.PublicKey) ([]byte, error) {
	// Convert G1 point
	g1Point := &bn254.G1Point{
		G1Affine: pubKey.GetG1Point(),
	}
	g1Bytes, err := g1Point.ToPrecompileFormat()
	if err != nil {
		return nil, fmt.Errorf("public key not in correct subgroup: %w", err)
	}

	keyRegG1 := IKeyRegistrar.BN254G1Point{
		X: new(big.Int).SetBytes(g1Bytes[0:32]),
		Y: new(big.Int).SetBytes(g1Bytes[32:64]),
	}

	g2Point := bn254.NewZeroG2Point().AddPublicKey(pubKey)
	g2Bytes, err := g2Point.ToPrecompileFormat()
	if err != nil {
		return nil, fmt.Errorf("public key not in correct subgroup: %w", err)
	}
	// Convert to IKeyRegistrar G2 point format
	keyRegG2 := IKeyRegistrar.BN254G2Point{
		X: [2]*big.Int{
			new(big.Int).SetBytes(g2Bytes[0:32]),
			new(big.Int).SetBytes(g2Bytes[32:64]),
		},
		Y: [2]*big.Int{
			new(big.Int).SetBytes(g2Bytes[64:96]),
			new(big.Int).SetBytes(g2Bytes[96:128]),
		},
	}

	return cc.keyRegistrar.EncodeBN254KeyData(
		&bind.CallOpts{},
		keyRegG1,
		keyRegG2,
	)
}

func (cc *ContractCaller) CreateOperatorRegistrationPayload(
	publicKey *bn254.PublicKey,
	signature *bn254.Signature,
	socket string,
) ([]byte, error) {
	return nil, nil
}

// ConfigureAVSOperatorSet is called on the KeyRegistry to configure an operator set for a given AVS,
// including specifying which curve type to use for the certificate verifier.
// NOTE: this needs to be called by the AVS
func (cc *ContractCaller) ConfigureAVSOperatorSet(ctx context.Context, avsAddress common.Address, operatorSetId uint32, curveType contractCaller.CurveType) (*types.Receipt, error) {
	noSendTxOpts, privateKey, err := cc.buildNoSendOptsWithPrivateKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction options: %w", err)
	}

	tx, err := cc.keyRegistrar.ConfigureOperatorSet(
		noSendTxOpts,
		IKeyRegistrar.OperatorSet{
			Avs: avsAddress,
			Id:  operatorSetId,
		},
		uint8(curveType),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	return cc.EstimateGasPriceAndLimitAndSendTx(ctx, noSendTxOpts.From, tx, privateKey, "ConfigureOperatorSet")
}

func (cc *ContractCaller) RegisterKeyWithKeyRegistrar(
	ctx context.Context,
	operatorAddress common.Address,
	avsAddress common.Address,
	operatorSetId uint32,
	signature *bn254.Signature,
	keyData []byte,
) (*types.Receipt, error) {
	noSendTxOpts, privateKey, err := cc.buildNoSendOptsWithPrivateKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction options: %w", err)
	}

	// Create operator set struct
	operatorSet := IKeyRegistrar.OperatorSet{
		Avs: avsAddress,
		Id:  operatorSetId,
	}

	g1Point := &bn254.G1Point{
		G1Affine: signature.GetG1Point(),
	}
	g1Bytes, err := g1Point.ToPrecompileFormat()
	if err != nil {
		return nil, fmt.Errorf("signature not in correct subgroup: %w", err)
	}

	cc.logger.Sugar().Debugw("Registering key with KeyRegistrar",
		"operatorAddress:", operatorAddress.String(),
		"avsAddress:", avsAddress.String(),
		"operatorSetId:", operatorSetId,
		"keyData", hexutil.Encode(keyData),
		"g1Bytes:", hexutil.Encode(g1Bytes),
	)

	tx, err := cc.keyRegistrar.RegisterKey(
		noSendTxOpts,
		operatorAddress,
		operatorSet,
		keyData,
		g1Bytes,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register key: %w", err)
	}

	return cc.EstimateGasPriceAndLimitAndSendTx(ctx, noSendTxOpts.From, tx, privateKey, "RegisterKey")
}

func (cc *ContractCaller) CreateOperatorAndRegisterWithAvs(
	ctx context.Context,
	avsAddress common.Address,
	operatorAddress common.Address,
	operatorSetIds []uint32,
	socket string,
	allocationDelay uint32,
	metadataUri string,
) (*types.Receipt, error) {
	createdOperator, err := cc.createOperator(ctx, operatorAddress, allocationDelay, metadataUri)
	if err != nil {
		return nil, fmt.Errorf("failed to register as operator: %w", err)
	}
	cc.logger.Sugar().Infow("Successfully created operator",
		zap.Any("receipt", createdOperator),
	)

	// Register socket with AVS
	cc.logger.Sugar().Infow("Registering operator socket with AVS")
	socketReceipt, err := cc.registerOperatorWithAvs(ctx, operatorAddress, avsAddress, operatorSetIds, socket)
	if err != nil {
		return nil, fmt.Errorf("failed to register operator socket with AVS: %w", err)
	}
	cc.logger.Sugar().Infow("Successfully registered operator socket with AVS",
		zap.Any("receipt", socketReceipt),
	)

	// Return the socket registration receipt as the primary receipt
	return socketReceipt, nil
}

func (cc *ContractCaller) getOperatorSetMembers(avsAddress string, operatorSetId uint32) ([]string, error) {
	avsAddr := common.HexToAddress(avsAddress)
	operatorSet, err := cc.allocationManager.GetMembers(&bind.CallOpts{}, IAllocationManager.OperatorSet{
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

func (cc *ContractCaller) createOperator(ctx context.Context, operatorAddress common.Address, allocationDelay uint32, metadataUri string) (*types.Receipt, error) {
	noSendTxOpts, privateKey, err := cc.buildNoSendOptsWithPrivateKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction options: %w", err)
	}

	exists, err := cc.delegationManager.IsOperator(&bind.CallOpts{}, operatorAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to check if operator exists: %w", err)
	}
	if exists {
		cc.logger.Sugar().Infow("Operator already exists, skipping creation",
			zap.String("operatorAddress", operatorAddress.String()),
		)
		return nil, nil
	}

	tx, err := cc.delegationManager.RegisterAsOperator(
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

func encodeString(str string) ([]byte, error) {
	// Define the ABI for a single string parameter
	stringType, _ := abi.NewType("string", "", nil)
	arguments := abi.Arguments{{Type: stringType}}

	// Encode the string
	encoded, err := arguments.Pack(str)
	if err != nil {
		return nil, err
	}

	return encoded, nil
}

func (cc *ContractCaller) registerOperatorWithAvs(
	ctx context.Context,
	operatorAddress common.Address,
	avsAddress common.Address,
	operatorSetIds []uint32,
	socket string,
) (*types.Receipt, error) {
	noSendTxOpts, privateKey, err := cc.buildNoSendOptsWithPrivateKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction options: %w", err)
	}

	encodedSocket, err := encodeString(socket)
	if err != nil {
		return nil, fmt.Errorf("failed to encode socket string: %w", err)
	}

	tx, err := cc.allocationManager.RegisterForOperatorSets(noSendTxOpts, operatorAddress, IAllocationManager.IAllocationManagerTypesRegisterParams{
		Avs:            avsAddress,
		OperatorSetIds: operatorSetIds,
		Data:           encodedSocket,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	return cc.EstimateGasPriceAndLimitAndSendTx(ctx, noSendTxOpts.From, tx, privateKey, "RegisterForOperatorSets")
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

func (cc *ContractCaller) GetSupportChainsForMultichain(ctx context.Context, referenceBlockNumber uint64) ([]*big.Int, []common.Address, error) {
	return cc.crossChainRegistry.GetSupportedChains(&bind.CallOpts{
		Context:     ctx,
		BlockNumber: new(big.Int).SetUint64(referenceBlockNumber),
	})
}

func (cc *ContractCaller) GetOperatorTableDataForOperatorSet(
	ctx context.Context,
	avsAddress common.Address,
	operatorSetId uint32,
	chainId config.ChainId,
	referenceBlocknumber uint64,
) (*contractCaller.OperatorTableData, error) {
	operatorSet := ICrossChainRegistry.OperatorSet{
		Avs: avsAddress,
		Id:  operatorSetId,
	}
	cc.logger.Sugar().Infow("Fetching operator table data",
		zap.String("avsAddress", avsAddress.String()),
		zap.Uint32("operatorSetId", operatorSetId),
	)
	otcAddr, err := cc.crossChainRegistry.GetOperatorTableCalculator(&bind.CallOpts{
		Context:     ctx,
		BlockNumber: new(big.Int).SetUint64(referenceBlocknumber),
	}, operatorSet)
	if err != nil {
		return nil, fmt.Errorf("failed to get operator table calculator address: %w", err)
	}

	cc.logger.Sugar().Infow("Operator table calculator address",
		zap.String("operatorTableCalculatorAddress", otcAddr.String()),
	)
	opTableCalculator, err := IOperatorTableCalculator.NewIOperatorTableCalculatorCaller(otcAddr, cc.ethclient)
	if err != nil {
		return nil, fmt.Errorf("failed to create operator table calculator caller: %w", err)
	}

	cc.logger.Sugar().Infow("Fetching operator weights for operator set",
		zap.String("avsAddress", avsAddress.String()),
		zap.Uint32("operatorSetId", operatorSetId),
	)
	operatorWeights, err := opTableCalculator.GetOperatorWeights(&bind.CallOpts{
		Context:     ctx,
		BlockNumber: new(big.Int).SetUint64(referenceBlocknumber),
	}, IOperatorTableCalculator.OperatorSet(operatorSet))
	if err != nil {
		return nil, fmt.Errorf("failed to get operator weights: %w", err)
	}

	cc.logger.Sugar().Infow("Fetching supported chains for multichain")
	chainIds, tableUpdaterAddresses, err := cc.GetSupportChainsForMultichain(ctx, referenceBlocknumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get supported chains for multichain: %w", err)
	}
	cc.logger.Sugar().Infow("Supported chains for multichain",
		zap.Any("chains", chainIds),
		zap.Any("tableUpdaterAddresses", tableUpdaterAddresses),
	)

	var tableUpdaterAddr common.Address
	for i, id := range chainIds {
		if id.Uint64() == uint64(chainId) {
			tableUpdaterAddr = tableUpdaterAddresses[i]
			break
		}
	}
	if tableUpdaterAddr == (common.Address{}) {
		return nil, fmt.Errorf("no table updater address found for chain ID %d", chainId)
	}

	tableUpdater, err := IOperatorTableUpdater.NewIOperatorTableUpdater(tableUpdaterAddr, cc.ethclient)
	if err != nil {
		return nil, fmt.Errorf("failed to create operator table updater: %w", err)
	}

	latestReferenceTimestamp, err := tableUpdater.GetLatestReferenceTimestamp(&bind.CallOpts{
		Context:     ctx,
		BlockNumber: new(big.Int).SetUint64(referenceBlocknumber),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get latest reference timestamp: %w", err)
	}

	return &contractCaller.OperatorTableData{
		OperatorWeights:          operatorWeights.Weights,
		Operators:                operatorWeights.Operators,
		LatestReferenceTimestamp: latestReferenceTimestamp,
	}, nil
}

func (cc *ContractCaller) GetActiveGenerationReservations() ([]ICrossChainRegistry.OperatorSet, error) {
	return cc.crossChainRegistry.GetActiveGenerationReservations(&bind.CallOpts{})
}

package caller

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/IAllocationManager"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/IBN254CertificateVerifier"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/ICrossChainRegistry"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/IDelegationManager"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/IECDSACertificateVerifier"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/IKeyRegistrar"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/IOperatorTableCalculator"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/IOperatorTableUpdater"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/ITaskMailbox"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/middleware-bindings/ITaskAVSRegistrarBase"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/middleware-bindings/TaskAVSRegistrarBase"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/aggregation"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionSigner"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

type ContractCaller struct {
	taskMailbox        *ITaskMailbox.ITaskMailbox
	allocationManager  *IAllocationManager.IAllocationManager
	delegationManager  *IDelegationManager.IDelegationManager
	crossChainRegistry *ICrossChainRegistry.ICrossChainRegistry
	keyRegistrar       *IKeyRegistrar.IKeyRegistrar
	ecdsaCertVerifier  *IECDSACertificateVerifier.IECDSACertificateVerifier
	bn254CertVerifier  *IBN254CertificateVerifier.IBN254CertificateVerifier
	ethclient          *ethclient.Client
	logger             *zap.Logger
	coreContracts      *config.CoreContractAddresses
	signer             transactionSigner.ITransactionSigner
}

func NewContractCallerFromEthereumClient(
	ethClient *ethereum.Client,
	signer transactionSigner.ITransactionSigner,
	logger *zap.Logger,
) (*ContractCaller, error) {
	client, err := ethClient.GetEthereumContractCaller()
	if err != nil {
		return nil, err
	}

	return NewContractCaller(client, signer, logger)
}

func NewContractCaller(
	ethclient *ethclient.Client,
	signer transactionSigner.ITransactionSigner,
	logger *zap.Logger,
) (*ContractCaller, error) {
	logger.Sugar().Debugw("Creating contract caller")

	chainId, err := ethclient.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	coreContracts, err := config.GetCoreContractsForChainId(config.ChainId(chainId.Uint64()))
	if err != nil {
		return nil, fmt.Errorf("failed to get core contracts: %w", err)
	}

	keyRegistrar, err := IKeyRegistrar.NewIKeyRegistrar(common.HexToAddress(coreContracts.KeyRegistrar), ethclient)
	if err != nil {
		return nil, fmt.Errorf("failed to create KeyRegistrar: %w", err)
	}

	allocationManager, err := IAllocationManager.NewIAllocationManager(common.HexToAddress(coreContracts.AllocationManager), ethclient)
	if err != nil {
		return nil, fmt.Errorf("failed to create AllocationManager: %w", err)
	}

	crossChainRegistry, err := ICrossChainRegistry.NewICrossChainRegistry(common.HexToAddress(coreContracts.CrossChainRegistry), ethclient)
	if err != nil {
		return nil, fmt.Errorf("failed to create CrossChainRegistry: %w", err)
	}

	delegationManager, err := IDelegationManager.NewIDelegationManager(common.HexToAddress(coreContracts.DelegationManager), ethclient)
	if err != nil {
		return nil, fmt.Errorf("failed to create DelegationManager transactor: %w", err)
	}

	ecdsaCertVerifier, err := IECDSACertificateVerifier.NewIECDSACertificateVerifier(common.HexToAddress(coreContracts.ECDSACertificateVerifier), ethclient)
	if err != nil {
		return nil, fmt.Errorf("failed to create ECDSACertificateVerifier: %w", err)
	}

	bn254CertVerifier, err := IBN254CertificateVerifier.NewIBN254CertificateVerifier(common.HexToAddress(coreContracts.BN254CertificateVerifier), ethclient)
	if err != nil {
		return nil, fmt.Errorf("failed to create BN254CertificateVerifier: %w", err)
	}

	taskMailbox, err := ITaskMailbox.NewITaskMailbox(common.HexToAddress(coreContracts.TaskMailbox), ethclient)
	if err != nil {
		return nil, fmt.Errorf("failed to create TaskMailbox: %w", err)
	}

	return &ContractCaller{
		taskMailbox:        taskMailbox,
		allocationManager:  allocationManager,
		keyRegistrar:       keyRegistrar,
		delegationManager:  delegationManager,
		crossChainRegistry: crossChainRegistry,
		ecdsaCertVerifier:  ecdsaCertVerifier,
		bn254CertVerifier:  bn254CertVerifier,
		ethclient:          ethclient,
		coreContracts:      coreContracts,
		logger:             logger,
		signer:             signer,
	}, nil
}

func (cc *ContractCaller) SubmitBN254TaskResultRetryable(
	ctx context.Context,
	aggCert *aggregation.AggregatedBN254Certificate,
	globalTableRootReferenceTimestamp uint32,
) (*types.Receipt, error) {
	backoffs := []int{1, 3, 5, 10, 20}
	for i, backoff := range backoffs {
		res, err := cc.SubmitBN254TaskResult(ctx, aggCert, globalTableRootReferenceTimestamp)
		if err != nil {
			if i == len(backoffs)-1 {
				cc.logger.Sugar().Errorw("failed to submit task result after retries",
					zap.String("taskId", hexutil.Encode(aggCert.TaskId)),
					zap.Error(err),
				)
				return nil, fmt.Errorf("failed to submit task result: %w", err)
			}
			cc.logger.Sugar().Errorw("failed to submit task result, retrying",
				zap.Error(err),
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

func (cc *ContractCaller) SubmitBN254TaskResult(
	ctx context.Context,
	aggCert *aggregation.AggregatedBN254Certificate,
	globalTableRootReferenceTimestamp uint32,
) (*types.Receipt, error) {
	noSendTxOpts, err := cc.buildTransactionOpts(ctx)
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
		zap.String("mailboxAddress", cc.coreContracts.TaskMailbox),
		zap.Uint32("globalTableRootReferenceTimestamp", globalTableRootReferenceTimestamp),
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

	digest := aggCert.TaskResponseDigest

	// Populate NonSignerWitnesses from the sorted non-signer operators
	nonSignerWitnesses := make([]ITaskMailbox.IBN254CertificateVerifierTypesBN254OperatorInfoWitness, 0, len(aggCert.NonSignerOperators))
	for _, nonSigner := range aggCert.NonSignerOperators {
		// For now, we only provide the operator index
		// The contract can look up cached operator info or we can provide proofs later
		witness := ITaskMailbox.IBN254CertificateVerifierTypesBN254OperatorInfoWitness{
			OperatorIndex: nonSigner.OperatorIndex,
			// OperatorInfoProof and OperatorInfo can be empty if the operator is already cached
			// in the contract from previous verifications
			OperatorInfoProof: []byte{},
			OperatorInfo:      ITaskMailbox.IOperatorTableCalculatorTypesBN254OperatorInfo{
				// Empty for now - contract will use cached data
			},
		}
		nonSignerWitnesses = append(nonSignerWitnesses, witness)
	}

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
		NonSignerWitnesses: nonSignerWitnesses,
	}

	certBytes, err := cc.taskMailbox.GetBN254CertificateBytes(&bind.CallOpts{}, cert)
	if err != nil {
		return nil, fmt.Errorf("failed to get BN254 certificate bytes: %w", err)
	}

	tx, err := cc.taskMailbox.SubmitResult(noSendTxOpts, taskId, certBytes, aggCert.TaskResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	return cc.signAndSendTransaction(ctx, tx, "SubmitTaskSession")
}

func (cc *ContractCaller) SubmitECDSATaskResultRetryable(
	ctx context.Context,
	aggCert *aggregation.AggregatedECDSACertificate,
	globalTableRootReferenceTimestamp uint32,
) (*types.Receipt, error) {
	backoffs := []int{1, 3, 5, 10, 20}
	for i, backoff := range backoffs {
		res, err := cc.SubmitECDSATaskResult(ctx, aggCert, globalTableRootReferenceTimestamp)
		if err != nil {
			if i == len(backoffs)-1 {
				cc.logger.Sugar().Errorw("failed to submit task result after retries",
					zap.String("taskId", hexutil.Encode(aggCert.TaskId)),
					zap.Error(err),
				)
				return nil, fmt.Errorf("failed to submit task result: %w", err)
			}
			cc.logger.Sugar().Errorw("failed to submit task result, retrying",
				zap.Error(err),
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

func (cc *ContractCaller) SubmitECDSATaskResult(
	ctx context.Context,
	aggCert *aggregation.AggregatedECDSACertificate,
	globalTableRootReferenceTimestamp uint32,
) (*types.Receipt, error) {
	noSendTxOpts, err := cc.buildTransactionOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction options: %w", err)
	}

	if len(aggCert.TaskId) != 32 {
		return nil, fmt.Errorf("taskId must be 32 bytes, got %d", len(aggCert.TaskId))
	}
	var taskId [32]byte
	copy(taskId[:], aggCert.TaskId)

	finalSig, err := aggCert.GetFinalSignature()
	if err != nil {
		return nil, fmt.Errorf("failed to get final signature: %w", err)
	}

	cc.logger.Sugar().Infow("submitting task result",
		zap.String("taskId", hexutil.Encode(taskId[:])),
		zap.String("mailboxAddress", cc.coreContracts.TaskMailbox),
		zap.Uint32("globalTableRootReferenceTimestamp", globalTableRootReferenceTimestamp),
		zap.String("finalSig", hexutil.Encode(finalSig[:])),
	)

	cert := ITaskMailbox.IECDSACertificateVerifierTypesECDSACertificate{
		ReferenceTimestamp: globalTableRootReferenceTimestamp,
		MessageHash:        aggCert.GetTaskMessageHash(),
		Sig:                finalSig,
	}

	certBytes, err := cc.taskMailbox.GetECDSACertificateBytes(&bind.CallOpts{}, cert)
	if err != nil {
		return nil, fmt.Errorf("failed to call GetECDSACertificateBytes: %w", err)
	}

	tx, err := cc.taskMailbox.SubmitResult(noSendTxOpts, taskId, certBytes, aggCert.TaskResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	return cc.signAndSendTransaction(ctx, tx, "SubmitTaskSession")
}

func (cc *ContractCaller) CalculateECDSACertificateDigestBytes(ctx context.Context, referenceTimestamp uint32, messageHash [32]byte) ([]byte, error) {
	return cc.ecdsaCertVerifier.CalculateCertificateDigestBytes(&bind.CallOpts{}, referenceTimestamp, messageHash)
}

func (cc *ContractCaller) CalculateBN254CertificateDigestBytes(ctx context.Context, referenceTimestamp uint32, messageHash [32]byte) ([]byte, error) {
	digest, err := cc.bn254CertVerifier.CalculateCertificateDigest(&bind.CallOpts{}, referenceTimestamp, messageHash)
	if err != nil {
		return nil, err
	}
	return digest[:], nil
}

func (cc *ContractCaller) GetExecutorOperatorSetTaskConfig(ctx context.Context, avsAddress common.Address, opsetId uint32) (*contractCaller.TaskMailboxExecutorOperatorSetConfig, error) {
	res, err := cc.taskMailbox.GetExecutorOperatorSetTaskConfig(&bind.CallOpts{}, ITaskMailbox.OperatorSet{
		Avs: avsAddress,
		Id:  opsetId,
	})
	if err != nil {
		return nil, err
	}

	return &contractCaller.TaskMailboxExecutorOperatorSetConfig{
		TaskHook:     res.TaskHook,
		TaskSLA:      res.TaskSLA,
		FeeToken:     res.FeeToken,
		CurveType:    res.CurveType,
		FeeCollector: res.FeeCollector,
		Consensus:    res.Consensus,
		TaskMetadata: res.TaskMetadata,
	}, nil
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
		// Set the operator's index in this operator set based on their position in the members array
		operatorSetInfo.OperatorIndex = uint32(i)

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

	// Get the AVS registrar address from the allocation manager
	avsAddr := common.HexToAddress(avsAddress)
	avsRegistrarAddress, err := cc.allocationManager.GetAVSRegistrar(&bind.CallOpts{}, avsAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to get AVS registrar address: %w", err)
	}

	// Create new registrar caller
	caller, err := TaskAVSRegistrarBase.NewTaskAVSRegistrarBaseCaller(avsRegistrarAddress, cc.ethclient)
	if err != nil {
		return nil, fmt.Errorf("failed to create AVS registrar caller: %w", err)
	}
	socket, err := caller.GetOperatorSocket(&bind.CallOpts{}, operatorAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get operator socket: %w", err)
	}

	curveTypeSolidity, err := cc.keyRegistrar.GetOperatorSetCurveType(&bind.CallOpts{}, opset)
	if err != nil {
		cc.logger.Sugar().Errorf("failed to get operator set curve type: %v", err)
		return nil, err
	}

	curveType, err := config.ConvertSolidityEnumToCurveType(curveTypeSolidity)
	if err != nil {
		cc.logger.Sugar().Errorf("failed to convert curve type: %v", err)
		return nil, fmt.Errorf("failed to convert curve type: %w", err)
	}

	peeringOpset := &peering.OperatorSet{
		OperatorSetID:  operatorSetId,
		NetworkAddress: socket,
		CurveType:      curveType,
	}

	if curveType == config.CurveTypeBN254 {
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
		peeringOpset.WrappedPublicKey = peering.WrappedPublicKey{
			PublicKey: pubKey,
		}
		return peeringOpset, nil
	}

	if curveType == config.CurveTypeECDSA {
		ecdsaAddr, err := cc.keyRegistrar.GetECDSAAddress(&bind.CallOpts{}, opset, operatorAddress)
		if err != nil {
			cc.logger.Sugar().Errorf("failed to get operator set public key: %v", err)
			return nil, err
		}
		peeringOpset.WrappedPublicKey = peering.WrappedPublicKey{
			ECDSAAddress: ecdsaAddr,
		}
		return peeringOpset, nil
	}
	cc.logger.Sugar().Errorf("unsupported curve type: %s", curveType)
	return nil, fmt.Errorf("unsupported curve type: %s", curveType)
}

func (cc *ContractCaller) GetAVSConfig(avsAddress string) (*contractCaller.AVSConfig, error) {
	avsAddr := common.HexToAddress(avsAddress)

	avsRegistrarAddress, err := cc.allocationManager.GetAVSRegistrar(&bind.CallOpts{}, avsAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to get AVS registrar address: %w", err)
	}

	registrarCaller, err := ITaskAVSRegistrarBase.NewITaskAVSRegistrarBaseCaller(avsRegistrarAddress, cc.ethclient)
	if err != nil {
		return nil, fmt.Errorf("failed to create AVS registrar caller: %w", err)
	}

	avsConfig, err := registrarCaller.GetAvsConfig(&bind.CallOpts{})
	if err != nil {
		return nil, err
	}

	return &contractCaller.AVSConfig{
		AggregatorOperatorSetId: avsConfig.AggregatorOperatorSetId,
		ExecutorOperatorSetIds:  avsConfig.ExecutorOperatorSetIds,
	}, nil
}

func (cc *ContractCaller) GetOperatorSetCurveType(avsAddress string, operatorSetId uint32) (config.CurveType, error) {
	curveType, err := cc.keyRegistrar.GetOperatorSetCurveType(&bind.CallOpts{}, IKeyRegistrar.OperatorSet{
		Avs: common.HexToAddress(avsAddress),
		Id:  operatorSetId,
	})
	if err != nil {
		return config.CurveTypeUnknown, fmt.Errorf("failed to get operator set curve type: %w", err)
	}

	return config.ConvertSolidityEnumToCurveType(curveType)
}

func (cc *ContractCaller) PublishMessageToInbox(ctx context.Context, avsAddress string, operatorSetId uint32, payload []byte) (*types.Receipt, error) {
	noSendTxOpts, err := cc.buildTransactionOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction options: %w", err)
	}

	address := cc.signer.GetFromAddress()

	tx, err := cc.taskMailbox.CreateTask(noSendTxOpts, ITaskMailbox.ITaskMailboxTypesTaskParams{
		RefundCollector: address,
		ExecutorOperatorSet: ITaskMailbox.OperatorSet{
			Avs: common.HexToAddress(avsAddress),
			Id:  operatorSetId,
		},
		Payload: payload,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	receipt, err := cc.signAndSendTransaction(ctx, tx, "PublishMessageToInbox")
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}
	cc.logger.Sugar().Infow("Successfully published message to inbox",
		zap.String("transactionHash", receipt.TxHash.Hex()),
	)
	return receipt, nil
}

func (cc *ContractCaller) GetOperatorBN254KeyRegistrationMessageHash(
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

func (cc *ContractCaller) GetOperatorECDSAKeyRegistrationMessageHash(
	ctx context.Context,
	operatorAddress common.Address,
	avsAddress common.Address,
	operatorSetId uint32,
	signingKeyAddress common.Address,
) ([32]byte, error) {
	cc.logger.Sugar().Infow("Getting ECDSA key registration message hash",
		zap.String("operatorAddress", operatorAddress.String()),
		zap.String("avsAddress", avsAddress.Hex()),
		zap.Uint32("operatorSetId", operatorSetId),
		zap.String("signingKeyAddress", signingKeyAddress.String()),
	)
	return cc.keyRegistrar.GetECDSAKeyRegistrationMessageHash(&bind.CallOpts{Context: ctx}, operatorAddress, IKeyRegistrar.OperatorSet{
		Avs: avsAddress,
		Id:  operatorSetId,
	}, signingKeyAddress)
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
func (cc *ContractCaller) ConfigureAVSOperatorSet(
	ctx context.Context,
	avsAddress common.Address,
	operatorSetId uint32,
	curveType config.CurveType,
) (*types.Receipt, error) {
	txOpts, err := cc.buildTransactionOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction options: %w", err)
	}

	solidityCurveType, err := curveType.Uint8()
	if err != nil {
		return nil, fmt.Errorf("failed to convert curve type to uint8: %w", err)
	}

	cc.logger.Sugar().Infow("configuring AVS operator set",
		zap.String("avsAddress", avsAddress.String()),
		zap.Uint32("operatorSetId", operatorSetId),
		zap.String("curveType", curveType.String()),
		zap.Uint8("solidityCurveType", solidityCurveType),
	)
	tx, err := cc.keyRegistrar.ConfigureOperatorSet(
		txOpts,
		IKeyRegistrar.OperatorSet{
			Avs: avsAddress,
			Id:  operatorSetId,
		},
		solidityCurveType,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	return cc.signAndSendTransaction(ctx, tx, "ConfigureOperatorSet")
}

func (cc *ContractCaller) RegisterKeyWithKeyRegistrar(
	ctx context.Context,
	operatorAddress common.Address,
	avsAddress common.Address,
	operatorSetId uint32,
	sigBytes []byte,
	keyData []byte,
) (*types.Receipt, error) {
	txOpts, err := cc.buildTransactionOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction options: %w", err)
	}

	// Create operator set struct
	operatorSet := IKeyRegistrar.OperatorSet{
		Avs: avsAddress,
		Id:  operatorSetId,
	}

	cc.logger.Sugar().Debugw("Registering key with KeyRegistrar",
		"operatorAddress:", operatorAddress.String(),
		"avsAddress:", avsAddress.String(),
		"operatorSetId:", operatorSetId,
		"keyData", hexutil.Encode(keyData),
		"sigButes:", hexutil.Encode(sigBytes),
	)

	tx, err := cc.keyRegistrar.RegisterKey(
		txOpts,
		operatorAddress,
		operatorSet,
		keyData,
		sigBytes,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register key: %w", err)
	}

	return cc.signAndSendTransaction(ctx, tx, "ConfigureOperatorSet")
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
	noSendTxOpts, err := cc.buildTransactionOpts(ctx)
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
		common.Address{},
		allocationDelay,
		metadataUri,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	return cc.signAndSendTransaction(ctx, tx, "RegisterAsOperator")
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
	txOpts, err := cc.buildTransactionOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction options: %w", err)
	}

	encodedSocket, err := encodeString(socket)
	if err != nil {
		return nil, fmt.Errorf("failed to encode socket string: %w", err)
	}

	tx, err := cc.allocationManager.RegisterForOperatorSets(txOpts, operatorAddress, IAllocationManager.IAllocationManagerTypesRegisterParams{
		Avs:            avsAddress,
		OperatorSetIds: operatorSetIds,
		Data:           encodedSocket,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	return cc.signAndSendTransaction(ctx, tx, "registerOperatorWithAvs")
}

//nolint:unused
func (cc *ContractCaller) getTransactOpts(ctx context.Context) (*bind.TransactOpts, error) {
	return cc.signer.GetTransactOpts(ctx)
}

func (cc *ContractCaller) buildTransactionOpts(ctx context.Context) (*bind.TransactOpts, error) {
	return cc.signer.GetTransactOpts(ctx)
}

func (cc *ContractCaller) GetSupportedChainsForMultichain(ctx context.Context, referenceBlockNumber int64) ([]*big.Int, []common.Address, error) {
	opts := &bind.CallOpts{
		Context: ctx,
	}
	if referenceBlockNumber > 0 {
		opts.BlockNumber = new(big.Int).SetUint64(uint64(referenceBlockNumber))
	}
	return cc.crossChainRegistry.GetSupportedChains(opts)
}

func (cc *ContractCaller) GetOperatorTableDataForOperatorSet(
	ctx context.Context,
	avsAddress common.Address,
	operatorSetId uint32,
	chainId config.ChainId,
	atBlockNumber uint64,
) (*contractCaller.OperatorTableData, error) {
	operatorSet := ICrossChainRegistry.OperatorSet{
		Avs: avsAddress,
		Id:  operatorSetId,
	}
	cc.logger.Sugar().Infow("Fetching operator table data",
		zap.String("avsAddress", avsAddress.String()),
		zap.Uint32("operatorSetId", operatorSetId),
		zap.Uint64("atBlockNumber", atBlockNumber),
	)
	otcAddr, err := cc.crossChainRegistry.GetOperatorTableCalculator(&bind.CallOpts{
		Context:     ctx,
		BlockNumber: new(big.Int).SetUint64(atBlockNumber),
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
	operatorWeights, err := opTableCalculator.GetOperatorSetWeights(&bind.CallOpts{
		Context:     ctx,
		BlockNumber: new(big.Int).SetUint64(atBlockNumber),
	}, IOperatorTableCalculator.OperatorSet(operatorSet))
	if err != nil {
		return nil, fmt.Errorf("failed to get operator weights: %w", err)
	}

	cc.logger.Sugar().Infow("Fetching supported chains for multichain")
	chainIds, tableUpdaterAddresses, err := cc.GetSupportedChainsForMultichain(ctx, int64(atBlockNumber))
	if err != nil {
		return nil, fmt.Errorf("failed to get supported chains for multichain: %w", err)
	}
	cc.logger.Sugar().Infow("Supported chains for multichain",
		zap.Any("chains", chainIds),
		zap.Any("tableUpdaterAddresses", tableUpdaterAddresses),
	)

	var tableUpdaterAddr common.Address
	tableUpdaterAddressMap := make(map[uint64]common.Address)
	for i, id := range chainIds {
		tableUpdaterAddressMap[id.Uint64()] = tableUpdaterAddresses[i]

		if tableUpdaterAddr == (common.Address{}) && id.Uint64() == uint64(chainId) {
			tableUpdaterAddr = tableUpdaterAddresses[i]
		}
	}
	if tableUpdaterAddr == (common.Address{}) {
		return nil, fmt.Errorf("no table updater address found for chain ID %d", chainId)
	}

	latestReferenceTimeAndBlock, err := cc.GetTableUpdaterReferenceTimeAndBlock(ctx, tableUpdaterAddr, atBlockNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest reference time and block: %w", err)
	}

	return &contractCaller.OperatorTableData{
		OperatorWeights:            operatorWeights.Weights,
		Operators:                  operatorWeights.Operators,
		LatestReferenceTimestamp:   latestReferenceTimeAndBlock.LatestReferenceTimestamp,
		LatestReferenceBlockNumber: latestReferenceTimeAndBlock.LatestReferenceBlockNumber,
		TableUpdaterAddresses:      tableUpdaterAddressMap,
	}, nil
}

func (cc *ContractCaller) GetTableUpdaterReferenceTimeAndBlock(
	ctx context.Context,
	tableUpdaterAddr common.Address,
	atBlockNumber uint64,
) (*contractCaller.LatestReferenceTimeAndBlock, error) {
	tableUpdater, err := IOperatorTableUpdater.NewIOperatorTableUpdater(tableUpdaterAddr, cc.ethclient)
	if err != nil {
		return nil, fmt.Errorf("failed to create operator table updater: %w", err)
	}

	latestReferenceTimestamp, err := tableUpdater.GetLatestReferenceTimestamp(&bind.CallOpts{
		Context:     ctx,
		BlockNumber: new(big.Int).SetUint64(atBlockNumber),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get latest reference timestamp: %w", err)
	}
	latestReferenceBlockNumber, err := tableUpdater.GetReferenceBlockNumberByTimestamp(&bind.CallOpts{}, latestReferenceTimestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest reference block number: %w", err)
	}

	return &contractCaller.LatestReferenceTimeAndBlock{
		LatestReferenceTimestamp:   latestReferenceTimestamp,
		LatestReferenceBlockNumber: latestReferenceBlockNumber,
	}, nil
}

func (cc *ContractCaller) GetActiveGenerationReservations() ([]ICrossChainRegistry.OperatorSet, error) {
	return cc.crossChainRegistry.GetActiveGenerationReservations(&bind.CallOpts{})
}

func (cc *ContractCaller) SetupTaskMailboxForAvs(
	ctx context.Context,
	avsAddress common.Address,
	taskHookAddress common.Address,
	executorOperatorSetIds []uint32,
	curveTypes []config.CurveType,
) error {
	for i, id := range executorOperatorSetIds {
		curveTypeStr := curveTypes[i]

		mailboxCfg := ITaskMailbox.ITaskMailboxTypesExecutorOperatorSetTaskConfig{
			TaskHook: taskHookAddress,
			//FeeToken:                 common.HexToAddress("0x"),
			//FeeCollector:             common.HexToAddress("0x"),
			TaskSLA: big.NewInt(60),
			Consensus: ITaskMailbox.ITaskMailboxTypesConsensus{
				ConsensusType: 1,
				Value:         util.AbiEncodeUint16(6667), // 66.67% consensus threshold
			},
			TaskMetadata: nil,
		}

		solidityCurveType, err := curveTypeStr.Uint8()
		if err != nil {
			return fmt.Errorf("failed to convert curve type to uint8: %w", err)
		}
		mailboxCfg.CurveType = solidityCurveType

		noSendTxOpts, err := cc.buildTransactionOpts(ctx)
		if err != nil {
			return fmt.Errorf("failed to build transaction options: %w", err)
		}

		tx, err := cc.taskMailbox.SetExecutorOperatorSetTaskConfig(
			noSendTxOpts,
			ITaskMailbox.OperatorSet{
				Avs: avsAddress,
				Id:  id,
			},
			mailboxCfg,
		)
		if err != nil {
			return fmt.Errorf("failed to create transaction: %w", err)
		}
		receipt, err := cc.signAndSendTransaction(ctx, tx, "SetupTaskMailboxForAvs")
		if err != nil {
			return fmt.Errorf("failed to send transaction: %w", err)
		}
		cc.logger.Sugar().Infow("Successfully set up task mailbox for AVS",
			zap.String("avsAddress", avsAddress.String()),
			zap.Uint32("executorOperatorSetId", id),
			zap.String("transactionHash", receipt.TxHash.Hex()),
		)
	}
	return nil
}

func (cc *ContractCaller) DelegateToOperator(
	ctx context.Context,
	stakerAddress common.Address,
	operatorAddress common.Address,
) (*types.Receipt, error) {

	approver, err := cc.delegationManager.DelegationApprover(&bind.CallOpts{}, operatorAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get delegation approver: %w", err)
	}
	cc.logger.Sugar().Infow("Delegation approver for operator",
		zap.String("operatorAddress", operatorAddress.String()),
		zap.String("delegationApprover", approver.String()),
	)

	txOpts, err := cc.buildTransactionOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction options: %w", err)
	}
	cc.logger.Sugar().Infow("Delegating to operator",
		zap.String("operatorAddress", operatorAddress.String()),
		zap.Any("txOpts", txOpts),
	)

	approvalSignature := IDelegationManager.ISignatureUtilsMixinTypesSignatureWithExpiry{
		Signature: nil,
		Expiry:    big.NewInt(0),
	}
	var salt [32]byte

	tx, err := cc.delegationManager.DelegateTo(txOpts, operatorAddress, approvalSignature, salt)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	return cc.signAndSendTransaction(ctx, tx, "DelegateToOperator")
}

func (cc *ContractCaller) ModifyAllocations(
	ctx context.Context,
	operatorAddress common.Address,
	avsAddress common.Address,
	operatorSetId uint32,
	strategy common.Address,
) (interface{}, error) {
	cc.logger.Sugar().Infow("Modifying allocations",
		zap.String("operatorAddress", operatorAddress.String()),
		zap.String("avsAddress", avsAddress.String()),
		zap.Uint32("operatorSetId", operatorSetId),
		zap.String("strategy", strategy.String()),
	)
	alloactionDelay, err := cc.allocationManager.GetAllocationDelay(&bind.CallOpts{}, operatorAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get allocation delay: %w", err)
	}
	cc.logger.Sugar().Infow("allocation delay:",
		zap.Any("allocationDelay", alloactionDelay),
		zap.String("operatorAddress", operatorAddress.String()),
	)

	txOpts, err := cc.buildTransactionOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction options: %w", err)
	}

	tx, err := cc.allocationManager.ModifyAllocations(txOpts, operatorAddress, []IAllocationManager.IAllocationManagerTypesAllocateParams{
		{
			OperatorSet: IAllocationManager.OperatorSet{
				Avs: avsAddress,
				Id:  operatorSetId,
			},
			Strategies:    []common.Address{strategy},
			NewMagnitudes: []uint64{1e18},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}
	return cc.signAndSendTransaction(ctx, tx, "ModifyAllocations")
}

func (cc *ContractCaller) VerifyECDSACertificate(
	ctx context.Context,
	avsAddress common.Address,
	operatorSetId uint32,
	aggCert *aggregation.AggregatedECDSACertificate,
	globalTableRootReferenceTimestamp uint32,
	threshold uint16,
) (bool, []common.Address, error) {
	if len(aggCert.TaskId) != 32 {
		return false, nil, fmt.Errorf("taskId must be 32 bytes, got %d", len(aggCert.TaskId))
	}
	var taskId [32]byte
	copy(taskId[:], aggCert.TaskId)

	finalSig, err := aggCert.GetFinalSignature()
	if err != nil {
		return false, nil, fmt.Errorf("failed to get final signature: %w", err)
	}

	cc.logger.Sugar().Infow("verifying ECDSA certificate",
		zap.String("taskId", hexutil.Encode(taskId[:])),
		zap.String("mailboxAddress", cc.coreContracts.TaskMailbox),
		zap.Uint32("globalTableRootReferenceTimestamp", globalTableRootReferenceTimestamp),
		zap.String("finalSig", hexutil.Encode(finalSig[:])),
	)

	cert := IECDSACertificateVerifier.IECDSACertificateVerifierTypesECDSACertificate{
		ReferenceTimestamp: globalTableRootReferenceTimestamp,
		MessageHash:        aggCert.GetTaskMessageHash(),
		Sig:                finalSig,
	}

	return cc.ecdsaCertVerifier.VerifyCertificateProportion(
		&bind.CallOpts{},
		IECDSACertificateVerifier.OperatorSet{
			Avs: avsAddress,
			Id:  operatorSetId,
		},
		cert,
		[]uint16{threshold},
	)
}

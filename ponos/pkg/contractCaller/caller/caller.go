package caller

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"sort"
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
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/middleware-bindings/IBN254TableCalculator"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/middleware-bindings/ITaskAVSRegistrarBase"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/middleware-bindings/TaskAVSRegistrarBase"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionSigner"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	geth "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	merkletree "github.com/wealdtech/go-merkletree/v2"
	"github.com/wealdtech/go-merkletree/v2/keccak256"
	"go.uber.org/zap"
)

// Table calculator contract addresses
const (
	BN254TableCalculatorAddress = "0x797d076aB96a5d4104062C15c727447fD8b71eB0"
	ECDSATableCalculatorAddress = "0xbcff2Cb40eD4A80e3A9EB095840986F9c8395a38"
)

type OperatorInfo struct {
	PubkeyX *big.Int
	PubkeyY *big.Int
	Weights []*big.Int
}

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
	ethClient *ethereum.EthereumClient,
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
	params *contractCaller.BN254TaskResultParams,
	globalTableRootReferenceTimestamp uint32,
	operatorInfoTreeRoot [32]byte,
) (*types.Receipt, error) {
	backoffs := []int{1, 3, 5, 10, 20}
	for i, backoff := range backoffs {
		res, err := cc.SubmitBN254TaskResult(ctx, params, globalTableRootReferenceTimestamp, operatorInfoTreeRoot)
		if err != nil {
			if i == len(backoffs)-1 {
				cc.logger.Sugar().Errorw("failed to submit task result after retries",
					zap.String("taskId", hexutil.Encode(params.TaskId)),
					zap.Error(err),
				)
				return nil, fmt.Errorf("failed to submit task result: %w", err)
			}
			cc.logger.Sugar().Errorw("failed to submit task result, retrying",
				zap.Error(err),
				zap.String("taskId", hexutil.Encode(params.TaskId)),
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
	params *contractCaller.BN254TaskResultParams,
	globalTableRootReferenceTimestamp uint32,
	operatorInfoTreeRoot [32]byte,
) (*types.Receipt, error) {

	noSendTxOpts, err := cc.buildTransactionOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction options: %w", err)
	}

	if len(params.TaskId) != 32 {
		return nil, fmt.Errorf("taskId must be 32 bytes, got %d", len(params.TaskId))
	}

	var taskId [32]byte
	copy(taskId[:], params.TaskId)
	cc.logger.Sugar().Infow("submitting task result",
		zap.String("taskId", hexutil.Encode(taskId[:])),
		zap.String("mailboxAddress", cc.coreContracts.TaskMailbox),
		zap.Uint32("globalTableRootReferenceTimestamp", globalTableRootReferenceTimestamp),
	)

	g1Point := &bn254.G1Point{
		G1Affine: params.SignersSignature.GetG1Point(),
	}

	g1Bytes, err := g1Point.ToPrecompileFormat()
	if err != nil {
		return nil, fmt.Errorf("signature not in correct subgroup: %w", err)
	}

	g2Bytes, err := params.SignersPublicKey.ToPrecompileFormat()
	if err != nil {
		return nil, fmt.Errorf("public key not in correct subgroup: %w", err)
	}

	digest := params.TaskResponseDigest

	var allOperators []OperatorInfo
	if len(params.OperatorInfos) <= 0 {
		return nil, fmt.Errorf("OperatorInfos must be provided for BN254 task submission")
	}

	allOperators = make([]OperatorInfo, len(params.OperatorInfos))
	for i, info := range params.OperatorInfos {
		allOperators[i] = OperatorInfo{
			PubkeyX: info.PubkeyX,
			PubkeyY: info.PubkeyY,
			Weights: info.Weights,
		}
	}

	if operatorInfoTreeRoot == [32]byte{} {
		return nil, fmt.Errorf("failed to submit task result: operatorInfoTreeRoot is empty")
	}

	if len(allOperators) < 1 {
		return nil, fmt.Errorf("failed to submit task result: allOperators is empty (SortedOperatorsByIndex has %d operators)", len(params.SortedOperatorsByIndex))
	}

	cc.logger.Sugar().Infow("Generating merkle proofs for non-signers",
		zap.String("operatorInfoTreeRoot", hexutil.Encode(operatorInfoTreeRoot[:])),
		zap.Int("numNonSigners", len(params.NonSignerOperators)),
		zap.Int("numOperators", len(allOperators)),
	)

	proofs, err := cc.generateOperatorMerkleProofs(params.NonSignerOperators, operatorInfoTreeRoot, allOperators)
	if err != nil {
		cc.logger.Sugar().Warnw("Failed to generate merkle proofs", zap.Error(err))
		return nil, fmt.Errorf("failed to generate merkle proofs for non-signers: %w", err)
	}

	nonSignerWitnesses := make([]ITaskMailbox.IBN254CertificateVerifierTypesBN254OperatorInfoWitness, 0, len(params.NonSignerOperators))
	for _, nonSigner := range params.NonSignerOperators {

		// Check bounds before accessing SortedOperatorsByIndex
		var operatorWeights []*big.Int
		if int(nonSigner.OperatorIndex) >= len(params.SortedOperatorsByIndex) {
			return nil, fmt.Errorf("non-signer operator index %d out of range for SortedOperatorsByIndex (length %d)",
				nonSigner.OperatorIndex, len(params.SortedOperatorsByIndex))
		}
		operatorWeights = params.SortedOperatorsByIndex[nonSigner.OperatorIndex].Weights

		witness := ITaskMailbox.IBN254CertificateVerifierTypesBN254OperatorInfoWitness{
			OperatorIndex:     nonSigner.OperatorIndex,
			OperatorInfoProof: []byte{},
			OperatorInfo: ITaskMailbox.IOperatorTableCalculatorTypesBN254OperatorInfo{
				Pubkey: ITaskMailbox.BN254G1Point{
					X: new(big.Int).SetBytes(nonSigner.PublicKey[0:32]),
					Y: new(big.Int).SetBytes(nonSigner.PublicKey[32:64]),
				},
				Weights: operatorWeights,
			},
		}

		proof, ok := proofs[nonSigner.OperatorIndex]
		if !ok {
			return nil, fmt.Errorf("missing merkle proof for non-signer operator at index %d", nonSigner.OperatorIndex)
		}
		witness.OperatorInfoProof = proof

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

	cc.logger.Sugar().Infow("Creating BN254 Submit Result Transaction",
		zap.String("taskId", hexutil.Encode(taskId[:])),
		zap.String("Cert", hexutil.Encode(certBytes)),
		zap.String("TaskResponse", hexutil.Encode(params.TaskResponse)),
	)

	tx, err := cc.taskMailbox.SubmitResult(noSendTxOpts, taskId, certBytes, params.TaskResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	return cc.signAndSendTransaction(ctx, tx, "SubmitTaskSession")
}

func (cc *ContractCaller) SubmitECDSATaskResultRetryable(
	ctx context.Context,
	params *contractCaller.ECDSATaskResultParams,
	globalTableRootReferenceTimestamp uint32,
) (*types.Receipt, error) {
	backoffs := []int{1, 3, 5, 10, 20}
	for i, backoff := range backoffs {
		res, err := cc.SubmitECDSATaskResult(ctx, params, globalTableRootReferenceTimestamp)
		if err != nil {
			if i == len(backoffs)-1 {
				cc.logger.Sugar().Errorw("failed to submit task result after retries",
					zap.String("taskId", hexutil.Encode(params.TaskId)),
					zap.Error(err),
				)
				return nil, fmt.Errorf("failed to submit task result: %w", err)
			}
			cc.logger.Sugar().Errorw("failed to submit task result, retrying",
				zap.Error(err),
				zap.String("taskId", hexutil.Encode(params.TaskId)),
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
	params *contractCaller.ECDSATaskResultParams,
	globalTableRootReferenceTimestamp uint32,
) (*types.Receipt, error) {
	noSendTxOpts, err := cc.buildTransactionOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction options: %w", err)
	}

	if len(params.TaskId) != 32 {
		return nil, fmt.Errorf("taskId must be 32 bytes, got %d", len(params.TaskId))
	}
	var taskId [32]byte
	copy(taskId[:], params.TaskId)

	// Get final signature from params - concatenate signatures in sorted order
	finalSig, err := cc.getFinalECDSASignature(params.SignersSignatures)
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
		MessageHash:        params.TaskResponseDigest,
		Sig:                finalSig,
	}

	certBytes, err := cc.taskMailbox.GetECDSACertificateBytes(&bind.CallOpts{}, cert)
	if err != nil {
		return nil, fmt.Errorf("failed to call GetECDSACertificateBytes: %w", err)
	}
	cc.logger.Sugar().Infow("Creating ECDSA Submit Result Transaction",
		zap.String("taskId", hexutil.Encode(taskId[:])),
		zap.String("Cert", hexutil.Encode(certBytes)),
		zap.String("TaskResponse", hexutil.Encode(params.TaskResponse)),
	)
	tx, err := cc.taskMailbox.SubmitResult(noSendTxOpts, taskId, certBytes, params.TaskResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	return cc.signAndSendTransaction(ctx, tx, "SubmitTaskSession")
}

// getFinalECDSASignature concatenates ECDSA signatures in sorted order by address
func (cc *ContractCaller) getFinalECDSASignature(signatures map[common.Address][]byte) ([]byte, error) {
	if len(signatures) == 0 {
		return nil, fmt.Errorf("no signatures found")
	}

	// Collect addresses
	addresses := make([]common.Address, 0, len(signatures))
	for addr := range signatures {
		addresses = append(addresses, addr)
	}

	// Sort by raw bytes
	sort.Slice(addresses, func(i, j int) bool {
		return bytes.Compare(addresses[i][:], addresses[j][:]) < 0
	})

	// Concatenate signatures in sorted order
	var finalSignature []byte
	for _, addr := range addresses {
		sig := signatures[addr]
		if len(sig) != 65 {
			return nil, fmt.Errorf("signature for address %s has invalid length: expected 65, got %d",
				addr.Hex(), len(sig))
		}
		finalSignature = append(finalSignature, sig...)
	}

	return finalSignature, nil
}

func (cc *ContractCaller) CalculateECDSACertificateDigestBytes(
	_ context.Context,
	referenceTimestamp uint32,
	messageHash [32]byte,
) ([]byte, error) {
	return cc.ecdsaCertVerifier.CalculateCertificateDigestBytes(&bind.CallOpts{}, referenceTimestamp, messageHash)
}

func (cc *ContractCaller) CalculateTaskMessageHash(
	_ context.Context,
	taskHash [32]byte,
	result []byte,
) ([32]byte, error) {
	return cc.taskMailbox.GetMessageHash(&bind.CallOpts{}, taskHash, result)
}

func (cc *ContractCaller) CalculateBN254CertificateDigestBytes(
	_ context.Context,
	referenceTimestamp uint32,
	messageHash [32]byte,
) ([]byte, error) {
	digest, err := cc.bn254CertVerifier.CalculateCertificateDigest(&bind.CallOpts{}, referenceTimestamp, messageHash)
	if err != nil {
		return nil, err
	}
	return digest[:], nil
}

func (cc *ContractCaller) GetExecutorOperatorSetTaskConfig(
	_ context.Context,
	avsAddress common.Address,
	opsetId uint32,
	blockNumber uint64,
) (*contractCaller.TaskMailboxExecutorOperatorSetConfig, error) {
	blockHeightOpts := &bind.CallOpts{}
	if blockNumber > 0 {
		blockHeightOpts.BlockNumber = big.NewInt(int64(blockNumber))
	}
	res, err := cc.taskMailbox.GetExecutorOperatorSetTaskConfig(blockHeightOpts, ITaskMailbox.OperatorSet{
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

func (cc *ContractCaller) GetOperatorSetMembersWithPeering(avsAddress string, operatorSetId uint32, blockNumber uint64) ([]*peering.OperatorPeerInfo, error) {
	operatorSetStringAddrs, err := cc.getOperatorSetMembers(avsAddress, operatorSetId, blockNumber)
	if err != nil {
		return nil, err
	}

	operatorSetMemberAddrs := util.Map(operatorSetStringAddrs, func(address string, i uint64) common.Address {
		return common.HexToAddress(address)
	})

	allMembers := make([]*peering.OperatorPeerInfo, len(operatorSetMemberAddrs))
	for index, member := range operatorSetMemberAddrs {
		operatorSetInfo, err := cc.GetOperatorSetDetailsForOperator(member, avsAddress, operatorSetId, blockNumber)

		if err != nil {
			cc.logger.Sugar().Errorf("failed to get operator set details for operator %s: %v", member.Hex(), err)
			continue
		}
		operatorSetInfo.OperatorIndex = uint32(index)

		allMembers[index] = &peering.OperatorPeerInfo{
			OperatorAddress: operatorSetStringAddrs[index],
			OperatorSets:    []*peering.OperatorSet{operatorSetInfo},
		}
	}

	if len(allMembers) == 0 {
		return nil, fmt.Errorf("no valid operators found")
	}

	return allMembers, nil
}

func (cc *ContractCaller) GetOperatorSetDetailsForOperator(
	operatorAddress common.Address,
	avsAddress string,
	operatorSetId uint32,
	blockNumber uint64,
) (*peering.OperatorSet, error) {

	opSet := IKeyRegistrar.OperatorSet{
		Avs: common.HexToAddress(avsAddress),
		Id:  operatorSetId,
	}

	blockHeightOpts := &bind.CallOpts{}
	if blockNumber > 0 {
		blockHeightOpts.BlockNumber = big.NewInt(int64(blockNumber))
	}

	// Get the AVS registrar address from the allocation manager
	avsAddr := common.HexToAddress(avsAddress)
	avsRegistrarAddress, err := cc.allocationManager.GetAVSRegistrar(blockHeightOpts, avsAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to get AVS registrar address: %w", err)
	}

	// Create new registrar caller
	caller, err := TaskAVSRegistrarBase.NewTaskAVSRegistrarBaseCaller(avsRegistrarAddress, cc.ethclient)
	if err != nil {
		return nil, fmt.Errorf("failed to create AVS registrar caller: %w", err)
	}
	socket, err := caller.GetOperatorSocket(blockHeightOpts, operatorAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get operator socket: %w", err)
	}

	curveTypeSolidity, err := cc.keyRegistrar.GetOperatorSetCurveType(blockHeightOpts, opSet)
	if err != nil {
		cc.logger.Sugar().Errorf("failed to get operator set curve type: %v", err)
		return nil, err
	}

	curveType, err := config.ConvertSolidityEnumToCurveType(curveTypeSolidity)
	if err != nil {
		cc.logger.Sugar().Errorf("failed to convert curve type: %v", err)
		return nil, fmt.Errorf("failed to convert curve type: %w", err)
	}

	peeringOpSet := &peering.OperatorSet{
		OperatorSetID:  operatorSetId,
		NetworkAddress: socket,
		CurveType:      curveType,
	}

	if curveType == config.CurveTypeBN254 {
		solidityPubKey, err := cc.keyRegistrar.GetBN254Key(blockHeightOpts, opSet, operatorAddress)
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

		peeringOpSet.WrappedPublicKey = peering.WrappedPublicKey{
			PublicKey: pubKey,
		}

		return peeringOpSet, nil
	}

	if curveType == config.CurveTypeECDSA {
		ecdsaAddr, err := cc.keyRegistrar.GetECDSAAddress(blockHeightOpts, opSet, operatorAddress)
		if err != nil {
			cc.logger.Sugar().Errorf("failed to get operator set public key: %v", err)
			return nil, err
		}
		peeringOpSet.WrappedPublicKey = peering.WrappedPublicKey{
			ECDSAAddress: ecdsaAddr,
		}
		return peeringOpSet, nil
	}

	return nil, fmt.Errorf("unsupported curve type: %s", curveType)
}

func (cc *ContractCaller) GetAVSConfig(avsAddress string, blockNumber uint64) (*contractCaller.AVSConfig, error) {
	avsAddr := common.HexToAddress(avsAddress)

	// Support optional block number parameter for historical queries
	blockHeightOpts := &bind.CallOpts{}
	if blockNumber > 0 {
		blockHeightOpts.BlockNumber = big.NewInt(int64(blockNumber))
	}

	avsRegistrarAddress, err := cc.allocationManager.GetAVSRegistrar(blockHeightOpts, avsAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to get AVS registrar address: %w", err)
	}

	registrarCaller, err := ITaskAVSRegistrarBase.NewITaskAVSRegistrarBaseCaller(avsRegistrarAddress, cc.ethclient)
	if err != nil {
		return nil, fmt.Errorf("failed to create AVS registrar caller: %w", err)
	}

	avsConfig, err := registrarCaller.GetAvsConfig(blockHeightOpts)
	if err != nil {
		return nil, err
	}

	return &contractCaller.AVSConfig{
		AggregatorOperatorSetId: avsConfig.AggregatorOperatorSetId,
		ExecutorOperatorSetIds:  avsConfig.ExecutorOperatorSetIds,
	}, nil
}

func (cc *ContractCaller) GetOperatorSetCurveType(avsAddress string, operatorSetId uint32, blockNumber uint64) (config.CurveType, error) {
	blockHeightOpts := &bind.CallOpts{}
	if blockNumber > 0 {
		blockHeightOpts.BlockNumber = big.NewInt(int64(blockNumber))
	}

	curveType, err := cc.keyRegistrar.GetOperatorSetCurveType(blockHeightOpts, IKeyRegistrar.OperatorSet{
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

	cc.logger.Sugar().Infow("Registering operator socket with AVS")
	socketReceipt, err := cc.registerOperatorWithAvs(ctx, operatorAddress, avsAddress, operatorSetIds, socket)
	if err != nil {
		return nil, fmt.Errorf("failed to register operator socket with AVS: %w", err)
	}
	cc.logger.Sugar().Infow("Successfully registered operator socket with AVS",
		zap.Any("receipt", socketReceipt),
	)

	return socketReceipt, nil
}

func (cc *ContractCaller) getOperatorSetMembers(avsAddress string, operatorSetId uint32, blockNumber uint64) ([]string, error) {
	avsAddr := common.HexToAddress(avsAddress)
	blockHeightOpts := &bind.CallOpts{}
	if blockNumber > 0 {
		blockHeightOpts.BlockNumber = big.NewInt(int64(blockNumber))
	}

	operatorSet, err := cc.allocationManager.GetMembers(blockHeightOpts, IAllocationManager.OperatorSet{
		Avs: avsAddr,
		Id:  operatorSetId,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get operator set members: %w", err)
	}

	members := make([]string, len(operatorSet))
	for index, member := range operatorSet {
		members[index] = member.String()
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

	encodedSocket, err := util.EncodeString(socket)
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

func (cc *ContractCaller) buildTransactionOpts(ctx context.Context) (*bind.TransactOpts, error) {
	return cc.signer.GetTransactOpts(ctx)
}

func (cc *ContractCaller) GetSupportedChainsForMultichain(ctx context.Context, referenceBlockNumber uint64) ([]*big.Int, []common.Address, error) {
	opts := &bind.CallOpts{
		Context: ctx,
	}
	if referenceBlockNumber > 0 {
		opts.BlockNumber = new(big.Int).SetUint64(referenceBlockNumber)
	}
	return cc.crossChainRegistry.GetSupportedChains(opts)
}

func (cc *ContractCaller) GetOperatorTableDataForOperatorSet(ctx context.Context, avsAddress common.Address, operatorSetId uint32, curveType config.CurveType, chainId config.ChainId, atBlockNumber uint64, l2blocknumber uint64) (*contractCaller.OperatorTableData, error) {

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
	chainIds, tableUpdaterAddresses, err := cc.GetSupportedChainsForMultichain(ctx, atBlockNumber)

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

	operatorTableData := &contractCaller.OperatorTableData{
		OperatorWeights:            operatorWeights.Weights,
		Operators:                  operatorWeights.Operators,
		LatestReferenceTimestamp:   latestReferenceTimeAndBlock.LatestReferenceTimestamp,
		LatestReferenceBlockNumber: latestReferenceTimeAndBlock.LatestReferenceBlockNumber,
		TableUpdaterAddresses:      tableUpdaterAddressMap,
	}

	if curveType == config.CurveTypeBN254 {
		// Create L2 client
		ec := ethereum.NewEthereumClient(&ethereum.EthereumClientConfig{BaseUrl: "http://127.0.0.1:9545"}, cc.logger)
		calcClient, err := ec.GetEthereumContractCaller()
		if err != nil {
			return nil, fmt.Errorf("failed to get ethereum contract caller: %w", err)
		}

		// Get the L2 OperatorTableUpdater to find L2's reference timestamp
		l2TableUpdater, err := IOperatorTableUpdater.NewIOperatorTableUpdater(
			tableUpdaterAddr, // This is the L2 table updater address from the map
			calcClient,       // Use L2 client
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create L2 operator table updater: %w", err)
		}

		// Get L2's latest reference timestamp (not L1's!)
		l2LatestTimestamp, err := l2TableUpdater.GetLatestReferenceTimestamp(&bind.CallOpts{
			Context: ctx,
			// Don't specify block number - get the latest
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get L2 latest reference timestamp: %w", err)
		}

		// Get L2's reference block number for that timestamp
		l2RefBlockNumber, err := l2TableUpdater.GetReferenceBlockNumberByTimestamp(&bind.CallOpts{
			Context: ctx,
		}, l2LatestTimestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to get L2 reference block number: %w", err)
		}

		cc.logger.Sugar().Infow("L2 reference data",
			zap.Uint32("l2LatestTimestamp", l2LatestTimestamp),
			zap.Uint32("l2RefBlockNumber", l2RefBlockNumber),
			zap.Uint32("l1LatestTimestamp", latestReferenceTimeAndBlock.LatestReferenceTimestamp),
		)

		// Get operator infos from BN254TableCalculator (still using L1 client)
		bn254TableCalculator, err := IBN254TableCalculator.NewIBN254TableCalculatorCaller(otcAddr, cc.ethclient)
		if err != nil {
			return nil, fmt.Errorf("failed to create BN254 table calculator caller: %w", err)
		}

		operatorInfos, err := bn254TableCalculator.GetOperatorInfos(&bind.CallOpts{
			Context:     ctx,
			BlockNumber: new(big.Int).SetUint64(atBlockNumber),
		}, IBN254TableCalculator.OperatorSet{
			Avs: avsAddress,
			Id:  operatorSetId,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get operator infos from BN254TableCalculator: %w", err)
		}

		// Query L2 BN254CertificateVerifier with L2's reference data
		verifier, err := IBN254CertificateVerifier.NewIBN254CertificateVerifier(
			common.HexToAddress("0xff58A373c18268F483C1F5cA03Cf885c0C43373a"),
			calcClient, // L2 client
		)
		if err != nil {
			return nil, err
		}

		// Use L2's timestamp and query at L2's reference block
		operatorSetInfo, err := verifier.GetOperatorSetInfo(&bind.CallOpts{
			Context:     ctx,
			BlockNumber: new(big.Int).SetUint64(l2blocknumber),
		}, IBN254CertificateVerifier.OperatorSet{
			Avs: avsAddress,
			Id:  operatorSetId,
		}, l2LatestTimestamp) // Use L2 timestamp, not L1!
		if err != nil {
			return nil, fmt.Errorf("failed to get operator set info from BN254 verifier: %w", err)
		}

		// Populate operator infos
		operatorTableData.OperatorInfos = make([]contractCaller.BN254OperatorInfo, len(operatorInfos))
		for i, info := range operatorInfos {
			operatorTableData.OperatorInfos[i] = contractCaller.BN254OperatorInfo{
				PubkeyX: info.Pubkey.X,
				PubkeyY: info.Pubkey.Y,
				Weights: info.Weights,
			}
		}

		// This should now have the correct L2 operator info tree root
		operatorTableData.OperatorInfoTreeRoot = operatorSetInfo.OperatorInfoTreeRoot

		// Also update the timestamp to reflect L2's timestamp
		// This ensures consistency when used for verification
		operatorTableData.LatestReferenceTimestamp = l2LatestTimestamp
	}

	return operatorTableData, nil
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

	latestReferenceBlockNumber, err := tableUpdater.GetReferenceBlockNumberByTimestamp(&bind.CallOpts{
		Context:     ctx,
		BlockNumber: new(big.Int).SetUint64(atBlockNumber),
	}, latestReferenceTimestamp)
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
			TaskSLA:  big.NewInt(60),
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
	messageHash [32]byte,
	signature []byte,
	avsAddress common.Address,
	operatorSetId uint32,
	globalTableRootReferenceTimestamp uint32,
	threshold uint16,
) (bool, []common.Address, error) {
	cert := IECDSACertificateVerifier.IECDSACertificateVerifierTypesECDSACertificate{
		ReferenceTimestamp: globalTableRootReferenceTimestamp,
		MessageHash:        messageHash,
		Sig:                signature,
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

// CreateGenerationReservation creates a generation reservation for an operator set with table calculator and config
func (cc *ContractCaller) CreateGenerationReservation(
	ctx context.Context,
	avsAddress common.Address,
	operatorSetId uint32,
	operatorTableCalculatorAddress common.Address,
	owner common.Address,
	maxStalenessPeriod uint32,
) (*types.Receipt, error) {
	txOpts, err := cc.buildTransactionOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction options: %w", err)
	}

	operatorSet := ICrossChainRegistry.OperatorSet{
		Avs: avsAddress,
		Id:  operatorSetId,
	}

	config := ICrossChainRegistry.ICrossChainRegistryTypesOperatorSetConfig{
		Owner:              owner,
		MaxStalenessPeriod: maxStalenessPeriod,
	}

	cc.logger.Sugar().Infow("Creating generation reservation",
		zap.String("avsAddress", avsAddress.String()),
		zap.Uint32("operatorSetId", operatorSetId),
		zap.String("calculatorAddress", operatorTableCalculatorAddress.String()),
		zap.String("owner", owner.String()),
		zap.Uint32("maxStalenessPeriod", maxStalenessPeriod),
	)

	tx, err := cc.crossChainRegistry.CreateGenerationReservation(
		txOpts,
		operatorSet,
		operatorTableCalculatorAddress,
		config,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	return cc.signAndSendTransaction(ctx, tx, "CreateGenerationReservation")
}

// SetOperatorTableCalculator sets the operator table calculator for an operator set (requires existing reservation)
func (cc *ContractCaller) SetOperatorTableCalculator(
	ctx context.Context,
	avsAddress common.Address,
	operatorSetId uint32,
	operatorTableCalculatorAddress common.Address,
) (*types.Receipt, error) {
	txOpts, err := cc.buildTransactionOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction options: %w", err)
	}

	operatorSet := ICrossChainRegistry.OperatorSet{
		Avs: avsAddress,
		Id:  operatorSetId,
	}

	cc.logger.Sugar().Infow("Setting operator table calculator",
		zap.String("avsAddress", avsAddress.String()),
		zap.Uint32("operatorSetId", operatorSetId),
		zap.String("calculatorAddress", operatorTableCalculatorAddress.String()),
	)

	tx, err := cc.crossChainRegistry.SetOperatorTableCalculator(
		txOpts,
		operatorSet,
		operatorTableCalculatorAddress,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	return cc.signAndSendTransaction(ctx, tx, "SetOperatorTableCalculator")
}

// GetTableCalculatorAddress returns the appropriate table calculator address for a given curve type
func (cc *ContractCaller) GetTableCalculatorAddress(curveType config.CurveType) common.Address {
	switch curveType {
	case config.CurveTypeBN254:
		return common.HexToAddress(BN254TableCalculatorAddress)
	case config.CurveTypeECDSA:
		return common.HexToAddress(ECDSATableCalculatorAddress)
	default:
		cc.logger.Sugar().Warnw("Unknown curve type for table calculator",
			zap.String("curveType", curveType.String()),
		)
		return common.Address{}
	}
}

// VerifyBN254Certificate verifies a BN254 certificate directly with the BN254CertificateVerifier contract
// This replicates the certificate construction from SubmitBN254TaskResult and calls verifyCertificateProportion directly
// Returns true if the certificate meets the threshold percentage (e.g., 2500 for 25%)
//
// Testing use only
func (cc *ContractCaller) VerifyBN254Certificate(
	ctx context.Context,
	avsAddress common.Address,
	operatorSetId uint32,
	params *contractCaller.BN254TaskResultParams,
	globalTableRootReferenceTimestamp uint32,
	operatorInfoTreeRoot [32]byte,
	thresholdPercentage uint16,
) (bool, error) {
	g1Point := &bn254.G1Point{
		G1Affine: params.SignersSignature.GetG1Point(),
	}

	g1Bytes, err := g1Point.ToPrecompileFormat()
	if err != nil {
		return false, fmt.Errorf("signature not in correct subgroup: %w", err)
	}

	g2Bytes, err := params.SignersPublicKey.ToPrecompileFormat()
	if err != nil {
		return false, fmt.Errorf("public key not in correct subgroup: %w", err)
	}

	digest := params.TaskResponseDigest

	var allOperators []OperatorInfo
	if len(params.OperatorInfos) > 0 {
		allOperators = make([]OperatorInfo, len(params.OperatorInfos))
		for i, info := range params.OperatorInfos {
			allOperators[i] = OperatorInfo{
				PubkeyX: info.PubkeyX,
				PubkeyY: info.PubkeyY,
				Weights: info.Weights,
			}
		}
	} else {
		return false, fmt.Errorf("OperatorInfos must be provided for BN254 merkle proof generation")
	}

	proofs, err := cc.generateOperatorMerkleProofs(params.NonSignerOperators, operatorInfoTreeRoot, allOperators)
	if err != nil {
		return false, fmt.Errorf("failed to generate merkle proofs for non-signers: %w", err)
	}

	nonSignerWitnesses := make([]IBN254CertificateVerifier.IBN254CertificateVerifierTypesBN254OperatorInfoWitness, 0, len(params.NonSignerOperators))
	for _, nonSigner := range params.NonSignerOperators {
		if int(nonSigner.OperatorIndex) >= len(allOperators) {
			return false, fmt.Errorf("non-signer operator index %d out of range (have %d operators)",
				nonSigner.OperatorIndex, len(allOperators))
		}

		proof, ok := proofs[nonSigner.OperatorIndex]
		if !ok {
			return false, fmt.Errorf("missing merkle proof for non-signer operator at index %d", nonSigner.OperatorIndex)
		}

		operatorInfo := allOperators[nonSigner.OperatorIndex]

		witness := IBN254CertificateVerifier.IBN254CertificateVerifierTypesBN254OperatorInfoWitness{
			OperatorIndex:     nonSigner.OperatorIndex,
			OperatorInfoProof: proof,
			OperatorInfo: IBN254CertificateVerifier.IOperatorTableCalculatorTypesBN254OperatorInfo{
				Pubkey: IBN254CertificateVerifier.BN254G1Point{
					X: operatorInfo.PubkeyX,
					Y: operatorInfo.PubkeyY,
				},
				Weights: operatorInfo.Weights,
			},
		}
		nonSignerWitnesses = append(nonSignerWitnesses, witness)
	}

	cert := IBN254CertificateVerifier.IBN254CertificateVerifierTypesBN254Certificate{
		ReferenceTimestamp: globalTableRootReferenceTimestamp,
		MessageHash:        digest,
		Signature: IBN254CertificateVerifier.BN254G1Point{
			X: new(big.Int).SetBytes(g1Bytes[0:32]),
			Y: new(big.Int).SetBytes(g1Bytes[32:64]),
		},
		Apk: IBN254CertificateVerifier.BN254G2Point{
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

	// Call verifyCertificateProportion directly to get a bool result
	operatorSet := IBN254CertificateVerifier.OperatorSet{
		Avs: avsAddress,
		Id:  operatorSetId,
	}

	// Log what we're about to send to the verifier
	cc.logger.Sugar().Infow("Calling verifyCertificateProportion",
		"contractAddress", cc.coreContracts.BN254CertificateVerifier,
		"threshold", fmt.Sprintf("%d/10000 (%.1f%%)", thresholdPercentage, float64(thresholdPercentage)/100),
	)

	opts, err := cc.signer.GetTransactOpts(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get transaction options: %w", err)
	}

	// Get the parsed ABI from the metadata
	parsedABI, err := IBN254CertificateVerifier.IBN254CertificateVerifierMetaData.GetAbi()
	if err != nil {
		return false, fmt.Errorf("failed to get ABI: %w", err)
	}

	// Pack the method call data
	data, err := parsedABI.Pack(
		"verifyCertificateProportion",
		operatorSet,
		cert,
		[]uint16{thresholdPercentage},
	)
	if err != nil {
		return false, fmt.Errorf("failed to pack method data: %w", err)
	}

	contractAddr := common.HexToAddress(cc.coreContracts.BN254CertificateVerifier)
	callMsg := geth.CallMsg{
		From: opts.From,
		To:   &contractAddr,
		Data: data,
	}

	// TODO: set block number
	result, err := cc.ethclient.CallContract(ctx, callMsg, nil)
	if err != nil {
		return false, fmt.Errorf("failed to call contract: %w", err)
	}

	// Unpack the boolean result from the bytes
	var verified bool
	err = parsedABI.UnpackIntoInterface(&verified, "verifyCertificateProportion", result)
	if err != nil {
		return false, fmt.Errorf("failed to unpack result: %w", err)
	}

	cc.logger.Sugar().Infow("Certificate verification simulation result",
		"verified", verified,
		"avsAddress", avsAddress.Hex(),
		"operatorSetId", operatorSetId,
		"threshold", thresholdPercentage,
	)

	// Only proceed with transaction if verification would succeed
	if !verified {
		return false, nil
	}

	tx, err := cc.bn254CertVerifier.VerifyCertificateProportion(
		opts,
		operatorSet,
		cert,
		[]uint16{thresholdPercentage},
	)
	if err != nil {
		return false, fmt.Errorf("certificate verification failed: %w", err)
	}

	// Send the transaction and wait for receipt
	receipt, err := cc.signAndSendTransaction(ctx, tx, "VerifyBN254Certificate")
	if err != nil {
		return false, fmt.Errorf("failed to send verification transaction: %w", err)
	}

	cc.logger.Sugar().Infow("BN254 certificate verification result",
		"verified", verified,
		"avsAddress", avsAddress.Hex(),
		"operatorSetId", operatorSetId,
		"threshold", thresholdPercentage,
		"txHash", receipt.TxHash.Hex(),
		"gasUsed", receipt.GasUsed,
	)

	return verified, nil
}

func (cc *ContractCaller) generateOperatorMerkleProofs(
	nonSignerOperators []contractCaller.BN254NonSignerOperator,
	operatorInfoTreeRoot [32]byte,
	allOperators []OperatorInfo,
) (map[uint32][]byte, error) {

	leaves := make([][]byte, len(allOperators))
	for i, op := range allOperators {
		encodedLeaf, err := util.EncodeOperatorInfoLeaf(op.PubkeyX, op.PubkeyY, op.Weights)
		if err != nil {
			return nil, fmt.Errorf("failed to encode leaf for operator %d: %w", i, err)
		}

		leaves[i] = encodedLeaf
	}

	tree, err := merkletree.NewTree(
		merkletree.WithData(leaves),
		merkletree.WithHashType(keccak256.New()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create merkle tree: %w", err)
	}

	calculatedRoot := tree.Root()

	if !bytes.Equal(calculatedRoot, operatorInfoTreeRoot[:]) {
		cc.logger.Sugar().Errorw("Merkle root mismatch",
			"calculatedRoot", fmt.Sprintf("0x%x", calculatedRoot),
			"expectedRoot", fmt.Sprintf("0x%x", operatorInfoTreeRoot),
			"numLeaves", len(leaves),
		)

		return nil, fmt.Errorf("merkle root mismatch: calculated %x, expected %x",
			calculatedRoot, operatorInfoTreeRoot)
	}

	proofs := make(map[uint32][]byte)
	for _, nonSigner := range nonSignerOperators {
		proof, err := tree.GenerateProofWithIndex(uint64(nonSigner.OperatorIndex), 0)
		if err != nil {
			return nil, fmt.Errorf("failed to generate proof for operator %d: %w",
				nonSigner.OperatorIndex, err)
		}

		proofBytes := flattenProofHashes(proof.Hashes)
		proofs[nonSigner.OperatorIndex] = proofBytes
	}

	return proofs, nil
}

func flattenProofHashes(hashes [][]byte) []byte {
	if len(hashes) == 0 {
		return []byte{}
	}

	proofBytes := make([]byte, 0, len(hashes)*32)
	for _, hash := range hashes {
		proofBytes = append(proofBytes, hash...)
	}
	return proofBytes
}

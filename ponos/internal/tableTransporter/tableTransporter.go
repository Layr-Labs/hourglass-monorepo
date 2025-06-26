package tableTransporter

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	"github.com/Layr-Labs/multichain-go/pkg/blsSigner"
	"github.com/Layr-Labs/multichain-go/pkg/chainManager"
	"github.com/Layr-Labs/multichain-go/pkg/operatorTableCalculator"
	"github.com/Layr-Labs/multichain-go/pkg/transport"
	"github.com/Layr-Labs/multichain-go/pkg/txSigner"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
	"math/big"
)

func TransportTable(
	transporterPrivateKey string,
	l1RpcUrl string,
	l1ChainId uint64,
	l2RpcUrl string,
	l2ChainId uint64,
	crossChainRegistryAddress string,
	blsPrivateKey string,
	chainIdsToIgnore []*big.Int,
	l *zap.Logger,
) {
	ctx := context.Background()

	cm := chainManager.NewChainManager()

	holeskyAnvilConfig := &chainManager.ChainConfig{
		ChainID: l1ChainId,
		RPCUrl:  l1RpcUrl,
	}
	if err := cm.AddChain(holeskyAnvilConfig); err != nil {
		l.Sugar().Fatalf("Failed to add chain: %v", err)
	}
	holeskyClient, err := cm.GetChainForId(holeskyAnvilConfig.ChainID)
	if err != nil {
		l.Sugar().Fatalf("Failed to get chain for ID %d: %v", holeskyAnvilConfig.ChainID, err)
	}

	if l2RpcUrl != "" && l2ChainId != 0 {
		l2ChainConfig := &chainManager.ChainConfig{
			ChainID: l2ChainId,
			RPCUrl:  l2RpcUrl,
		}
		if err := cm.AddChain(l2ChainConfig); err != nil {
			l.Sugar().Fatalf("Failed to add L2 chain: %v", err)
		}
		l.Sugar().Infow("Added L2 chain",
			zap.Any("chainConfig", l2ChainConfig),
		)
	}

	txSign, err := txSigner.NewPrivateKeySigner(transporterPrivateKey)
	if err != nil {
		l.Sugar().Fatalf("Failed to create private key signer: %v", err)
	}

	l.Sugar().Infow("Using CrossChainRegistryAddress",
		zap.String("crossChainRegistryAddress", crossChainRegistryAddress),
	)

	tableCalc, err := operatorTableCalculator.NewStakeTableRootCalculator(&operatorTableCalculator.Config{
		CrossChainRegistryAddress: common.HexToAddress(crossChainRegistryAddress),
	}, holeskyClient.RPCClient, l)
	if err != nil {
		l.Sugar().Fatalf("Failed to create StakeTableRootCalculator: %v", err)
	}

	blockNumber, err := holeskyClient.RPCClient.BlockNumber(ctx)
	if err != nil {
		l.Sugar().Fatalf("Failed to get block number: %v", err)
	}
	// blockNumber = blockNumber - 2
	block, err := holeskyClient.RPCClient.BlockByNumber(ctx, big.NewInt(int64(blockNumber)))
	if err != nil {
		l.Sugar().Fatalf("Failed to get block by number: %v", err)
	}

	root, tree, dist, err := tableCalc.CalculateStakeTableRoot(ctx, block.NumberU64())
	if err != nil {
		l.Sugar().Fatalf("Failed to calculate stake table root: %v", err)
	}

	pk, err := bn254.NewPrivateKeyFromHexString(blsPrivateKey)
	if err != nil {
		l.Sugar().Fatalf("Failed to convert BLS private key: %v", err)
	}

	inMemSigner, err := blsSigner.NewInMemoryBLSSigner(pk)
	if err != nil {
		l.Sugar().Fatalf("Failed to create in-memory BLS signer: %v", err)
	}

	stakeTransport, err := transport.NewTransport(
		&transport.TransportConfig{
			L1CrossChainRegistryAddress: common.HexToAddress(crossChainRegistryAddress),
		},
		holeskyClient.RPCClient,
		inMemSigner,
		txSign,
		cm,
		l,
	)
	if err != nil {
		l.Sugar().Fatalf("Failed to create transport: %v", err)
	}

	referenceTimestamp := uint32(block.Time())

	err = stakeTransport.SignAndTransportGlobalTableRoot(
		ctx,
		root,
		referenceTimestamp,
		block.NumberU64(),
		chainIdsToIgnore,
	)
	if err != nil {
		l.Sugar().Fatalf("Failed to sign and transport global table root: %v", err)
	}

	opsets := dist.GetOperatorSets()
	if len(opsets) == 0 {
		l.Sugar().Infow("No operator sets found, skipping AVS stake table transport")
		return
	}
	fmt.Printf("Operatorsets to transport: %+v\n", opsets)
	for _, opset := range opsets {
		err = stakeTransport.SignAndTransportAvsStakeTable(
			ctx,
			referenceTimestamp,
			block.NumberU64(),
			opset,
			root,
			tree,
			dist,
			chainIdsToIgnore,
		)
		if err != nil {
			l.Sugar().Fatalf("Failed to sign and transport AVS stake table for opset %v: %v", opset, err)
		} else {
			l.Sugar().Infof("Successfully signed and transported AVS stake table for opset %v", opset)
		}
	}
}

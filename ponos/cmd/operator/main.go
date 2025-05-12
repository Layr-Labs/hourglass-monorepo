package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller/caller"
	cryptoUtils "github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/crypto"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/bn254"
	"github.com/ethereum/go-ethereum/common"
)

const (
	rpcUrl = "http://localhost:8545"
	// rpcUrl              = "https://virtual.mainnet.rpc.tenderly.co/a0e34c85-7746-4cd8-9856-371fe7c95465"
	avsAddress          = "0x70997970c51812dc3a010c7d01b50e0d17dc79c8"
	avsRegistrarAddress = "0xf4c5c29b14f0237131f7510a51684c8191f98e06"
	taskMailboxAddress  = "0x7306a649b451ae08781108445425bd4e8acf1e00"
	// operatorPrivateKey  = "0x5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a"
	operatorPrivateKey = "0x7c852118294e51e653712a81e05800f419141751be58f605c371e15141b007a6" // 0x90f79bf6eb2c4f870365e785982e1f101e93b906
)

func main() {
	ctx := context.Background()
	l, err := logger.NewLogger(&logger.LoggerConfig{
		Debug: false,
	})
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}

	opPrivateKey, err := cryptoUtils.StringToECDSAPrivateKey(operatorPrivateKey)
	if err != nil {
		l.Sugar().Fatalf("failed to convert private key: %v", err)
	}
	operatorAddress := cryptoUtils.DeriveAddress(opPrivateKey)
	fmt.Printf("\nOperator address: %s\n", operatorAddress.Hex())

	ethereumClient := ethereum.NewEthereumClient(&ethereum.EthereumClientConfig{
		BaseUrl: rpcUrl,
	}, l)

	ethClient, err := ethereumClient.GetEthereumContractCaller()
	if err != nil {
		l.Sugar().Fatalf("failed to get Ethereum contract caller: %v", err)
	}

	cc, err := caller.NewContractCaller(&caller.ContractCallerConfig{
		PrivateKey:          operatorPrivateKey,
		AVSRegistrarAddress: avsRegistrarAddress,
		TaskMailboxAddress:  taskMailboxAddress,
	}, ethClient, l)
	if err != nil {
		l.Sugar().Fatalf("failed to create contract caller: %v", err)
	}

	privateKey, publicKey, err := bn254.GenerateKeyPair()
	if err != nil {
		l.Sugar().Fatalf("failed to generate key pair: %v", err)
	}

	g1Point, err := cc.GetOperatorRegistrationMessageHash(ctx, operatorAddress)
	if err != nil {
		l.Sugar().Fatalf("failed to get operator registration message hash: %v", err)
	}

	// Create G1 point from contract coordinates
	hashPoint := bn254.NewG1Point(g1Point.X, g1Point.Y)

	// Sign the hash point
	signature, err := privateKey.SignG1Point(hashPoint.G1Affine)
	if err != nil {
		l.Sugar().Fatalf("failed to sign hash point: %v", err)
	}

	// Register with AVS
	result, err := cc.CreateOperatorAndRegisterWithAvs(
		ctx,
		common.HexToAddress(avsAddress),
		operatorAddress,
		[]uint32{0},
		publicKey,
		signature,
		"localhost:8545",
		7200,
		"http://localhost:8545",
	)
	if err != nil {
		l.Sugar().Fatalf("failed to register with AVS: %v", err)
	}
	l.Sugar().Infof("Successfully registered with AVS. Result: %v", result)
}

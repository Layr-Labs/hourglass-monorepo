package main

import (
	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller/caller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionSigner"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"go.uber.org/zap"
)

const (
	//RPCUrl = "https://practical-serene-mound.ethereum-sepolia.quiknode.pro/3aaa48bd95f3d6aed60e89a1a466ed1e2a440b61/"
	RPCUrl = "http://localhost:8545"
)

func main() {
	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	if err != nil {
		panic(err)
	}

	pk, _, err := bn254.GenerateKeyPair()
	if err != nil {
		panic(err)
	}
	privateKey, err := bn254.NewPrivateKeyFromBytes(pk.Bytes())
	if err != nil {
		panic(err)
	}

	ethereumClient := ethereum.NewEthereumClient(&ethereum.EthereumClientConfig{
		BaseUrl:   RPCUrl,
		BlockType: ethereum.BlockType_Latest,
	}, l)

	ethClient, err := ethereumClient.GetEthereumContractCaller()
	if err != nil {
		l.Sugar().Fatalf("failed to get Ethereum contract caller: %v", err)
		return
	}

	privateKeySigner, err := transactionSigner.NewPrivateKeySigner("0x90a7b1bcc84977a8b008fea51da40ad7e58b844095b13518f575ded17a4c67e4", ethClient, l)
	if err != nil {
		l.Sugar().Fatalf("failed to create private key signer: %v", err)
		return
	}

	aggregatorCc, err := caller.NewContractCaller(ethClient, privateKeySigner, l)
	if err != nil {
		l.Sugar().Fatalf("failed to create aggregator contract caller: %v", err)
		return
	}

	keyData, err := aggregatorCc.EncodeBN254KeyData(privateKey.Public())
	if err != nil {
		l.Sugar().Fatalf("failed to encode BN254 key data: %v", err)
		return
	}
	l.Sugar().Infow("Encoded BN254 key data",
		zap.String("keyData", hexutil.Encode(keyData)),
	)
}

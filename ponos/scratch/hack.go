package main

import (
	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller/caller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"go.uber.org/zap"
)

const (
	//RPCUrl = "https://special-yolo-river.ethereum-holesky.quiknode.pro/2d21099a19e7c896a22b9fcc23dc8ce80f2214a5/"
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

	aggregatorCc, err := caller.NewContractCaller(&caller.ContractCallerConfig{
		PrivateKey:          "0x90a7b1bcc84977a8b008fea51da40ad7e58b844095b13518f575ded17a4c67e4",
		AVSRegistrarAddress: "0x5897a9b8b746c78e0cae876962796949832e3357",
		TaskMailboxAddress:  "0xf481bf37a8e87898b03c5eccee79da7f20a0f58e",
		KeyRegistrarAddress: "0x1c84bb62fe7791e173014a879c706445fa893bbe",
	}, ethClient, l)
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

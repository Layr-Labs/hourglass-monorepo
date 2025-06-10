//nolint:all
package main

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/internal/testUtils"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller/caller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"math/big"
)

const (
	privateKey             = "5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a"
	mailboxContractAddress = "0x4B7099FD879435a087C364aD2f9E7B3f94d20bBe"
)

func main() {
	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	if err != nil {
		panic(err)
	}

	root := testUtils.GetProjectRootPath()
	chainConfig, err := testUtils.ReadChainConfig(root)
	_ = chainConfig

	client := ethereum.NewEthereumClient(&ethereum.EthereumClientConfig{
		BaseUrl:   "http://localhost:8545",
		BlockType: ethereum.BlockType_Latest,
	}, l)

	ethCaller, err := client.GetEthereumContractCaller()
	if err != nil {
		panic(err)
	}

	cc, err := caller.NewContractCaller(&caller.ContractCallerConfig{
		PrivateKey:          "0x3dd7c381f27775d9945f0fcf5bb914484c4d01681824603c71dd762259f43214",
		AVSRegistrarAddress: "0x5897a9b8b746c78e0cae876962796949832e3357",
		TaskMailboxAddress:  "0xf481bf37a8e87898b03c5eccee79da7f20a0f58e",
	}, ethCaller, l)
	if err != nil {
		panic(err)
	}

	payloadJsonBytes := util.BigIntToHex(new(big.Int).SetUint64(4))
	receipt, err := cc.PublishMessageToInbox(context.Background(), "0xCE2Ac75bE2E0951F1F7B288c7a6A9BfB6c331DC4", 1, payloadJsonBytes)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Receipt: %+v\n", receipt)

}

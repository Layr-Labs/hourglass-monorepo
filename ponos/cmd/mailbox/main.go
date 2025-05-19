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
		PrivateKey:          "0x4bbbf85ce3377467afe5d46f804f221813b2bb87f24d81f60f1fcdbf7cbf4356",
		AVSRegistrarAddress: "0x99aA73dA6309b8eC484eF2C95e96C131C1BBF7a0",
		TaskMailboxAddress:  "0x4B7099FD879435a087C364aD2f9E7B3f94d20bBe",
	}, ethCaller, l)
	if err != nil {
		panic(err)
	}

	payloadJsonBytes := util.BigIntToHex(new(big.Int).SetUint64(4))
	receipt, err := cc.PublishMessageToInbox(context.Background(), "0x70997970C51812dc3A010C7d01b50e0d17dc79C8", 1, payloadJsonBytes)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Receipt: %+v\n", receipt)

}

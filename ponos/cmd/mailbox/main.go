//nolint:all
package main

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/internal/testUtils"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller/caller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionSigner"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"math/big"
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
		BaseUrl:   "http://localhost:9545",
		BlockType: ethereum.BlockType_Latest,
	}, l)

	ethCaller, err := client.GetEthereumContractCaller()
	if err != nil {
		panic(err)
	}

	privateKeySigner, err := transactionSigner.NewPrivateKeySigner(chainConfig.AVSAccountPrivateKey, ethCaller, l)
	if err != nil {
		panic(err)
	}

	cc, err := caller.NewContractCaller(ethCaller, privateKeySigner, l)
	if err != nil {
		panic(err)
	}

	payloadJsonBytes := util.BigIntToHex(new(big.Int).SetUint64(4))
	receipt, err := cc.PublishMessageToInbox(context.Background(), chainConfig.AVSAccountAddress, 1, payloadJsonBytes)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Receipt: %+v\n", receipt)

}

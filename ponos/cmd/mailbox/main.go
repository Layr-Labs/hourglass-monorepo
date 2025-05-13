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
	mailboxContractAddress = "0x7306a649b451ae08781108445425bd4e8acf1e00"
)

func main() {
	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	if err != nil {
		panic(err)
	}

	root := testUtils.GetProjectRootPath()
	chainConfig, err := testUtils.ReadChainConfig(root)

	client := ethereum.NewEthereumClient(&ethereum.EthereumClientConfig{
		BaseUrl:   "http://localhost:8545",
		BlockType: ethereum.BlockType_Latest,
	}, l)

	ethCaller, err := client.GetEthereumContractCaller()
	if err != nil {
		panic(err)
	}

	cc, err := caller.NewContractCaller(&caller.ContractCallerConfig{
		PrivateKey:          chainConfig.AppAccountPrivateKey,
		AVSRegistrarAddress: chainConfig.AVSTaskRegistrarAddress,
		TaskMailboxAddress:  chainConfig.MailboxContractAddress,
	}, ethCaller, l)
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

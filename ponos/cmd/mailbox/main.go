package main

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller/caller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
)

const (
	privateKey             = "59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d"
	mailboxContractAddress = "0x74E7CF978C61685dB8527086CD66316Ce7aF295c"
)

func main() {
	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	if err != nil {
		panic(err)
	}

	client := ethereum.NewEthereumClient(&ethereum.EthereumClientConfig{
		BaseUrl:   "http://localhost:8545",
		BlockType: ethereum.BlockType_Latest,
	}, l)

	ethCaller, err := client.GetEthereumContractCaller()
	if err != nil {
		panic(err)
	}

	cc, err := caller.NewContractCaller(&caller.ContractCallerConfig{
		PrivateKey:          privateKey,
		AVSRegistrarAddress: "0x1234567890abcdef1234567890abcdef12345678",
		TaskMailboxAddress:  mailboxContractAddress,
	}, ethCaller, l)
	if err != nil {
		panic(err)
	}

	payloadJsonBytes := []byte(`{ "numberToBeSquared": 4 }`)

	receipt, err := cc.PublishMessageToInbox(context.Background(), "0x1234", 1, payloadJsonBytes)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Receipt: %+v\n", receipt)

}

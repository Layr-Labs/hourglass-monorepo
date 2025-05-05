package main

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller/caller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
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
		AVSRegistrarAddress: "0xf4c5c29b14f0237131f7510a51684c8191f98e06",
		TaskMailboxAddress:  mailboxContractAddress,
	}, ethCaller, l)
	if err != nil {
		panic(err)
	}

	payloadJsonBytes := []byte(`{ "numberToBeSquared": 4 }`)

	receipt, err := cc.PublishMessageToInbox(context.Background(), "0x70997970c51812dc3a010c7d01b50e0d17dc79c8", 1, payloadJsonBytes)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Receipt: %+v\n", receipt)

}

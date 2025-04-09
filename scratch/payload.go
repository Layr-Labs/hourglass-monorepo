package main

import (
	"fmt"
	"github.com/Layr-Labs/go-ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"google.golang.org/protobuf/proto"
)

func main() {

	p := &executor.PeerInfo{
		NetworkAddress: "some-potentially-really-long-address.com",
		Port:           5432,
		PublicKey:      "some really long public key",
	}

	pBytes, err := proto.Marshal(p)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Serialized bytes: %d\n", len(pBytes))

}

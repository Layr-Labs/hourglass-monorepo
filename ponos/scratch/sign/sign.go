package main

import (
	"fmt"
	"github.com/Layr-Labs/crypto-libs/pkg/ecdsa"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/iden3/go-iden3-crypto/keccak256"
	"strings"
)

func main() {
	privateKeyStr := "0x3dd7c381f27775d9945f0fcf5bb914484c4d01681824603c71dd762259f43214"
	expectedAddress := "0x6B58f6762689DF33fe8fa3FC40Fb5a3089D3a8cc"

	privateKey, err := ecdsa.NewPrivateKeyFromHexString(privateKeyStr)
	if err != nil {
		panic(err)
	}
	derivedAddress, err := privateKey.DeriveAddress()
	if err != nil {
		panic(err)
	}
	if !strings.EqualFold(expectedAddress, derivedAddress.String()) {
		fmt.Printf("Expected address: %s\n", expectedAddress)
		fmt.Printf("Derived address:  %s\n", derivedAddress.String())
		panic("Address does not match expected address")
	}

	messageHash, err := hexutil.Decode("0x4a94a75be50eeaf849804270cebf9370d4dc4793c895807a9625014f7af115f1")
	if err != nil {
		panic(fmt.Errorf("failed to decode message hash: %w", err))
	}

	signature, err := privateKey.Sign(messageHash)
	if err != nil {
		panic(fmt.Errorf("failed to sign message: %w", err))
	}
	fmt.Printf("Raw sig: %+v\n", signature)
	fmt.Printf("Signature: %s\n", hexutil.Encode(signature.Bytes()))

	isValid, err := signature.VerifyWithAddress(messageHash, derivedAddress)
	if err != nil {
		panic(fmt.Errorf("failed to verify signature: %w", err))
	}
	if !isValid {
		panic("Signature verification failed")
	} else {
		fmt.Println("Signature verification succeeded")
	}

	testMessage := []byte("test message")

	var otherHash [32]byte
	copy(otherHash[:], keccak256.Hash(testMessage))
	sig2, err := privateKey.Sign(otherHash[:])
	if err != nil {
		panic(fmt.Errorf("failed to sign test message: %w", err))
	}
	valid, err := sig2.VerifyWithAddress(otherHash[:], derivedAddress)
	if err != nil {
		panic(fmt.Errorf("failed to verify test message signature: %w", err))
	}
	if !valid {
		panic("Test message signature verification failed")
	} else {
		fmt.Println("Test message signature verification succeeded")
	}

}

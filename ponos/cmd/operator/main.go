package main

import (
	"context"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller/caller"
	cryptoUtils "github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/crypto"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/bn254"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"log"
	"math/big"
)

const (
	// rpcUrl              = "http://localhost:8545"
	rpcUrl              = "https://virtual.mainnet.rpc.tenderly.co/9d86e329-5246-4485-86fe-f8dc7875bfca"
	avsAddress          = "0x70997970c51812dc3a010c7d01b50e0d17dc79c8"
	avsRegistrarAddress = "0xf4c5c29b14f0237131f7510a51684c8191f98e06"
	taskMailboxAddress  = "0x7306a649b451ae08781108445425bd4e8acf1e00"
	operatorPrivateKey  = "0x5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a"
)

var (
	domainName     = "TaskAVSRegistrar"
	domainVersion  = "v0.1.0"
	typehashString = "BN254PubkeyRegistration(address operator)"
)

// calculateDomainSeparator calculates the EIP-712 domain separator
func calculateDomainSeparator(chainID *big.Int, contractAddress common.Address) []byte {
	// EIP-712 domain separator: keccak256(abi.encode(
	//     keccak256("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"),
	//     keccak256(bytes(name)),
	//     keccak256(bytes(version)),
	//     chainId,
	//     verifyingContract))
	domainTypehash := crypto.Keccak256([]byte("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"))
	nameHash := crypto.Keccak256([]byte(domainName))
	versionHash := crypto.Keccak256([]byte(domainVersion))

	// Encode the domain data with contract address
	chainIDPadded := common.LeftPadBytes(chainID.Bytes(), 32)
	contractAddressPadded := common.LeftPadBytes(contractAddress.Bytes(), 32)

	// Calculate the domain separator with contract address
	return crypto.Keccak256(
		append(
			append(
				append(
					append(
						domainTypehash,
						nameHash...,
					),
					versionHash...,
				),
				chainIDPadded...,
			),
			contractAddressPadded...,
		),
	)
}

func main() {
	ctx := context.Background()
	l, err := logger.NewLogger(&logger.LoggerConfig{
		Debug: false,
	})
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}

	opPrivateKey, err := cryptoUtils.StringToECDSAPrivateKey(operatorPrivateKey)
	if err != nil {
		l.Sugar().Fatalf("failed to convert private key: %v", err)
	}
	operatorAddress := cryptoUtils.DeriveAddress(opPrivateKey)

	ethereumClient := ethereum.NewEthereumClient(&ethereum.EthereumClientConfig{
		BaseUrl: rpcUrl,
	}, l)

	ethClient, err := ethereumClient.GetEthereumContractCaller()
	if err != nil {
		l.Sugar().Fatalf("failed to get Ethereum contract caller: %v", err)
	}

	cc, err := caller.NewContractCaller(&caller.ContractCallerConfig{
		PrivateKey:          operatorPrivateKey,
		AVSRegistrarAddress: avsRegistrarAddress,
		TaskMailboxAddress:  taskMailboxAddress,
	}, ethClient, l)
	if err != nil {
		l.Sugar().Fatalf("failed to create contract caller: %v", err)
	}

	privateKey, publicKey, err := bn254.GenerateKeyPair()
	if err != nil {
		l.Sugar().Fatalf("failed to generate key pair: %v", err)
	}

	g1Point, err := cc.GetOperatorRegistrationMessageHash(ctx, operatorAddress)
	if err != nil {
		l.Sugar().Fatalf("failed to get operator registration message hash: %v", err)
	}

	hashPoint := bn254.NewG1Point(g1Point.X, g1Point.Y)

	sig, err := privateKey.SignG1Point(hashPoint.G1Affine)
	if err != nil {
		l.Sugar().Fatalf("failed to sign hash point: %v", err)
	}

	var sigBytes [32]byte
	copy(sigBytes[:], sig.Bytes())

	_, err = cc.CreateOperatorAndRegisterWithAvs(
		ctx,
		common.HexToAddress(avsAddress),
		operatorAddress,
		[]uint32{0},
		publicKey,
		sig,
		"localhost:8545",
		7200,
		"http://localhost:8545",
	)
	if err != nil {
		l.Sugar().Fatalf("failed to create operator and register with AVS: %v", err)
	}
}

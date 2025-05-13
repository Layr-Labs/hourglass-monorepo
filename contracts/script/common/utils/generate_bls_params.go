package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// Domain values from the contract
var (
	domainName     = "TaskAVSRegistrar"
	domainVersion  = "v0.1.0"
	typehashString = "BN254PubkeyRegistration(address operator)"
)

func main() {
	if len(os.Args) < 6 {
		fmt.Println("Usage: go run generate_bls_params.go <hash_point_x> <hash_point_y>")
		fmt.Println("The hash_point_x and hash_point_y values should be obtained from calling pubkeyRegistrationMessageHash on the contract.")
		os.Exit(1)
	}

	// Parse the hash point coordinates from command line
	hashPointX, ok := new(big.Int).SetString(os.Args[4], 10)
	if !ok {
		fmt.Println("Invalid hash point X coordinate")
		os.Exit(1)
	}

	hashPointY, ok := new(big.Int).SetString(os.Args[5], 10)
	if !ok {
		fmt.Println("Invalid hash point Y coordinate")
		os.Exit(1)
	}

	// Generate a random private key for BLS signing
	var privateKeyBytes [32]byte
	_, err := rand.Read(privateKeyBytes[:])
	if err != nil {
		fmt.Printf("Error generating private key: %v\n", err)
		os.Exit(1)
	}

	// Convert to scalar in Fr field
	var privateKey fr.Element
	privateKey.SetBytes(privateKeyBytes[:])
	privateKeyBigInt := privateKey.BigInt(new(big.Int))

	// Create a proper G1 generator point
	var g1Gen bn254.G1Affine
	_, err = g1Gen.X.SetString("1")
	if err != nil {
		fmt.Printf("Error setting G1 generator X: %v\n", err)
		os.Exit(1)
	}
	_, err = g1Gen.Y.SetString("2")
	if err != nil {
		fmt.Printf("Error setting G1 generator Y: %v\n", err)
		os.Exit(1)
	}

	// Create a proper G2 generator point
	var g2Gen bn254.G2Affine
	// These are the coordinates of the standard BN254 G2 generator
	_, err = g2Gen.X.A0.SetString("10857046999023057135944570762232829481370756359578518086990519993285655852781")
	if err != nil {
		fmt.Printf("Error setting G2 generator X.A0: %v\n", err)
		os.Exit(1)
	}
	_, err = g2Gen.X.A1.SetString("11559732032986387107991004021392285783925812861821192530917403151452391805634")
	if err != nil {
		fmt.Printf("Error setting G2 generator X.A1: %v\n", err)
		os.Exit(1)
	}
	_, err = g2Gen.Y.A0.SetString("8495653923123431417604973247489272438418190587263600148770280649306958101930")
	if err != nil {
		fmt.Printf("Error setting G2 generator Y.A0: %v\n", err)
		os.Exit(1)
	}
	_, err = g2Gen.Y.A1.SetString("4082367875863433681332203403145435568316851327593401208105741076214120093531")
	if err != nil {
		fmt.Printf("Error setting G2 generator Y.A1: %v\n", err)
		os.Exit(1)
	}

	// Calculate public keys
	// G1 public key
	var pubkeyG1 bn254.G1Affine
	pubkeyG1.ScalarMultiplication(&g1Gen, privateKeyBigInt)

	// G2 public key
	var pubkeyG2 bn254.G2Affine
	pubkeyG2.ScalarMultiplication(&g2Gen, privateKeyBigInt)

	// Use the hash point coordinates directly from the contract
	var hashPoint bn254.G1Affine

	// Set the point coordinates directly using SetString
	_, err = hashPoint.X.SetString(hashPointX.String())
	if err != nil {
		fmt.Printf("Error setting hash point X: %v\n", err)
		os.Exit(1)
	}

	_, err = hashPoint.Y.SetString(hashPointY.String())
	if err != nil {
		fmt.Printf("Error setting hash point Y: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Hash point: E([%s,%s])\n", hashPointX.String(), hashPointY.String())

	// Sign the message (scalar multiplication of the hash point by private key)
	var signature bn254.G1Affine
	signature.ScalarMultiplication(&hashPoint, privateKeyBigInt)

	// Format for Solidity
	pubkeyRegistrationParams := formatForSolidity(signature, pubkeyG1, pubkeyG2)

	fmt.Println("PUBKEY_REGISTRATION_PARAMS=", pubkeyRegistrationParams)
	fmt.Println("\nPrivate Key (keep secure):", privateKeyBigInt.String())
}

// calculatePubkeyRegistrationMessageHash calculates the EIP-712 hash for BLS public key registration
func calculatePubkeyRegistrationMessageHash(operator common.Address, chainID *big.Int, contractAddress common.Address) []byte {
	// Calculate PUBKEY_REGISTRATION_TYPEHASH = keccak256("BN254PubkeyRegistration(address operator)")
	pubkeyRegistrationTypehash := crypto.Keccak256([]byte(typehashString))

	// Calculate the domain separator for EIP-712
	domainSeparator := calculateDomainSeparator(chainID, contractAddress)

	// Encode the message: keccak256(abi.encode(PUBKEY_REGISTRATION_TYPEHASH, operator))
	encodedMessage := crypto.Keccak256(
		append(pubkeyRegistrationTypehash, common.LeftPadBytes(operator.Bytes(), 32)...),
	)

	// Calculate _hashTypedDataV4: keccak256(0x1901 + domainSeparator + hashStruct)
	result := crypto.Keccak256(
		append(
			append([]byte{0x19, 0x01}, domainSeparator...),
			encodedMessage...,
		),
	)

	return result
}

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

// Format the BLS parameters for Solidity
func formatForSolidity(signature, pubkeyG1 bn254.G1Affine, pubkeyG2 bn254.G2Affine) string {
	// Get G1 point coordinates
	sigX, sigY := signature.X, signature.Y
	pkG1X, pkG1Y := pubkeyG1.X, pubkeyG1.Y

	// Get G2 point coordinates
	// Note: G2 points use two coordinates for each of X and Y
	pkG2X0, pkG2X1 := pubkeyG2.X.A0, pubkeyG2.X.A1
	pkG2Y0, pkG2Y1 := pubkeyG2.Y.A0, pubkeyG2.Y.A1

	// Convert to bytes and format as hex strings
	params := []string{
		// Signature (G1 point)
		formatBigInt(sigX.BigInt(new(big.Int))),
		formatBigInt(sigY.BigInt(new(big.Int))),

		// PubkeyG1 (G1 point)
		formatBigInt(pkG1X.BigInt(new(big.Int))),
		formatBigInt(pkG1Y.BigInt(new(big.Int))),

		// PubkeyG2 (G2 point)
		// Note: Order matters due to the way EVM expects G2 points
		formatBigInt(pkG2X1.BigInt(new(big.Int))), // X.A1 first
		formatBigInt(pkG2X0.BigInt(new(big.Int))), // X.A0 second
		formatBigInt(pkG2Y1.BigInt(new(big.Int))), // Y.A1 first
		formatBigInt(pkG2Y0.BigInt(new(big.Int))), // Y.A0 second
	}

	// Join all hex strings without separators, add 0x prefix
	return "0x" + strings.Join(params, "")
}

// Format a big.Int as a 32-byte hex string without 0x prefix
func formatBigInt(b *big.Int) string {
	// Ensure 32 bytes (64 hex chars)
	return fmt.Sprintf("%064s", hex.EncodeToString(b.Bytes()))
}

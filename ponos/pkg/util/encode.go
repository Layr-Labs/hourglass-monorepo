package util

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

const (
	operatorInfoLeafSalt = 0x75
)

// The Solidity struct is:
// struct BN254OperatorInfo {
//     BN254.G1Point pubkey;  // struct { uint256 X; uint256 Y; }
//     uint256[] weights;
// }
//
// When we call abi.encode(operatorInfo) in Solidity, it produces:
// - The pubkey.X value (32 bytes)
// - The pubkey.Y value (32 bytes)
// - Offset to weights array (32 bytes) = 0x60
// - Length of weights array (32 bytes)
// - Weight values (32 bytes each)

// EncodeOperatorInfoLeaf Use go-ethereum's ABI library to encode this properly
// The Solidity contract uses: abi.encodePacked(OPERATOR_INFO_LEAF_SALT, abi.encode(operatorInfo))
// where operatorInfo is a struct with pubkey (G1Point) and weights (uint256[])
func EncodeOperatorInfoLeaf(pubkeyX, pubkeyY *big.Int, weights []*big.Int) ([]byte, error) {

	// Define the operatorInfo struct as a single tuple argument
	// This matches Solidity's abi.encode(operatorInfo) which treats the struct as one argument
	operatorInfoType, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{
			Name: "pubkey",
			Type: "tuple",
			Components: []abi.ArgumentMarshaling{
				{Name: "X", Type: "uint256"},
				{Name: "Y", Type: "uint256"},
			},
		},
		{Name: "weights", Type: "uint256[]"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create operator info type: %w", err)
	}

	// Single argument of type tuple - this adds a 32-byte offset at the beginning
	args := abi.Arguments{
		{Type: operatorInfoType},
	}

	// Create the operator info struct
	operatorInfo := struct {
		Pubkey struct {
			X *big.Int
			Y *big.Int
		}
		Weights []*big.Int
	}{}
	operatorInfo.Pubkey.X = pubkeyX
	operatorInfo.Pubkey.Y = pubkeyY
	operatorInfo.Weights = weights

	encoded, err := args.Pack(operatorInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to ABI encode operator info: %w", err)
	}

	// Prepend the salt (matching Solidity's abi.encodePacked(OPERATOR_INFO_LEAF_SALT, abi.encode(operatorInfo)))
	// abi.encodePacked just concatenates the bytes, so we use append
	result := append([]byte{operatorInfoLeafSalt}, encoded...)
	return result, nil
}

func EncodeString(str string) ([]byte, error) {
	// Define the ABI for a single string parameter
	stringType, _ := abi.NewType("string", "", nil)
	arguments := abi.Arguments{{Type: stringType}}

	// Encode the string
	encoded, err := arguments.Pack(str)
	if err != nil {
		return nil, err
	}

	return encoded, nil
}

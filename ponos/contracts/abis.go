package contracts

import (
	"embed"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"strings"
)

//go:embed *
var abis embed.FS

func GetContractAbi(contractName string, version string) (*abi.ABI, error) {
	abiFile := fmt.Sprintf("abi/%s/%s.abi.json", version, contractName)
	abiBytes, err := abis.ReadFile(abiFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded ABI file %s: %w", abiFile, err)
	}

	parsedABI, err := abi.JSON(strings.NewReader(string(abiBytes)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}
	return &parsedABI, nil
}

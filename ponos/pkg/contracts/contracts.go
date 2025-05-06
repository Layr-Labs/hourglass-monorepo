package contracts

import (
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"regexp"
	"strings"
)

type Contract struct {
	Name        string         `json:"name"`
	Address     string         `json:"address"`
	AbiVersions []string       `json:"abiVersions"`
	ChainId     config.ChainId `json:"chainId"`
}

func (c *Contract) GetCombinedAbis() (string, error) {
	return combineAbis(c.AbiVersions)
}

func (c *Contract) GetAbi() (*abi.ABI, error) {
	combinedAbi, err := c.GetCombinedAbis()
	if err != nil {
		return nil, fmt.Errorf("failed to combine ABIs: %w", err)
	}

	parsedAbi, err := unmarshalJsonToAbi(combinedAbi)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	return parsedAbi, nil
}

func combineAbis(abis []string) (string, error) {
	abisToCombine := make([]string, 0)

	for _, contractAbi := range abis {
		strippedContractAbi := contractAbi[1 : len(contractAbi)-1]
		abisToCombine = append(abisToCombine, strippedContractAbi)
	}

	combinedAbi := fmt.Sprintf("[%s]", strings.Join(abisToCombine, ","))
	return combinedAbi, nil
}

func unmarshalJsonToAbi(json string) (*abi.ABI, error) {
	a := &abi.ABI{}

	err := a.UnmarshalJSON([]byte(json))

	if err != nil {
		foundMatch := false
		// patterns that we're fine to ignore and not treat as an error
		patterns := []*regexp.Regexp{
			regexp.MustCompile(`only single receive is allowed`),
			regexp.MustCompile(`only single fallback is allowed`),
		}

		for _, pattern := range patterns {
			if pattern.MatchString(err.Error()) {
				foundMatch = true
				break
			}
		}

		// If the error isnt one that we can ignore, return it
		if !foundMatch {
			return nil, err
		}
	}

	return a, nil
}

package config

import "k8s.io/apimachinery/pkg/util/validation/field"

type ChainId uint

const (
	ChainId_EthereumMainnet ChainId = 1
	ChainId_EthereumHolesky ChainId = 17000
	ChainId_EthereumHoodi   ChainId = 560048
)

var (
	SupportedChainIds = []ChainId{
		ChainId_EthereumMainnet,
		ChainId_EthereumHolesky,
		ChainId_EthereumHoodi,
	}
)

type SigningKey struct {
	Keystore string `json:"keystore"`
	Password string `json:"password"`
}

type SigningKeys struct {
	BLS *SigningKey `json:"bls"`
}

func (sk *SigningKeys) Validate() error {
	var allErrors field.ErrorList
	if sk.BLS == nil {
		allErrors = append(allErrors, field.Required(field.NewPath("bls"), "bls is required"))
	}
	if len(allErrors) > 0 {
		return allErrors.ToAggregate()
	}
	return nil
}

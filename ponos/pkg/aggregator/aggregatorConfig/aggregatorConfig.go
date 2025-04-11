package aggregatorConfig

import (
	"encoding/json"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/yaml"
	"slices"
)

const (
	EnvPrefix = "AGGREGATOR_"
)

type Chain struct {
	Name    string `json:"name"`
	Network string `json:"network"`
	ChainID uint   `json:"chainId"`
	RpcURL  string `json:"rpcUrl"`
}

func (c *Chain) Validate() field.ErrorList {
	var allErrors field.ErrorList
	if c.Name == "" {
		allErrors = append(allErrors, field.Required(field.NewPath("name"), "name is required"))
	}
	if c.Network == "" {
		allErrors = append(allErrors, field.Required(field.NewPath("network"), "network is required"))
	}
	if c.ChainID == 0 {
		allErrors = append(allErrors, field.Required(field.NewPath("chainId"), "chainId is required"))
	}
	if !slices.Contains(config.SupportedChainIds, config.ChainID(c.ChainID)) {
		allErrors = append(allErrors, field.Invalid(field.NewPath("chainId"), c.ChainID, "unsupported chainId"))
	}

	if c.RpcURL == "" {
		allErrors = append(allErrors, field.Required(field.NewPath("rpcUrl"), "rpcUrl is required"))
	}
	return nil
}

type AggregatorConfig struct {
	Chains []*Chain `json:"chains" yaml:"chains"`
}

func (ac *AggregatorConfig) Validate() error {
	var allErrors field.ErrorList
	for _, chain := range ac.Chains {
		if chainErrors := chain.Validate(); len(chainErrors) > 0 {
			allErrors = append(allErrors, field.Invalid(field.NewPath("chains"), chain, "invalid chain config"))
		}
	}
	return allErrors.ToAggregate()
}

func NewAggregatorConfigFromJsonBytes(data []byte) (*AggregatorConfig, error) {
	var c *AggregatorConfig

	if err := json.Unmarshal(data, &c); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal AggregatorConfig from JSON")
	}
	return c, nil
}

func NewAggregatorConfigFromYamlBytes(data []byte) (*AggregatorConfig, error) {
	var c *AggregatorConfig

	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal AggregatorConfig from YAML")
	}
	return c, nil
}

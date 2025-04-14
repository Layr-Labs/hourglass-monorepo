package aggregatorConfig

import (
	"encoding/json"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/yaml"
	"slices"
)

const (
	EnvPrefix = "AGGREGATOR_"

	Debug     = "debug"
	Simulated = "simulated"

	SimulatedPort        = "simulated.port"
	SimulatedDefaultPort = 8080
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

type AggregatorRuntimeConfig struct {
	Chains []*Chain `json:"chains"`
}

func (arc *AggregatorRuntimeConfig) Validate() error {
	var allErrors field.ErrorList
	for _, chain := range arc.Chains {
		if chainErrors := chain.Validate(); len(chainErrors) > 0 {
			allErrors = append(allErrors, field.Invalid(field.NewPath("chains"), chain, "invalid chain config"))
		}
	}
	return allErrors.ToAggregate()
}

func NewAggregatorConfigFromJsonBytes(data []byte) (*AggregatorRuntimeConfig, error) {
	var c *AggregatorRuntimeConfig

	if err := json.Unmarshal(data, &c); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal AggregatorRuntimeConfig from JSON")
	}
	return c, nil
}

func NewAggregatorConfigFromYamlBytes(data []byte) (*AggregatorRuntimeConfig, error) {
	var c *AggregatorRuntimeConfig

	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal AggregatorRuntimeConfig from YAML")
	}
	return c, nil
}

type AggregatorConfig struct {
	Debug         bool
	Simulated     bool
	SimulatedPort int
	Runtime       AggregatorRuntimeConfig
}

func NewAggregatorConfig() *AggregatorConfig {
	return &AggregatorConfig{
		Debug:         viper.GetBool(config.NormalizeFlagName(Debug)),
		Simulated:     viper.GetBool(config.NormalizeFlagName(Simulated)),
		SimulatedPort: config.DefaultInt(viper.GetInt(config.NormalizeFlagName(SimulatedPort)), SimulatedDefaultPort),
	}
}

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

	Debug = "debug"
)

type Chain struct {
	Name    string         `json:"name" yaml:"name"`
	Network string         `json:"network" yaml:"network"`
	ChainID config.ChainId `json:"chainId" yaml:"chainId"`
	RpcURL  string         `json:"rpcUrl" yaml:"rpcUrl"`
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
	if !slices.Contains(config.SupportedChainIds, c.ChainID) {
		allErrors = append(allErrors, field.Invalid(field.NewPath("chainId"), c.ChainID, "unsupported chainId"))
	}
	if c.RpcURL == "" {
		allErrors = append(allErrors, field.Required(field.NewPath("rpcUrl"), "rpcUrl is required"))
	}
	return allErrors
}

type AggregatorAvs struct {
	Address               string `json:"address" yaml:"address"`
	PrivateKey            string `json:"privateKey" yaml:"privateKey"`
	PrivateSigningKey     string `json:"privateSigningKey" yaml:"privateSigningKey"`
	PrivateSigningKeyType string `json:"privateSigningKeyType" yaml:"privateSigningKeyType"`
	ResponseTimeout       int    `json:"responseTimeout" yaml:"responseTimeout"`
	ChainIds              []uint `json:"chainIds" yaml:"chainIds"`
}

type ExecutorPeerConfig struct {
	NetworkAddress string `json:"network_address" yaml:"network_address"`
	Port           int    `json:"port" yaml:"port"`
	PublicKey      string `json:"public_key" yaml:"public_key"`
}

type SimulationConfig struct {
	Enabled             bool                 `json:"enabled" yaml:"enabled"`
	Port                int                  `json:"port" yaml:"port"`
	SecureConnection    bool                 `json:"secure_connection" yaml:"secure_connection"`
	ExecutorPeerConfigs []ExecutorPeerConfig `json:"executor_peer_configs" yaml:"executor_peer_configs"`
}

type ServerConfig struct {
	Port             int  `json:"port" yaml:"port"`
	SecureConnection bool `json:"secure_connection" yaml:"secure_connection"`
}

type AggregatorConfig struct {
	Debug            bool             `json:"debug" yaml:"debug"`
	SimulationConfig SimulationConfig `json:"simulation_config" yaml:"simulation_config"`
	ServerConfig     ServerConfig     `json:"server_config" yaml:"server_config"`
	Chains           []Chain          `json:"chains" yaml:"chains"`
	Avss             []AggregatorAvs  `json:"avss" yaml:"avss"`
}

func (arc *AggregatorConfig) Validate() error {
	var allErrors field.ErrorList
	if len(arc.Chains) == 0 {
		allErrors = append(allErrors, field.Required(field.NewPath("chains"), "at least one chain is required"))
	} else {
		for _, chain := range arc.Chains {
			if chainErrors := chain.Validate(); len(chainErrors) > 0 {
				allErrors = append(allErrors, field.Invalid(field.NewPath("chains"), chain, "invalid chain config"))
			}
		}
	}
	if len(arc.Avss) == 0 {
		allErrors = append(allErrors, field.Required(field.NewPath("avss"), "at least one avs is required"))
	}
	return allErrors.ToAggregate()
}

func NewAggregatorConfigFromJsonBytes(data []byte) (*AggregatorConfig, error) {
	var c AggregatorConfig
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal AggregatorConfig from JSON")
	}
	return &c, nil
}

func NewAggregatorConfigFromYamlBytes(data []byte) (*AggregatorConfig, error) {
	var c AggregatorConfig
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal AggregatorConfig from YAML")
	}
	return &c, nil
}

func NewAggregatorConfig() *AggregatorConfig {
	return &AggregatorConfig{
		Debug: viper.GetBool(config.NormalizeFlagName(Debug)),
		SimulationConfig: SimulationConfig{
			Enabled:          viper.GetBool("enabled"),
			Port:             viper.GetInt("port"),
			SecureConnection: viper.GetBool("secure_connection"),
		},
	}
}

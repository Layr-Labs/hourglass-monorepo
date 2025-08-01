package aggregatorConfig

import (
	"encoding/json"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/yaml"
	"slices"
	"strings"
)

const (
	Debug = "debug"
)

type Chain struct {
	Name                string         `json:"name" yaml:"name"`
	Version             string         `json:"version" yaml:"version"`
	ChainId             config.ChainId `json:"chainId" yaml:"chainId"`
	RpcURL              string         `json:"rpcUrl" yaml:"rpcUrl"`
	PollIntervalSeconds int            `json:"pollIntervalSeconds" yaml:"pollIntervalSeconds"`
}

func (c *Chain) Validate() field.ErrorList {
	var allErrors field.ErrorList
	if c.Name == "" {
		allErrors = append(allErrors, field.Required(field.NewPath("name"), "name is required"))
	}
	if c.ChainId == 0 {
		allErrors = append(allErrors, field.Required(field.NewPath("chainId"), "chainId is required"))
	}
	if !slices.Contains(config.SupportedChainIds, c.ChainId) {
		allErrors = append(allErrors, field.Invalid(field.NewPath("chainId"), c.ChainId, "unsupported chainId"))
	}
	if c.RpcURL == "" {
		allErrors = append(allErrors, field.Required(field.NewPath("rpcUrl"), "rpcUrl is required"))
	}
	return allErrors
}

func (c *Chain) IsAnvilRpc() bool {
	return strings.Contains(c.RpcURL, "127.0.0.1:8545")
}

type AggregatorAvs struct {
	Address  string `json:"address" yaml:"address"`
	ChainIds []uint `json:"chainIds" yaml:"chainIds"`
}

func (aa *AggregatorAvs) Validate() error {
	var allErrors field.ErrorList
	if aa.Address == "" {
		allErrors = append(allErrors, field.Required(field.NewPath("address"), "address is required"))
	}
	if len(allErrors) > 0 {
		return allErrors.ToAggregate()
	}
	return nil
}

type ServerConfig struct {
	Port          int    `json:"port" yaml:"port"`
	AggregatorUrl string `json:"aggregatorUrl" yaml:"aggregatorUrl"`
}

type AggregatorConfig struct {
	Debug bool `json:"debug" yaml:"debug"`

	// Operator represents who is actually running the aggregator for the AVS
	Operator *config.OperatorConfig `json:"operator" yaml:"operator"`

	L1ChainId config.ChainId `json:"l1ChainId" yaml:"l1ChainId"`

	// Chains contains the list of chains that the aggregator supports
	Chains []*Chain `json:"chains" yaml:"chains"`

	// Avss contains the list of AVSs that the aggregator is collecting and distributing tasks for
	Avss []*AggregatorAvs `json:"avss" yaml:"avss"`

	// Contracts is an optional field to override the addresses and ABIs for the core contracts that are loaded
	Contracts json.RawMessage `json:"contracts" yaml:"contracts"`

	OverrideContracts *config.OverrideContracts `json:"overrideContracts" yaml:"overrideContracts"`
}

func (arc *AggregatorConfig) Validate() error {
	var allErrors field.ErrorList
	if arc.Operator == nil {
		allErrors = append(allErrors, field.Required(field.NewPath("operator"), "operator is required"))
	} else {
		if err := arc.Operator.Validate(); err != nil {
			allErrors = append(allErrors, field.Invalid(field.NewPath("operator"), arc.Operator, err.Error()))
		}
	}

	if len(arc.Chains) == 0 {
		allErrors = append(allErrors, field.Required(field.NewPath("chains"), "at least one chain is required"))
	} else {
		for _, chain := range arc.Chains {
			if chainErrors := chain.Validate(); len(chainErrors) > 0 {
				allErrors = append(allErrors, field.Invalid(field.NewPath("chains"), chain, "invalid chain config"))
			}
		}
	}

	if arc.L1ChainId == 0 {
		allErrors = append(allErrors, field.Required(field.NewPath("l1ChainId"), "l1ChainId is required"))
	} else {
		found := util.Find(arc.Chains, func(c *Chain) bool {
			return c.ChainId == arc.L1ChainId
		})
		if found == nil {
			allErrors = append(allErrors, field.Invalid(field.NewPath("l1ChainId"), arc.L1ChainId, "l1ChainId must be one of the configured chains"))
		}
	}

	if len(arc.Avss) == 0 {
		allErrors = append(allErrors, field.Required(field.NewPath("avss"), "at least one avs is required"))
	} else {
		for _, avs := range arc.Avss {
			if err := avs.Validate(); err != nil {
				allErrors = append(allErrors, field.Invalid(field.NewPath("avss"), avs, "invalid avs config"))
			}
		}
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
	}
}

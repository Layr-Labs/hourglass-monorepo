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

type ChainSimulation struct {
	Enabled         bool `json:"enabled" yaml:"enabled"`
	Port            int  `json:"port" yaml:"port"`
	AutomaticPoller bool `json:"automaticPoller" yaml:"automaticPoller"`
}

type Chain struct {
	Name                string           `json:"name" yaml:"name"`
	Version             string           `json:"version" yaml:"version"`
	ChainId             config.ChainId   `json:"chainId" yaml:"chainId"`
	RpcURL              string           `json:"rpcUrl" yaml:"rpcUrl"`
	PollIntervalSeconds int              `json:"pollIntervalSeconds" yaml:"pollIntervalSeconds"`
	Simulation          *ChainSimulation `json:"simulation" yaml:"simulation"`
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
	Address         string `json:"address" yaml:"address"`
	PrivateKey      string `json:"privateKey" yaml:"privateKey"`
	ResponseTimeout int    `json:"responseTimeout" yaml:"responseTimeout"`
	ChainIds        []uint `json:"chainIds" yaml:"chainIds"`
	SigningCurve    string `json:"signingCurve" yaml:"signingCurve"`
}

func (aa *AggregatorAvs) Validate() error {
	var allErrors field.ErrorList
	if aa.Address == "" {
		allErrors = append(allErrors, field.Required(field.NewPath("address"), "address is required"))
	}
	if aa.SigningCurve == "" {
		allErrors = append(allErrors, field.Required(field.NewPath("signingCurve"), "signingCurve is required"))
	} else if !slices.Contains([]string{"bn254", "bls381"}, aa.SigningCurve) {
		allErrors = append(allErrors, field.Invalid(field.NewPath("signingCurve"), aa.SigningCurve, "signingCurve must be one of [bn254, bls381]"))
	}
	if len(allErrors) > 0 {
		return allErrors.ToAggregate()
	}
	return nil
}

type ExecutorPeerConfig struct {
	Port      int    `json:"port" yaml:"port"`
	PublicKey string `json:"publicKey" yaml:"publicKey"`
}

type SimulationConfig struct {
	// Enabled indicates whether the simulation mode is enabled
	Enabled bool `json:"enabled" yaml:"enabled"`

	// SimulateExecutors generates a number of fake executors to simulate the behavior of real executors
	SimulateExecutors bool `json:"simulateExecutors" yaml:"simulateExecutors"`

	// SimulatePeering is used by the LocalPeeringDataFetcher to simulate fetching peering data on-chain
	SimulatePeering *config.SimulatedPeeringConfig `json:"simulatePeering" yaml:"simulatePeering"`

	// WriteDelaySeconds is used to slow the aggregator's submission to meet TaskMailbox validation requirements.
	WriteDelaySeconds int64 `json:"writeDelaySeconds" yaml:"writeDelaySeconds"`
}

type ServerConfig struct {
	Port             int    `json:"port" yaml:"port"`
	SecureConnection bool   `json:"secureConnection" yaml:"secureConnection"`
	AggregatorUrl    string `json:"aggregatorUrl" yaml:"aggregatorUrl"`
}

type AggregatorConfig struct {
	Debug bool `json:"debug" yaml:"debug"`

	// Operator represents who is actually running the aggregator for the AVS
	Operator *config.OperatorConfig `json:"operator" yaml:"operator"`

	// ServerConfig contains the configuration for the ExecutionManager grpc server
	ServerConfig ServerConfig `json:"serverConfig" yaml:"serverConfig"`

	L1ChainId config.ChainId `json:"l1ChainId" yaml:"l1ChainId"`

	// Chains contains the list of chains that the aggregator supports
	Chains []*Chain `json:"chains" yaml:"chains"`

	// Avss contains the list of AVSs that the aggregator is collecting and distributing tasks for
	Avss []*AggregatorAvs `json:"avss" yaml:"avss"`

	// SimulationConfig contains the configuration for the simulation mode
	SimulationConfig SimulationConfig `json:"simulationConfig" yaml:"simulationConfig"`

	// Contracts is an optional field to override the addresses and ABIs for the core contracts that are loaded
	Contracts json.RawMessage `json:"contracts" yaml:"contracts"`
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
		SimulationConfig: SimulationConfig{
			Enabled: viper.GetBool("enabled"),
		},
	}
}

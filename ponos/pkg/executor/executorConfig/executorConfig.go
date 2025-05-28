package executorConfig

import (
	"encoding/json"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/avsPerformer"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/yaml"
)

const (
	EnvPrefix = "EXECUTOR_"

	Debug                = "debug"
	GrpcPort             = "grpc-port"
	PerformerNetworkName = "performer-network-name"
)

type SimulationConfig struct {
	SimulatePeering *config.SimulatedPeeringConfig `json:"simulatePeering" yaml:"simulatePeering"`
}

type ExecutorConfig struct {
	Debug                bool
	GrpcPort             int                                `json:"grpcPort" yaml:"grpcPort"`
	PerformerNetworkName string                             `json:"performerNetworkName" yaml:"performerNetworkName"`
	Operator             *config.OperatorConfig             `json:"operator" yaml:"operator"`
	AvsPerformers        []*avsPerformer.AvsPerformerConfig `json:"avsPerformers" yaml:"avsPerformers"`
	Simulation           *SimulationConfig                  `json:"simulation" yaml:"simulation"`

	// Chains configuration
	L1Chain *config.Chain `json:"l1Chain" yaml:"l1Chain"`

	Contracts         json.RawMessage           `json:"contracts" yaml:"contracts"`
	OverrideContracts *config.OverrideContracts `json:"overrideContracts" yaml:"overrideContracts"`
}

func (ec *ExecutorConfig) Validate() error {
	var allErrors field.ErrorList
	if ec.Operator == nil {
		allErrors = append(allErrors, field.Required(field.NewPath("operator"), "operator is required"))
	} else {
		if err := ec.Operator.Validate(); err != nil {
			allErrors = append(allErrors, field.Invalid(field.NewPath("operator"), ec.Operator, err.Error()))
		}
	}

	if len(ec.AvsPerformers) == 0 {
		allErrors = append(allErrors, field.Required(field.NewPath("avss"), "at least one AVS performer is required"))
	} else {
		for _, avs := range ec.AvsPerformers {
			if err := avs.Validate(); err != nil {
				allErrors = append(allErrors, field.Invalid(field.NewPath("avsPerformers"), avs, err.Error()))
			}
		}
	}

	if ec.L1Chain == nil {
		allErrors = append(allErrors, field.Required(field.NewPath("chain"), "a chain is required for the executor"))
	}

	if len(allErrors) > 0 {
		return allErrors.ToAggregate()
	}
	return nil
}

func NewExecutorConfig() *ExecutorConfig {
	return &ExecutorConfig{
		Debug:    viper.GetBool(config.NormalizeFlagName(Debug)),
		GrpcPort: viper.GetInt(config.NormalizeFlagName(GrpcPort)),
		// PerformerNetworkName: viper.GetString(config.NormalizeFlagName(PerformerNetworkName)),
	}
}
func NewExecutorConfigFromYamlBytes(data []byte) (*ExecutorConfig, error) {
	var ec *ExecutorConfig
	if err := yaml.Unmarshal(data, &ec); err != nil {
		return nil, err
	}
	return ec, nil
}

func NewExecutorConfigFromJsonBytes(data []byte) (*ExecutorConfig, error) {
	var ec *ExecutorConfig
	if err := json.Unmarshal(data, &ec); err != nil {
		return nil, err
	}
	return ec, nil
}

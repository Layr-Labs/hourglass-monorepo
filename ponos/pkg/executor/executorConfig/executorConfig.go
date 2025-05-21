package executorConfig

import (
	"encoding/json"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/yaml"
	"slices"
)

const (
	EnvPrefix = "EXECUTOR_"

	Debug                = "debug"
	GrpcPort             = "grpc-port"
	PerformerNetworkName = "performer-network-name"
)

type PerformerImage struct {
	Repository string
	Tag        string
}

type AvsPerformerConfig struct {
	Image        *PerformerImage
	ProcessType  string
	AvsAddress   string
	WorkerCount  int
	SigningCurve string // bn254, bls381, etc
}

func (ap *AvsPerformerConfig) Validate() error {
	var allErrors field.ErrorList
	if ap.AvsAddress == "" {
		allErrors = append(allErrors, field.Required(field.NewPath("avsAddress"), "avsAddress is required"))
	}
	if ap.Image == nil {
		allErrors = append(allErrors, field.Required(field.NewPath("image"), "image is required"))
	} else {
		if ap.Image.Repository == "" {
			allErrors = append(allErrors, field.Required(field.NewPath("image.repository"), "image.repository is required"))
		}
		if ap.Image.Tag == "" {
			allErrors = append(allErrors, field.Required(field.NewPath("image.tag"), "image.tag is required"))
		}
	}
	if ap.SigningCurve == "" {
		allErrors = append(allErrors, field.Required(field.NewPath("signingCurve"), "signingCurve is required"))
	} else if !slices.Contains([]string{"bn254", "bls381"}, ap.SigningCurve) {
		allErrors = append(allErrors, field.Invalid(field.NewPath("signingCurve"), ap.SigningCurve, "signingCurve must be one of [bn254, bls381]"))
	}

	if ap.WorkerCount == 0 {
		allErrors = append(allErrors, field.Required(field.NewPath("workerCount"), "workerCount is required"))
	}
	if len(allErrors) > 0 {
		return allErrors.ToAggregate()
	}
	return nil
}

type SimulationConfig struct {
	SimulatePeering *config.SimulatedPeeringConfig `json:"simulatePeering" yaml:"simulatePeering"`
}

type ExecutorConfig struct {
	Debug                bool
	GrpcPort             int                    `json:"grpcPort" yaml:"grpcPort"`
	PerformerNetworkName string                 `json:"performerNetworkName" yaml:"performerNetworkName"`
	Operator             *config.OperatorConfig `json:"operator" yaml:"operator"`
	AvsPerformers        []*AvsPerformerConfig  `json:"avsPerformers" yaml:"avsPerformers"`
	Simulation           *SimulationConfig      `json:"simulation" yaml:"simulation"`
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

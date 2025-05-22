package executorConfig

import (
	"encoding/json"
	"slices"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
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

// ImageConfig contains configuration for container images
type ImageConfig struct {
	// Repository for the container image
	Repository string `json:"repository"`

	// Tag for the container image
	Tag string `json:"tag"`
}

// AvsPerformerConfig contains configuration for an AVS performer
type AvsPerformerConfig struct {
	// Address of the AVS performer
	AvsAddress string `json:"avsAddress"`

	// Number of worker instances to run
	WorkerCount int `json:"workerCount"`

	// Container image configuration
	Image *ImageConfig `json:"image"`

	// Additional environment variables for the performer
	Env map[string]string `json:"env,omitempty"`

	// Signing curve for the performer
	SigningCurve string `json:"signingCurve,omitempty"`

	// Process type for the performer
	ProcessType string `json:"processType,omitempty"`
}

func (ap *AvsPerformerConfig) Validate() error {
	var allErrors field.ErrorList
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

	// Chains configuration
	Chain *config.Chain `json:"chain" yaml:"chain"`

	// Contract addresses for operator set and executor registration tracking
	AvsArtifactRegistry string `json:"avsArtifactRegistry" yaml:"avsArtifactRegistry"`

	// Contracts JSON data for runtime configuration
	// TODO: double check why we need this configured here. Should be able to load from contracts.json
	Contracts []byte `json:"contracts" yaml:"contracts"`
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

	if ec.Chain == nil {
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

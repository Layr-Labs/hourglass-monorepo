package executorConfig

import (
	"encoding/json"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/yaml"
)

const (
	EnvPrefix = "EXECUTOR_"

	Debug    = "debug"
	GrpcPort = "grpc-port"
)

type PerformerImage struct {
	Repository string
	Tag        string
}

type AvsPerformerConfig struct {
	Image       *PerformerImage
	ProcessType string
	AvsAddress  string
	WorkerCount int
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
	if ap.WorkerCount == 0 {
		allErrors = append(allErrors, field.Required(field.NewPath("workerCount"), "workerCount is required"))
	}
	if len(allErrors) > 0 {
		return allErrors.ToAggregate()
	}
	return nil
}

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

type OperatorConfig struct {
	Address            string      `json:"address" yaml:"address"`
	OperatorPrivateKey string      `json:"operatorPrivateKey" yaml:"operatorPrivateKey"`
	SigningKeys        SigningKeys `json:"signingKeys" yaml:"signingKeys"`
}

func (oc *OperatorConfig) Validate() error {
	var allErrors field.ErrorList
	if oc.Address == "" {
		allErrors = append(allErrors, field.Required(field.NewPath("address"), "address is required"))
	}
	if oc.OperatorPrivateKey == "" {
		allErrors = append(allErrors, field.Required(field.NewPath("operatorPrivateKey"), "operatorPrivateKey is required"))
	}
	if err := oc.SigningKeys.Validate(); err != nil {
		allErrors = append(allErrors, field.Invalid(field.NewPath("signingKeys"), oc.SigningKeys, err.Error()))
	}
	if len(allErrors) > 0 {
		return allErrors.ToAggregate()
	}
	return nil
}

type ExecutorConfig struct {
	Debug         bool
	GrpcPort      int                   `json:"grpcPort" yaml:"grpcPort"`
	Operator      *OperatorConfig       `json:"operator" yaml:"operator"`
	AvsPerformers []*AvsPerformerConfig `json:"avsPerformers" yaml:"avsPerformers"`
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

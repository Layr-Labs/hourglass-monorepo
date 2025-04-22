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

	Debug = "debug"
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

type SingingKeys struct {
	ECDSA string `json:"ecdsa"`
	BLS   string `json:"bls"`
}

type OperatorConfig struct {
	Address            string      `json:"address" yaml:"address"`
	OperatorPrivateKey string      `json:"operatorPrivateKey" yaml:"operatorPrivateKey"`
	SigningKeys        SingingKeys `json:"signingKeys" yaml:"signingKeys"`
}

type ExecutorConfig struct {
	Debug         bool
	Operator      *OperatorConfig       `json:"operator" yaml:"operator"`
	AvsPerformers []*AvsPerformerConfig `json:"avsPerformers" yaml:"avsPerformers"`
}

func (ec *ExecutorConfig) Validate() error {
	var allErrors field.ErrorList
	if ec.Operator == nil {
		allErrors = append(allErrors, field.Required(field.NewPath("operator"), "operator is required"))
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
		Debug: viper.GetBool(config.NormalizeFlagName(Debug)),
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

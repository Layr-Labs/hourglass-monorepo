package executorConfig

import (
	"encoding/json"
	"time"

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

// DeploymentMode represents the deployment mode for performers
type DeploymentMode string

const (
	DeploymentModeDocker     DeploymentMode = "docker"
	DeploymentModeKubernetes DeploymentMode = "kubernetes"
)

// KubernetesConfig contains configuration for Kubernetes deployment mode
type KubernetesConfig struct {
	// Namespace is the Kubernetes namespace where performers will be deployed
	Namespace string `json:"namespace" yaml:"namespace"`

	// OperatorNamespace is the namespace where the Hourglass operator is deployed
	OperatorNamespace string `json:"operatorNamespace" yaml:"operatorNamespace"`

	// CRDGroup is the API group for Performer CRDs
	CRDGroup string `json:"crdGroup" yaml:"crdGroup"`

	// CRDVersion is the API version for Performer CRDs
	CRDVersion string `json:"crdVersion" yaml:"crdVersion"`

	// ConnectionTimeout is the timeout for Kubernetes API connections
	ConnectionTimeout time.Duration `json:"connectionTimeout" yaml:"connectionTimeout"`

	// KubeConfigPath is the path to the kubeconfig file (optional, for out-of-cluster)
	KubeConfigPath string `json:"kubeConfigPath,omitempty" yaml:"kubeConfigPath,omitempty"`

	// InCluster indicates whether the executor is running inside a Kubernetes cluster
	InCluster bool `json:"inCluster" yaml:"inCluster"`
}

// Validate validates the KubernetesConfig
func (kc *KubernetesConfig) Validate() error {
	var allErrors field.ErrorList

	if kc.Namespace == "" {
		allErrors = append(allErrors, field.Required(field.NewPath("namespace"), "namespace is required"))
	}

	if kc.OperatorNamespace == "" {
		allErrors = append(allErrors, field.Required(field.NewPath("operatorNamespace"), "operatorNamespace is required"))
	}

	if kc.CRDGroup == "" {
		allErrors = append(allErrors, field.Required(field.NewPath("crdGroup"), "crdGroup is required"))
	}

	if kc.CRDVersion == "" {
		allErrors = append(allErrors, field.Required(field.NewPath("crdVersion"), "crdVersion is required"))
	}

	if kc.ConnectionTimeout == 0 {
		allErrors = append(allErrors, field.Required(field.NewPath("connectionTimeout"), "connectionTimeout is required"))
	}

	// If not in cluster, kubeconfig path should be specified
	if !kc.InCluster && kc.KubeConfigPath == "" {
		allErrors = append(allErrors, field.Required(field.NewPath("kubeConfigPath"), "kubeConfigPath is required when not running in cluster"))
	}

	if len(allErrors) > 0 {
		return allErrors.ToAggregate()
	}
	return nil
}

// NewDefaultKubernetesConfig creates a KubernetesConfig with sensible defaults
func NewDefaultKubernetesConfig() *KubernetesConfig {
	return &KubernetesConfig{
		Namespace:         "default",
		OperatorNamespace: "hourglass-system",
		CRDGroup:          "hourglass.eigenlayer.io",
		CRDVersion:        "v1alpha1",
		ConnectionTimeout: 30 * time.Second,
		InCluster:         true,
	}
}

type PerformerImage struct {
	Repository string
	Tag        string
}

type AvsPerformerConfig struct {
	Image               *PerformerImage
	ProcessType         string
	AvsAddress          string
	WorkerCount         int
	SigningCurve        string // bn254, bls381, etc
	AVSRegistrarAddress string
	Envs                []config.AVSPerformerEnv
	DeploymentMode      DeploymentMode `json:"deploymentMode" yaml:"deploymentMode"`
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

	if ap.AVSRegistrarAddress == "" {
		allErrors = append(allErrors, field.Required(field.NewPath("avsRegistrarAddress"), "avsRegistrarAddress is required"))
	}
	for i, env := range ap.Envs {
		if err := env.Validate(); err != nil {
			allErrors = append(allErrors, field.Invalid(field.NewPath("envs").Index(i), env, err.Error()))
		}
	}

	// Validate deployment mode - default to docker if not specified
	if ap.DeploymentMode == "" {
		ap.DeploymentMode = DeploymentModeDocker
	} else if !slices.Contains([]DeploymentMode{DeploymentModeDocker, DeploymentModeKubernetes}, ap.DeploymentMode) {
		allErrors = append(allErrors, field.Invalid(field.NewPath("deploymentMode"), ap.DeploymentMode, "deploymentMode must be one of [docker, kubernetes]"))
	}

	if len(allErrors) > 0 {
		return allErrors.ToAggregate()
	}
	return nil
}

type SimulationConfig struct {
	SimulatePeering *config.SimulatedPeeringConfig `json:"simulatePeering" yaml:"simulatePeering"`
}

type Chain struct {
	RpcUrl  string         `json:"rpcUrl" yaml:"rpcUrl"`
	ChainId config.ChainId `json:"chainId" yaml:"chainId"`
}

type ExecutorConfig struct {
	Debug                bool
	GrpcPort             int                       `json:"grpcPort" yaml:"grpcPort"`
	PerformerNetworkName string                    `json:"performerNetworkName" yaml:"performerNetworkName"`
	Operator             *config.OperatorConfig    `json:"operator" yaml:"operator"`
	AvsPerformers        []*AvsPerformerConfig     `json:"avsPerformers" yaml:"avsPerformers"`
	Simulation           *SimulationConfig         `json:"simulation" yaml:"simulation"`
	L1Chain              *Chain                    `json:"l1Chain" yaml:"l1Chain"`
	Contracts            json.RawMessage           `json:"contracts" yaml:"contracts"`
	OverrideContracts    *config.OverrideContracts `json:"overrideContracts" yaml:"overrideContracts"`
	Kubernetes           *KubernetesConfig         `json:"kubernetes,omitempty" yaml:"kubernetes,omitempty"`
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
		// Validate single runtime configuration approach
		dockerCount := 0
		kubernetesCount := 0
		for _, avs := range ec.AvsPerformers {
			if err := avs.Validate(); err != nil {
				allErrors = append(allErrors, field.Invalid(field.NewPath("avsPerformers"), avs, err.Error()))
			}
			if avs.DeploymentMode == DeploymentModeDocker {
				dockerCount++
			} else if avs.DeploymentMode == DeploymentModeKubernetes {
				kubernetesCount++
			}
		}

		// Enforce single runtime configuration: all performers must use the same deployment mode
		if dockerCount > 0 && kubernetesCount > 0 {
			allErrors = append(allErrors, field.Invalid(field.NewPath("avsPerformers"), ec.AvsPerformers, "mixed deployment modes not supported: all performers must use the same deployment mode (either 'docker' or 'kubernetes')"))
		}

		// If any performer uses Kubernetes mode, validate Kubernetes config
		if kubernetesCount > 0 {
			if ec.Kubernetes == nil {
				allErrors = append(allErrors, field.Required(field.NewPath("kubernetes"), "kubernetes configuration is required when using kubernetes deployment mode"))
			} else {
				if err := ec.Kubernetes.Validate(); err != nil {
					allErrors = append(allErrors, field.Invalid(field.NewPath("kubernetes"), ec.Kubernetes, err.Error()))
				}
			}
		}
	}

	if ec.L1Chain == nil {
		allErrors = append(allErrors, field.Required(field.NewPath("l1Chain"), "l1Chain is required"))
	} else {
		if ec.L1Chain.RpcUrl == "" {
			allErrors = append(allErrors, field.Required(field.NewPath("chain.rpcUrl"), "rpcUrl is required"))
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

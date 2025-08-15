package executorConfig

import (
	"encoding/json"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/auth"
	"slices"
	"time"

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

// DeploymentMode represents the deployment mode for performers
type DeploymentMode string

func (d DeploymentMode) String() string {
	return string(d)
}

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

func (k *KubernetesConfig) UnmarshalJSON(data []byte) error {
	// Create an alias to avoid infinite recursion
	type Alias KubernetesConfig

	// Create a temporary struct with string for the duration field
	aux := &struct {
		ConnectionTimeout interface{} `json:"connectionTimeout"`
		*Alias
	}{
		Alias: (*Alias)(k),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// Handle the ConnectionTimeout field specially
	if aux.ConnectionTimeout != nil {
		switch v := aux.ConnectionTimeout.(type) {
		case string:
			duration, err := time.ParseDuration(v)
			if err != nil {
				return fmt.Errorf("invalid duration format for connectionTimeout: %w", err)
			}
			k.ConnectionTimeout = duration
		case float64:
			k.ConnectionTimeout = time.Duration(v)
		default:
			return fmt.Errorf("connectionTimeout must be a string (e.g., '30s') or number")
		}
	}

	return nil
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

// AvsPerformerKubernetesConfig contains Kubernetes-specific configuration for an AVS performer
type AvsPerformerKubernetesConfig struct {
	// ServiceAccountName is the name of the ServiceAccount to use for the performer pod
	ServiceAccountName string `json:"serviceAccountName,omitempty" yaml:"serviceAccountName,omitempty"`
	EndpointOverride   string `json:"endpointOverride,omitempty" yaml:"endpointOverride,omitempty"` // Optional: Override auto-detected endpoint (for testing)
}

func (apc *AvsPerformerKubernetesConfig) Validate() error {
	return nil
}

type AvsPerformerConfig struct {
	Image       *PerformerImage
	ProcessType string
	AvsAddress  string
	Envs        []config.AVSPerformerEnv
	Kubernetes  *AvsPerformerKubernetesConfig `json:"kubernetes,omitempty" yaml:"kubernetes,omitempty"`
}

func (ap *AvsPerformerConfig) Validate(deploymentMode DeploymentMode) error {
	var allErrors field.ErrorList
	if ap.AvsAddress == "" {
		allErrors = append(allErrors, field.Required(field.NewPath("avsAddress"), "avsAddress is required"))
	}
	if ap.Image != nil {
		if ap.Image.Repository == "" {
			allErrors = append(allErrors, field.Required(field.NewPath("image.repository"), "image.repository is required"))
		}
		if ap.Image.Tag == "" {
			allErrors = append(allErrors, field.Required(field.NewPath("image.tag"), "image.tag is required"))
		}
	}

	for i, env := range ap.Envs {
		if err := env.Validate(); err != nil {
			allErrors = append(allErrors, field.Invalid(field.NewPath("envs").Index(i), env, err.Error()))
		}
	}

	// Validate Kubernetes config if in Kubernetes mode
	if deploymentMode == DeploymentModeKubernetes && ap.Kubernetes != nil {
		if err := ap.Kubernetes.Validate(); err != nil {
			allErrors = append(allErrors, field.Invalid(field.NewPath("kubernetes"), ap.Kubernetes, err.Error()))
		}
	}

	if len(allErrors) > 0 {
		return allErrors.ToAggregate()
	}
	return nil
}

type Chain struct {
	RpcUrl  string         `json:"rpcUrl" yaml:"rpcUrl"`
	ChainId config.ChainId `json:"chainId" yaml:"chainId"`
}

// StorageConfig contains configuration for the storage layer
type StorageConfig struct {
	Type         string        `json:"type" yaml:"type"` // "memory" or "badger"
	BadgerConfig *BadgerConfig `json:"badger,omitempty" yaml:"badger,omitempty"`
}

// BadgerConfig contains configuration for BadgerDB storage
type BadgerConfig struct {
	// Directory where BadgerDB will store its data
	Dir string `json:"dir" yaml:"dir"`
	// InMemory runs BadgerDB in memory-only mode (for testing)
	InMemory bool `json:"inMemory,omitempty" yaml:"inMemory,omitempty"`
	// ValueLogFileSize sets the maximum size of a single value log file
	ValueLogFileSize int64 `json:"valueLogFileSize,omitempty" yaml:"valueLogFileSize,omitempty"`
	// NumVersionsToKeep sets how many versions to keep for each key
	NumVersionsToKeep int `json:"numVersionsToKeep,omitempty" yaml:"numVersionsToKeep,omitempty"`
	// NumLevelZeroTables sets the maximum number of level zero tables before stalling
	NumLevelZeroTables int `json:"numLevelZeroTables,omitempty" yaml:"numLevelZeroTables,omitempty"`
	// NumLevelZeroTablesStall sets the number of level zero tables that will trigger a stall
	NumLevelZeroTablesStall int `json:"numLevelZeroTablesStall,omitempty" yaml:"numLevelZeroTablesStall,omitempty"`
}

// Validate validates the StorageConfig
func (sc *StorageConfig) Validate() error {
	var allErrors field.ErrorList

	if sc.Type == "" {
		sc.Type = "memory" // Default to memory if not specified
	}

	if sc.Type != "memory" && sc.Type != "badger" {
		allErrors = append(allErrors, field.Invalid(field.NewPath("type"), sc.Type, "type must be 'memory' or 'badger'"))
	}

	if sc.Type == "badger" {
		if sc.BadgerConfig == nil {
			allErrors = append(allErrors, field.Required(field.NewPath("badger"), "badger configuration is required when type is 'badger'"))
		} else if sc.BadgerConfig.Dir == "" {
			allErrors = append(allErrors, field.Required(field.NewPath("badger.dir"), "badger directory is required"))
		}
	}

	if len(allErrors) > 0 {
		return allErrors.ToAggregate()
	}
	return nil
}

type ExecutorConfig struct {
	Debug                    bool
	GrpcPort                 int                       `json:"grpcPort" yaml:"grpcPort"`
	ManagementServerGrpcPort int                       `json:"managementServerGrpcPort" yaml:"managementServerGrpcPort"`
	PerformerNetworkName     string                    `json:"performerNetworkName" yaml:"performerNetworkName"`
	Operator                 *config.OperatorConfig    `json:"operator" yaml:"operator"`
	AvsPerformers            []*AvsPerformerConfig     `json:"avsPerformers" yaml:"avsPerformers"`
	L1Chain                  *Chain                    `json:"l1Chain" yaml:"l1Chain"`
	Contracts                json.RawMessage           `json:"contracts" yaml:"contracts"`
	OverrideContracts        *config.OverrideContracts `json:"overrideContracts" yaml:"overrideContracts"`
	Kubernetes               *KubernetesConfig         `json:"kubernetes,omitempty" yaml:"kubernetes,omitempty"`
	Storage                  *StorageConfig            `json:"storage,omitempty" yaml:"storage,omitempty"`
	AuthConfig               *auth.Config              `json:"authentication,omitempty" yaml:"authentication,omitempty"`
	DeploymentMode           DeploymentMode            `json:"deploymentMode" yaml:"deploymentMode"` // Deployment mode for the executor (docker or kubernetes)
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

	if ec.DeploymentMode == "" {
		ec.DeploymentMode = DeploymentModeDocker // Default to Docker if not specified
	}
	if !slices.Contains([]DeploymentMode{DeploymentModeDocker, DeploymentModeKubernetes}, ec.DeploymentMode) {
		allErrors = append(allErrors, field.Invalid(field.NewPath("deploymentMode"), ec.DeploymentMode, "deploymentMode must be one of [docker, kubernetes]"))
	}

	for _, avs := range ec.AvsPerformers {
		if err := avs.Validate(ec.DeploymentMode); err != nil {
			allErrors = append(allErrors, field.Invalid(field.NewPath("avsPerformers"), avs, err.Error()))
		}
	}

	// If we're using Kubernetes deployment mode, validate the Kubernetes configuration
	if ec.DeploymentMode == DeploymentModeKubernetes {
		if ec.Kubernetes == nil {
			allErrors = append(allErrors, field.Required(field.NewPath("kubernetes"), "kubernetes configuration is required when using kubernetes deployment mode"))
		} else {
			if err := ec.Kubernetes.Validate(); err != nil {
				allErrors = append(allErrors, field.Invalid(field.NewPath("kubernetes"), ec.Kubernetes, err.Error()))
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

	// Validate storage configuration if present
	if ec.Storage != nil {
		if err := ec.Storage.Validate(); err != nil {
			allErrors = append(allErrors, field.Invalid(field.NewPath("storage"), ec.Storage, err.Error()))
		}
	}

	if ec.ManagementServerGrpcPort == 0 {
		ec.ManagementServerGrpcPort = ec.GrpcPort
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

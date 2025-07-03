package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server       ServerConfig       `yaml:"server"`
	Orchestrator OrchestratorConfig `yaml:"orchestrator"`
	Registry     RegistryConfig     `yaml:"registry"`
	Proxy        ProxyConfig        `yaml:"proxy"`
	Logging      LoggingConfig      `yaml:"logging"`
	Monitoring   MonitoringConfig   `yaml:"monitoring"`
}

type ServerConfig struct {
	Port            int           `yaml:"port"`
	GRPCPort        int           `yaml:"grpcPort"`
	MaxConnections  int           `yaml:"maxConnections"`
	ReadTimeout     time.Duration `yaml:"readTimeout"`
	WriteTimeout    time.Duration `yaml:"writeTimeout"`
	ShutdownTimeout time.Duration `yaml:"shutdownTimeout"`
}

type OrchestratorConfig struct {
	Resources ResourcesConfig `yaml:"resources"`
	Queue     QueueConfig     `yaml:"queue"`
	Container ContainerConfig `yaml:"container"`
}

type ResourcesConfig struct {
	MaxCPU        string `yaml:"maxCpu"`
	MaxMemory     string `yaml:"maxMemory"`
	MaxDisk       string `yaml:"maxDisk"`
	MaxContainers int    `yaml:"maxContainers"`
}

type QueueConfig struct {
	MaxQueueSize      int              `yaml:"maxQueueSize"`
	TaskTimeout       time.Duration    `yaml:"taskTimeout"`
	RetryPolicy       RetryPolicyConfig `yaml:"retryPolicy"`
}

type RetryPolicyConfig struct {
	MaxRetries        int     `yaml:"maxRetries"`
	BackoffMultiplier float64 `yaml:"backoffMultiplier"`
}

type ContainerConfig struct {
	Runtime    string            `yaml:"runtime"`
	Network    string            `yaml:"network"`
	PullPolicy string            `yaml:"pullPolicy"`
	Env        []EnvVar          `yaml:"env"`
	Labels     map[string]string `yaml:"labels"`
}

type EnvVar struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type RegistryConfig struct {
	Registries []RegistryEntry  `yaml:"registries"`
	Cache      CacheConfig      `yaml:"cache"`
	Security   SecurityConfig   `yaml:"security"`
}

type RegistryEntry struct {
	Name             string `yaml:"name"`
	Type             string `yaml:"type"`
	URL              string `yaml:"url"`
	Region           string `yaml:"region"`
	CredentialSource string `yaml:"credentialSource"`
	SecretName       string `yaml:"secretName"`
}

type CacheConfig struct {
	MaxSize         string        `yaml:"maxSize"`
	TTL             time.Duration `yaml:"ttl"`
	CleanupInterval time.Duration `yaml:"cleanupInterval"`
}

type SecurityConfig struct {
	VulnerabilityScanEnabled   bool `yaml:"vulnerabilityScanEnabled"`
	MaxCriticalVulnerabilities int  `yaml:"maxCriticalVulnerabilities"`
	MaxHighVulnerabilities     int  `yaml:"maxHighVulnerabilities"`
	SignatureVerification      bool `yaml:"signatureVerification"`
}

type ProxyConfig struct {
	Server   ProxyServerConfig `yaml:"server"`
	Backends []BackendConfig   `yaml:"backends"`
	Logging  ProxyLoggingConfig `yaml:"logging"`
}

type ProxyServerConfig struct {
	Port           int           `yaml:"port"`
	MaxConnections int           `yaml:"maxConnections"`
	ReadTimeout    time.Duration `yaml:"readTimeout"`
	WriteTimeout   time.Duration `yaml:"writeTimeout"`
}

type BackendConfig struct {
	Name           string            `yaml:"name"`
	URL            string            `yaml:"url"`
	RateLimits     RateLimitsConfig  `yaml:"rateLimits"`
	AllowedMethods []string          `yaml:"allowedMethods"`
	Filters        []FilterConfig    `yaml:"filters"`
}

type RateLimitsConfig struct {
	RequestsPerSecond int `yaml:"requestsPerSecond"`
	Burst             int `yaml:"burst"`
}

type FilterConfig struct {
	Type      string `yaml:"type"`
	Parameter string `yaml:"parameter"`
	Value     string `yaml:"value"`
}

type ProxyLoggingConfig struct {
	LogRequests   bool    `yaml:"logRequests"`
	LogResponses  bool    `yaml:"logResponses"`
	SamplingRate  float64 `yaml:"samplingRate"`
}

type LoggingConfig struct {
	Level      string `yaml:"level"`
	Format     string `yaml:"format"`
	OutputPath string `yaml:"outputPath"`
}

type MonitoringConfig struct {
	MetricsEnabled bool   `yaml:"metricsEnabled"`
	MetricsPort    int    `yaml:"metricsPort"`
	TracingEnabled bool   `yaml:"tracingEnabled"`
	TracingEndpoint string `yaml:"tracingEndpoint"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var config Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

func (c *Config) Validate() error {
	if c.Server.Port <= 0 {
		return fmt.Errorf("server port must be positive")
	}
	if c.Server.GRPCPort <= 0 {
		return fmt.Errorf("gRPC port must be positive")
	}
	if c.Orchestrator.Resources.MaxContainers <= 0 {
		return fmt.Errorf("max containers must be positive")
	}
	if c.Orchestrator.Queue.MaxQueueSize <= 0 {
		return fmt.Errorf("max queue size must be positive")
	}
	return nil
}
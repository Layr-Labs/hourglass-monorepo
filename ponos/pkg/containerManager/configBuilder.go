package containerManager

import "time"

// ConfigBuilder helps build and validate container manager configurations
type ConfigBuilder struct{}

// NewConfigBuilder creates a new configuration builder
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{}
}

// BuildContainerManagerConfig builds a complete ContainerManagerConfig with defaults
func (cb *ConfigBuilder) BuildContainerManagerConfig(config *ContainerManagerConfig) *ContainerManagerConfig {
	if config == nil {
		config = &ContainerManagerConfig{}
	}

	// Set default timeouts
	if config.DefaultStartTimeout == 0 {
		config.DefaultStartTimeout = DefaultStartTimeout
	}
	if config.DefaultStopTimeout == 0 {
		config.DefaultStopTimeout = DefaultStopTimeout
	}

	// Build default health check config
	if config.DefaultHealthCheckConfig == nil {
		config.DefaultHealthCheckConfig = cb.BuildHealthCheckConfig(nil)
	}

	// Build default liveness config
	if config.DefaultLivenessConfig == nil {
		config.DefaultLivenessConfig = cb.BuildLivenessConfig(nil, config.DefaultHealthCheckConfig)
	}

	return config
}

// BuildHealthCheckConfig builds a complete HealthCheckConfig with defaults
func (cb *ConfigBuilder) BuildHealthCheckConfig(config *HealthCheckConfig) *HealthCheckConfig {
	if config == nil {
		config = &HealthCheckConfig{}
	}

	// Set defaults for zero values
	if config.Interval == 0 {
		config.Interval = DefaultHealthInterval
	}
	if config.Timeout == 0 {
		config.Timeout = DefaultHealthTimeout
	}
	if config.Retries == 0 {
		config.Retries = DefaultHealthRetries
	}
	if config.StartPeriod == 0 {
		config.StartPeriod = DefaultHealthStartPeriod
	}
	if config.FailureThreshold == 0 {
		config.FailureThreshold = DefaultFailureThreshold
	}

	// Enabled defaults to false - must be explicitly set to true to enable health checks
	// No automatic enabling - this makes the behavior explicit and predictable

	return config
}

// BuildLivenessConfig builds a complete LivenessConfig with defaults
func (cb *ConfigBuilder) BuildLivenessConfig(config *LivenessConfig, defaultHealthConfig *HealthCheckConfig) *LivenessConfig {
	if config == nil {
		config = &LivenessConfig{}
	}

	// Build health check config
	if defaultHealthConfig != nil {
		config.HealthCheckConfig = *cb.mergeHealthCheckConfig(&config.HealthCheckConfig, defaultHealthConfig)
	} else {
		config.HealthCheckConfig = *cb.BuildHealthCheckConfig(&config.HealthCheckConfig)
	}

	// Build restart policy
	config.RestartPolicy = *cb.BuildRestartPolicy(&config.RestartPolicy)

	// Build resource thresholds
	config.ResourceThresholds = *cb.BuildResourceThresholds(&config.ResourceThresholds)

	// Set default intervals
	if config.ResourceCheckInterval == 0 {
		config.ResourceCheckInterval = DefaultResourceInterval
	}

	return config
}

// BuildRestartPolicy builds a complete RestartPolicy with defaults
func (cb *ConfigBuilder) BuildRestartPolicy(policy *RestartPolicy) *RestartPolicy {
	if policy == nil {
		policy = &RestartPolicy{}
	}

	// Set defaults for zero values
	if policy.MaxRestarts == 0 {
		policy.MaxRestarts = DefaultMaxRestarts
	}
	if policy.RestartDelay == 0 {
		policy.RestartDelay = DefaultRestartDelay
	}
	if policy.BackoffMultiplier == 0 {
		policy.BackoffMultiplier = DefaultBackoffMultiplier
	}
	if policy.MaxBackoffDelay == 0 {
		policy.MaxBackoffDelay = DefaultMaxBackoffDelay
	}
	if policy.RestartTimeout == 0 {
		policy.RestartTimeout = DefaultRestartTimeout
	}

	// Set default behaviors
	if !policy.RestartOnCrash && !policy.RestartOnOOM && !policy.RestartOnUnhealthy {
		policy.RestartOnCrash = true
		policy.RestartOnOOM = true
		policy.RestartOnUnhealthy = false // Let application decide
	}

	return policy
}

// BuildResourceThresholds builds complete ResourceThresholds with defaults
func (cb *ConfigBuilder) BuildResourceThresholds(thresholds *ResourceThresholds) *ResourceThresholds {
	if thresholds == nil {
		thresholds = &ResourceThresholds{}
	}

	if thresholds.CPUThreshold == 0 {
		thresholds.CPUThreshold = DefaultCPUThreshold
	}
	if thresholds.MemoryThreshold == 0 {
		thresholds.MemoryThreshold = DefaultMemoryThreshold
	}

	return thresholds
}

// mergeHealthCheckConfig merges user config with defaults
func (cb *ConfigBuilder) mergeHealthCheckConfig(userConfig, defaultConfig *HealthCheckConfig) *HealthCheckConfig {
	merged := *defaultConfig

	// Always apply user's Enabled value (whether true or false)
	merged.Enabled = userConfig.Enabled
	if userConfig.Interval != 0 {
		merged.Interval = userConfig.Interval
	}
	if userConfig.Timeout != 0 {
		merged.Timeout = userConfig.Timeout
	}
	if userConfig.Retries != 0 {
		merged.Retries = userConfig.Retries
	}
	if userConfig.StartPeriod != 0 {
		merged.StartPeriod = userConfig.StartPeriod
	}
	if userConfig.FailureThreshold != 0 {
		merged.FailureThreshold = userConfig.FailureThreshold
	}

	return &merged
}

// ValidateConfig validates a container manager configuration
func (cb *ConfigBuilder) ValidateConfig(config *ContainerManagerConfig) error {
	if config.DefaultStartTimeout < 0 {
		return ErrInvalidStartTimeout
	}
	if config.DefaultStopTimeout < 0 {
		return ErrInvalidStopTimeout
	}

	if config.DefaultHealthCheckConfig != nil {
		if err := cb.ValidateHealthCheckConfig(config.DefaultHealthCheckConfig); err != nil {
			return err
		}
	}

	if config.DefaultLivenessConfig != nil {
		if err := cb.ValidateLivenessConfig(config.DefaultLivenessConfig); err != nil {
			return err
		}
	}

	return nil
}

// ValidateHealthCheckConfig validates a health check configuration
func (cb *ConfigBuilder) ValidateHealthCheckConfig(config *HealthCheckConfig) error {
	if config.Interval < time.Second {
		return ErrInvalidHealthInterval
	}
	if config.FailureThreshold < 1 {
		return ErrInvalidFailureThreshold
	}
	return nil
}

// ValidateLivenessConfig validates a liveness configuration
func (cb *ConfigBuilder) ValidateLivenessConfig(config *LivenessConfig) error {
	if err := cb.ValidateHealthCheckConfig(&config.HealthCheckConfig); err != nil {
		return err
	}

	if config.RestartPolicy.MaxRestarts < 0 {
		return ErrInvalidMaxRestarts
	}
	if config.RestartPolicy.RestartDelay < 0 {
		return ErrInvalidRestartDelay
	}
	if config.ResourceCheckInterval < time.Second {
		return ErrInvalidResourceInterval
	}

	return nil
}

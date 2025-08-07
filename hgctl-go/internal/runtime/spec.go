package runtime

// Spec represents an EigenRuntime specification
type Spec struct {
	APIVersion string                   `yaml:"apiVersion"`
	Kind       string                   `yaml:"kind"`
	Name       string                   `yaml:"name"`
	Version    string                   `yaml:"version"`
	Spec       map[string]ComponentSpec `yaml:"spec"`
}

// ComponentSpec represents a component in the runtime spec
type ComponentSpec struct {
	Registry  string                `yaml:"registry"`
	Digest    string                `yaml:"digest"`
	Command   []string              `yaml:"command,omitempty"`
	Env       []EnvVar              `yaml:"env,omitempty"`
	Ports     []int                 `yaml:"ports,omitempty"`
	Resources *ResourceRequirements `yaml:"resources,omitempty"`
}

// EnvVar represents an environment variable
type EnvVar struct {
	Name     string `yaml:"name"`
	Kind     string `yaml:"kind"`
	Value    string `yaml:"value"`
	Required bool   `yaml:"required,omitempty"`
}

// ResourceRequirements represents resource requirements for a component
type ResourceRequirements struct {
	TEEEnabled bool `yaml:"tee_enabled,omitempty"`
	GPUEnabled bool `yaml:"gpu_enabled,omitempty"`
}

// ResourceLimits represents CPU and memory limits
type ResourceLimits struct {
	CPU    string `yaml:"cpu,omitempty"`
	Memory string `yaml:"memory,omitempty"`
}

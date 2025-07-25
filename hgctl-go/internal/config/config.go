package config

import (
    "fmt"
    "net/url"
    "os"
    "path/filepath"
    
    "github.com/ethereum/go-ethereum/common"
    "gopkg.in/yaml.v3"
)

type Context struct {
    Name                  string `yaml:"-"`
    ExecutorAddress       string `yaml:"executorAddress"`
    AVSAddress           string `yaml:"avsAddress,omitempty"`
    OperatorSetID        uint32 `yaml:"operatorSetId,omitempty"`
    NetworkID            uint64 `yaml:"networkId,omitempty"`
    RPCUrl               string `yaml:"rpcUrl,omitempty"`
    ReleaseManagerAddress string `yaml:"releaseManagerAddress,omitempty"`
}

type Config struct {
    CurrentContext string              `yaml:"currentContext"`
    Contexts       map[string]*Context `yaml:"contexts"`
}

// Global config instance
var globalConfig *Config

func LoadConfig() (*Config, error) {
    configPath := getConfigPath()
    
    data, err := os.ReadFile(configPath)
    if err != nil {
        if os.IsNotExist(err) {
            cfg := defaultConfig()
            globalConfig = cfg
            return cfg, nil
        }
        return nil, fmt.Errorf("failed to read config: %w", err)
    }
    
    var config Config
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("failed to parse config: %w", err)
    }
    
    // Set context names
    for name, ctx := range config.Contexts {
        ctx.Name = name
    }
    
    globalConfig = &config
    return &config, nil
}

func SaveConfig(config *Config) error {
    configPath := getConfigPath()
    
    // Ensure config directory exists
    configDir := filepath.Dir(configPath)
    if err := os.MkdirAll(configDir, 0755); err != nil {
        return fmt.Errorf("failed to create config directory: %w", err)
    }
    
    data, err := yaml.Marshal(config)
    if err != nil {
        return fmt.Errorf("failed to marshal config: %w", err)
    }
    
    if err := os.WriteFile(configPath, data, 0644); err != nil {
        return fmt.Errorf("failed to write config: %w", err)
    }
    
    globalConfig = config
    return nil
}

func GetCurrentContext() (*Context, error) {
    cfg, err := LoadConfig()
    if err != nil {
        return nil, err
    }
    
    ctx, ok := cfg.Contexts[cfg.CurrentContext]
    if !ok {
        return nil, fmt.Errorf("current context '%s' not found", cfg.CurrentContext)
    }
    
    return ctx, nil
}

func (c *Context) Validate() error {
    if c.ExecutorAddress == "" {
        return fmt.Errorf("executor address is required")
    }
    
    if c.RPCUrl != "" {
        if _, err := url.Parse(c.RPCUrl); err != nil {
            return fmt.Errorf("invalid RPC URL: %w", err)
        }
    }
    
    if c.ReleaseManagerAddress != "" {
        if !common.IsHexAddress(c.ReleaseManagerAddress) {
            return fmt.Errorf("invalid release manager address")
        }
    }
    
    if c.AVSAddress != "" {
        if !common.IsHexAddress(c.AVSAddress) {
            return fmt.Errorf("invalid AVS address")
        }
    }
    
    return nil
}

func (c *Context) ApplyOverrides(executorAddr, rpcURL, releaseManagerAddr string, operatorSetID *uint32) {
    if executorAddr != "" {
        c.ExecutorAddress = executorAddr
    }
    if rpcURL != "" {
        c.RPCUrl = rpcURL
    }
    if releaseManagerAddr != "" {
        c.ReleaseManagerAddress = releaseManagerAddr
    }
    if operatorSetID != nil {
        c.OperatorSetID = *operatorSetID
    }
}

func getConfigPath() string {
    home, _ := os.UserHomeDir()
    return filepath.Join(home, ".hgctl", "config.yaml")
}

func defaultConfig() *Config {
    return &Config{
        CurrentContext: "default",
        Contexts: map[string]*Context{
            "default": {
                Name:            "default",
                ExecutorAddress: "127.0.0.1:9090",
            },
        },
    }
}

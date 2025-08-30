package aggregatorConfig

import (
	"encoding/json"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/auth"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/yaml"
	"slices"
	"strings"
)

const (
	Debug = "debug"
)

type Chain struct {
	Name                string         `json:"name" yaml:"name"`
	Version             string         `json:"version" yaml:"version"`
	ChainId             config.ChainId `json:"chainId" yaml:"chainId"`
	RpcURL              string         `json:"rpcUrl" yaml:"rpcUrl"`
	PollIntervalSeconds int            `json:"pollIntervalSeconds" yaml:"pollIntervalSeconds"`
}

func (c *Chain) Validate() field.ErrorList {
	var allErrors field.ErrorList
	if c.Name == "" {
		allErrors = append(allErrors, field.Required(field.NewPath("name"), "name is required"))
	}
	if c.ChainId == 0 {
		allErrors = append(allErrors, field.Required(field.NewPath("chainId"), "chainId is required"))
	}
	if !slices.Contains(config.SupportedChainIds, c.ChainId) {
		allErrors = append(allErrors, field.Invalid(field.NewPath("chainId"), c.ChainId, "unsupported chainId"))
	}
	if c.RpcURL == "" {
		allErrors = append(allErrors, field.Required(field.NewPath("rpcUrl"), "rpcUrl is required"))
	}
	return allErrors
}

func (c *Chain) IsAnvilRpc() bool {
	return strings.Contains(c.RpcURL, "127.0.0.1:8545")
}

type AggregatorAvs struct {
	Address  string `json:"address" yaml:"address"`
	ChainIds []uint `json:"chainIds" yaml:"chainIds"`
}

func (aa *AggregatorAvs) Validate() error {
	var allErrors field.ErrorList
	if aa.Address == "" {
		allErrors = append(allErrors, field.Required(field.NewPath("address"), "address is required"))
	}
	if len(allErrors) > 0 {
		return allErrors.ToAggregate()
	}
	return nil
}

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
	NumLevelZeroTablesStall int    `json:"numLevelZeroTablesStall,omitempty" yaml:"numLevelZeroTablesStall,omitempty"`
	ValueLogMaxEntries      uint32 `json:"valueLogMaxEntries,omitempty" yaml:"valueLogMaxEntries,omitempty"`
}

type StorageConfig struct {
	// Type of storage backend: "memory" or "badger"
	Type string `json:"type" yaml:"type"`
	// BadgerConfig contains BadgerDB-specific configuration (only used when Type is "badger")
	BadgerConfig *BadgerConfig `json:"badger,omitempty" yaml:"badger,omitempty"`
}

func (sc *StorageConfig) Validate() field.ErrorList {
	var allErrors field.ErrorList
	if sc.Type == "" {
		allErrors = append(allErrors, field.Required(field.NewPath("type"), "storage type is required"))
	} else if sc.Type != "memory" && sc.Type != "badger" {
		allErrors = append(allErrors, field.Invalid(field.NewPath("type"), sc.Type, "storage type must be 'memory' or 'badger'"))
	}

	if sc.Type == "badger" && sc.BadgerConfig == nil {
		allErrors = append(allErrors, field.Required(field.NewPath("badger"), "badger config is required when type is 'badger'"))
	} else if sc.Type == "badger" && sc.BadgerConfig != nil {
		if sc.BadgerConfig.Dir == "" {
			allErrors = append(allErrors, field.Required(field.NewPath("badger", "dir"), "badger directory is required"))
		}
	}

	return allErrors
}

type AggregatorConfig struct {
	Debug bool `json:"debug" yaml:"debug"`

	TLSEnabled bool `json:"tlsEnabled" yaml:"tlsEnabled"`

	ManagementServerGrpcPort int `json:"managementServerGrpcPort" yaml:"managementServerGrpcPort"`

	// Operator represents who is actually running the aggregator for the AVS
	Operator *config.OperatorConfig `json:"operator" yaml:"operator"`

	L1ChainId config.ChainId `json:"l1ChainId" yaml:"l1ChainId"`

	// Chains contains the list of chains that the aggregator supports
	Chains []*Chain `json:"chains" yaml:"chains"`

	// Avss contains the list of AVSs that the aggregator is collecting and distributing tasks for
	Avss []*AggregatorAvs `json:"avss" yaml:"avss"`

	// Contracts is an optional field to override the addresses and ABIs for the core contracts that are loaded
	Contracts json.RawMessage `json:"contracts" yaml:"contracts"`

	OverrideContracts *config.OverrideContracts `json:"overrideContracts" yaml:"overrideContracts"`

	// Storage configuration for persistence
	Storage *StorageConfig `json:"storage,omitempty" yaml:"storage,omitempty"`

	// Authentication configuration for mgmt apis
	Authentication auth.Config `json:"authentication,omitempty" yaml:"authentication,omitempty"`
}

func (arc *AggregatorConfig) Validate() error {
	var allErrors field.ErrorList

	if arc.ManagementServerGrpcPort == 0 {
		allErrors = append(allErrors, field.Required(field.NewPath("managementServerGrpcPort"), "managementServerGrpcPort is required"))
	}

	if arc.Operator == nil {
		allErrors = append(allErrors, field.Required(field.NewPath("operator"), "operator is required"))
	} else {
		if err := arc.Operator.Validate(); err != nil {
			allErrors = append(allErrors, field.Invalid(field.NewPath("operator"), arc.Operator, err.Error()))
		}
	}

	if len(arc.Chains) == 0 {
		allErrors = append(allErrors, field.Required(field.NewPath("chains"), "at least one chain is required"))
	} else {
		for _, chain := range arc.Chains {
			if chainErrors := chain.Validate(); len(chainErrors) > 0 {
				allErrors = append(allErrors, field.Invalid(field.NewPath("chains"), chain, "invalid chain config"))
			}
		}
	}

	if arc.L1ChainId == 0 {
		allErrors = append(allErrors, field.Required(field.NewPath("l1ChainId"), "l1ChainId is required"))
	} else {
		found := util.Find(arc.Chains, func(c *Chain) bool {
			return c.ChainId == arc.L1ChainId
		})
		if found == nil {
			allErrors = append(allErrors, field.Invalid(field.NewPath("l1ChainId"), arc.L1ChainId, "l1ChainId must be one of the configured chains"))
		}
	}

	for _, avs := range arc.Avss {
		if err := avs.Validate(); err != nil {
			allErrors = append(allErrors, field.Invalid(field.NewPath("avss"), avs, "invalid avs config"))
		}
	}

	// Validate storage config if provided
	if arc.Storage != nil {
		if storageErrors := arc.Storage.Validate(); len(storageErrors) > 0 {
			allErrors = append(allErrors, storageErrors...)
		}
	}

	return allErrors.ToAggregate()
}

func NewAggregatorConfigFromJsonBytes(data []byte) (*AggregatorConfig, error) {
	var c AggregatorConfig
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal AggregatorConfig from JSON")
	}
	return &c, nil
}

func NewAggregatorConfigFromYamlBytes(data []byte) (*AggregatorConfig, error) {
	var c AggregatorConfig
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal AggregatorConfig from YAML")
	}
	return &c, nil
}

func NewAggregatorConfig() *AggregatorConfig {
	return &AggregatorConfig{
		Debug: viper.GetBool(config.NormalizeFlagName(Debug)),
	}
}

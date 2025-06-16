package blsHelperConfig

import (
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/spf13/viper"
)

const (
	Debug           = "debug"
	KeyfilePath     = "keyfile-path"
	KeyPassword     = "key-password"
	RpcUrl          = "rpc-url"
	OperatorAddress = "operator-address"
	AvsAddress      = "avs-address"
	Socket          = "socket"
	OperatorSetId   = "operator-set-id"
)

type BlsHelperConfig struct {
	Debug           bool   `mapstructure:"debug"`
	KeyfilePath     string `mapstructure:"keyfile_path"`
	KeyPassword     string `mapstructure:"key_password"`
	RpcUrl          string `mapstructure:"rpc_url"`
	OperatorAddress string `mapstructure:"operator_address"`
	AvsAddress      string `mapstructure:"avs_address"`
	OperatorSetId   uint32 `mapstructure:"operator_set_id"`
	Socket          string `mapstructure:"socket"`
}

func NewBlsHelperConfig() *BlsHelperConfig {
	return &BlsHelperConfig{
		Debug:           viper.GetBool(config.NormalizeFlagName(Debug)),
		KeyfilePath:     viper.GetString(config.NormalizeFlagName(KeyfilePath)),
		KeyPassword:     viper.GetString(config.NormalizeFlagName(KeyPassword)),
		RpcUrl:          viper.GetString(config.NormalizeFlagName(RpcUrl)),
		OperatorAddress: viper.GetString(config.NormalizeFlagName(OperatorAddress)),
		AvsAddress:      viper.GetString(config.NormalizeFlagName(AvsAddress)),
		Socket:          viper.GetString(config.NormalizeFlagName(Socket)),
		OperatorSetId:   viper.GetUint32(config.NormalizeFlagName(OperatorSetId)),
	}
}

package blsHelperConfig

import (
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/spf13/viper"
)

const (
	Debug               = "debug"
	KeyfilePath         = "keyfile-path"
	KeyPassword         = "key-password"
	RpcUrl              = "rpc-url"
	OperatorAddress     = "operator-address"
	AvsRegistrarAddress = "avs-registrar-address"
	Socket              = "socket"
)

type BlsHelperConfig struct {
	Debug               bool   `mapstructure:"debug"`
	KeyfilePath         string `mapstructure:"keyfile_path"`
	KeyPassword         string `mapstructure:"key_password"`
	RpcUrl              string `mapstructure:"rpc_url"`
	OperatorAddress     string `mapstructure:"operator_address"`
	AvsRegistrarAddress string `mapstructure:"avs_registrar_address"`
	Socket              string `mapstructure:"socket"`
}

func NewBlsHelperConfig() *BlsHelperConfig {
	return &BlsHelperConfig{
		Debug:               viper.GetBool(config.NormalizeFlagName(Debug)),
		KeyfilePath:         viper.GetString(config.NormalizeFlagName(KeyfilePath)),
		KeyPassword:         viper.GetString(config.NormalizeFlagName(KeyPassword)),
		RpcUrl:              viper.GetString(config.NormalizeFlagName(RpcUrl)),
		OperatorAddress:     viper.GetString(config.NormalizeFlagName(OperatorAddress)),
		AvsRegistrarAddress: viper.GetString(config.NormalizeFlagName(AvsRegistrarAddress)),
		Socket:              viper.GetString(config.NormalizeFlagName(Socket)),
	}
}

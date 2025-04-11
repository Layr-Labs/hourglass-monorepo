package config

type ChainID uint

const (
	ChainID_EthereumMainnet ChainID = 1
	ChainID_EthereumHolesky ChainID = 17000
	ChainID_EthereumHoodi   ChainID = 560048
)

var (
	SupportedChainIds = []ChainID{
		ChainID_EthereumMainnet,
		ChainID_EthereumHolesky,
		ChainID_EthereumHoodi,
	}
)

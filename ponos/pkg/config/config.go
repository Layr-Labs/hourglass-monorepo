package config

type ChainId uint

const (
	ChainId_EthereumMainnet ChainId = 1
	ChainId_EthereumHolesky ChainId = 17000
	ChainId_EthereumHoodi   ChainId = 560048
)

var (
	SupportedChainIds = []ChainId{
		ChainId_EthereumMainnet,
		ChainId_EthereumHolesky,
		ChainId_EthereumHoodi,
	}
)

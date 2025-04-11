package localPeeringDataFetcher

import (
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/peering"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type LocalPeeringDataFetcherConfig struct {
	Url string
}

type LocalPeeringDataFetcher struct {
	config *LocalPeeringDataFetcherConfig
	logger *zap.Logger
}

func NewLocalPeeringDataFetcher(
	config *LocalPeeringDataFetcherConfig,
	logger *zap.Logger,
	httpClient *http.Client,
) *LocalPeeringDataFetcher {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 5 * time.Second,
		}
	}
	return &LocalPeeringDataFetcher{
		config: config,
		logger: logger,
	}
}

func (lpdf *LocalPeeringDataFetcher) ListExecutorOperators() ([]*peering.ExecutorOperatorPeerInfo, error) {
	return []*peering.ExecutorOperatorPeerInfo{
		{
			NetworkAddress: "localhost",
			Port:           8400,
			PublicKey:      "totally a public key",
		},
	}, nil
}

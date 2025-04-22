package aggregatorClient

import (
	aggregatorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/aggregator"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients"
)

func NewAggregatorClient(fullUrl string, insecureConn bool) (aggregatorV1.AggregatorServiceClient, error) {
	grpcClient, err := clients.NewGrpcClient(fullUrl, insecureConn)
	if err != nil {
		return nil, err
	}
	return aggregatorV1.NewAggregatorServiceClient(grpcClient), nil
}

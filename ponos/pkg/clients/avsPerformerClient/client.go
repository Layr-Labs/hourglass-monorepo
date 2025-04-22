package avsPerformerClient

import (
	performerV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/performer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients"
)

func NewAvsPerformerClient(fullUrl string, insecureConn bool) (performerV1.PerformerServiceClient, error) {
	grpcClient, err := clients.NewGrpcClient(fullUrl, insecureConn)
	if err != nil {
		return nil, err
	}
	return performerV1.NewPerformerServiceClient(grpcClient), nil
}

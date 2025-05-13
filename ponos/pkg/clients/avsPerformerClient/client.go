package avsPerformerClient

import (
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients"
	performerV1 "github.com/Layr-Labs/protocol-apis/gen/protos/eigenlayer/hourglass/v1/performer"
)

func NewAvsPerformerClient(fullUrl string, insecureConn bool) (performerV1.PerformerServiceClient, error) {
	grpcClient, err := clients.NewGrpcClient(fullUrl, insecureConn)
	if err != nil {
		return nil, err
	}
	return performerV1.NewPerformerServiceClient(grpcClient), nil
}

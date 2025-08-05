package avsPerformerClient

import (
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients"
	performerV1 "github.com/Layr-Labs/protocol-apis/gen/protos/eigenlayer/hourglass/v1/performer"
	healthV1 "github.com/Layr-Labs/protocol-apis/gen/protos/grpc/health/v1"
	"google.golang.org/grpc"
)

type PerformerClient struct {
	HealthClient    healthV1.HealthClient
	PerformerClient performerV1.PerformerServiceClient
}

func NewAvsPerformerClient(fullUrl string, insecureConn bool) (*PerformerClient, error) {
	grpcClient, err := clients.NewGrpcClient(fullUrl, insecureConn)
	if err != nil {
		return nil, err
	}

	return NewAvsPerformerClientWithConn(grpcClient)
}

func NewAvsPerformerClientWithConn(conn *grpc.ClientConn) (*PerformerClient, error) {
	if conn == nil {
		return nil, fmt.Errorf("connection cannot be nil")
	}
	hc := healthV1.NewHealthClient(conn)
	pc := performerV1.NewPerformerServiceClient(conn)

	return &PerformerClient{
		HealthClient:    hc,
		PerformerClient: pc,
	}, nil
}

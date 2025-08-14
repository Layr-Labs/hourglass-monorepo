package aggregatorClient

import (
	"context"
	"crypto/tls"

	aggregatorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/aggregator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// NewAggregatorManagementClient creates a new aggregator management client
func NewAggregatorManagementClient(fullUrl string, insecureConn bool) (aggregatorV1.AggregatorManagementServiceClient, error) {
	var opts []grpc.DialOption

	if insecureConn {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		config := &tls.Config{}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(config)))
	}

	conn, err := grpc.NewClient(fullUrl, opts...)
	if err != nil {
		return nil, err
	}

	return aggregatorV1.NewAggregatorManagementServiceClient(conn), nil
}

// Helper function to close the connection
func CloseConnection(ctx context.Context, client aggregatorV1.AggregatorManagementServiceClient) error {
	if gc, ok := client.(interface{ Close() error }); ok {
		return gc.Close()
	}
	return nil
}

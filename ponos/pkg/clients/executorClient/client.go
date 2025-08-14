package executorClient

import (
	executorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients"
)

func NewExecutorClient(fullUrl string, insecureConn bool) (executorV1.ExecutorServiceClient, error) {
	grpcClient, err := clients.NewGrpcClientWithRetry(fullUrl, insecureConn, clients.DefaultRetryConfig())
	if err != nil {
		return nil, err
	}
	return executorV1.NewExecutorServiceClient(grpcClient), nil
}

func NewExecutorManagementClient(fullUrl string, insecureConn bool) (executorV1.ExecutorManagementServiceClient, error) {
	grpcClient, err := clients.NewGrpcClientWithRetry(fullUrl, insecureConn, clients.DefaultRetryConfig())
	if err != nil {
		return nil, err
	}
	return executorV1.NewExecutorManagementServiceClient(grpcClient), nil
}

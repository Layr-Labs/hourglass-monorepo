package avsPerformerClient

import (
	"crypto/tls"
	performerV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/performer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"math"
	"strings"
)

func newGrpcClient(url string, insecureConn bool) (*grpc.ClientConn, error) {
	var creds grpc.DialOption
	if strings.Contains(url, "localhost:") || strings.Contains(url, "127.0.0.1:") || insecureConn {
		creds = grpc.WithTransportCredentials(insecure.NewCredentials())
	} else {
		creds = grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: false}))
	}

	opts := []grpc.DialOption{
		creds,
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(math.MaxInt32)),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(math.MaxInt32)),
	}

	return grpc.NewClient(url, opts...)
}

func NewAvsPerformerClient(fullUrl string, insecureConn bool) (performerV1.PerformerServiceClient, error) {
	grpcClient, err := newGrpcClient(fullUrl, insecureConn)
	if err != nil {
		return nil, err
	}
	return performerV1.NewPerformerServiceClient(grpcClient), nil
}

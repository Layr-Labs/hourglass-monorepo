package client

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
	pb "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/aggregator"
)

// AggregatorClient provides a client for the Aggregator Management Service
type AggregatorClient struct {
	conn   *grpc.ClientConn
	client pb.AggregatorManagementServiceClient
	logger logger.Logger
}

// NewAggregatorClient creates a new AggregatorClient
func NewAggregatorClient(address string, log logger.Logger) (*AggregatorClient, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to create aggregator client: %w", err)
	}

	return &AggregatorClient{
		conn:   conn,
		client: pb.NewAggregatorManagementServiceClient(conn),
		logger: log,
	}, nil
}

// RegisterAvs registers an AVS with the aggregator
func (c *AggregatorClient) RegisterAvs(ctx context.Context, avsAddress string, chainIDs []uint32) error {
	c.logger.Info("Registering AVS with aggregator",
		zap.String("avsAddress", avsAddress),
		zap.Uint32s("chainIDs", chainIDs))

	req := &pb.RegisterAvsRequest{
		AvsAddress: avsAddress,
		ChainIds:   chainIDs,
	}

	resp, err := c.client.RegisterAvs(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to register AVS: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("AVS registration returned unsuccessful")
	}

	c.logger.Info("Successfully registered AVS with aggregator")
	return nil
}

// DeRegisterAvs deregisters an AVS from the aggregator
func (c *AggregatorClient) DeRegisterAvs(ctx context.Context, avsAddress string) error {
	c.logger.Info("Deregistering AVS from aggregator", zap.String("avsAddress", avsAddress))

	req := &pb.DeRegisterAvsRequest{
		AvsAddress: avsAddress,
	}

	resp, err := c.client.DeRegisterAvs(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to deregister AVS: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("AVS deregistration returned unsuccessful")
	}

	c.logger.Info("Successfully deregistered AVS from aggregator")
	return nil
}

// Close closes the gRPC connection
func (c *AggregatorClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

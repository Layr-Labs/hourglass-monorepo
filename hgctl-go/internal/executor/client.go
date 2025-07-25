package executor

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
	pb "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
)

type Client struct {
	conn   *grpc.ClientConn
	client pb.ExecutorServiceClient
	logger logger.Logger
}

func NewClient(address string, log logger.Logger) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to executor: %w", err)
	}

	return &Client{
		conn:   conn,
		client: pb.NewExecutorServiceClient(conn),
		logger: log,
	}, nil
}

func (c *Client) DeployArtifact(ctx context.Context, avsAddress, digest, registryName string) (string, error) {
	req := &pb.DeployArtifactRequest{
		AvsAddress:  avsAddress,
		Digest:      digest,
		RegistryUrl: registryName,
	}

	resp, err := c.client.DeployArtifact(ctx, req)
	if err != nil {
		return "", fmt.Errorf("deploy artifact RPC failed: %w", err)
	}

	if !resp.Success {
		return "", fmt.Errorf("deployment failed: %s", resp.Message)
	}

	return resp.DeploymentId, nil
}

func (c *Client) GetPerformers(ctx context.Context, avsAddress string) ([]*pb.Performer, error) {
	req := &pb.ListPerformersRequest{
		AvsAddress: avsAddress,
	}

	resp, err := c.client.ListPerformers(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("list performers RPC failed: %w", err)
	}

	return resp.Performers, nil
}

func (c *Client) RemovePerformer(ctx context.Context, performerID string) error {
	req := &pb.RemovePerformerRequest{
		PerformerId: performerID,
	}

	resp, err := c.client.RemovePerformer(ctx, req)
	if err != nil {
		return fmt.Errorf("remove performer RPC failed: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("removal failed: %s", resp.Message)
	}

	return nil
}

func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

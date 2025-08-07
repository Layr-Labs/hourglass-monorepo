package client

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
	pb "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
)

type Client struct {
	conn   *grpc.ClientConn
	client pb.ExecutorManagementServiceClient
	logger logger.Logger
}

func NewExecutorClient(address string, log logger.Logger) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &Client{
		conn:   conn,
		client: pb.NewExecutorManagementServiceClient(conn),
		logger: log,
	}, nil
}

// DeployPerformer deploys a performer to the Executor
func (c *Client) DeployPerformer(ctx context.Context, avsAddress, digest, registryName string) (string, error) {
	return c.DeployPerformerWithEnv(ctx, avsAddress, digest, registryName, nil)
}

// DeployPerformerWithEnv deploys an artifact with environment variables
func (c *Client) DeployPerformerWithEnv(
	ctx context.Context,
	avsAddress string,
	digest string,
	registryName string,
	envVars map[string]string) (string, error) {

	var performerEnvs []*pb.PerformerEnv
	for key, value := range envVars {
		performerEnvs = append(performerEnvs, &pb.PerformerEnv{
			Name:  key,
			Value: value,
		})
	}

	req := &pb.DeployArtifactRequest{
		AvsAddress:  avsAddress,
		Digest:      digest,
		RegistryUrl: registryName,
		Env:         performerEnvs,
	}

	c.logger.Info("Deploying performer",
		zap.String("registry", registryName),
		zap.Int("envCount", len(performerEnvs)))

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

package executorClient

import (
	"context"
	"fmt"

	commonV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/common"
	executorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/auth"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer"
)

// AuthenticatedExecutorClient wraps the executor client with authentication
type AuthenticatedExecutorClient struct {
	taskClient       executorV1.ExecutorServiceClient
	managementClient executorV1.ExecutorManagementServiceClient
	operatorAddress  string
	signer           signer.ISigner
}

// NewAuthenticatedExecutorClient creates a new authenticated executor client
func NewAuthenticatedExecutorClient(fullUrl string, operatorAddress string, signer signer.ISigner, insecureConn bool) (*AuthenticatedExecutorClient, error) {
	taskClient, err := NewExecutorClient(fullUrl, insecureConn)
	if err != nil {
		return nil, err
	}

	managementClient, err := NewExecutorManagementClient(fullUrl, insecureConn)
	if err != nil {
		return nil, err
	}

	return &AuthenticatedExecutorClient{
		taskClient:       taskClient,
		managementClient: managementClient,
		operatorAddress:  operatorAddress,
		signer:           signer,
	}, nil
}

// createAuthSignature creates an authentication signature for a request
func (c *AuthenticatedExecutorClient) createAuthSignature(ctx context.Context) (*commonV1.AuthSignature, error) {
	// First, get a challenge token from the server
	tokenResp, err := c.managementClient.GetChallengeToken(ctx, &executorV1.GetChallengeTokenRequest{
		OperatorAddress: c.operatorAddress,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get challenge token: %w", err)
	}

	// Simplified authentication: only sign the challenge token
	signedMessage := auth.ConstructSignedMessage(tokenResp.ChallengeToken)

	// Sign the message
	signature, err := c.signer.SignMessage(signedMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to sign message: %w", err)
	}

	return &commonV1.AuthSignature{
		ChallengeToken: tokenResp.ChallengeToken,
		Signature:      signature,
	}, nil
}

// DeployArtifact deploys an artifact with authentication
func (c *AuthenticatedExecutorClient) DeployArtifact(ctx context.Context, req *executorV1.DeployArtifactRequest) (*executorV1.DeployArtifactResponse, error) {
	// Create auth signature
	auth, err := c.createAuthSignature(ctx)
	if err != nil {
		return nil, err
	}

	// Set auth field
	req.Auth = auth

	// Make the authenticated request
	return c.managementClient.DeployArtifact(ctx, req)
}

// ListPerformers lists performers with authentication
func (c *AuthenticatedExecutorClient) ListPerformers(ctx context.Context, req *executorV1.ListPerformersRequest) (*executorV1.ListPerformersResponse, error) {
	// Create auth signature
	auth, err := c.createAuthSignature(ctx)
	if err != nil {
		return nil, err
	}

	// Set auth field
	req.Auth = auth

	// Make the authenticated request
	return c.managementClient.ListPerformers(ctx, req)
}

// RemovePerformer removes a performer with authentication
func (c *AuthenticatedExecutorClient) RemovePerformer(ctx context.Context, req *executorV1.RemovePerformerRequest) (*executorV1.RemovePerformerResponse, error) {
	// Create auth signature
	auth, err := c.createAuthSignature(ctx)
	if err != nil {
		return nil, err
	}

	// Set auth field
	req.Auth = auth

	// Make the authenticated request
	return c.managementClient.RemovePerformer(ctx, req)
}

// SubmitTask submits a task without authentication (unchanged)
func (c *AuthenticatedExecutorClient) SubmitTask(ctx context.Context, req *executorV1.TaskSubmission) (*executorV1.TaskResult, error) {
	return c.taskClient.SubmitTask(ctx, req)
}

// GetTaskClient returns the underlying executor task client
func (c *AuthenticatedExecutorClient) GetTaskClient() executorV1.ExecutorServiceClient {
	return c.taskClient
}

// GetManagementClient returns the underlying executor management client
func (c *AuthenticatedExecutorClient) GetManagementClient() executorV1.ExecutorManagementServiceClient {
	return c.managementClient
}

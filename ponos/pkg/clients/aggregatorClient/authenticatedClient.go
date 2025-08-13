package aggregatorClient

import (
	"context"
	"fmt"

	aggregatorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/aggregator"
	commonV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/common"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/auth"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer"
)

// AuthenticatedAggregatorClient wraps the aggregator client with authentication
type AuthenticatedAggregatorClient struct {
	managementClient  aggregatorV1.AggregatorManagementServiceClient
	aggregatorAddress string
	signer            signer.ISigner
}

// NewAuthenticatedAggregatorClient creates a new authenticated aggregator client
func NewAuthenticatedAggregatorClient(fullUrl string, aggregatorAddress string, signer signer.ISigner, insecureConn bool) (*AuthenticatedAggregatorClient, error) {
	managementClient, err := NewAggregatorManagementClient(fullUrl, insecureConn)
	if err != nil {
		return nil, err
	}

	return &AuthenticatedAggregatorClient{
		managementClient:  managementClient,
		aggregatorAddress: aggregatorAddress,
		signer:            signer,
	}, nil
}

// createAuthSignature creates an authentication signature for a request
func (c *AuthenticatedAggregatorClient) createAuthSignature(ctx context.Context) (*commonV1.AuthSignature, error) {
	// First, get a challenge token from the server
	tokenResp, err := c.managementClient.GetChallengeToken(ctx, &aggregatorV1.AggregatorGetChallengeTokenRequest{
		AggregatorAddress: c.aggregatorAddress,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get challenge token: %w", err)
	}

	// Construct and sign the message (just the token)
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

// RegisterAvs registers an AVS with authentication
func (c *AuthenticatedAggregatorClient) RegisterAvs(ctx context.Context, req *aggregatorV1.RegisterAvsRequest) (*aggregatorV1.RegisterAvsResponse, error) {
	// Create auth signature
	auth, err := c.createAuthSignature(ctx)
	if err != nil {
		return nil, err
	}

	// Set auth field
	req.Auth = auth

	// Make the authenticated request
	return c.managementClient.RegisterAvs(ctx, req)
}

// DeRegisterAvs deregisters an AVS with authentication
func (c *AuthenticatedAggregatorClient) DeRegisterAvs(ctx context.Context, req *aggregatorV1.DeRegisterAvsRequest) (*aggregatorV1.DeRegisterAvsResponse, error) {
	// Create auth signature
	auth, err := c.createAuthSignature(ctx)
	if err != nil {
		return nil, err
	}

	// Set auth field
	req.Auth = auth

	// Make the authenticated request
	return c.managementClient.DeRegisterAvs(ctx, req)
}

// GetManagementClient returns the underlying aggregator management client
func (c *AuthenticatedAggregatorClient) GetManagementClient() aggregatorV1.AggregatorManagementServiceClient {
	return c.managementClient
}
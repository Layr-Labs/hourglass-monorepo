package rewards

import (
	"context"
	"errors"
	"testing"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/client"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const (
	errFailedToGetProof  = "failed to get proof"
	errFailedToClaim     = "failed to claim"
)

func TestClaimAction_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRewardsClient := mocks.NewMockRewardsClientInterface(ctrl)
	mockContractClient := mocks.NewMockContractClientInterface(ctrl)
	log := logger.NewTestLogger()
	ctx := context.Background()
	currentCtx := &config.Context{
		Name:            testContextName,
		OperatorAddress: testOperatorAddress,
	}

	proof := &client.ClaimProof{
		RootIndex:   1,
		EarnerIndex: 2,
		TokenLeaves: []client.TokenLeaf{
			{Token: "0xtoken1", CumulativeEarnings: "1000"},
			{Token: "0xtoken2", CumulativeEarnings: "2000"},
		},
	}

	mockRewardsClient.EXPECT().
		GetClaimProof(ctx, testEarnerAddress).
		Return(proof, nil)

	mockContractClient.EXPECT().
		ProcessClaim(ctx, testRewardsCoordinator, proof).
		Return(testTxHash, nil)

	err := executeClaim(ctx, mockRewardsClient, mockContractClient, log, currentCtx, testEarnerAddress, testRewardsCoordinator)
	assert.NoError(t, err)
}

func TestClaimAction_DefaultsToOperatorAddress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRewardsClient := mocks.NewMockRewardsClientInterface(ctrl)
	mockContractClient := mocks.NewMockContractClientInterface(ctrl)
	log := logger.NewTestLogger()
	ctx := context.Background()
	currentCtx := &config.Context{
		Name:            testContextName,
		OperatorAddress: testOperatorAddress,
	}

	proof := &client.ClaimProof{
		RootIndex:   1,
		EarnerIndex: 2,
		TokenLeaves: []client.TokenLeaf{
			{Token: "0xtoken1", CumulativeEarnings: "1000"},
		},
	}

	mockRewardsClient.EXPECT().
		GetClaimProof(ctx, testOperatorAddress).
		Return(proof, nil)

	mockContractClient.EXPECT().
		ProcessClaim(ctx, testRewardsCoordinator, proof).
		Return(testTxHash, nil)

	err := executeClaim(ctx, mockRewardsClient, mockContractClient, log, currentCtx, "", testRewardsCoordinator)
	assert.NoError(t, err)
}

func TestClaimAction_NoEarnerAddress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRewardsClient := mocks.NewMockRewardsClientInterface(ctrl)
	mockContractClient := mocks.NewMockContractClientInterface(ctrl)
	log := logger.NewTestLogger()
	ctx := context.Background()
	currentCtx := &config.Context{
		Name:            testContextName,
		OperatorAddress: "",
	}

	err := executeClaim(ctx, mockRewardsClient, mockContractClient, log, currentCtx, "", testRewardsCoordinator)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "earner address not provided")
}

func TestClaimAction_GetClaimProofError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRewardsClient := mocks.NewMockRewardsClientInterface(ctrl)
	mockContractClient := mocks.NewMockContractClientInterface(ctrl)
	log := logger.NewTestLogger()
	ctx := context.Background()
	currentCtx := &config.Context{
		Name:            testContextName,
		OperatorAddress: testOperatorAddress,
	}

	mockRewardsClient.EXPECT().
		GetClaimProof(ctx, testEarnerAddress).
		Return(nil, errors.New(errFailedToGetProof))

	err := executeClaim(ctx, mockRewardsClient, mockContractClient, log, currentCtx, testEarnerAddress, testRewardsCoordinator)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get claim proof")
	assert.Contains(t, err.Error(), errFailedToGetProof)
}

func TestClaimAction_ProcessClaimError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRewardsClient := mocks.NewMockRewardsClientInterface(ctrl)
	mockContractClient := mocks.NewMockContractClientInterface(ctrl)
	log := logger.NewTestLogger()
	ctx := context.Background()
	currentCtx := &config.Context{
		Name:            testContextName,
		OperatorAddress: testOperatorAddress,
	}

	proof := &client.ClaimProof{
		RootIndex:   1,
		EarnerIndex: 2,
		TokenLeaves: []client.TokenLeaf{
			{Token: "0xtoken1", CumulativeEarnings: "1000"},
		},
	}

	mockRewardsClient.EXPECT().
		GetClaimProof(ctx, testEarnerAddress).
		Return(proof, nil)

	mockContractClient.EXPECT().
		ProcessClaim(ctx, testRewardsCoordinator, proof).
		Return("", errors.New(errFailedToClaim))

	err := executeClaim(ctx, mockRewardsClient, mockContractClient, log, currentCtx, testEarnerAddress, testRewardsCoordinator)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to process claim")
	assert.Contains(t, err.Error(), errFailedToClaim)
}
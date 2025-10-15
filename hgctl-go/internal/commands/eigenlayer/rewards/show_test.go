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
	errFailedToGetRewards = "failed to get rewards"
)

func TestShowRewardsAction_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRewardsClient := mocks.NewMockRewardsClientInterface(ctrl)
	log := logger.NewTestLogger()
	ctx := context.Background()
	currentCtx := &config.Context{
		Name:            testContextName,
		OperatorAddress: testOperatorAddress,
	}

	summary := &client.RewardsSummary{
		Earner: testEarnerAddress,
		Tokens: []*client.TokenReward{
			{Token: "0xtoken1", Earned: "1000", Claimed: "500", Claimable: "500"},
			{Token: "0xtoken2", Earned: "2000", Claimed: "1000", Claimable: "1000"},
		},
	}

	mockRewardsClient.EXPECT().
		GetSummarizedRewards(ctx, testEarnerAddress).
		Return(summary, nil)

	err := executeShowRewards(ctx, mockRewardsClient, log, currentCtx, testEarnerAddress)
	assert.NoError(t, err)
}

func TestShowRewardsAction_DefaultsToOperatorAddress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRewardsClient := mocks.NewMockRewardsClientInterface(ctrl)
	log := logger.NewTestLogger()
	ctx := context.Background()
	currentCtx := &config.Context{
		Name:            testContextName,
		OperatorAddress: testOperatorAddress,
	}

	summary := &client.RewardsSummary{
		Earner: testOperatorAddress,
		Tokens: []*client.TokenReward{
			{Token: "0xtoken1", Earned: "1000", Claimed: "500", Claimable: "500"},
		},
	}

	mockRewardsClient.EXPECT().
		GetSummarizedRewards(ctx, testOperatorAddress).
		Return(summary, nil)

	err := executeShowRewards(ctx, mockRewardsClient, log, currentCtx, "")
	assert.NoError(t, err)
}

func TestShowRewardsAction_NoRewardsFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRewardsClient := mocks.NewMockRewardsClientInterface(ctrl)
	log := logger.NewTestLogger()
	ctx := context.Background()
	currentCtx := &config.Context{
		Name:            testContextName,
		OperatorAddress: testOperatorAddress,
	}

	summary := &client.RewardsSummary{
		Earner: testEarnerAddress,
		Tokens: []*client.TokenReward{},
	}

	mockRewardsClient.EXPECT().
		GetSummarizedRewards(ctx, testEarnerAddress).
		Return(summary, nil)

	err := executeShowRewards(ctx, mockRewardsClient, log, currentCtx, testEarnerAddress)
	assert.NoError(t, err)
}

func TestShowRewardsAction_NoEarnerAddress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRewardsClient := mocks.NewMockRewardsClientInterface(ctrl)
	log := logger.NewTestLogger()
	ctx := context.Background()
	currentCtx := &config.Context{
		Name:            testContextName,
		OperatorAddress: "",
	}

	err := executeShowRewards(ctx, mockRewardsClient, log, currentCtx, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "earner address not provided")
}

func TestShowRewardsAction_GetSummarizedRewardsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRewardsClient := mocks.NewMockRewardsClientInterface(ctrl)
	log := logger.NewTestLogger()
	ctx := context.Background()
	currentCtx := &config.Context{
		Name:            testContextName,
		OperatorAddress: testOperatorAddress,
	}

	mockRewardsClient.EXPECT().
		GetSummarizedRewards(ctx, testEarnerAddress).
		Return(nil, errors.New(errFailedToGetRewards))

	err := executeShowRewards(ctx, mockRewardsClient, log, currentCtx, testEarnerAddress)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get rewards")
	assert.Contains(t, err.Error(), errFailedToGetRewards)
}

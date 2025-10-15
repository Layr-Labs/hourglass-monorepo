package rewards

import (
	"context"
	"errors"
	"testing"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const (
	testEarnerAddress      = "0x1111111111111111111111111111111111111111"
	testClaimerAddress     = "0x2222222222222222222222222222222222222222"
	testRewardsCoordinator = "0x3333333333333333333333333333333333333333"
	testOperatorAddress    = "0x9999999999999999999999999999999999999999"
	testTxHash             = "0xabc123def456"
	testContextName        = "test"
	errTransactionReverted = "transaction reverted"
)

func TestSetClaimerAction_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockContractClient := mocks.NewMockContractClientInterface(ctrl)
	log := logger.NewTestLogger()
	ctx := context.Background()
	currentCtx := &config.Context{
		Name:            testContextName,
		OperatorAddress: testOperatorAddress,
	}

	mockContractClient.EXPECT().
		SetClaimerFor(ctx, testRewardsCoordinator, testEarnerAddress, testClaimerAddress).
		Return(testTxHash, nil)

	err := executeSetClaimer(ctx, mockContractClient, log, currentCtx, testEarnerAddress, testClaimerAddress, testRewardsCoordinator)
	assert.NoError(t, err)
}

func TestSetClaimerAction_DefaultsToOperatorAddress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockContractClient := mocks.NewMockContractClientInterface(ctrl)
	log := logger.NewTestLogger()
	ctx := context.Background()
	currentCtx := &config.Context{
		Name:            testContextName,
		OperatorAddress: testOperatorAddress,
	}

	mockContractClient.EXPECT().
		SetClaimerFor(ctx, testRewardsCoordinator, testOperatorAddress, testClaimerAddress).
		Return(testTxHash, nil)

	err := executeSetClaimer(ctx, mockContractClient, log, currentCtx, "", testClaimerAddress, testRewardsCoordinator)
	assert.NoError(t, err)
}

func TestSetClaimerAction_NoEarnerAddress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockContractClient := mocks.NewMockContractClientInterface(ctrl)
	log := logger.NewTestLogger()
	ctx := context.Background()
	currentCtx := &config.Context{
		Name:            testContextName,
		OperatorAddress: "",
	}

	err := executeSetClaimer(ctx, mockContractClient, log, currentCtx, "", testClaimerAddress, testRewardsCoordinator)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "earner address not provided")
}

func TestSetClaimerAction_ContractClientError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockContractClient := mocks.NewMockContractClientInterface(ctrl)
	log := logger.NewTestLogger()
	ctx := context.Background()
	currentCtx := &config.Context{
		Name:            testContextName,
		OperatorAddress: testOperatorAddress,
	}

	mockContractClient.EXPECT().
		SetClaimerFor(ctx, testRewardsCoordinator, testEarnerAddress, testClaimerAddress).
		Return("", errors.New(errTransactionReverted))

	err := executeSetClaimer(ctx, mockContractClient, log, currentCtx, testEarnerAddress, testClaimerAddress, testRewardsCoordinator)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to set claimer")
	assert.Contains(t, err.Error(), errTransactionReverted)
}
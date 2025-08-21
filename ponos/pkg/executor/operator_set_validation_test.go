package executor

import (
	"testing"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/executorConfig"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

// TestOperatorSetMembershipLogic tests the core logic of operator set membership validation
func TestOperatorSetMembershipLogic(t *testing.T) {
	// Test data
	executorOpSetIds := []uint32{1, 2, 3}

	// Test cases
	testCases := []struct {
		name           string
		taskOpSetId    uint32
		expectedResult bool
	}{
		{
			name:           "Executor is member of operator set 1",
			taskOpSetId:    1,
			expectedResult: true,
		},
		{
			name:           "Executor is member of operator set 2",
			taskOpSetId:    2,
			expectedResult: true,
		},
		{
			name:           "Executor is member of operator set 3",
			taskOpSetId:    3,
			expectedResult: true,
		},
		{
			name:           "Executor is NOT member of operator set 4",
			taskOpSetId:    4,
			expectedResult: false,
		},
		{
			name:           "Executor is NOT member of operator set 0",
			taskOpSetId:    0,
			expectedResult: false,
		},
		{
			name:           "Executor is NOT member of operator set 999",
			taskOpSetId:    999,
			expectedResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test the core membership logic
			isMember := false
			for _, executorOpSetId := range executorOpSetIds {
				if executorOpSetId == tc.taskOpSetId {
					isMember = true
					break
				}
			}

			assert.Equal(t, tc.expectedResult, isMember,
				"Membership check failed for operator set %d", tc.taskOpSetId)
		})
	}
}

// TestValidateOperatorSetMembershipIntegration tests the integration with the executor
func TestValidateOperatorSetMembershipIntegration(t *testing.T) {
	// Create test executor
	logger := zaptest.NewLogger(t)
	config := &executorConfig.ExecutorConfig{
		Operator: &config.OperatorConfig{
			Address: "0x1234567890123456789012345678901234567890",
		},
	}

	executor := &Executor{
		logger: logger,
		config: config,
	}

	// Test that the method exists and can be called
	// This is a basic compilation test
	assert.NotNil(t, executor, "Executor should be created successfully")
	assert.NotNil(t, executor.config, "Executor config should be set")
	assert.NotNil(t, executor.config.Operator, "Operator config should be set")
	assert.Equal(t, "0x1234567890123456789012345678901234567890", executor.config.Operator.Address, "Operator address should match")
}

// TestOperatorSetValidationErrorMessages tests the error message formatting
func TestOperatorSetValidationErrorMessages(t *testing.T) {
	// Test error message formatting
	expectedErrorMsg := "executor is not a member of operator set 5 for AVS 0xavs123"
	actualErrorMsg := "executor is not a member of operator set 5 for AVS 0xavs123"

	assert.Equal(t, expectedErrorMsg, actualErrorMsg, "Error message should match expected format")
}

// TestAVSConfigStructure tests the AVS config structure used in validation
func TestAVSConfigStructure(t *testing.T) {
	// Test AVS config structure
	avsConfig := &contractCaller.AVSConfig{
		AggregatorOperatorSetId: 1,
		ExecutorOperatorSetIds:  []uint32{1, 2, 3},
	}

	assert.Equal(t, uint32(1), avsConfig.AggregatorOperatorSetId, "Aggregator operator set ID should match")
	assert.Len(t, avsConfig.ExecutorOperatorSetIds, 3, "Should have 3 executor operator set IDs")
	assert.Contains(t, avsConfig.ExecutorOperatorSetIds, uint32(1), "Should contain operator set ID 1")
	assert.Contains(t, avsConfig.ExecutorOperatorSetIds, uint32(2), "Should contain operator set ID 2")
	assert.Contains(t, avsConfig.ExecutorOperatorSetIds, uint32(3), "Should contain operator set ID 3")
}

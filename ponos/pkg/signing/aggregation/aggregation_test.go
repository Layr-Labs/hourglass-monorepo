package aggregation

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/bn254"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Aggregation(t *testing.T) {
	// Create test operators with key pairs
	operators := make([]*Operator, 3)
	for i := 0; i < 3; i++ {
		privKey, pubKey, err := bn254.GenerateKeyPair()
		require.NoError(t, err)
		operators[i] = &Operator{
			Address:    fmt.Sprintf("0x%d", i+1), // Simple address format for testing
			PublicKey:  pubKey,
			privateKey: privKey,
		}
	}

	// Initialize new task
	taskId := []byte("test-task-1")
	taskData := []byte("test-data")
	agg, err := InitializeNewTaskWithWindow(
		context.Background(),
		taskId,
		100, // taskCreatedBlock
		1,   // operatorSetId
		66,  // thresholdPercentage (2/3)
		taskData,
		5*time.Minute,
		operators,
	)
	require.NoError(t, err)
	require.NotNil(t, agg)

	// Create a common response payload
	commonPayload := []byte("test-response-payload")
	digest := util.GetKeccak256Digest(commonPayload)

	// Store individual signatures for verification
	individualSigs := make([]*bn254.Signature, 3)

	// Simulate receiving responses from all operators
	for i, operator := range operators {
		// Create task result
		taskResult := &types.TaskResult{
			OperatorAddress: operator.Address,
			Output:          commonPayload,
		}

		// Sign the response
		sig, err := operator.privateKey.Sign(digest[:])
		require.NoError(t, err)
		taskResult.Signature = sig.Bytes()
		individualSigs[i] = sig

		// Process the signature
		err = agg.ProcessNewSignature(context.Background(), taskId, taskResult)
		require.NoError(t, err)
	}

	// Verify threshold is met
	assert.True(t, agg.SigningThresholdMet())

	// Generate final certificate
	cert, err := agg.GenerateFinalCertificate()
	require.NoError(t, err)
	require.NotNil(t, cert)

	// Verify the aggregated signature
	signersPubKey, err := bn254.NewPublicKeyFromBytes(cert.SignersPublicKey.Marshal())
	require.NoError(t, err)
	verified, err := cert.SignersSignature.Verify(signersPubKey, cert.TaskResponseDigest)
	require.NoError(t, err)
	assert.True(t, verified, "Aggregated signature verification failed")

	// Verify all responses match
	assert.Equal(t, commonPayload, cert.TaskResponse)
	assert.Equal(t, 0, len(cert.NonSignersPubKeys), "Should have no non-signers")
	assert.Equal(t, 3, len(cert.AllOperatorsPubKeys), "Should have all operators' public keys")

	// Test: Verify if an operator's signature was included
	for i, operator := range operators {
		// Get the operator's individual signature
		individualSig := individualSigs[i]
		require.NotNil(t, individualSig)

		// Create a copy of the aggregated signature
		remainingSig, err := bn254.NewSignatureFromBytes(cert.SignersSignature.Bytes())
		require.NoError(t, err)
		require.NotNil(t, remainingSig)

		// Subtract the operator's signature from the aggregated signature
		remainingSig.Sub(individualSig)

		// Create a copy of the aggregated public key
		remainingPubKey, err := bn254.NewPublicKeyFromBytes(agg.AggregatePublicKey.Bytes())
		require.NoError(t, err)
		require.NotNil(t, remainingPubKey)

		// Subtract the operator's public key from the aggregated public key
		remainingPubKey.Sub(operator.PublicKey)

		// Verify the remaining signature against the remaining public key
		verified, err := remainingSig.Verify(remainingPubKey, cert.TaskResponseDigest)
		require.NoError(t, err)
		assert.True(t, verified, "Operator %d's signature was not included in aggregation", i)
	}
}

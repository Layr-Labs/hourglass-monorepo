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
	operators := make([]*Operator, 4) // Changed to 4 operators
	privateKeys := make([]*bn254.PrivateKey, 4)
	for i := 0; i < 4; i++ {
		privKey, pubKey, err := bn254.GenerateKeyPair()
		require.NoError(t, err)
		operators[i] = &Operator{
			Address:   fmt.Sprintf("0x%d", i+1), // Simple address format for testing
			PublicKey: pubKey,
		}
		privateKeys[i] = privKey
	}

	// Initialize new task
	taskId := "0x29cebefe301c6ce1bb36b58654fea275e1cacc83"
	taskData := []byte("test-data")

	deadline := time.Now().Add(10 * time.Minute)

	agg, err := NewTaskResultAggregator(
		context.Background(),
		taskId,
		100, // taskCreatedBlock
		1,   // operatorSetId
		75,  // thresholdPercentage (3/4)
		taskData,
		&deadline,
		operators,
	)
	require.NoError(t, err)
	require.NotNil(t, agg)

	// Create a common response payload
	commonPayload := []byte("test-response-payload")
	digest := util.GetKeccak256Digest(commonPayload)

	// Store individual signatures for verification
	individualSigs := make([]*bn254.Signature, 3) // Only store 3 signatures since one operator won't sign
	remainingPubKeys := make([]*bn254.PublicKey, 3)
	remainingSigs := make([]*bn254.Signature, 3)

	// Simulate receiving responses from all operators except the last one
	for i := 0; i < 3; i++ { // Only process first 3 operators
		operator := operators[i]
		// Create task result
		taskResult := &types.TaskResult{
			OperatorAddress: operator.Address,
			Output:          commonPayload,
		}

		// Sign the response
		sig, err := privateKeys[i].Sign(digest[:])
		require.NoError(t, err)
		taskResult.Signature = sig.Bytes()
		individualSigs[i] = sig
		remainingPubKeys[i] = operator.PublicKey
		remainingSigs[i] = sig

		// Process the signature
		err = agg.ProcessNewSignature(context.Background(), taskId, taskResult)
		require.NoError(t, err)
	}

	// Verify threshold is met (3/4 operators signed)
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
	assert.Equal(t, 1, len(cert.NonSignersPubKeys), "Should have one non-signer")
	assert.Equal(t, 4, len(cert.AllOperatorsPubKeys), "Should have all operators' public keys")

	// Verify the non-signer is correctly identified
	nonSignerPubKey := cert.NonSignersPubKeys[0]
	assert.Equal(t, operators[3].PublicKey.Bytes(), nonSignerPubKey.Bytes(), "Non-signer public key should match the last operator")

	// Test: Verify if an operator's signature was included
	// We can verify this by checking that the remaining signatures verify correctly
	verified, err = bn254.BatchVerify(remainingPubKeys, cert.TaskResponseDigest, remainingSigs)
	require.NoError(t, err)
	assert.True(t, verified, "Remaining signatures should verify correctly")

	// Test: Verify that the non-signer's signature is not included
	// Create a new signature array including the non-signer's signature
	allSigs := append(remainingSigs, individualSigs[0])            // Add a duplicate signature
	allPubKeys := append(remainingPubKeys, operators[3].PublicKey) // Add non-signer's public key
	verified, err = bn254.BatchVerify(allPubKeys, cert.TaskResponseDigest, allSigs)
	require.NoError(t, err)
	assert.False(t, verified, "Verification should fail when including non-signer's public key")
}

package aggregation

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Layr-Labs/crypto-libs/pkg/signing"
	"github.com/ethereum/go-ethereum/common"

	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	cryptoLibsEcdsa "github.com/Layr-Labs/crypto-libs/pkg/ecdsa"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper constants for tests
const testAvsAddress = "0x1234567890123456789012345678901234567890"

// Helper function to create a properly signed TaskResult for BN254
func createSignedBN254TaskResult(taskId string, operator *Operator[signing.PublicKey], operatorSetId uint32, output []byte, privateKey *bn254.PrivateKey) (*types.TaskResult, error) {
	taskResult := &types.TaskResult{
		TaskId:          taskId,
		AvsAddress:      testAvsAddress,
		OperatorAddress: operator.Address,
		OperatorSetId:   operatorSetId,
		Output:          output,
	}

	// Step 1: Sign the result (same for all operators)
	// We need to sign hash(output), and SignSolidityCompatible expects a [32]byte
	outputDigest := util.GetKeccak256Digest(output)
	resultSig, err := privateKey.SignSolidityCompatible(outputDigest)
	if err != nil {
		return nil, err
	}
	taskResult.ResultSignature = resultSig.Bytes()

	// Step 2: Sign the auth data (unique per operator)
	resultSigDigest := util.GetKeccak256Digest(taskResult.ResultSignature)
	authData := &types.AuthSignatureData{
		TaskId:          taskResult.TaskId,
		AvsAddress:      taskResult.AvsAddress,
		OperatorAddress: taskResult.OperatorAddress,
		OperatorSetId:   taskResult.OperatorSetId,
		ResultSigDigest: resultSigDigest,
	}
	authBytes := authData.ToSigningBytes()
	authDigest := util.GetKeccak256Digest(authBytes)
	authSig, err := privateKey.SignSolidityCompatible(authDigest)
	if err != nil {
		return nil, err
	}
	taskResult.AuthSignature = authSig.Bytes()

	return taskResult, nil
}

// Helper function to create a properly signed TaskResult for ECDSA
func createSignedECDSATaskResult(taskId string, operator *Operator[common.Address], operatorSetId uint32, output []byte, privateKey *cryptoLibsEcdsa.PrivateKey) (*types.TaskResult, error) {
	taskResult := &types.TaskResult{
		TaskId:          taskId,
		AvsAddress:      testAvsAddress,
		OperatorAddress: operator.Address,
		OperatorSetId:   operatorSetId,
		Output:          output,
	}

	// Step 1: Sign the result (same for all operators)
	// We need to sign hash(output), and Sign expects []byte
	outputDigest := util.GetKeccak256Digest(output)
	resultSig, err := privateKey.Sign(outputDigest[:])
	if err != nil {
		return nil, err
	}
	taskResult.ResultSignature = resultSig.Bytes()

	// Step 2: Sign the auth data (unique per operator)
	resultSigDigest := util.GetKeccak256Digest(taskResult.ResultSignature)
	authData := &types.AuthSignatureData{
		TaskId:          taskResult.TaskId,
		AvsAddress:      taskResult.AvsAddress,
		OperatorAddress: taskResult.OperatorAddress,
		OperatorSetId:   taskResult.OperatorSetId,
		ResultSigDigest: resultSigDigest,
	}
	authBytes := authData.ToSigningBytes()
	authDigest := util.GetKeccak256Digest(authBytes)
	authSig, err := privateKey.Sign(authDigest[:])
	if err != nil {
		return nil, err
	}
	taskResult.AuthSignature = authSig.Bytes()

	return taskResult, nil
}

func Test_Aggregation(t *testing.T) {
	t.Run("BN254", func(t *testing.T) {
		// Create test operators with key pairs
		operators := make([]*Operator[signing.PublicKey], 4) // Changed to 4 operators

		privateKeys := make([]*bn254.PrivateKey, 4)
		for i := 0; i < 4; i++ {
			privKey, pubKey, err := bn254.GenerateKeyPair()
			require.NoError(t, err)
			operators[i] = &Operator[signing.PublicKey]{
				Address:       fmt.Sprintf("0x%d", i+1), // Simple address format for testing
				PublicKey:     pubKey,
				OperatorIndex: uint32(i),
			}
			privateKeys[i] = privKey
		}

		// Initialize new task
		taskId := "0x29cebefe301c6ce1bb36b58654fea275e1cacc83"
		taskData := []byte("test-data")

		deadline := time.Now().Add(10 * time.Minute)

		agg, err := NewBN254TaskResultAggregator(
			context.Background(),
			taskId,
			1,    // operatorSetId
			7500, // thresholdBips (3/4 = 7500 bips)
			taskData,
			&deadline,
			operators,
		)
		require.NoError(t, err)
		require.NotNil(t, agg)

		// Create a common response payload
		commonPayload := []byte("test-response-payload")

		// Simulate receiving responses from all operators except the last one
		for i := 0; i < 3; i++ { // Only process first 3 operators
			operator := operators[i]
			// Create properly signed task result using the helper function
			taskResult, err := createSignedBN254TaskResult(
				taskId,
				operator,
				1, // operatorSetId
				commonPayload,
				privateKeys[i],
			)
			require.NoError(t, err)

			// Process the signature
			err = agg.ProcessNewSignature(context.Background(), taskResult)
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
		responseDigest := cert.TaskResponseDigest
		verified, err := cert.SignersSignature.VerifySolidityCompatible(signersPubKey, responseDigest)
		require.NoError(t, err)
		assert.True(t, verified, "Aggregated signature verification failed")

		// Verify all responses match
		assert.Equal(t, commonPayload, cert.TaskResponse)
		assert.Equal(t, 1, len(cert.NonSignersPubKeys), "Should have one non-signer")
		assert.Equal(t, 4, len(cert.AllOperatorsPubKeys), "Should have all operators' public keys")

		// Verify the non-signer is correctly identified
		nonSignerPubKey := cert.NonSignersPubKeys[0]
		assert.Equal(t, operators[3].PublicKey.Bytes(), nonSignerPubKey.Bytes(), "Non-signer public key should match the last operator")

		// Test: Verify that the aggregated signature works correctly
		// The aggregated signature should verify against the aggregated public key
		assert.NotNil(t, cert.SignersSignature, "Should have aggregated signature")
		assert.NotNil(t, cert.SignersPublicKey, "Should have aggregated public key")
	})
	t.Run("ECDSA", func(t *testing.T) {
		// Create test operators with key pairs
		operators := make([]*Operator[common.Address], 4) // Changed to 4 operators
		privateKeys := make([]*cryptoLibsEcdsa.PrivateKey, 4)
		for i := 0; i < 4; i++ {
			privKey, _, err := cryptoLibsEcdsa.GenerateKeyPair()
			require.NoError(t, err)
			derivedAddress, err := privKey.DeriveAddress()
			if err != nil {
				t.Fatalf("Failed to derive address for operator %d: %v", i, err)
			}

			operators[i] = &Operator[common.Address]{
				Address:       derivedAddress.String(),
				PublicKey:     derivedAddress,
				OperatorIndex: uint32(i),
			}
			privateKeys[i] = privKey
		}

		// Initialize new task
		taskId := "0x29cebefe301c6ce1bb36b58654fea275e1cacc83"
		taskData := []byte("test-data")

		deadline := time.Now().Add(10 * time.Minute)

		agg, err := NewECDSATaskResultAggregator(
			context.Background(),
			taskId,
			1,    // operatorSetId
			7500, // thresholdBips (3/4 = 7500 bips)
			taskData,
			&deadline,
			operators,
		)
		require.NoError(t, err)
		require.NotNil(t, agg)

		// Create a common response payload
		commonPayload := []byte("test-response-payload")

		// Simulate receiving responses from all operators except the last one
		for i := 0; i < 3; i++ { // Only process first 3 operators
			operator := operators[i]
			// Create properly signed task result using the helper function
			taskResult, err := createSignedECDSATaskResult(
				taskId,
				operator,
				1, // operatorSetId
				commonPayload,
				privateKeys[i],
			)
			require.NoError(t, err)

			// Process the signature
			err = agg.ProcessNewSignature(context.Background(), taskResult)
			require.NoError(t, err)
		}

		// Verify threshold is met (3/4 operators signed)
		assert.True(t, agg.SigningThresholdMet())

		// Generate final certificate
		cert, err := agg.GenerateFinalCertificate()
		require.NoError(t, err)
		require.NotNil(t, cert)

		// Verify all responses match
		assert.Equal(t, commonPayload, cert.TaskResponse)
		assert.Equal(t, 3, len(cert.SignersSignatures), "Should have three signers")

		// Verify certificate uses implementation's hash calculation method
		taskMessageHash := cert.GetTaskMessageHash()
		expectedHash := util.GetKeccak256Digest(commonPayload)
		assert.Equal(t, expectedHash, taskMessageHash, "Certificate should calculate correct task message hash")

		finalCert, err := agg.GenerateFinalCertificate()
		require.NoError(t, err)
		assert.NotNil(t, finalCert)
	})
}

func Test_MostCommonDigestTracking(t *testing.T) {
	t.Run("BN254 - Multiple Digests", func(t *testing.T) {
		// Create test operators with key pairs
		operators := make([]*Operator[signing.PublicKey], 5)
		privateKeys := make([]*bn254.PrivateKey, 5)
		for i := 0; i < 5; i++ {
			privKey, pubKey, err := bn254.GenerateKeyPair()
			require.NoError(t, err)
			operators[i] = &Operator[signing.PublicKey]{
				Address:       fmt.Sprintf("0x%d", i+1),
				PublicKey:     pubKey,
				OperatorIndex: uint32(i),
			}
			privateKeys[i] = privKey
		}

		// Initialize new task
		taskId := "0x29cebefe301c6ce1bb36b58654fea275e1cacc83"
		taskData := []byte("test-data")
		deadline := time.Now().Add(10 * time.Minute)

		agg, err := NewBN254TaskResultAggregator(
			context.Background(),
			taskId,
			1,    // operatorSetId
			6000, // thresholdBips (3/5 = 6000 bips)
			taskData,
			&deadline,
			operators,
		)
		require.NoError(t, err)
		require.NotNil(t, agg)

		// Create different response payloads
		payloadA := []byte("response-A")
		payloadB := []byte("response-B")
		payloadC := []byte("response-C")

		digestA := util.GetKeccak256Digest(payloadA)
		digestB := util.GetKeccak256Digest(payloadB)
		digestC := util.GetKeccak256Digest(payloadC)

		// Test scenario:
		// - Operator 0: submits digest A
		// - Operator 1: submits digest B
		// - Operator 2: submits digest A (A now has 2 votes, should become most common)
		// - Operator 3: submits digest C
		// - Operator 4: submits digest B (B now has 2 votes, but A should remain most common as it got 2 first)

		// Operator 0 submits digest A
		taskResult0, err := createSignedBN254TaskResult(taskId, operators[0], 1, payloadA, privateKeys[0])
		require.NoError(t, err)
		err = agg.ProcessNewSignature(context.Background(), taskResult0)
		require.NoError(t, err)

		// Verify mostCommonCount is 1 and points to digest A
		assert.Equal(t, 1, agg.aggregatedOperators.mostCommonCount)
		assert.Equal(t, digestA, agg.aggregatedOperators.mostCommonDigest)

		// Operator 1 submits digest B
		taskResult1, err := createSignedBN254TaskResult(taskId, operators[1], 1, payloadB, privateKeys[1])
		require.NoError(t, err)
		err = agg.ProcessNewSignature(context.Background(), taskResult1)
		require.NoError(t, err)

		// Most common should still be A (both have count 1, A came first)
		assert.Equal(t, 1, agg.aggregatedOperators.mostCommonCount)
		assert.Equal(t, digestA, agg.aggregatedOperators.mostCommonDigest)

		// Operator 2 submits digest A
		taskResult2, err := createSignedBN254TaskResult(taskId, operators[2], 1, payloadA, privateKeys[2])
		require.NoError(t, err)
		err = agg.ProcessNewSignature(context.Background(), taskResult2)
		require.NoError(t, err)

		// Now A should have count 2 and be most common
		assert.Equal(t, 2, agg.aggregatedOperators.mostCommonCount)
		assert.Equal(t, digestA, agg.aggregatedOperators.mostCommonDigest)
		assert.Equal(t, payloadA, agg.aggregatedOperators.digestGroups[agg.aggregatedOperators.mostCommonDigest].response.TaskResult.Output)

		// Operator 3 submits digest C
		taskResult3, err := createSignedBN254TaskResult(taskId, operators[3], 1, payloadC, privateKeys[3])
		require.NoError(t, err)
		err = agg.ProcessNewSignature(context.Background(), taskResult3)
		require.NoError(t, err)

		// A should still be most common with count 2
		assert.Equal(t, 2, agg.aggregatedOperators.mostCommonCount)
		assert.Equal(t, digestA, agg.aggregatedOperators.mostCommonDigest)

		// Operator 4 submits digest B
		taskResult4, err := createSignedBN254TaskResult(taskId, operators[4], 1, payloadB, privateKeys[4])
		require.NoError(t, err)
		err = agg.ProcessNewSignature(context.Background(), taskResult4)
		require.NoError(t, err)

		// A should still be most common (both A and B have count 2, but A reached 2 first)
		assert.Equal(t, 2, agg.aggregatedOperators.mostCommonCount)
		assert.Equal(t, digestA, agg.aggregatedOperators.mostCommonDigest)

		// Verify digest counts through digest groups
		assert.Equal(t, 2, agg.aggregatedOperators.digestGroups[digestA].count)
		assert.Equal(t, 2, agg.aggregatedOperators.digestGroups[digestB].count)
		assert.Equal(t, 1, agg.aggregatedOperators.digestGroups[digestC].count)

		// Verify threshold IS met (5/5 operators participated, need 3/5 for 60% participation)
		// Total participation is what matters, not consensus on the same message
		assert.True(t, agg.SigningThresholdMet())

		// Generate final certificate and verify it uses the most common response
		cert, err := agg.GenerateFinalCertificate()
		require.NoError(t, err)
		require.NotNil(t, cert)

		// Certificate should use payload A (the most common)
		assert.Equal(t, payloadA, cert.TaskResponse)
		assert.Equal(t, digestA, cert.TaskResponseDigest)
	})

	t.Run("BN254 - Single Operator", func(t *testing.T) {
		// Create a single test operator
		privKey, pubKey, err := bn254.GenerateKeyPair()
		require.NoError(t, err)

		operators := []*Operator[signing.PublicKey]{
			{
				Address:       "0x1",
				PublicKey:     pubKey,
				OperatorIndex: 0,
			},
		}

		// Initialize task with 100% threshold (requiring the single operator)
		taskId := "0x29cebefe301c6ce1bb36b58654fea275e1cacc83"
		taskData := []byte("test-data")
		deadline := time.Now().Add(10 * time.Minute)

		agg, err := NewBN254TaskResultAggregator(
			context.Background(),
			taskId,
			1,
			10000, // 100% threshold (10000 bips) - requires the single operator
			taskData,
			&deadline,
			operators,
		)
		require.NoError(t, err)

		// Single operator submits response
		payload := []byte("single-operator-response")
		digest := util.GetKeccak256Digest(payload)

		// Create properly signed task result using the helper function
		taskResult, err := createSignedBN254TaskResult(
			taskId,
			operators[0],
			1, // operatorSetId
			payload,
			privKey,
		)
		require.NoError(t, err)

		// Process the single signature
		err = agg.ProcessNewSignature(context.Background(), taskResult)
		require.NoError(t, err)

		// Verify most common digest tracking is set correctly
		assert.NotNil(t, agg.aggregatedOperators.digestGroups[digest])
		assert.Equal(t, 1, agg.aggregatedOperators.mostCommonCount)
		assert.Equal(t, digest, agg.aggregatedOperators.mostCommonDigest)
		assert.Equal(t, payload, agg.aggregatedOperators.digestGroups[agg.aggregatedOperators.mostCommonDigest].response.TaskResult.Output)

		// Verify threshold is met
		assert.True(t, agg.SigningThresholdMet())

		// Generate certificate and verify it works with single operator
		cert, err := agg.GenerateFinalCertificate()
		require.NoError(t, err)
		require.NotNil(t, cert)

		// Certificate should use the single operator's response
		assert.Equal(t, payload, cert.TaskResponse)
		assert.Equal(t, digest, cert.TaskResponseDigest)

		// Verify signature
		signersPubKey, err := bn254.NewPublicKeyFromBytes(cert.SignersPublicKey.Marshal())
		require.NoError(t, err)
		verified, err := cert.SignersSignature.VerifySolidityCompatible(signersPubKey, cert.TaskResponseDigest)
		require.NoError(t, err)
		assert.True(t, verified, "Single operator signature verification failed")

		// Should have no non-signers and one total operator
		assert.Equal(t, 0, len(cert.NonSignersPubKeys), "Should have no non-signers")
		assert.Equal(t, 1, len(cert.AllOperatorsPubKeys), "Should have one total operator")
	})

	t.Run("BN254 - Unanimous Agreement", func(t *testing.T) {
		// Create test operators
		operators := make([]*Operator[signing.PublicKey], 3)
		privateKeys := make([]*bn254.PrivateKey, 3)
		for i := 0; i < 3; i++ {
			privKey, pubKey, err := bn254.GenerateKeyPair()
			require.NoError(t, err)
			operators[i] = &Operator[signing.PublicKey]{
				Address:       fmt.Sprintf("0x%d", i+1),
				PublicKey:     pubKey,
				OperatorIndex: uint32(i),
			}
			privateKeys[i] = privKey
		}

		// Initialize task
		taskId := "0x29cebefe301c6ce1bb36b58654fea275e1cacc83"
		taskData := []byte("test-data")
		deadline := time.Now().Add(10 * time.Minute)

		agg, err := NewBN254TaskResultAggregator(
			context.Background(),
			taskId,
			1,
			10000, // 100% threshold (10000 bips)
			taskData,
			&deadline,
			operators,
		)
		require.NoError(t, err)

		// All operators submit the same response
		commonPayload := []byte("unanimous-response")
		digest := util.GetKeccak256Digest(commonPayload)

		for i := 0; i < 3; i++ {
			// Create properly signed task result using the helper function
			taskResult, err := createSignedBN254TaskResult(
				taskId,
				operators[i],
				1, // operatorSetId
				commonPayload,
				privateKeys[i],
			)
			require.NoError(t, err)

			err = agg.ProcessNewSignature(context.Background(), taskResult)
			require.NoError(t, err)

			// Most common count should increase with each submission
			assert.Equal(t, i+1, agg.aggregatedOperators.mostCommonCount)
			assert.Equal(t, digest, agg.aggregatedOperators.mostCommonDigest)
		}

		// Generate certificate
		cert, err := agg.GenerateFinalCertificate()
		require.NoError(t, err)
		assert.Equal(t, commonPayload, cert.TaskResponse)
		assert.Equal(t, digest, cert.TaskResponseDigest)
	})

	t.Run("BN254 - Insufficient Participation", func(t *testing.T) {
		// This test verifies that threshold is NOT met when total participation is below threshold,
		// even if all participating operators agree on the same message

		// Create test operators with key pairs
		operators := make([]*Operator[signing.PublicKey], 5)
		privateKeys := make([]*bn254.PrivateKey, 5)
		for i := 0; i < 5; i++ {
			privKey, pubKey, err := bn254.GenerateKeyPair()
			require.NoError(t, err)
			operators[i] = &Operator[signing.PublicKey]{
				Address:       fmt.Sprintf("0x%d", i+1),
				PublicKey:     pubKey,
				OperatorIndex: uint32(i),
			}
			privateKeys[i] = privKey
		}

		// Initialize new task
		taskId := "0x29cebefe301c6ce1bb36b58654fea275e1cacc83"
		taskData := []byte("test-data")
		deadline := time.Now().Add(10 * time.Minute)

		agg, err := NewBN254TaskResultAggregator(
			context.Background(),
			taskId,
			1,    // operatorSetId
			6000, // thresholdBips (3/5 = 6000 bips = 60% participation required)
			taskData,
			&deadline,
			operators,
		)
		require.NoError(t, err)
		require.NotNil(t, agg)

		// Create a common response payload
		commonPayload := []byte("unanimous-response")
		digest := util.GetKeccak256Digest(commonPayload)

		// Only 2 operators participate (both signing the same message)
		// This is below the 60% threshold (need 3 out of 5)
		for i := 0; i < 2; i++ {
			taskResult, err := createSignedBN254TaskResult(
				taskId,
				operators[i],
				1, // operatorSetId
				commonPayload,
				privateKeys[i],
			)
			require.NoError(t, err)

			err = agg.ProcessNewSignature(context.Background(), taskResult)
			require.NoError(t, err)
		}

		// Verify tracking
		assert.Equal(t, 2, agg.aggregatedOperators.mostCommonCount)
		assert.Equal(t, digest, agg.aggregatedOperators.mostCommonDigest)
		assert.Equal(t, 2, agg.aggregatedOperators.totalSignerCount)

		// Verify threshold is NOT met (only 2/5 operators participated, need 3/5 for 60%)
		assert.False(t, agg.SigningThresholdMet(),
			"Threshold should not be met with only 2/5 operators participating")

		// Attempting to generate certificate should still work but with only 2 signers
		cert, err := agg.GenerateFinalCertificate()
		require.NoError(t, err)
		require.NotNil(t, cert)

		// Certificate should still use the unanimous response
		assert.Equal(t, commonPayload, cert.TaskResponse)
		assert.Equal(t, digest, cert.TaskResponseDigest)

		// Should have 3 non-signers (operators 2, 3, 4)
		assert.Equal(t, 3, len(cert.NonSignerOperators))
	})

	t.Run("ECDSA - Multiple Digests", func(t *testing.T) {
		// Create test operators with key pairs
		operators := make([]*Operator[common.Address], 5)
		privateKeys := make([]*cryptoLibsEcdsa.PrivateKey, 5)
		for i := 0; i < 5; i++ {
			privKey, _, err := cryptoLibsEcdsa.GenerateKeyPair()
			require.NoError(t, err)
			derivedAddress, err := privKey.DeriveAddress()
			require.NoError(t, err)

			operators[i] = &Operator[common.Address]{
				Address:       derivedAddress.String(),
				PublicKey:     derivedAddress,
				OperatorIndex: uint32(i),
			}
			privateKeys[i] = privKey
		}

		// Initialize new task
		taskId := "0x29cebefe301c6ce1bb36b58654fea275e1cacc83"
		taskData := []byte("test-data")
		deadline := time.Now().Add(10 * time.Minute)

		agg, err := NewECDSATaskResultAggregator(
			context.Background(),
			taskId,
			1,    // operatorSetId
			6000, // thresholdBips (3/5 = 6000 bips)
			taskData,
			&deadline,
			operators,
		)
		require.NoError(t, err)
		require.NotNil(t, agg)

		// Create different response payloads
		payloadA := []byte("response-A")
		payloadB := []byte("response-B")
		payloadC := []byte("response-C")

		digestA := util.GetKeccak256Digest(payloadA)
		digestB := util.GetKeccak256Digest(payloadB)
		digestC := util.GetKeccak256Digest(payloadC)

		// Test scenario:
		// - Operator 0: submits digest A
		// - Operator 1: submits digest B
		// - Operator 2: submits digest A (A now has 2 votes, should become most common)
		// - Operator 3: submits digest C
		// - Operator 4: submits digest B (B now has 2 votes, but A should remain most common as it got 2 first)

		// Operator 0 submits digest A
		taskResult0, err := createSignedECDSATaskResult(taskId, operators[0], 1, payloadA, privateKeys[0])
		require.NoError(t, err)
		err = agg.ProcessNewSignature(context.Background(), taskResult0)
		require.NoError(t, err)

		// Verify mostCommonCount is 1 and points to digest A
		assert.Equal(t, 1, agg.aggregatedOperators.mostCommonCount)
		assert.Equal(t, digestA, agg.aggregatedOperators.mostCommonDigest)

		// Operator 1 submits digest B
		taskResult1, err := createSignedECDSATaskResult(taskId, operators[1], 1, payloadB, privateKeys[1])
		require.NoError(t, err)
		err = agg.ProcessNewSignature(context.Background(), taskResult1)
		require.NoError(t, err)

		// Most common should still be A (both have count 1, A came first)
		assert.Equal(t, 1, agg.aggregatedOperators.mostCommonCount)
		assert.Equal(t, digestA, agg.aggregatedOperators.mostCommonDigest)

		// Operator 2 submits digest A
		taskResult2, err := createSignedECDSATaskResult(taskId, operators[2], 1, payloadA, privateKeys[2])
		require.NoError(t, err)
		err = agg.ProcessNewSignature(context.Background(), taskResult2)
		require.NoError(t, err)

		// Now A should have count 2 and be most common
		assert.Equal(t, 2, agg.aggregatedOperators.mostCommonCount)
		assert.Equal(t, digestA, agg.aggregatedOperators.mostCommonDigest)
		assert.Equal(t, payloadA, agg.aggregatedOperators.digestGroups[agg.aggregatedOperators.mostCommonDigest].response.TaskResult.Output)

		// Operator 3 submits digest C
		taskResult3, err := createSignedECDSATaskResult(taskId, operators[3], 1, payloadC, privateKeys[3])
		require.NoError(t, err)
		err = agg.ProcessNewSignature(context.Background(), taskResult3)
		require.NoError(t, err)

		// A should still be most common with count 2
		assert.Equal(t, 2, agg.aggregatedOperators.mostCommonCount)
		assert.Equal(t, digestA, agg.aggregatedOperators.mostCommonDigest)

		// Operator 4 submits digest B
		taskResult4, err := createSignedECDSATaskResult(taskId, operators[4], 1, payloadB, privateKeys[4])
		require.NoError(t, err)
		err = agg.ProcessNewSignature(context.Background(), taskResult4)
		require.NoError(t, err)

		// A should still be most common (both A and B have count 2, but A reached 2 first)
		assert.Equal(t, 2, agg.aggregatedOperators.mostCommonCount)
		assert.Equal(t, digestA, agg.aggregatedOperators.mostCommonDigest)

		// Verify digest counts
		assert.Equal(t, 2, agg.aggregatedOperators.digestGroups[digestA].count)
		assert.Equal(t, 2, agg.aggregatedOperators.digestGroups[digestB].count)
		assert.Equal(t, 1, agg.aggregatedOperators.digestGroups[digestC].count)

		// ECDSA threshold check is based on most common (not total participation)
		// Only 2/5 signed same message, need 3/5 for 60%
		assert.False(t, agg.SigningThresholdMet())

		// Generate final certificate and verify it uses the most common response
		cert, err := agg.GenerateFinalCertificate()
		require.NoError(t, err)
		require.NotNil(t, cert)

		// Certificate should use payload A (the most common)
		assert.Equal(t, payloadA, cert.TaskResponse)
		assert.Equal(t, digestA, cert.TaskResponseDigest)
	})

	t.Run("ECDSA - Unanimous Agreement", func(t *testing.T) {
		// Create test operators
		operators := make([]*Operator[common.Address], 3)
		privateKeys := make([]*cryptoLibsEcdsa.PrivateKey, 3)
		for i := 0; i < 3; i++ {
			privKey, _, err := cryptoLibsEcdsa.GenerateKeyPair()
			require.NoError(t, err)
			derivedAddress, err := privKey.DeriveAddress()
			require.NoError(t, err)

			operators[i] = &Operator[common.Address]{
				Address:       derivedAddress.String(),
				PublicKey:     derivedAddress,
				OperatorIndex: uint32(i),
			}
			privateKeys[i] = privKey
		}

		// Initialize task
		taskId := "0x29cebefe301c6ce1bb36b58654fea275e1cacc83"
		taskData := []byte("test-data")
		deadline := time.Now().Add(10 * time.Minute)

		agg, err := NewECDSATaskResultAggregator(
			context.Background(),
			taskId,
			1,
			10000, // 100% threshold (10000 bips)
			taskData,
			&deadline,
			operators,
		)
		require.NoError(t, err)

		// All operators submit the same response
		commonPayload := []byte("unanimous-response")
		digest := util.GetKeccak256Digest(commonPayload)

		for i := 0; i < 3; i++ {
			// Create properly signed task result using the helper function
			taskResult, err := createSignedECDSATaskResult(
				taskId,
				operators[i],
				1, // operatorSetId
				commonPayload,
				privateKeys[i],
			)
			require.NoError(t, err)

			err = agg.ProcessNewSignature(context.Background(), taskResult)
			require.NoError(t, err)

			// Most common count should increase with each submission
			assert.Equal(t, i+1, agg.aggregatedOperators.mostCommonCount)
			assert.Equal(t, digest, agg.aggregatedOperators.mostCommonDigest)
		}

		// Generate certificate
		cert, err := agg.GenerateFinalCertificate()
		require.NoError(t, err)
		assert.Equal(t, commonPayload, cert.TaskResponse)
		assert.Equal(t, digest, cert.TaskResponseDigest)
	})
}

func Test_OutputDigestSecurityValidation(t *testing.T) {
	t.Run("BN254 - Malicious OutputDigest Ignored", func(t *testing.T) {
		// Create test operator
		privKey, pubKey, err := bn254.GenerateKeyPair()
		require.NoError(t, err)

		operators := []*Operator[signing.PublicKey]{
			{
				Address:       "0x1",
				PublicKey:     pubKey,
				OperatorIndex: 0,
			},
		}

		// Initialize task
		taskId := "0x29cebefe301c6ce1bb36b58654fea275e1cacc83"
		taskData := []byte("test-data")
		deadline := time.Now().Add(10 * time.Minute)

		agg, err := NewBN254TaskResultAggregator(
			context.Background(),
			taskId,
			1,
			10000, // 100% threshold (10000 bips)
			taskData,
			&deadline,
			operators,
		)
		require.NoError(t, err)

		// Create legitimate output and sign it properly
		legitimateOutput := []byte("legitimate-response")

		// Create a properly signed TaskResult
		taskResult, err := createSignedBN254TaskResult(taskId, operators[0], 1, legitimateOutput, privKey)
		require.NoError(t, err)

		// Calculate the legitimate output digest (for consensus)
		legitimateOutputDigest := util.GetKeccak256Digest(legitimateOutput)

		// Process the signature
		err = agg.ProcessNewSignature(context.Background(), taskResult)
		require.NoError(t, err)

		// Verify that the aggregator stored the correct response
		legitimateDigest := util.GetKeccak256Digest(legitimateOutput)
		assert.NotNil(t, agg.aggregatedOperators.digestGroups[legitimateDigest])
		// Verify the output is correct
		assert.Equal(t, legitimateOutput, agg.aggregatedOperators.digestGroups[legitimateDigest].response.TaskResult.Output)

		// Generate certificate and verify it uses the calculated digest
		cert, err := agg.GenerateFinalCertificate()
		require.NoError(t, err)

		// Certificate should contain the legitimate output and its output digest
		assert.Equal(t, legitimateOutput, cert.TaskResponse)
		assert.Equal(t, legitimateOutputDigest, cert.TaskResponseDigest)

		// Verify the aggregated signature - all operators sign the same output digest
		signersPubKey, err := bn254.NewPublicKeyFromBytes(cert.SignersPublicKey.Marshal())
		require.NoError(t, err)
		verified, err := cert.SignersSignature.VerifySolidityCompatible(signersPubKey, legitimateOutputDigest)
		require.NoError(t, err)
		assert.True(t, verified, "Certificate signature should be valid")
	})

	t.Run("ECDSA - Malicious OutputDigest Ignored", func(t *testing.T) {
		// Create test operator
		privKey, _, err := cryptoLibsEcdsa.GenerateKeyPair()
		require.NoError(t, err)
		derivedAddress, err := privKey.DeriveAddress()
		require.NoError(t, err)

		operators := []*Operator[common.Address]{
			{
				Address:   derivedAddress.String(),
				PublicKey: derivedAddress,
			},
		}

		// Initialize task
		taskId := "0x29cebefe301c6ce1bb36b58654fea275e1cacc83"
		taskData := []byte("test-data")
		deadline := time.Now().Add(10 * time.Minute)

		agg, err := NewECDSATaskResultAggregator(
			context.Background(),
			taskId,
			1,
			10000, // 100% threshold (10000 bips)
			taskData,
			&deadline,
			operators,
		)
		require.NoError(t, err)

		// Create legitimate output and sign it properly
		legitimateOutput := []byte("legitimate-response")

		// Create a properly signed TaskResult
		taskResult, err := createSignedECDSATaskResult(taskId, operators[0], 1, legitimateOutput, privKey)
		require.NoError(t, err)

		// Calculate the legitimate output digest (for consensus)
		legitimateOutputDigest := util.GetKeccak256Digest(legitimateOutput)

		// Process the properly signed signature
		err = agg.ProcessNewSignature(context.Background(), taskResult)
		require.NoError(t, err)

		// Verify that the aggregator stored the correct response
		legitimateDigest := util.GetKeccak256Digest(legitimateOutput)
		assert.NotNil(t, agg.aggregatedOperators.digestGroups[legitimateDigest])
		// Verify the output is correct
		assert.Equal(t, legitimateOutput, agg.aggregatedOperators.digestGroups[legitimateDigest].response.TaskResult.Output)

		// Generate certificate and verify it uses the calculated digest
		cert, err := agg.GenerateFinalCertificate()
		require.NoError(t, err)

		// Certificate should contain the legitimate output and its calculated digest
		assert.Equal(t, legitimateOutput, cert.TaskResponse)
		assert.Equal(t, legitimateOutputDigest, cert.TaskResponseDigest)
	})
}

func Test_NonSignerOrdering(t *testing.T) {
	t.Run("BN254 - Non-Signers Sorted by OperatorIndex", func(t *testing.T) {
		// Create test operators with specific operator indices in non-sequential order
		// This tests that sorting is by OperatorIndex, not by address or order of creation
		operatorIndices := []uint32{4, 2, 0, 3, 1} // Deliberately non-sequential
		operators := make([]*Operator[signing.PublicKey], 5)
		privateKeys := make([]*bn254.PrivateKey, 5)

		for i := 0; i < 5; i++ {
			privKey, pubKey, err := bn254.GenerateKeyPair()
			require.NoError(t, err)
			operators[i] = &Operator[signing.PublicKey]{
				Address:       fmt.Sprintf("0x%d", i+1),
				PublicKey:     pubKey,
				OperatorIndex: operatorIndices[i],
			}
			privateKeys[i] = privKey
		}

		// Initialize task
		taskId := "0x29cebefe301c6ce1bb36b58654fea275e1cacc83"
		taskData := []byte("test-data")
		deadline := time.Now().Add(10 * time.Minute)

		agg, err := NewBN254TaskResultAggregator(
			context.Background(),
			taskId,
			1,
			6000, // 60% threshold (3/5 operators)
			taskData,
			&deadline,
			operators,
		)
		require.NoError(t, err)
		require.NotNil(t, agg)

		// Have operators at positions 0, 2, and 4 sign (with indices 4, 0, 1)
		// This means operators with indices 2 and 3 will be non-signers
		signingOperatorPositions := []int{0, 2, 4} // Positions in operators array
		expectedNonSignerIndices := []uint32{2, 3} // Expected non-signer operator indices (sorted)

		commonPayload := []byte("test-response")

		for _, pos := range signingOperatorPositions {
			operator := operators[pos]
			taskResult, err := createSignedBN254TaskResult(
				taskId,
				operator,
				1, // operatorSetId
				commonPayload,
				privateKeys[pos],
			)
			require.NoError(t, err)

			err = agg.ProcessNewSignature(context.Background(), taskResult)
			require.NoError(t, err)
		}

		// Verify threshold is met
		assert.True(t, agg.SigningThresholdMet())

		// Generate final certificate
		cert, err := agg.GenerateFinalCertificate()
		require.NoError(t, err)
		require.NotNil(t, cert)

		// Verify non-signers count
		assert.Equal(t, 2, len(cert.NonSignerOperators), "Should have 2 non-signers")
		assert.Equal(t, 2, len(cert.NonSignersPubKeys), "Should have 2 non-signer public keys")

		// Verify non-signers are sorted by OperatorIndex
		for i, nonSigner := range cert.NonSignerOperators {
			assert.Equal(t, expectedNonSignerIndices[i], nonSigner.OperatorIndex,
				"Non-signer at position %d should have OperatorIndex %d", i, expectedNonSignerIndices[i])
		}

		// Verify the actual non-signer operators are correct
		nonSignerAddresses := make(map[string]bool)
		for _, ns := range cert.NonSignerOperators {
			nonSignerAddresses[ns.Address] = true
		}

		// Operators at positions 1 and 3 (with indices 2 and 3) should be non-signers
		assert.True(t, nonSignerAddresses[operators[1].Address], "Operator at position 1 should be non-signer")
		assert.True(t, nonSignerAddresses[operators[3].Address], "Operator at position 3 should be non-signer")
	})

	t.Run("BN254 - All Operators Sign", func(t *testing.T) {
		// Create operators with random indices
		operatorIndices := []uint32{2, 0, 1}
		operators := make([]*Operator[signing.PublicKey], 3)
		privateKeys := make([]*bn254.PrivateKey, 3)

		for i := 0; i < 3; i++ {
			privKey, pubKey, err := bn254.GenerateKeyPair()
			require.NoError(t, err)
			operators[i] = &Operator[signing.PublicKey]{
				Address:       fmt.Sprintf("0x%d", i+1),
				PublicKey:     pubKey,
				OperatorIndex: operatorIndices[i],
			}
			privateKeys[i] = privKey
		}

		taskId := "0x29cebefe301c6ce1bb36b58654fea275e1cacc83"
		taskData := []byte("test-data")
		deadline := time.Now().Add(10 * time.Minute)

		agg, err := NewBN254TaskResultAggregator(
			context.Background(),
			taskId,
			1,
			10000, // 100% threshold
			taskData,
			&deadline,
			operators,
		)
		require.NoError(t, err)

		commonPayload := []byte("test-response")

		// All operators sign
		for i := 0; i < 3; i++ {
			taskResult, err := createSignedBN254TaskResult(
				taskId,
				operators[i],
				1,
				commonPayload,
				privateKeys[i],
			)
			require.NoError(t, err)
			err = agg.ProcessNewSignature(context.Background(), taskResult)
			require.NoError(t, err)
		}

		cert, err := agg.GenerateFinalCertificate()
		require.NoError(t, err)

		// Should have no non-signers
		assert.Equal(t, 0, len(cert.NonSignerOperators), "Should have no non-signers when all sign")
		assert.Equal(t, 0, len(cert.NonSignersPubKeys), "Should have no non-signer public keys")
	})

	t.Run("BN254 - No Operators Sign", func(t *testing.T) {
		// Create operators with indices that would need sorting
		operatorIndices := []uint32{3, 1, 2, 0}
		operators := make([]*Operator[signing.PublicKey], 4)

		for i := 0; i < 4; i++ {
			_, pubKey, err := bn254.GenerateKeyPair()
			require.NoError(t, err)
			operators[i] = &Operator[signing.PublicKey]{
				Address:       fmt.Sprintf("0x%d", i+1),
				PublicKey:     pubKey,
				OperatorIndex: operatorIndices[i],
			}
		}

		taskId := "0x29cebefe301c6ce1bb36b58654fea275e1cacc83"
		taskData := []byte("test-data")
		deadline := time.Now().Add(10 * time.Minute)

		agg, err := NewBN254TaskResultAggregator(
			context.Background(),
			taskId,
			1,
			2500, // 25% threshold (can be met with 0 signers if we want to test)
			taskData,
			&deadline,
			operators,
		)
		require.NoError(t, err)

		// Don't process any signatures, go straight to certificate generation
		// This would fail in practice due to no signatures, but we're testing the sorting
		// We need to initialize aggregatedOperators to avoid nil pointer
		dummyDigest := [32]byte{}
		agg.aggregatedOperators = &aggregatedBN254Operators{
			digestGroups: map[[32]byte]*digestGroup{
				dummyDigest: {
					signers: make(map[string]*signerInfo),
					response: &ReceivedBN254ResponseWithDigest{
						TaskResult:   &types.TaskResult{Output: []byte("dummy")},
						OutputDigest: dummyDigest,
					},
					count: 0,
				},
			},
			mostCommonDigest: dummyDigest,
			mostCommonCount:  0,
		}

		cert, err := agg.GenerateFinalCertificate()
		// Should fail because there are no signatures to aggregate
		require.Error(t, err, "Should fail to generate certificate with no signatures")
		assert.Contains(t, err.Error(), "no signatures for winning digest")
		assert.Nil(t, cert)

		// Can't verify sorting since certificate generation failed (which is expected)
	})

	t.Run("BN254 - Single Non-Signer", func(t *testing.T) {
		// Create operators where only one doesn't sign
		operators := make([]*Operator[signing.PublicKey], 3)
		privateKeys := make([]*bn254.PrivateKey, 3)

		for i := 0; i < 3; i++ {
			privKey, pubKey, err := bn254.GenerateKeyPair()
			require.NoError(t, err)
			operators[i] = &Operator[signing.PublicKey]{
				Address:       fmt.Sprintf("0x%d", i+1),
				PublicKey:     pubKey,
				OperatorIndex: uint32(i),
			}
			privateKeys[i] = privKey
		}

		taskId := "0x29cebefe301c6ce1bb36b58654fea275e1cacc83"
		taskData := []byte("test-data")
		deadline := time.Now().Add(10 * time.Minute)

		agg, err := NewBN254TaskResultAggregator(
			context.Background(),
			taskId,
			1,
			6667, // 66.67% threshold (2/3)
			taskData,
			&deadline,
			operators,
		)
		require.NoError(t, err)

		commonPayload := []byte("test-response")

		// Operators 0 and 2 sign, operator 1 doesn't
		for _, i := range []int{0, 2} {
			taskResult, err := createSignedBN254TaskResult(
				taskId,
				operators[i],
				1,
				commonPayload,
				privateKeys[i],
			)
			require.NoError(t, err)
			err = agg.ProcessNewSignature(context.Background(), taskResult)
			require.NoError(t, err)
		}

		cert, err := agg.GenerateFinalCertificate()
		require.NoError(t, err)

		// Should have exactly one non-signer (operator at index 1)
		assert.Equal(t, 1, len(cert.NonSignerOperators), "Should have exactly one non-signer")
		assert.Equal(t, uint32(1), cert.NonSignerOperators[0].OperatorIndex, "Non-signer should have OperatorIndex 1")
		assert.Equal(t, operators[1].Address, cert.NonSignerOperators[0].Address, "Non-signer should be operator 1")
	})
}

func Test_DigestBasedAggregation(t *testing.T) {
	t.Run("BN254 - Different Messages Only Aggregate Same Digest", func(t *testing.T) {
		// Create 3 test operators
		operators := make([]*Operator[signing.PublicKey], 3)
		privateKeys := make([]*bn254.PrivateKey, 3)

		for i := 0; i < 3; i++ {
			privKey, pubKey, err := bn254.GenerateKeyPair()
			require.NoError(t, err)
			operators[i] = &Operator[signing.PublicKey]{
				Address:       fmt.Sprintf("0x%d", i+1),
				PublicKey:     pubKey,
				OperatorIndex: uint32(i),
			}
			privateKeys[i] = privKey
		}

		// Initialize task
		taskId := "0xabc123def456789012345678901234567890abcd"
		taskData := []byte("test-task-data")
		deadline := time.Now().Add(10 * time.Minute)

		// Create aggregator with 66.67% threshold (2/3 operators needed)
		agg, err := NewBN254TaskResultAggregator(
			context.Background(),
			taskId,
			1,    // operatorSetId
			6667, // thresholdBips (2/3 = 66.67%)
			taskData,
			&deadline,
			operators,
		)
		require.NoError(t, err)
		require.NotNil(t, agg)

		// Define two different messages
		messageA := []byte("message-A-consensus")
		messageB := []byte("message-B-different")

		// Operators 0 and 1 sign message A
		for i := 0; i < 2; i++ {
			taskResult, err := createSignedBN254TaskResult(
				taskId,
				operators[i],
				1, // operatorSetId
				messageA,
				privateKeys[i],
			)
			require.NoError(t, err)

			err = agg.ProcessNewSignature(context.Background(), taskResult)
			require.NoError(t, err)
		}

		// Operator 2 signs message B (different message)
		taskResultB, err := createSignedBN254TaskResult(
			taskId,
			operators[2],
			1, // operatorSetId
			messageB,
			privateKeys[2],
		)
		require.NoError(t, err)

		err = agg.ProcessNewSignature(context.Background(), taskResultB)
		require.NoError(t, err)

		// Verify that the threshold is met (2 out of 3 operators signed the same message)
		assert.True(t, agg.SigningThresholdMet(), "Threshold should be met with 2/3 operators signing same message")

		// Generate the certificate
		cert, err := agg.GenerateFinalCertificate()
		require.NoError(t, err)
		require.NotNil(t, cert)

		// Critical verification: Certificate should only aggregate signatures for message A
		assert.Equal(t, messageA, cert.TaskResponse, "Certificate should contain message A (the majority message)")

		// Verify the aggregated signature only includes operators who signed message A
		digestA := util.GetKeccak256Digest(messageA)
		assert.Equal(t, digestA, cert.TaskResponseDigest, "Certificate should use digest of message A")

		// Verify operator 2 (who signed message B) is in the non-signers list
		assert.Equal(t, 1, len(cert.NonSignerOperators), "Should have 1 non-signer")
		assert.Equal(t, operators[2].Address, cert.NonSignerOperators[0].Address,
			"Operator 2 who signed different message should be treated as non-signer")
		assert.Equal(t, uint32(2), cert.NonSignerOperators[0].OperatorIndex,
			"Non-signer should be operator with index 2")

		// Verify the aggregated signature is valid for message A only
		signersPubKey, err := bn254.NewPublicKeyFromBytes(cert.SignersPublicKey.Marshal())
		require.NoError(t, err)

		// The aggregated signature should verify against digest A (not B)
		verified, err := cert.SignersSignature.VerifySolidityCompatible(signersPubKey, digestA)
		require.NoError(t, err)
		assert.True(t, verified, "Aggregated signature should be valid for message A")

		// Additional verification: check digest groups to ensure proper segregation
		digestB := util.GetKeccak256Digest(messageB)
		assert.NotNil(t, agg.aggregatedOperators.digestGroups[digestA], "Digest A group should exist")
		assert.NotNil(t, agg.aggregatedOperators.digestGroups[digestB], "Digest B group should exist")
		assert.Equal(t, 2, agg.aggregatedOperators.digestGroups[digestA].count,
			"Digest A should have 2 operators")
		assert.Equal(t, 1, agg.aggregatedOperators.digestGroups[digestB].count,
			"Digest B should have 1 operator")

		// The most common digest should be A
		assert.Equal(t, digestA, agg.aggregatedOperators.mostCommonDigest,
			"Most common digest should be A")
		assert.Equal(t, 2, agg.aggregatedOperators.mostCommonCount,
			"Most common count should be 2")
	})
}

func Test_TaskIDMismatchValidation(t *testing.T) {
	t.Run("BN254 - Task ID Mismatch", func(t *testing.T) {
		// Create test operators with key pairs
		operators := make([]*Operator[signing.PublicKey], 2)
		privateKeys := make([]*bn254.PrivateKey, 2)
		for i := 0; i < 2; i++ {
			privKey, pubKey, err := bn254.GenerateKeyPair()
			require.NoError(t, err)
			operators[i] = &Operator[signing.PublicKey]{
				Address:       fmt.Sprintf("0x%d", i+1),
				PublicKey:     pubKey,
				OperatorIndex: uint32(i),
			}
			privateKeys[i] = privKey
		}

		// Initialize task with specific ID
		taskId := "0x29cebefe301c6ce1bb36b58654fea275e1cacc83"
		taskData := []byte("test-data")
		deadline := time.Now().Add(10 * time.Minute)

		agg, err := NewBN254TaskResultAggregator(
			context.Background(),
			taskId,
			1,
			5000, // 50% threshold (5000 bips)
			taskData,
			&deadline,
			operators,
		)
		require.NoError(t, err)
		require.NotNil(t, agg)

		// Create a task result with mismatched task ID
		mismatchedTaskId := "0xdifferenttaskid1234567890abcdef12345678"
		payload := []byte("test-response")

		// Create a TaskResult with mismatched task ID - note we sign with the wrong ID
		taskResult, err := createSignedBN254TaskResult(mismatchedTaskId, operators[0], 1, payload, privateKeys[0])
		require.NoError(t, err)

		// Process the signature with mismatched task ID
		err = agg.ProcessNewSignature(context.Background(), taskResult)

		// Should return an error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "task ID mismatch")
		assert.Contains(t, err.Error(), taskId)
		assert.Contains(t, err.Error(), mismatchedTaskId)

		// Verify that no signature was recorded
		assert.Equal(t, 0, len(agg.ReceivedSignatures))
		assert.False(t, agg.SigningThresholdMet())

		// Now submit with correct task ID to verify aggregator works properly
		correctTaskResult, err := createSignedBN254TaskResult(taskId, operators[0], 1, payload, privateKeys[0])
		require.NoError(t, err)

		err = agg.ProcessNewSignature(context.Background(), correctTaskResult)
		require.NoError(t, err)

		// Verify signature was recorded
		assert.Equal(t, 1, len(agg.ReceivedSignatures))
		assert.True(t, agg.SigningThresholdMet()) // 1/2 = 50% threshold met
	})

	t.Run("ECDSA - Task ID Mismatch", func(t *testing.T) {
		// Create test operators with key pairs
		operators := make([]*Operator[common.Address], 2)
		privateKeys := make([]*cryptoLibsEcdsa.PrivateKey, 2)
		for i := 0; i < 2; i++ {
			privKey, _, err := cryptoLibsEcdsa.GenerateKeyPair()
			require.NoError(t, err)
			derivedAddress, err := privKey.DeriveAddress()
			require.NoError(t, err)

			operators[i] = &Operator[common.Address]{
				Address:       derivedAddress.String(),
				PublicKey:     derivedAddress,
				OperatorIndex: uint32(i),
			}
			privateKeys[i] = privKey
		}

		// Initialize task with specific ID
		taskId := "0x29cebefe301c6ce1bb36b58654fea275e1cacc83"
		taskData := []byte("test-data")
		deadline := time.Now().Add(10 * time.Minute)

		agg, err := NewECDSATaskResultAggregator(
			context.Background(),
			taskId,
			1,
			5000, // 50% threshold (5000 bips)
			taskData,
			&deadline,
			operators,
		)
		require.NoError(t, err)
		require.NotNil(t, agg)

		// Create a task result with mismatched task ID
		mismatchedTaskId := "0xdifferenttaskid1234567890abcdef12345678"
		payload := []byte("test-response")

		// Create a TaskResult with mismatched task ID - note we sign with the wrong ID
		taskResult, err := createSignedECDSATaskResult(mismatchedTaskId, operators[0], 1, payload, privateKeys[0])
		require.NoError(t, err)

		// Process the signature with mismatched task ID
		err = agg.ProcessNewSignature(context.Background(), taskResult)

		// Should return an error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "task ID mismatch")
		assert.Contains(t, err.Error(), taskId)
		assert.Contains(t, err.Error(), mismatchedTaskId)

		// Verify that no signature was recorded
		assert.Equal(t, 0, len(agg.ReceivedSignatures))
		assert.False(t, agg.SigningThresholdMet())

		// Now submit with correct task ID to verify aggregator works properly
		correctTaskResult, err := createSignedECDSATaskResult(taskId, operators[0], 1, payload, privateKeys[0])
		require.NoError(t, err)

		err = agg.ProcessNewSignature(context.Background(), correctTaskResult)
		require.NoError(t, err)

		// Verify signature was recorded
		assert.Equal(t, 1, len(agg.ReceivedSignatures))
		assert.True(t, agg.SigningThresholdMet()) // 1/2 = 50% threshold met
	})
}

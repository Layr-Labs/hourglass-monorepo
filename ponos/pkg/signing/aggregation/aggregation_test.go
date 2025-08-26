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
			100,  // taskCreatedBlock
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
			100,  // taskCreatedBlock
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
		assert.Equal(t, 1, len(cert.NonSignersPubKeys), "Should have one non-signer")
		assert.Equal(t, 4, len(cert.AllOperatorsPubKeys), "Should have all operators' public keys")

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
			100,  // taskCreatedBlock
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
		assert.Equal(t, digestA, agg.aggregatedOperators.mostCommonResponse.OutputDigest)

		// Operator 1 submits digest B
		taskResult1, err := createSignedBN254TaskResult(taskId, operators[1], 1, payloadB, privateKeys[1])
		require.NoError(t, err)
		err = agg.ProcessNewSignature(context.Background(), taskResult1)
		require.NoError(t, err)

		// Most common should still be A (both have count 1, A came first)
		assert.Equal(t, 1, agg.aggregatedOperators.mostCommonCount)
		assert.Equal(t, digestA, agg.aggregatedOperators.mostCommonResponse.OutputDigest)

		// Operator 2 submits digest A
		taskResult2, err := createSignedBN254TaskResult(taskId, operators[2], 1, payloadA, privateKeys[2])
		require.NoError(t, err)
		err = agg.ProcessNewSignature(context.Background(), taskResult2)
		require.NoError(t, err)

		// Now A should have count 2 and be most common
		assert.Equal(t, 2, agg.aggregatedOperators.mostCommonCount)
		assert.Equal(t, digestA, agg.aggregatedOperators.mostCommonResponse.OutputDigest)
		assert.Equal(t, payloadA, agg.aggregatedOperators.mostCommonResponse.TaskResult.Output)

		// Operator 3 submits digest C
		taskResult3, err := createSignedBN254TaskResult(taskId, operators[3], 1, payloadC, privateKeys[3])
		require.NoError(t, err)
		err = agg.ProcessNewSignature(context.Background(), taskResult3)
		require.NoError(t, err)

		// A should still be most common with count 2
		assert.Equal(t, 2, agg.aggregatedOperators.mostCommonCount)
		assert.Equal(t, digestA, agg.aggregatedOperators.mostCommonResponse.OutputDigest)

		// Operator 4 submits digest B
		taskResult4, err := createSignedBN254TaskResult(taskId, operators[4], 1, payloadB, privateKeys[4])
		require.NoError(t, err)
		err = agg.ProcessNewSignature(context.Background(), taskResult4)
		require.NoError(t, err)

		// A should still be most common (both A and B have count 2, but A reached 2 first)
		assert.Equal(t, 2, agg.aggregatedOperators.mostCommonCount)
		assert.Equal(t, digestA, agg.aggregatedOperators.mostCommonResponse.OutputDigest)

		// Verify digest counts
		var digestArrayA, digestArrayB, digestArrayC [32]byte
		copy(digestArrayA[:], digestA[:])
		copy(digestArrayB[:], digestB[:])
		copy(digestArrayC[:], digestC[:])
		assert.Equal(t, 2, agg.aggregatedOperators.digestCounts[digestArrayA])
		assert.Equal(t, 2, agg.aggregatedOperators.digestCounts[digestArrayB])
		assert.Equal(t, 1, agg.aggregatedOperators.digestCounts[digestArrayC])

		// Verify threshold is met
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
			100,
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

		// Verify mostCommonResponse is set correctly
		assert.NotNil(t, agg.aggregatedOperators.mostCommonResponse)
		assert.Equal(t, 1, agg.aggregatedOperators.mostCommonCount)
		assert.Equal(t, digest, agg.aggregatedOperators.mostCommonResponse.OutputDigest)
		assert.Equal(t, payload, agg.aggregatedOperators.mostCommonResponse.TaskResult.Output)

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
			100,
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
			assert.Equal(t, digest, agg.aggregatedOperators.mostCommonResponse.OutputDigest)
		}

		// Generate certificate
		cert, err := agg.GenerateFinalCertificate()
		require.NoError(t, err)
		assert.Equal(t, commonPayload, cert.TaskResponse)
		assert.Equal(t, digest, cert.TaskResponseDigest)
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
			100,  // taskCreatedBlock
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
		assert.Equal(t, digestA, agg.aggregatedOperators.mostCommonResponse.OutputDigest)

		// Operator 1 submits digest B
		taskResult1, err := createSignedECDSATaskResult(taskId, operators[1], 1, payloadB, privateKeys[1])
		require.NoError(t, err)
		err = agg.ProcessNewSignature(context.Background(), taskResult1)
		require.NoError(t, err)

		// Most common should still be A (both have count 1, A came first)
		assert.Equal(t, 1, agg.aggregatedOperators.mostCommonCount)
		assert.Equal(t, digestA, agg.aggregatedOperators.mostCommonResponse.OutputDigest)

		// Operator 2 submits digest A
		taskResult2, err := createSignedECDSATaskResult(taskId, operators[2], 1, payloadA, privateKeys[2])
		require.NoError(t, err)
		err = agg.ProcessNewSignature(context.Background(), taskResult2)
		require.NoError(t, err)

		// Now A should have count 2 and be most common
		assert.Equal(t, 2, agg.aggregatedOperators.mostCommonCount)
		assert.Equal(t, digestA, agg.aggregatedOperators.mostCommonResponse.OutputDigest)
		assert.Equal(t, payloadA, agg.aggregatedOperators.mostCommonResponse.TaskResult.Output)

		// Operator 3 submits digest C
		taskResult3, err := createSignedECDSATaskResult(taskId, operators[3], 1, payloadC, privateKeys[3])
		require.NoError(t, err)
		err = agg.ProcessNewSignature(context.Background(), taskResult3)
		require.NoError(t, err)

		// A should still be most common with count 2
		assert.Equal(t, 2, agg.aggregatedOperators.mostCommonCount)
		assert.Equal(t, digestA, agg.aggregatedOperators.mostCommonResponse.OutputDigest)

		// Operator 4 submits digest B
		taskResult4, err := createSignedECDSATaskResult(taskId, operators[4], 1, payloadB, privateKeys[4])
		require.NoError(t, err)
		err = agg.ProcessNewSignature(context.Background(), taskResult4)
		require.NoError(t, err)

		// A should still be most common (both A and B have count 2, but A reached 2 first)
		assert.Equal(t, 2, agg.aggregatedOperators.mostCommonCount)
		assert.Equal(t, digestA, agg.aggregatedOperators.mostCommonResponse.OutputDigest)

		// Verify digest counts
		assert.Equal(t, 2, agg.aggregatedOperators.digestCounts[digestA])
		assert.Equal(t, 2, agg.aggregatedOperators.digestCounts[digestB])
		assert.Equal(t, 1, agg.aggregatedOperators.digestCounts[digestC])

		// Verify threshold is met
		assert.True(t, agg.SigningThresholdMet())

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
			100,
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
			assert.Equal(t, digest, agg.aggregatedOperators.mostCommonResponse.OutputDigest)
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
			100,
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
		assert.NotNil(t, agg.aggregatedOperators.mostCommonResponse)
		// Verify the output is correct
		assert.Equal(t, legitimateOutput, agg.aggregatedOperators.mostCommonResponse.TaskResult.Output)

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
			100,
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
		assert.NotNil(t, agg.aggregatedOperators.mostCommonResponse)
		// Verify the output is correct
		assert.Equal(t, legitimateOutput, agg.aggregatedOperators.mostCommonResponse.TaskResult.Output)

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
			100,
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
		expectedNonSignerIndices := []uint32{2, 3}  // Expected non-signer operator indices (sorted)
		
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
			100,
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
			100,
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
		agg.aggregatedOperators = &aggregatedBN254Operators{
			signersOperatorSet: make(map[string]bool),
			digestCounts:       make(map[[32]byte]int),
			digestResponses:    make(map[[32]byte]*ReceivedBN254ResponseWithDigest),
			signersG2:          bn254.NewZeroG2Point(),
			signersAggSig:      &bn254.Signature{},
			mostCommonResponse: &ReceivedBN254ResponseWithDigest{
				TaskResult:   &types.TaskResult{Output: []byte("dummy")},
				OutputDigest: [32]byte{},
			},
		}

		cert, err := agg.GenerateFinalCertificate()
		require.NoError(t, err)
		
		// All operators should be non-signers, sorted by OperatorIndex
		assert.Equal(t, 4, len(cert.NonSignerOperators), "Should have all operators as non-signers")
		
		// Verify they're sorted by OperatorIndex (should be 0, 1, 2, 3)
		for i := 0; i < 4; i++ {
			assert.Equal(t, uint32(i), cert.NonSignerOperators[i].OperatorIndex,
				"Non-signer at position %d should have OperatorIndex %d", i, i)
		}
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
			100,
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
			100,
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
			100,
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

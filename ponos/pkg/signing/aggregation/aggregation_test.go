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

func Test_Aggregation(t *testing.T) {
	t.Run("BN254", func(t *testing.T) {
		// Create test operators with key pairs
		operators := make([]*Operator[signing.PublicKey], 4) // Changed to 4 operators

		privateKeys := make([]*bn254.PrivateKey, 4)
		for i := 0; i < 4; i++ {
			privKey, pubKey, err := bn254.GenerateKeyPair()
			require.NoError(t, err)
			operators[i] = &Operator[signing.PublicKey]{
				Address:   fmt.Sprintf("0x%d", i+1), // Simple address format for testing
				PublicKey: pubKey,
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
		remainingPubKeys := make([]signing.PublicKey, 3)
		remainingSigs := make([]*bn254.Signature, 3)

		// Simulate receiving responses from all operators except the last one
		for i := 0; i < 3; i++ { // Only process first 3 operators
			operator := operators[i]
			// Create task result
			taskResult := &types.TaskResult{
				TaskId:          taskId,
				OperatorAddress: operator.Address,
				Output:          commonPayload,
				OutputDigest:    digest[:],
			}

			// Sign what the implementation will calculate: keccak256(Output)
			expectedDigest := util.GetKeccak256Digest(commonPayload)
			sig, err := privateKeys[i].SignSolidityCompatible(expectedDigest)
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

		// Test: Verify if an operator's signature was included
		// We can verify this by checking that the remaining signatures verify correctly
		remainingBn254PubKeys := util.Map(remainingPubKeys, func(pk signing.PublicKey, i uint64) *bn254.PublicKey {
			return pk.(*bn254.PublicKey)
		})
		verified, err = bn254.BatchVerifySolidityCompatible(remainingBn254PubKeys, responseDigest, remainingSigs)
		require.NoError(t, err)
		assert.True(t, verified, "Remaining signatures should verify correctly")

		// Test: Verify that the non-signer's signature is not included
		// Create a new signature array including the non-signer's signature
		allSigs := append(remainingSigs, individualSigs[0])            // Add a duplicate signature
		allPubKeys := append(remainingPubKeys, operators[3].PublicKey) // Add non-signer's public key
		allBn254PubKeys := util.Map(allPubKeys, func(pk signing.PublicKey, i uint64) *bn254.PublicKey {
			return pk.(*bn254.PublicKey)
		})
		verified, err = bn254.BatchVerifySolidityCompatible(allBn254PubKeys, responseDigest, allSigs)
		require.NoError(t, err)
		assert.False(t, verified, "Verification should fail when including non-signer's public key")
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
				Address:   derivedAddress.String(),
				PublicKey: derivedAddress,
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
		individualSigs := make([][]byte, 3) // Only store 3 signatures since one operator won't sign
		remainingPubKeys := make([]common.Address, 3)
		remainingSigs := make([][]byte, 3)

		// Simulate receiving responses from all operators except the last one
		for i := 0; i < 3; i++ { // Only process first 3 operators
			operator := operators[i]
			// Create task result
			taskResult := &types.TaskResult{
				TaskId:          taskId,
				OperatorAddress: operator.Address,
				Output:          commonPayload,
				OutputDigest:    digest[:],
			}

			// Sign what the implementation will calculate: keccak256(Output)
			expectedDigest := util.GetKeccak256Digest(commonPayload)
			sig, err := privateKeys[i].Sign(expectedDigest[:])
			require.NoError(t, err)
			taskResult.Signature = sig.Bytes()

			individualSigs[i] = sig.Bytes()
			remainingPubKeys[i] = operator.PublicKey
			remainingSigs[i] = sig.Bytes()

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
				Address:   fmt.Sprintf("0x%d", i+1),
				PublicKey: pubKey,
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
			100, // taskCreatedBlock
			1,   // operatorSetId
			60,  // thresholdPercentage (3/5)
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
		taskResult0 := &types.TaskResult{
			TaskId:          taskId,
			OperatorAddress: operators[0].Address,
			Output:          payloadA,
			OutputDigest:    digestA[:],
		}
		sig0, err := privateKeys[0].SignSolidityCompatible(digestA)
		require.NoError(t, err)
		taskResult0.Signature = sig0.Bytes()
		err = agg.ProcessNewSignature(context.Background(), taskId, taskResult0)
		require.NoError(t, err)

		// Verify mostCommonCount is 1 and points to digest A
		assert.Equal(t, 1, agg.aggregatedOperators.mostCommonCount)
		assert.Equal(t, digestA, agg.aggregatedOperators.mostCommonResponse.Digest)

		// Operator 1 submits digest B
		taskResult1 := &types.TaskResult{
			TaskId:          taskId,
			OperatorAddress: operators[1].Address,
			Output:          payloadB,
			OutputDigest:    digestB[:],
		}
		sig1, err := privateKeys[1].SignSolidityCompatible(digestB)
		require.NoError(t, err)
		taskResult1.Signature = sig1.Bytes()
		err = agg.ProcessNewSignature(context.Background(), taskId, taskResult1)
		require.NoError(t, err)

		// Most common should still be A (both have count 1, A came first)
		assert.Equal(t, 1, agg.aggregatedOperators.mostCommonCount)
		assert.Equal(t, digestA, agg.aggregatedOperators.mostCommonResponse.Digest)

		// Operator 2 submits digest A
		taskResult2 := &types.TaskResult{
			TaskId:          taskId,
			OperatorAddress: operators[2].Address,
			Output:          payloadA,
			OutputDigest:    digestA[:],
		}
		sig2, err := privateKeys[2].SignSolidityCompatible(digestA)
		require.NoError(t, err)
		taskResult2.Signature = sig2.Bytes()
		err = agg.ProcessNewSignature(context.Background(), taskId, taskResult2)
		require.NoError(t, err)

		// Now A should have count 2 and be most common
		assert.Equal(t, 2, agg.aggregatedOperators.mostCommonCount)
		assert.Equal(t, digestA, agg.aggregatedOperators.mostCommonResponse.Digest)
		assert.Equal(t, payloadA, agg.aggregatedOperators.mostCommonResponse.TaskResult.Output)

		// Operator 3 submits digest C
		taskResult3 := &types.TaskResult{
			TaskId:          taskId,
			OperatorAddress: operators[3].Address,
			Output:          payloadC,
			OutputDigest:    digestC[:],
		}
		sig3, err := privateKeys[3].SignSolidityCompatible(digestC)
		require.NoError(t, err)
		taskResult3.Signature = sig3.Bytes()
		err = agg.ProcessNewSignature(context.Background(), taskId, taskResult3)
		require.NoError(t, err)

		// A should still be most common with count 2
		assert.Equal(t, 2, agg.aggregatedOperators.mostCommonCount)
		assert.Equal(t, digestA, agg.aggregatedOperators.mostCommonResponse.Digest)

		// Operator 4 submits digest B
		taskResult4 := &types.TaskResult{
			TaskId:          taskId,
			OperatorAddress: operators[4].Address,
			Output:          payloadB,
			OutputDigest:    digestB[:],
		}
		sig4, err := privateKeys[4].SignSolidityCompatible(digestB)
		require.NoError(t, err)
		taskResult4.Signature = sig4.Bytes()
		err = agg.ProcessNewSignature(context.Background(), taskId, taskResult4)
		require.NoError(t, err)

		// A should still be most common (both A and B have count 2, but A reached 2 first)
		assert.Equal(t, 2, agg.aggregatedOperators.mostCommonCount)
		assert.Equal(t, digestA, agg.aggregatedOperators.mostCommonResponse.Digest)

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
				Address:   "0x1",
				PublicKey: pubKey,
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
			100, // 100% threshold - requires the single operator
			taskData,
			&deadline,
			operators,
		)
		require.NoError(t, err)

		// Single operator submits response
		payload := []byte("single-operator-response")
		digest := util.GetKeccak256Digest(payload)

		taskResult := &types.TaskResult{
			TaskId:          taskId,
			OperatorAddress: operators[0].Address,
			Output:          payload,
			OutputDigest:    digest[:],
		}
		sig, err := privKey.SignSolidityCompatible(digest)
		require.NoError(t, err)
		taskResult.Signature = sig.Bytes()

		// Process the single signature
		err = agg.ProcessNewSignature(context.Background(), taskId, taskResult)
		require.NoError(t, err)

		// Verify mostCommonResponse is set correctly
		assert.NotNil(t, agg.aggregatedOperators.mostCommonResponse)
		assert.Equal(t, 1, agg.aggregatedOperators.mostCommonCount)
		assert.Equal(t, digest, agg.aggregatedOperators.mostCommonResponse.Digest)
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
				Address:   fmt.Sprintf("0x%d", i+1),
				PublicKey: pubKey,
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
			100, // 100% threshold
			taskData,
			&deadline,
			operators,
		)
		require.NoError(t, err)

		// All operators submit the same response
		commonPayload := []byte("unanimous-response")
		digest := util.GetKeccak256Digest(commonPayload)

		for i := 0; i < 3; i++ {
			taskResult := &types.TaskResult{
				TaskId:          taskId,
				OperatorAddress: operators[i].Address,
				Output:          commonPayload,
				OutputDigest:    digest[:],
			}
			sig, err := privateKeys[i].SignSolidityCompatible(digest)
			require.NoError(t, err)
			taskResult.Signature = sig.Bytes()

			err = agg.ProcessNewSignature(context.Background(), taskId, taskResult)
			require.NoError(t, err)

			// Most common count should increase with each submission
			assert.Equal(t, i+1, agg.aggregatedOperators.mostCommonCount)
			assert.Equal(t, digest, agg.aggregatedOperators.mostCommonResponse.Digest)
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
				Address:   derivedAddress.String(),
				PublicKey: derivedAddress,
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
			100, // taskCreatedBlock
			1,   // operatorSetId
			60,  // thresholdPercentage (3/5)
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
		taskResult0 := &types.TaskResult{
			TaskId:          taskId,
			OperatorAddress: operators[0].Address,
			Output:          payloadA,
			OutputDigest:    digestA[:],
		}
		sig0, err := privateKeys[0].Sign(digestA[:])
		require.NoError(t, err)
		taskResult0.Signature = sig0.Bytes()
		err = agg.ProcessNewSignature(context.Background(), taskId, taskResult0)
		require.NoError(t, err)

		// Verify mostCommonCount is 1 and points to digest A
		assert.Equal(t, 1, agg.aggregatedOperators.mostCommonCount)
		assert.Equal(t, digestA, agg.aggregatedOperators.mostCommonResponse.Digest)

		// Operator 1 submits digest B
		taskResult1 := &types.TaskResult{
			TaskId:          taskId,
			OperatorAddress: operators[1].Address,
			Output:          payloadB,
			OutputDigest:    digestB[:],
		}
		sig1, err := privateKeys[1].Sign(digestB[:])
		require.NoError(t, err)
		taskResult1.Signature = sig1.Bytes()
		err = agg.ProcessNewSignature(context.Background(), taskId, taskResult1)
		require.NoError(t, err)

		// Most common should still be A (both have count 1, A came first)
		assert.Equal(t, 1, agg.aggregatedOperators.mostCommonCount)
		assert.Equal(t, digestA, agg.aggregatedOperators.mostCommonResponse.Digest)

		// Operator 2 submits digest A
		taskResult2 := &types.TaskResult{
			TaskId:          taskId,
			OperatorAddress: operators[2].Address,
			Output:          payloadA,
			OutputDigest:    digestA[:],
		}
		sig2, err := privateKeys[2].Sign(digestA[:])
		require.NoError(t, err)
		taskResult2.Signature = sig2.Bytes()
		err = agg.ProcessNewSignature(context.Background(), taskId, taskResult2)
		require.NoError(t, err)

		// Now A should have count 2 and be most common
		assert.Equal(t, 2, agg.aggregatedOperators.mostCommonCount)
		assert.Equal(t, digestA, agg.aggregatedOperators.mostCommonResponse.Digest)
		assert.Equal(t, payloadA, agg.aggregatedOperators.mostCommonResponse.TaskResult.Output)

		// Operator 3 submits digest C
		taskResult3 := &types.TaskResult{
			TaskId:          taskId,
			OperatorAddress: operators[3].Address,
			Output:          payloadC,
			OutputDigest:    digestC[:],
		}
		sig3, err := privateKeys[3].Sign(digestC[:])
		require.NoError(t, err)
		taskResult3.Signature = sig3.Bytes()
		err = agg.ProcessNewSignature(context.Background(), taskId, taskResult3)
		require.NoError(t, err)

		// A should still be most common with count 2
		assert.Equal(t, 2, agg.aggregatedOperators.mostCommonCount)
		assert.Equal(t, digestA, agg.aggregatedOperators.mostCommonResponse.Digest)

		// Operator 4 submits digest B
		taskResult4 := &types.TaskResult{
			TaskId:          taskId,
			OperatorAddress: operators[4].Address,
			Output:          payloadB,
			OutputDigest:    digestB[:],
		}
		sig4, err := privateKeys[4].Sign(digestB[:])
		require.NoError(t, err)
		taskResult4.Signature = sig4.Bytes()
		err = agg.ProcessNewSignature(context.Background(), taskId, taskResult4)
		require.NoError(t, err)

		// A should still be most common (both A and B have count 2, but A reached 2 first)
		assert.Equal(t, 2, agg.aggregatedOperators.mostCommonCount)
		assert.Equal(t, digestA, agg.aggregatedOperators.mostCommonResponse.Digest)

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
				Address:   derivedAddress.String(),
				PublicKey: derivedAddress,
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
			100, // 100% threshold
			taskData,
			&deadline,
			operators,
		)
		require.NoError(t, err)

		// All operators submit the same response
		commonPayload := []byte("unanimous-response")
		digest := util.GetKeccak256Digest(commonPayload)

		for i := 0; i < 3; i++ {
			taskResult := &types.TaskResult{
				TaskId:          taskId,
				OperatorAddress: operators[i].Address,
				Output:          commonPayload,
				OutputDigest:    digest[:],
			}
			sig, err := privateKeys[i].Sign(digest[:])
			require.NoError(t, err)
			taskResult.Signature = sig.Bytes()

			err = agg.ProcessNewSignature(context.Background(), taskId, taskResult)
			require.NoError(t, err)

			// Most common count should increase with each submission
			assert.Equal(t, i+1, agg.aggregatedOperators.mostCommonCount)
			assert.Equal(t, digest, agg.aggregatedOperators.mostCommonResponse.Digest)
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
				Address:   "0x1",
				PublicKey: pubKey,
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
			100, // 100% threshold
			taskData,
			&deadline,
			operators,
		)
		require.NoError(t, err)

		// Create legitimate output and sign it properly
		legitimateOutput := []byte("legitimate-response")
		legitimateDigest := util.GetKeccak256Digest(legitimateOutput)
		legitimateSignature, err := privKey.SignSolidityCompatible(legitimateDigest)
		require.NoError(t, err)

		// Create malicious digest (different from legitimate output)
		maliciousOutput := []byte("malicious-response")
		maliciousDigest := util.GetKeccak256Digest(maliciousOutput)

		// Create task result with legitimate signature but malicious output digest
		// This simulates the attack where OutputDigest doesn't match Output
		taskResult := &types.TaskResult{
			TaskId:          taskId,
			OperatorAddress: operators[0].Address,
			Output:          legitimateOutput,            // Legitimate output
			OutputDigest:    maliciousDigest[:],          // Malicious digest (doesn't match output)
			Signature:       legitimateSignature.Bytes(), // Valid signature for legitimate digest
		}

		// Process the signature - our implementation should calculate digest from Output
		// and ignore the malicious OutputDigest, so this should succeed
		err = agg.ProcessNewSignature(context.Background(), taskId, taskResult)
		require.NoError(t, err)

		// Verify that the aggregator used the calculated digest, not the malicious one
		assert.NotNil(t, agg.aggregatedOperators.mostCommonResponse)
		assert.Equal(t, legitimateDigest, agg.aggregatedOperators.mostCommonResponse.Digest)

		// Generate certificate and verify it uses the calculated digest
		cert, err := agg.GenerateFinalCertificate()
		require.NoError(t, err)

		// Certificate should contain the legitimate output and its calculated digest
		assert.Equal(t, legitimateOutput, cert.TaskResponse)
		assert.Equal(t, legitimateDigest, cert.TaskResponseDigest)

		// Verify the signature in the certificate is valid
		signersPubKey, err := bn254.NewPublicKeyFromBytes(cert.SignersPublicKey.Marshal())
		require.NoError(t, err)
		verified, err := cert.SignersSignature.VerifySolidityCompatible(signersPubKey, legitimateDigest)
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
			100, // 100% threshold
			taskData,
			&deadline,
			operators,
		)
		require.NoError(t, err)

		// Create legitimate output and sign it properly
		legitimateOutput := []byte("legitimate-response")
		legitimateDigest := util.GetKeccak256Digest(legitimateOutput)
		legitimateSignature, err := privKey.Sign(legitimateDigest[:])
		require.NoError(t, err)

		// Create malicious digest (different from legitimate output)
		maliciousOutput := []byte("malicious-response")
		maliciousDigest := util.GetKeccak256Digest(maliciousOutput)

		// Create task result with legitimate signature but malicious output digest
		taskResult := &types.TaskResult{
			TaskId:          taskId,
			OperatorAddress: operators[0].Address,
			Output:          legitimateOutput,            // Legitimate output
			OutputDigest:    maliciousDigest[:],          // Malicious digest (doesn't match output)
			Signature:       legitimateSignature.Bytes(), // Valid signature for legitimate digest
		}

		// Process the signature - our implementation should calculate digest from Output
		// and ignore the malicious OutputDigest, so this should succeed
		err = agg.ProcessNewSignature(context.Background(), taskId, taskResult)
		require.NoError(t, err)

		// Verify that the aggregator used the calculated digest, not the malicious one
		assert.NotNil(t, agg.aggregatedOperators.mostCommonResponse)
		assert.Equal(t, legitimateDigest, agg.aggregatedOperators.mostCommonResponse.Digest)

		// Generate certificate and verify it uses the calculated digest
		cert, err := agg.GenerateFinalCertificate()
		require.NoError(t, err)

		// Certificate should contain the legitimate output and its calculated digest
		assert.Equal(t, legitimateOutput, cert.TaskResponse)
		assert.Equal(t, legitimateDigest, cert.TaskResponseDigest)
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
				Address:   fmt.Sprintf("0x%d", i+1),
				PublicKey: pubKey,
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
			50, // 50% threshold
			taskData,
			&deadline,
			operators,
		)
		require.NoError(t, err)
		require.NotNil(t, agg)

		// Create a task result with mismatched task ID
		mismatchedTaskId := "0xdifferenttaskid1234567890abcdef12345678"
		payload := []byte("test-response")
		digest := util.GetKeccak256Digest(payload)

		taskResult := &types.TaskResult{
			TaskId:          mismatchedTaskId, // Wrong task ID
			OperatorAddress: operators[0].Address,
			Output:          payload,
			OutputDigest:    digest[:],
		}
		sig, err := privateKeys[0].SignSolidityCompatible(digest)
		require.NoError(t, err)
		taskResult.Signature = sig.Bytes()

		// Process the signature with mismatched task ID
		err = agg.ProcessNewSignature(context.Background(), taskId, taskResult)

		// Should return an error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "task ID mismatch")
		assert.Contains(t, err.Error(), taskId)
		assert.Contains(t, err.Error(), mismatchedTaskId)

		// Verify that no signature was recorded
		assert.Equal(t, 0, len(agg.ReceivedSignatures))
		assert.False(t, agg.SigningThresholdMet())

		// Now submit with correct task ID to verify aggregator works properly
		correctTaskResult := &types.TaskResult{
			TaskId:          taskId, // Correct task ID
			OperatorAddress: operators[0].Address,
			Output:          payload,
			OutputDigest:    digest[:],
		}
		correctTaskResult.Signature = sig.Bytes()

		err = agg.ProcessNewSignature(context.Background(), taskId, correctTaskResult)
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
				Address:   derivedAddress.String(),
				PublicKey: derivedAddress,
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
			50, // 50% threshold
			taskData,
			&deadline,
			operators,
		)
		require.NoError(t, err)
		require.NotNil(t, agg)

		// Create a task result with mismatched task ID
		mismatchedTaskId := "0xdifferenttaskid1234567890abcdef12345678"
		payload := []byte("test-response")
		digest := util.GetKeccak256Digest(payload)

		taskResult := &types.TaskResult{
			TaskId:          mismatchedTaskId, // Wrong task ID
			OperatorAddress: operators[0].Address,
			Output:          payload,
			OutputDigest:    digest[:],
		}
		sig, err := privateKeys[0].Sign(digest[:])
		require.NoError(t, err)
		taskResult.Signature = sig.Bytes()

		// Process the signature with mismatched task ID
		err = agg.ProcessNewSignature(context.Background(), taskId, taskResult)

		// Should return an error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "task ID mismatch")
		assert.Contains(t, err.Error(), taskId)
		assert.Contains(t, err.Error(), mismatchedTaskId)

		// Verify that no signature was recorded
		assert.Equal(t, 0, len(agg.ReceivedSignatures))
		assert.False(t, agg.SigningThresholdMet())

		// Now submit with correct task ID to verify aggregator works properly
		correctTaskResult := &types.TaskResult{
			TaskId:          taskId, // Correct task ID
			OperatorAddress: operators[0].Address,
			Output:          payload,
			OutputDigest:    digest[:],
		}
		correctTaskResult.Signature = sig.Bytes()

		err = agg.ProcessNewSignature(context.Background(), taskId, correctTaskResult)
		require.NoError(t, err)

		// Verify signature was recorded
		assert.Equal(t, 1, len(agg.ReceivedSignatures))
		assert.True(t, agg.SigningThresholdMet()) // 1/2 = 50% threshold met
	})
}

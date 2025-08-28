package aggregation

import (
	"context"
	"fmt"
	"math/big"
	"sort"
	"testing"

	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	"github.com/Layr-Labs/crypto-libs/pkg/signing"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/ITaskMailbox"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// MockTaskMailbox mocks the ITaskMailbox contract binding
type MockTaskMailbox struct {
	mock.Mock

	// Store the last submitted certificate for verification
	LastSubmittedTaskId       [32]byte
	LastSubmittedCertBytes    []byte
	LastSubmittedTaskResponse []byte
	LastBN254Certificate      ITaskMailbox.IBN254CertificateVerifierTypesBN254Certificate
}

func (m *MockTaskMailbox) GetBN254CertificateBytes(opts *bind.CallOpts, cert ITaskMailbox.IBN254CertificateVerifierTypesBN254Certificate) ([]byte, error) {
	// Store the certificate for later verification
	m.LastBN254Certificate = cert

	args := m.Called(opts, cert)

	// Simulate actual certificate byte encoding
	certBytes := []byte("mocked_cert_bytes")
	return certBytes, args.Error(0)
}

func (m *MockTaskMailbox) SubmitResult(opts *bind.TransactOpts, taskId [32]byte, certBytes []byte, taskResponse []byte) (*types.Transaction, error) {
	// Store the submitted data for verification
	m.LastSubmittedTaskId = taskId
	m.LastSubmittedCertBytes = certBytes
	m.LastSubmittedTaskResponse = taskResponse

	args := m.Called(opts, taskId, certBytes, taskResponse)

	// Return a mock transaction
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Transaction), args.Error(1)
}

// TestBN254ContractSubmission tests the actual data that would be submitted to the contract
func TestBN254ContractSubmission(t *testing.T) {
	ctx := context.Background()
	taskId := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	operatorSetId := uint32(1)

	// Create operators with keys
	operators := make([]*Operator[signing.PublicKey], 4)
	privateKeys := make([]*bn254.PrivateKey, 4)

	for i := 0; i < 4; i++ {
		privateKey, publicKey, err := bn254.GenerateKeyPair()
		require.NoError(t, err)
		privateKeys[i] = privateKey

		operators[i] = &Operator[signing.PublicKey]{
			Address:       fmt.Sprintf("0x%040d", i+1),
			PublicKey:     publicKey,
			OperatorIndex: uint32(i),
		}
	}

	// Create aggregator with 75% threshold (3 out of 4)
	aggregator, err := NewBN254TaskResultAggregator(
		ctx,
		taskId,
		operatorSetId,
		7500,
		[]byte("task data"),
		nil,
		operators,
	)
	require.NoError(t, err)

	// Have 3 operators sign the same result
	output := []byte("consensus result")
	for i := 0; i < 3; i++ {
		taskResult, err := createSignedBN254TaskResult(
			taskId,
			operators[i],
			operatorSetId,
			output,
			privateKeys[i],
		)
		require.NoError(t, err)

		err = aggregator.ProcessNewSignature(ctx, taskResult)
		require.NoError(t, err)
	}

	// Generate certificate
	cert, err := aggregator.GenerateFinalCertificate()
	require.NoError(t, err)

	// Create mock TaskMailbox
	mockMailbox := new(MockTaskMailbox)

	// Set up expectations for GetBN254CertificateBytes
	mockMailbox.On("GetBN254CertificateBytes", mock.Anything, mock.Anything).Return(nil)

	// Set up expectations for SubmitResult
	mockTx := &types.Transaction{}
	mockMailbox.On("SubmitResult", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockTx, nil)

	// Create a mock-enabled contract caller
	logger, _ := zap.NewDevelopment()
	cc := &MockContractCaller{
		taskMailbox: mockMailbox,
		logger:      logger,
	}

	// Submit the certificate through our mock
	globalTableRootReferenceTimestamp := uint32(1000)
	err = cc.SubmitBN254TaskResult(ctx, cert, globalTableRootReferenceTimestamp)
	require.NoError(t, err)

	// Verify the certificate data submitted to the contract
	submittedCert := mockMailbox.LastBN254Certificate

	// 1. Verify reference timestamp
	assert.Equal(t, globalTableRootReferenceTimestamp, submittedCert.ReferenceTimestamp)

	// 2. Verify message hash (task response digest)
	assert.Equal(t, cert.TaskResponseDigest, submittedCert.MessageHash)

	// 3. Verify aggregated signature (G1 point)
	g1Point := &bn254.G1Point{
		G1Affine: cert.SignersSignature.GetG1Point(),
	}
	g1Bytes, err := g1Point.ToPrecompileFormat()
	require.NoError(t, err)

	expectedSigX := new(big.Int).SetBytes(g1Bytes[0:32])
	expectedSigY := new(big.Int).SetBytes(g1Bytes[32:64])
	assert.Equal(t, expectedSigX, submittedCert.Signature.X)
	assert.Equal(t, expectedSigY, submittedCert.Signature.Y)

	// 4. Verify aggregated public key (G2 point)
	g2Bytes, err := cert.SignersPublicKey.ToPrecompileFormat()
	require.NoError(t, err)

	assert.Equal(t, new(big.Int).SetBytes(g2Bytes[0:32]), submittedCert.Apk.X[0])
	assert.Equal(t, new(big.Int).SetBytes(g2Bytes[32:64]), submittedCert.Apk.X[1])
	assert.Equal(t, new(big.Int).SetBytes(g2Bytes[64:96]), submittedCert.Apk.Y[0])
	assert.Equal(t, new(big.Int).SetBytes(g2Bytes[96:128]), submittedCert.Apk.Y[1])

	// 5. Verify non-signer witnesses
	assert.Len(t, submittedCert.NonSignerWitnesses, 1) // Only operator 3 didn't sign

	nonSignerWitness := submittedCert.NonSignerWitnesses[0]
	assert.Equal(t, uint32(3), nonSignerWitness.OperatorIndex)

	// Verify non-signer public key
	nonSignerPubKeyBytes := cert.NonSignerOperators[0].PublicKey.Bytes()
	expectedPubKeyX := new(big.Int).SetBytes(nonSignerPubKeyBytes[0:32])
	expectedPubKeyY := new(big.Int).SetBytes(nonSignerPubKeyBytes[32:64])
	assert.Equal(t, expectedPubKeyX, nonSignerWitness.OperatorInfo.Pubkey.X)
	assert.Equal(t, expectedPubKeyY, nonSignerWitness.OperatorInfo.Pubkey.Y)

	// 6. Verify task response was submitted
	assert.Equal(t, output, mockMailbox.LastSubmittedTaskResponse)

	// 7. Verify task ID
	taskIdBytes, _ := hexutil.Decode(taskId)
	assert.Equal(t, taskIdBytes, mockMailbox.LastSubmittedTaskId[:])
}

// TestBN254ThresholdEdgeCases tests edge cases for threshold calculation
func TestBN254ThresholdEdgeCases(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name                   string
		numOperators           int
		thresholdBips          uint16
		signersByResult        map[string][]int // result -> operator indices
		expectedThresholdMet   bool
		expectedWinner         string
		expectedNonSignerCount int
	}{
		{
			name:          "Majority consensus meets participation threshold",
			numOperators:  5,
			thresholdBips: 6000, // 60% = 3 operators
			signersByResult: map[string][]int{
				"resultA": {0, 1, 2}, // 3 signers for A
				// operators 3, 4 don't sign
			},
			expectedThresholdMet:   true,
			expectedWinner:         "resultA",
			expectedNonSignerCount: 2,
		},
		{
			name:          "Split results meet participation threshold",
			numOperators:  6,
			thresholdBips: 5000, // 50% = 3 operators
			signersByResult: map[string][]int{
				"resultA": {0, 1}, // 2 signers for A
				"resultB": {2},    // 1 signer for B
				// Total = 3 signers (meets threshold)
			},
			expectedThresholdMet:   true,
			expectedWinner:         "resultA", // Most common
			expectedNonSignerCount: 4,         // B signer + 3 non-signers
		},
		{
			name:          "Minority result wins with sufficient participation",
			numOperators:  10,
			thresholdBips: 6000, // 60% = 6 operators
			signersByResult: map[string][]int{
				"resultA": {0, 1}, // 2 signers
				"resultB": {2, 3}, // 2 signers
				"resultC": {4, 5}, // 2 signers
				// Total = 6 signers (meets threshold)
			},
			expectedThresholdMet:   true,
			expectedWinner:         "resultA", // First to reach 2
			expectedNonSignerCount: 8,         // All except A signers
		},
		{
			name:          "Threshold not met",
			numOperators:  5,
			thresholdBips: 8000, // 80% = 4 operators
			signersByResult: map[string][]int{
				"resultA": {0, 1, 2}, // 3 signers
				// Only 3 < 4 required
			},
			expectedThresholdMet: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			taskId := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
			operatorSetId := uint32(1)

			// Create operators
			operators := make([]*Operator[signing.PublicKey], tc.numOperators)
			privateKeys := make([]*bn254.PrivateKey, tc.numOperators)

			for i := 0; i < tc.numOperators; i++ {
				privateKey, publicKey, err := bn254.GenerateKeyPair()
				require.NoError(t, err)
				privateKeys[i] = privateKey

				operators[i] = &Operator[signing.PublicKey]{
					Address:       fmt.Sprintf("0x%040d", i+1),
					PublicKey:     publicKey,
					OperatorIndex: uint32(i),
				}
			}

			// Create aggregator
			aggregator, err := NewBN254TaskResultAggregator(
				ctx,
				taskId,
				operatorSetId,
				tc.thresholdBips,
				[]byte("task data"),
				nil,
				operators,
			)
			require.NoError(t, err)

			// Process signatures in deterministic order by sorting results alphabetically
			// This ensures consistent test behavior regardless of map iteration order
			var results []string
			for result := range tc.signersByResult {
				results = append(results, result)
			}
			sort.Strings(results)

			for _, result := range results {
				signerIndices := tc.signersByResult[result]
				output := []byte(result)
				for _, idx := range signerIndices {
					taskResult, err := createSignedBN254TaskResult(
						taskId,
						operators[idx],
						operatorSetId,
						output,
						privateKeys[idx],
					)
					require.NoError(t, err)

					err = aggregator.ProcessNewSignature(ctx, taskResult)
					require.NoError(t, err)
				}
			}

			// Check threshold
			assert.Equal(t, tc.expectedThresholdMet, aggregator.SigningThresholdMet(),
				"Threshold met status mismatch")

			// If threshold is met, generate certificate and verify
			if tc.expectedThresholdMet {
				cert, err := aggregator.GenerateFinalCertificate()
				require.NoError(t, err)

				assert.Equal(t, []byte(tc.expectedWinner), cert.TaskResponse,
					"Wrong winning result")
				assert.Len(t, cert.NonSignerOperators, tc.expectedNonSignerCount,
					"Wrong number of non-signers")

				// Verify non-signers are sorted by index
				for i := 1; i < len(cert.NonSignerOperators); i++ {
					assert.Less(t, cert.NonSignerOperators[i-1].OperatorIndex,
						cert.NonSignerOperators[i].OperatorIndex,
						"Non-signers not sorted by operator index")
				}
			}
		})
	}
}

// MockContractCaller implements a simplified version of contract caller for testing
type MockContractCaller struct {
	taskMailbox *MockTaskMailbox
	logger      *zap.Logger
}

func (m *MockContractCaller) SubmitBN254TaskResult(
	ctx context.Context,
	aggCert *AggregatedBN254Certificate,
	globalTableRootReferenceTimestamp uint32,
) error {
	// Simulate what the real contract caller does

	// Convert signature to G1 point
	g1Point := &bn254.G1Point{
		G1Affine: aggCert.SignersSignature.GetG1Point(),
	}
	g1Bytes, err := g1Point.ToPrecompileFormat()
	if err != nil {
		return fmt.Errorf("signature not in correct subgroup: %w", err)
	}

	// Convert public key to G2 point
	g2Bytes, err := aggCert.SignersPublicKey.ToPrecompileFormat()
	if err != nil {
		return fmt.Errorf("public key not in correct subgroup: %w", err)
	}

	// Build non-signer witnesses
	nonSignerWitnesses := make([]ITaskMailbox.IBN254CertificateVerifierTypesBN254OperatorInfoWitness, 0, len(aggCert.NonSignerOperators))
	for _, nonSigner := range aggCert.NonSignerOperators {
		witness := ITaskMailbox.IBN254CertificateVerifierTypesBN254OperatorInfoWitness{
			OperatorIndex:     nonSigner.OperatorIndex,
			OperatorInfoProof: []byte{},
			OperatorInfo: ITaskMailbox.IOperatorTableCalculatorTypesBN254OperatorInfo{
				Pubkey: ITaskMailbox.BN254G1Point{
					X: new(big.Int).SetBytes(nonSigner.PublicKey.Bytes()[0:32]),
					Y: new(big.Int).SetBytes(nonSigner.PublicKey.Bytes()[32:64]),
				},
			},
		}
		nonSignerWitnesses = append(nonSignerWitnesses, witness)
	}

	// Create certificate struct for contract
	cert := ITaskMailbox.IBN254CertificateVerifierTypesBN254Certificate{
		ReferenceTimestamp: globalTableRootReferenceTimestamp,
		MessageHash:        aggCert.TaskResponseDigest,
		Signature: ITaskMailbox.BN254G1Point{
			X: new(big.Int).SetBytes(g1Bytes[0:32]),
			Y: new(big.Int).SetBytes(g1Bytes[32:64]),
		},
		Apk: ITaskMailbox.BN254G2Point{
			X: [2]*big.Int{
				new(big.Int).SetBytes(g2Bytes[0:32]),
				new(big.Int).SetBytes(g2Bytes[32:64]),
			},
			Y: [2]*big.Int{
				new(big.Int).SetBytes(g2Bytes[64:96]),
				new(big.Int).SetBytes(g2Bytes[96:128]),
			},
		},
		NonSignerWitnesses: nonSignerWitnesses,
	}

	// Get certificate bytes
	certBytes, err := m.taskMailbox.GetBN254CertificateBytes(&bind.CallOpts{}, cert)
	if err != nil {
		return fmt.Errorf("failed to get BN254 certificate bytes: %w", err)
	}

	// Submit result
	var taskId [32]byte
	copy(taskId[:], aggCert.TaskId)

	_, err = m.taskMailbox.SubmitResult(
		&bind.TransactOpts{From: common.Address{}},
		taskId,
		certBytes,
		aggCert.TaskResponse,
	)

	return err
}

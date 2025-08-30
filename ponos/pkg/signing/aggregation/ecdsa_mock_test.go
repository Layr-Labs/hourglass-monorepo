package aggregation

import (
	"bytes"
	"crypto/rand"
	"sort"
	"testing"

	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/IECDSACertificateVerifier"
	"github.com/Layr-Labs/eigenlayer-contracts/pkg/bindings/ITaskMailbox"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockECDSATaskMailbox mocks the ITaskMailbox contract binding for ECDSA
type MockECDSATaskMailbox struct {
	mock.Mock

	// Store the last submitted data for verification
	LastSubmittedTaskId       [32]byte
	LastSubmittedCertBytes    []byte
	LastSubmittedTaskResponse []byte
	LastECDSACertificate      IECDSACertificateVerifier.IECDSACertificateVerifierTypesECDSACertificate
}

func (m *MockECDSATaskMailbox) GetECDSACertificateBytes(opts *bind.CallOpts, cert ITaskMailbox.IECDSACertificateVerifierTypesECDSACertificate) ([]byte, error) {
	// Store for verification (convert between similar types)
	m.LastECDSACertificate = IECDSACertificateVerifier.IECDSACertificateVerifierTypesECDSACertificate{
		ReferenceTimestamp: cert.ReferenceTimestamp,
		MessageHash:        cert.MessageHash,
		Sig:                cert.Sig,
	}

	args := m.Called(opts, cert)

	// Simulate actual certificate byte encoding
	certBytes := []byte("mocked_ecdsa_cert_bytes")
	return certBytes, args.Error(0)
}

func (m *MockECDSATaskMailbox) SubmitResult(opts *bind.TransactOpts, taskId [32]byte, certBytes []byte, taskResponse []byte) (*types.Transaction, error) {
	// Store the submitted data for verification
	m.LastSubmittedTaskId = taskId
	m.LastSubmittedCertBytes = certBytes
	m.LastSubmittedTaskResponse = taskResponse

	args := m.Called(opts, taskId, certBytes, taskResponse)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Transaction), args.Error(1)
}

// TestECDSAGetFinalSignatureSorting tests that signatures are sorted correctly by address
func TestECDSAGetFinalSignatureSorting(t *testing.T) {
	// Create test addresses with known sorting order
	// These addresses will sort in a specific order when compared byte-by-byte
	addresses := []common.Address{
		common.HexToAddress("0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"), // Should be last
		common.HexToAddress("0x0000000000000000000000000000000000000001"), // Should be first
		common.HexToAddress("0x5555555555555555555555555555555555555555"), // Should be middle
		common.HexToAddress("0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"), // Should be second to last
		common.HexToAddress("0x1111111111111111111111111111111111111111"), // Should be second
	}

	// Create signatures for each address (65 bytes each)
	signatures := make(map[common.Address][]byte)
	for i, addr := range addresses {
		// Create a unique signature for each address
		sig := make([]byte, 65)
		// Fill with pattern that makes it easy to verify order
		for j := range sig {
			sig[j] = byte(i + 1) // Use i+1 so we don't have zeros
		}
		signatures[addr] = sig
	}

	// Create certificate with unsorted signatures
	cert := &AggregatedECDSACertificate{
		SignersSignatures: signatures,
	}

	// Get final signature
	finalSig, err := cert.GetFinalSignature()
	require.NoError(t, err)

	// Expected order after sorting
	expectedOrder := []common.Address{
		common.HexToAddress("0x0000000000000000000000000000000000000001"), // Index 1 in original
		common.HexToAddress("0x1111111111111111111111111111111111111111"), // Index 4 in original
		common.HexToAddress("0x5555555555555555555555555555555555555555"), // Index 2 in original
		common.HexToAddress("0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"), // Index 3 in original
		common.HexToAddress("0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"), // Index 0 in original
	}

	// Verify final signature has correct order
	assert.Len(t, finalSig, 65*5, "Final signature should have 5 * 65 bytes")

	// Check each signature is in the correct position
	for i, expectedAddr := range expectedOrder {
		startIdx := i * 65
		endIdx := startIdx + 65
		sigSegment := finalSig[startIdx:endIdx]

		// Compare with the original signature for this address
		originalSig := signatures[expectedAddr]
		assert.Equal(t, originalSig, sigSegment,
			"Signature for address %s should be at position %d", expectedAddr.Hex(), i)
	}
}

//// TestECDSAContractSubmission tests the actual data that would be submitted to the contract
//func TestECDSAContractSubmission(t *testing.T) {
//	ctx := context.Background()
//	taskId := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
//	operatorSetId := uint32(1)
//
//	// Create operators with ECDSA keys
//	operators := make([]*Operator[common.Address], 4)
//	privateKeys := make([]*ecdsa.PrivateKey, 4)
//
//	for i := 0; i < 4; i++ {
//		privKey, _, err := ecdsa.GenerateKeyPair()
//		require.NoError(t, err)
//		privateKeys[i] = privKey
//
//		derivedAddress, err := privKey.DeriveAddress()
//		require.NoError(t, err)
//
//		operators[i] = &Operator[common.Address]{
//			Address:       derivedAddress.String(),
//			PublicKey:     derivedAddress,
//			OperatorIndex: uint32(i),
//		}
//	}
//
//	// Create mock TaskMailbox
//	mockMailbox := new(MockECDSATaskMailbox)
//
//	// Set up expectations
//	mockMailbox.On("GetECDSACertificateBytes", mock.Anything, mock.Anything).Return(nil)
//	mockTx := &types.Transaction{}
//	mockMailbox.On("SubmitResult", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockTx, nil)
//
//	// Create a mock-enabled contract caller
//	logger, _ := zap.NewDevelopment()
//	cc := &MockECDSAContractCaller{
//		taskMailbox: mockMailbox,
//		logger:      logger,
//	}
//
//	// Create aggregator with 75% threshold (3 out of 4)
//	aggregator, err := NewECDSATaskResultAggregator(
//		ctx,
//		taskId,
//		operatorSetId,
//		1,
//		7500,
//		cc,
//		[]byte("task data"),
//		nil,
//		operators,
//	)
//	require.NoError(t, err)
//
//	// Have 3 operators sign the same result
//	output := []byte("consensus result")
//	for i := 0; i < 3; i++ {
//		taskResult, err := createSignedECDSATaskResult(
//			taskId,
//			operators[i],
//			operatorSetId,
//			output,
//			privateKeys[i],
//			1000,
//			cc,
//		)
//		require.NoError(t, err)
//
//		err = aggregator.ProcessNewSignature(ctx, taskResult)
//		require.NoError(t, err)
//	}
//
//	// Generate certificate
//	cert, err := aggregator.GenerateFinalCertificate()
//	require.NoError(t, err)
//
//	// Verify GetFinalSignature produces correctly sorted signatures
//	finalSig, err := cert.GetFinalSignature()
//	require.NoError(t, err)
//
//	// Should have 3 signatures * 65 bytes each
//	assert.Len(t, finalSig, 3*65)
//
//	// Submit the certificate through our mock
//	globalTableRootReferenceTimestamp := uint32(1000)
//	err = cc.SubmitECDSATaskResult(ctx, cert, globalTableRootReferenceTimestamp)
//	require.NoError(t, err)
//
//	// Verify the certificate data submitted to the contract
//	submittedCert := mockMailbox.LastECDSACertificate
//
//	// 1. Verify reference timestamp
//	assert.Equal(t, globalTableRootReferenceTimestamp, submittedCert.ReferenceTimestamp)
//
//	// 2. Verify message hash
//	assert.Equal(t, cert.GetTaskMessageHash(), submittedCert.MessageHash)
//
//	// 3. Verify final signature (should be sorted)
//	assert.Equal(t, finalSig, submittedCert.Sig[:])
//
//	// 4. Verify task response was submitted
//	assert.Equal(t, output, mockMailbox.LastSubmittedTaskResponse)
//
//	// 5. Verify task ID
//	taskIdBytes, _ := hexutil.Decode(taskId)
//	assert.Equal(t, taskIdBytes, mockMailbox.LastSubmittedTaskId[:])
//}

// TestECDSASignatureSortingDeterministic tests that sorting is deterministic
func TestECDSASignatureSortingDeterministic(t *testing.T) {
	// Generate random addresses
	numAddresses := 10
	addresses := make([]common.Address, numAddresses)
	signatures := make(map[common.Address][]byte)

	for i := 0; i < numAddresses; i++ {
		// Generate random address
		addr := make([]byte, 20)
		_, err := rand.Read(addr)
		require.NoError(t, err)
		addresses[i] = common.BytesToAddress(addr)

		// Generate signature
		sig := make([]byte, 65)
		_, err = rand.Read(sig)
		require.NoError(t, err)
		signatures[common.BytesToAddress(addr)] = sig
	}

	// Create multiple certificates with the same signatures
	results := make([][]byte, 5)
	for i := 0; i < 5; i++ {
		cert := &AggregatedECDSACertificate{
			SignersSignatures: signatures,
		}

		finalSig, err := cert.GetFinalSignature()
		require.NoError(t, err)
		results[i] = finalSig
	}

	// All results should be identical (deterministic)
	for i := 1; i < 5; i++ {
		assert.Equal(t, results[0], results[i],
			"Sorting should be deterministic - result %d differs from result 0", i)
	}
}

// TestECDSASignatureConcatenationOrder verifies the exact byte order of concatenation
func TestECDSASignatureConcatenationOrder(t *testing.T) {
	// Use specific addresses with known hex values
	addr1 := common.HexToAddress("0x1000000000000000000000000000000000000001")
	addr2 := common.HexToAddress("0x2000000000000000000000000000000000000002")
	addr3 := common.HexToAddress("0x3000000000000000000000000000000000000003")

	// Create signatures with identifiable patterns
	sig1 := bytes.Repeat([]byte{0x11}, 65)
	sig2 := bytes.Repeat([]byte{0x22}, 65)
	sig3 := bytes.Repeat([]byte{0x33}, 65)

	cert := &AggregatedECDSACertificate{
		SignersSignatures: map[common.Address][]byte{
			addr3: sig3, // Add in reverse order to test sorting
			addr1: sig1,
			addr2: sig2,
		},
	}

	finalSig, err := cert.GetFinalSignature()
	require.NoError(t, err)

	// Addresses should be sorted in byte order: addr1 < addr2 < addr3
	expectedSig := append(append(sig1, sig2...), sig3...)
	assert.Equal(t, expectedSig, finalSig,
		"Signatures should be concatenated in sorted address order")

	// Verify each segment
	assert.Equal(t, sig1, finalSig[0:65], "First 65 bytes should be sig1")
	assert.Equal(t, sig2, finalSig[65:130], "Second 65 bytes should be sig2")
	assert.Equal(t, sig3, finalSig[130:195], "Third 65 bytes should be sig3")
}

// TestECDSAInvalidSignatureLength tests error handling for invalid signature lengths
func TestECDSAInvalidSignatureLength(t *testing.T) {
	addr := common.HexToAddress("0x1234567890123456789012345678901234567890")

	testCases := []struct {
		name        string
		sigLength   int
		shouldError bool
	}{
		{"Valid 65 bytes", 65, false},
		{"Too short 64 bytes", 64, true},
		{"Too long 66 bytes", 66, true},
		{"Empty signature", 0, true},
		{"Way too long", 130, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cert := &AggregatedECDSACertificate{
				SignersSignatures: map[common.Address][]byte{
					addr: make([]byte, tc.sigLength),
				},
			}

			finalSig, err := cert.GetFinalSignature()

			if tc.shouldError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid length")
				assert.Nil(t, finalSig)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, finalSig)
				assert.Len(t, finalSig, 65)
			}
		})
	}
}

// TestECDSAEmptySignatures tests error handling for empty signatures map
func TestECDSAEmptySignatures(t *testing.T) {
	cert := &AggregatedECDSACertificate{
		SignersSignatures: map[common.Address][]byte{},
	}

	finalSig, err := cert.GetFinalSignature()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no signatures found")
	assert.Nil(t, finalSig)
}

// TestECDSALargeScaleSorting tests sorting with many addresses
func TestECDSALargeScaleSorting(t *testing.T) {
	// Generate 100 random addresses
	numAddresses := 100
	addresses := make([]common.Address, numAddresses)
	signatures := make(map[common.Address][]byte)

	for i := 0; i < numAddresses; i++ {
		privateKey, err := crypto.GenerateKey()
		require.NoError(t, err)

		addr := crypto.PubkeyToAddress(privateKey.PublicKey)
		addresses[i] = addr

		// Create signature with index encoded for verification
		sig := make([]byte, 65)
		// Put index in first byte for tracking
		sig[0] = byte(i)
		_, err = rand.Read(sig[1:])
		require.NoError(t, err)
		signatures[addr] = sig
	}

	cert := &AggregatedECDSACertificate{
		SignersSignatures: signatures,
	}

	finalSig, err := cert.GetFinalSignature()
	require.NoError(t, err)

	assert.Len(t, finalSig, numAddresses*65)

	// Extract addresses and verify they're sorted
	extractedAddresses := make([]common.Address, 0, numAddresses)
	for addr := range signatures {
		extractedAddresses = append(extractedAddresses, addr)
	}

	// Sort using same method as GetFinalSignature
	sortedAddresses := make([]common.Address, len(extractedAddresses))
	copy(sortedAddresses, extractedAddresses)
	sort.Slice(sortedAddresses, func(i, j int) bool {
		return bytes.Compare(sortedAddresses[i][:], sortedAddresses[j][:]) < 0
	})

	// Verify the signatures are in the correct order
	for i, addr := range sortedAddresses {
		expectedSig := signatures[addr]
		actualSig := finalSig[i*65 : (i+1)*65]
		assert.Equal(t, expectedSig, actualSig,
			"Signature at position %d should match address %s", i, addr.Hex())
	}
}

// BenchmarkECDSAGetFinalSignature benchmarks the sorting performance
func BenchmarkECDSAGetFinalSignature(b *testing.B) {
	// Create signatures for benchmarking
	numSigners := 100
	signatures := make(map[common.Address][]byte)

	for i := 0; i < numSigners; i++ {
		addr := make([]byte, 20)
		_, err := rand.Read(addr)
		if err != nil {
			b.Fatal(err)
		}
		sig := make([]byte, 65)
		_, err = rand.Read(sig)
		if err != nil {
			b.Fatal(err)
		}
		signatures[common.BytesToAddress(addr)] = sig
	}

	cert := &AggregatedECDSACertificate{
		SignersSignatures: signatures,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cert.GetFinalSignature()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// TestECDSAHexConsistency verifies addresses maintain consistent hex representation
func TestECDSAHexConsistency(t *testing.T) {
	// Test addresses that might have tricky hex representations
	testAddresses := []string{
		"0x0000000000000000000000000000000000000001",
		"0x00000000000000000000000000000000000000ff",
		"0xffffffffffffffffffffffffffffffffffffffff",
		"0x1234567890abcdef1234567890abcdef12345678",
		"0xABCDEF1234567890ABCDEF1234567890ABCDEF12",
	}

	signatures := make(map[common.Address][]byte)
	for i, addrStr := range testAddresses {
		addr := common.HexToAddress(addrStr)
		sig := make([]byte, 65)
		sig[0] = byte(i) // Mark with index for identification
		signatures[addr] = sig

		// Verify address parsing is consistent
		assert.Equal(t, 20, len(addr), "Address should be 20 bytes")
	}

	cert := &AggregatedECDSACertificate{
		SignersSignatures: signatures,
	}

	finalSig, err := cert.GetFinalSignature()
	require.NoError(t, err)
	assert.Len(t, finalSig, len(testAddresses)*65)

	// Verify addresses maintain their identity through the process
	for i := 0; i < len(testAddresses); i++ {
		segment := finalSig[i*65 : (i+1)*65]
		// Find which original signature this is by checking first byte
		found := false
		for _, origSig := range signatures {
			if segment[0] == origSig[0] {
				found = true
				assert.Equal(t, origSig, segment, "Signature integrity should be maintained")
				break
			}
		}
		assert.True(t, found, "Each signature should be found in the final result")
	}
}

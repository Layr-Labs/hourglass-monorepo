package operatorManager

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/Layr-Labs/crypto-libs/pkg/bn254"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/contractCaller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/peering"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signing/aggregation"
	"github.com/ethereum/go-ethereum/common"
	ethereumTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// MockContractCaller is a mock implementation of IContractCaller
type MockContractCaller struct {
	mock.Mock
}

func (m *MockContractCaller) SubmitBN254TaskResult(ctx context.Context, aggCert *aggregation.AggregatedBN254Certificate, globalTableRootReferenceTimestamp uint32) (*ethereumTypes.Receipt, error) {
	args := m.Called(ctx, aggCert, globalTableRootReferenceTimestamp)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ethereumTypes.Receipt), args.Error(1)
}

func (m *MockContractCaller) SubmitBN254TaskResultRetryable(ctx context.Context, aggCert *aggregation.AggregatedBN254Certificate, globalTableRootReferenceTimestamp uint32) (*ethereumTypes.Receipt, error) {
	args := m.Called(ctx, aggCert, globalTableRootReferenceTimestamp)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ethereumTypes.Receipt), args.Error(1)
}

func (m *MockContractCaller) SubmitECDSATaskResult(ctx context.Context, aggCert *aggregation.AggregatedECDSACertificate, globalTableRootReferenceTimestamp uint32) (*ethereumTypes.Receipt, error) {
	args := m.Called(ctx, aggCert, globalTableRootReferenceTimestamp)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ethereumTypes.Receipt), args.Error(1)
}

func (m *MockContractCaller) SubmitECDSATaskResultRetryable(ctx context.Context, aggCert *aggregation.AggregatedECDSACertificate, globalTableRootReferenceTimestamp uint32) (*ethereumTypes.Receipt, error) {
	args := m.Called(ctx, aggCert, globalTableRootReferenceTimestamp)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ethereumTypes.Receipt), args.Error(1)
}

func (m *MockContractCaller) GetAVSConfig(avsAddress string) (*contractCaller.AVSConfig, error) {
	args := m.Called(avsAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*contractCaller.AVSConfig), args.Error(1)
}

func (m *MockContractCaller) GetOperatorSetCurveType(avsAddress string, operatorSetId uint32) (config.CurveType, error) {
	args := m.Called(avsAddress, operatorSetId)
	return args.Get(0).(config.CurveType), args.Error(1)
}

func (m *MockContractCaller) GetOperatorSetMembersWithPeering(avsAddress string, operatorSetId uint32) ([]*peering.OperatorPeerInfo, error) {
	args := m.Called(avsAddress, operatorSetId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*peering.OperatorPeerInfo), args.Error(1)
}

func (m *MockContractCaller) GetOperatorSetDetailsForOperator(operatorAddress common.Address, avsAddress string, operatorSetId uint32) (*peering.OperatorSet, error) {
	args := m.Called(operatorAddress, avsAddress, operatorSetId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*peering.OperatorSet), args.Error(1)
}

func (m *MockContractCaller) PublishMessageToInbox(ctx context.Context, avsAddress string, operatorSetId uint32, payload []byte) (*ethereumTypes.Receipt, error) {
	args := m.Called(ctx, avsAddress, operatorSetId, payload)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ethereumTypes.Receipt), args.Error(1)
}

func (m *MockContractCaller) GetOperatorTableDataForOperatorSet(ctx context.Context, avsAddress common.Address, operatorSetId uint32, chainId config.ChainId, referenceBlocknumber uint64) (*contractCaller.OperatorTableData, error) {
	args := m.Called(ctx, avsAddress, operatorSetId, chainId, referenceBlocknumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*contractCaller.OperatorTableData), args.Error(1)
}

func (m *MockContractCaller) GetTableUpdaterReferenceTimeAndBlock(ctx context.Context, tableUpdaterAddr common.Address, atBlockNumber uint64) (*contractCaller.LatestReferenceTimeAndBlock, error) {
	args := m.Called(ctx, tableUpdaterAddr, atBlockNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*contractCaller.LatestReferenceTimeAndBlock), args.Error(1)
}

func (m *MockContractCaller) GetSupportedChainsForMultichain(ctx context.Context, referenceBlockNumber int64) ([]*big.Int, []common.Address, error) {
	args := m.Called(ctx, referenceBlockNumber)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).([]*big.Int), args.Get(1).([]common.Address), args.Error(2)
}

func (m *MockContractCaller) CalculateECDSACertificateDigestBytes(ctx context.Context, referenceTimestamp uint32, messageHash [32]byte) ([]byte, error) {
	args := m.Called(ctx, referenceTimestamp, messageHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockContractCaller) CalculateBN254CertificateDigestBytes(ctx context.Context, referenceTimestamp uint32, messageHash [32]byte) ([]byte, error) {
	args := m.Called(ctx, referenceTimestamp, messageHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockContractCaller) GetExecutorOperatorSetTaskConfig(ctx context.Context, avsAddress common.Address, opsetId uint32) (*contractCaller.TaskMailboxExecutorOperatorSetConfig, error) {
	args := m.Called(ctx, avsAddress, opsetId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*contractCaller.TaskMailboxExecutorOperatorSetConfig), args.Error(1)
}

func (m *MockContractCaller) GetOperatorBN254KeyRegistrationMessageHash(ctx context.Context, operatorAddress common.Address, avsAddress common.Address, operatorSetId uint32, keyData []byte) ([32]byte, error) {
	args := m.Called(ctx, operatorAddress, avsAddress, operatorSetId, keyData)
	return args.Get(0).([32]byte), args.Error(1)
}

func (m *MockContractCaller) GetOperatorECDSAKeyRegistrationMessageHash(ctx context.Context, operatorAddress common.Address, avsAddress common.Address, operatorSetId uint32, signingKeyAddress common.Address) ([32]byte, error) {
	args := m.Called(ctx, operatorAddress, avsAddress, operatorSetId, signingKeyAddress)
	return args.Get(0).([32]byte), args.Error(1)
}

func (m *MockContractCaller) ConfigureAVSOperatorSet(ctx context.Context, avsAddress common.Address, operatorSetId uint32, curveType config.CurveType) (*ethereumTypes.Receipt, error) {
	args := m.Called(ctx, avsAddress, operatorSetId, curveType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ethereumTypes.Receipt), args.Error(1)
}

func (m *MockContractCaller) RegisterKeyWithKeyRegistrar(ctx context.Context, operatorAddress common.Address, avsAddress common.Address, operatorSetId uint32, sigBytes []byte, keyData []byte) (*ethereumTypes.Receipt, error) {
	args := m.Called(ctx, operatorAddress, avsAddress, operatorSetId, sigBytes, keyData)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ethereumTypes.Receipt), args.Error(1)
}

func (m *MockContractCaller) CreateOperatorAndRegisterWithAvs(ctx context.Context, avsAddress common.Address, operatorAddress common.Address, operatorSetIds []uint32, socket string, allocationDelay uint32, metadataUri string) (*ethereumTypes.Receipt, error) {
	args := m.Called(ctx, avsAddress, operatorAddress, operatorSetIds, socket, allocationDelay, metadataUri)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ethereumTypes.Receipt), args.Error(1)
}

func (m *MockContractCaller) EncodeBN254KeyData(pubKey *bn254.PublicKey) ([]byte, error) {
	args := m.Called(pubKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockContractCaller) SetupTaskMailboxForAvs(ctx context.Context, avsAddress common.Address, taskHookAddress common.Address, executorOperatorSetIds []uint32, curveTypes []config.CurveType) error {
	args := m.Called(ctx, avsAddress, taskHookAddress, executorOperatorSetIds, curveTypes)
	return args.Error(0)
}

// MockPeeringDataFetcher is a mock implementation of IPeeringDataFetcher
type MockPeeringDataFetcher struct {
	mock.Mock
}

func (m *MockPeeringDataFetcher) ListExecutorOperators(ctx context.Context, avsAddress string) ([]*peering.OperatorPeerInfo, error) {
	args := m.Called(ctx, avsAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*peering.OperatorPeerInfo), args.Error(1)
}

func (m *MockPeeringDataFetcher) ListAggregatorOperators(ctx context.Context, avsAddress string) ([]*peering.OperatorPeerInfo, error) {
	args := m.Called(ctx, avsAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*peering.OperatorPeerInfo), args.Error(1)
}

// Helper function to create test configuration
func createTestConfig() *OperatorManagerConfig {
	return &OperatorManagerConfig{
		AvsAddress:     "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
		OperatorSetIds: []uint32{1, 2},
		ChainIds:       []config.ChainId{1, 10},
		L1ChainId:      1,
	}
}

// Helper function to create test operator peer info
func createTestOperatorPeerInfo(address string, operatorSetIds []uint32) *peering.OperatorPeerInfo {
	operatorSets := make([]*peering.OperatorSet, len(operatorSetIds))
	for i, id := range operatorSetIds {
		operatorSets[i] = &peering.OperatorSet{
			OperatorSetID:  id,
			NetworkAddress: fmt.Sprintf("http://operator-%d:8080", id),
			CurveType:      config.CurveTypeBN254,
		}
	}
	return &peering.OperatorPeerInfo{
		OperatorAddress: address,
		OperatorSets:    operatorSets,
	}
}

func TestNewOperatorManager(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := createTestConfig()
	
	mockL1CC := new(MockContractCaller)
	mockL2CC := new(MockContractCaller)
	contractCallers := map[config.ChainId]contractCaller.IContractCaller{
		1:  mockL1CC,
		10: mockL2CC,
	}
	
	mockPDF := new(MockPeeringDataFetcher)
	
	om := NewOperatorManager(cfg, contractCallers, mockPDF, logger)
	
	assert.NotNil(t, om)
	assert.Equal(t, cfg, om.config)
	assert.Equal(t, contractCallers, om.contractCallers)
	assert.Equal(t, mockPDF, om.peeringDataFetcher)
	assert.Equal(t, logger, om.logger)
}

func TestGetCurveTypeForOperatorSet(t *testing.T) {
	tests := []struct {
		name          string
		avsAddress    string
		operatorSetId uint32
		setupMocks    func(*MockContractCaller)
		expectedCurve config.CurveType
		expectedError bool
		errorContains string
	}{
		{
			name:          "successful curve type retrieval",
			avsAddress:    "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
			operatorSetId: 1,
			setupMocks: func(mockCC *MockContractCaller) {
				mockCC.On("GetOperatorSetCurveType", "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0", uint32(1)).
					Return(config.CurveTypeBN254, nil)
			},
			expectedCurve: config.CurveTypeBN254,
			expectedError: false,
		},
		{
			name:          "contract caller not found for L1 chain",
			avsAddress:    "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
			operatorSetId: 1,
			setupMocks:    func(mockCC *MockContractCaller) {},
			expectedCurve: config.CurveTypeUnknown,
			expectedError: true,
			errorContains: "no contract caller found",
		},
		{
			name:          "error from contract caller",
			avsAddress:    "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
			operatorSetId: 1,
			setupMocks: func(mockCC *MockContractCaller) {
				mockCC.On("GetOperatorSetCurveType", "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0", uint32(1)).
					Return(config.CurveTypeUnknown, errors.New("contract error"))
			},
			expectedCurve: config.CurveTypeUnknown,
			expectedError: true,
			errorContains: "contract error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			cfg := createTestConfig()
			
			mockL1CC := new(MockContractCaller)
			tt.setupMocks(mockL1CC)
			
			contractCallers := map[config.ChainId]contractCaller.IContractCaller{}
			if tt.name != "contract caller not found for L1 chain" {
				contractCallers[1] = mockL1CC
			}
			
			mockPDF := new(MockPeeringDataFetcher)
			om := NewOperatorManager(cfg, contractCallers, mockPDF, logger)
			
			ctx := context.Background()
			curveType, err := om.GetCurveTypeForOperatorSet(ctx, tt.avsAddress, tt.operatorSetId)
			
			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCurve, curveType)
			}
			
			mockL1CC.AssertExpectations(t)
		})
	}
}

func TestGetExecutorPeersAndWeightsForBlock_L1Chain(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := createTestConfig()
	
	mockL1CC := new(MockContractCaller)
	contractCallers := map[config.ChainId]contractCaller.IContractCaller{
		1: mockL1CC,
	}
	
	mockPDF := new(MockPeeringDataFetcher)
	om := NewOperatorManager(cfg, contractCallers, mockPDF, logger)
	
	ctx := context.Background()
	chainId := config.ChainId(1) // L1
	taskBlockNumber := uint64(1000)
	operatorSetId := uint32(1)
	
	// Setup mock expectations
	supportedChains := []*big.Int{big.NewInt(1), big.NewInt(10)}
	tableUpdaterAddresses := []common.Address{
		common.HexToAddress("0x1111111111111111111111111111111111111111"),
		common.HexToAddress("0x2222222222222222222222222222222222222222"),
	}
	
	mockL1CC.On("GetSupportedChainsForMultichain", ctx, int64(taskBlockNumber)).
		Return(supportedChains, tableUpdaterAddresses, nil)
	
	// For L1, GetTableUpdaterReferenceTimeAndBlock is still called
	latestRefTimeAndBlock := &contractCaller.LatestReferenceTimeAndBlock{
		LatestReferenceTimestamp:   12345,
		LatestReferenceBlockNumber: 999,
	}
	
	mockL1CC.On("GetTableUpdaterReferenceTimeAndBlock", ctx, tableUpdaterAddresses[0], taskBlockNumber).
		Return(latestRefTimeAndBlock, nil)
	
	operatorAddresses := []common.Address{
		common.HexToAddress("0xaaa1111111111111111111111111111111111111"),
		common.HexToAddress("0xbbb2222222222222222222222222222222222222"),
	}
	operatorWeights := [][]*big.Int{
		{big.NewInt(100), big.NewInt(200)},
		{big.NewInt(150), big.NewInt(250)},
	}
	
	tableData := &contractCaller.OperatorTableData{
		Operators:                operatorAddresses,
		OperatorWeights:          operatorWeights,
		LatestReferenceTimestamp: 12345,
		LatestReferenceBlockNumber: 999,
	}
	
	mockL1CC.On("GetOperatorTableDataForOperatorSet", ctx, common.HexToAddress(cfg.AvsAddress), operatorSetId, cfg.L1ChainId, taskBlockNumber).
		Return(tableData, nil)
	
	mockL1CC.On("GetOperatorSetCurveType", cfg.AvsAddress, operatorSetId).
		Return(config.CurveTypeBN254, nil)
	
	operators := []*peering.OperatorPeerInfo{
		createTestOperatorPeerInfo("0xaaa1111111111111111111111111111111111111", []uint32{1}),
		createTestOperatorPeerInfo("0xbbb2222222222222222222222222222222222222", []uint32{1}),
		createTestOperatorPeerInfo("0xccc3333333333333333333333333333333333333", []uint32{2}), // Different operator set
	}
	
	mockPDF.On("ListExecutorOperators", ctx, cfg.AvsAddress).Return(operators, nil)
	
	result, err := om.GetExecutorPeersAndWeightsForBlock(ctx, chainId, taskBlockNumber, operatorSetId)
	
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, taskBlockNumber, result.BlockNumber)
	assert.Equal(t, chainId, result.ChainId)
	assert.Equal(t, operatorSetId, result.OperatorSetId)
	assert.Equal(t, uint32(12345), result.RootReferenceTimestamp) // For L1, uses tableData's timestamp
	assert.Equal(t, config.CurveTypeBN254, result.CurveType)
	assert.Len(t, result.Operators, 2) // Only operators in the correct operator set
	assert.Len(t, result.Weights, 2)
	
	mockL1CC.AssertExpectations(t)
	mockPDF.AssertExpectations(t)
}

func TestGetExecutorPeersAndWeightsForBlock_L2Chain(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := createTestConfig()
	
	mockL1CC := new(MockContractCaller)
	mockL2CC := new(MockContractCaller)
	contractCallers := map[config.ChainId]contractCaller.IContractCaller{
		1:  mockL1CC,
		10: mockL2CC,
	}
	
	mockPDF := new(MockPeeringDataFetcher)
	om := NewOperatorManager(cfg, contractCallers, mockPDF, logger)
	
	ctx := context.Background()
	chainId := config.ChainId(10) // L2
	taskBlockNumber := uint64(2000)
	operatorSetId := uint32(1)
	
	// Setup mock expectations
	supportedChains := []*big.Int{big.NewInt(1), big.NewInt(10)}
	tableUpdaterAddresses := []common.Address{
		common.HexToAddress("0x1111111111111111111111111111111111111111"),
		common.HexToAddress("0x2222222222222222222222222222222222222222"),
	}
	
	mockL1CC.On("GetSupportedChainsForMultichain", ctx, int64(-1)).
		Return(supportedChains, tableUpdaterAddresses, nil)
	
	latestRefTimeAndBlock := &contractCaller.LatestReferenceTimeAndBlock{
		LatestReferenceTimestamp:   54321,
		LatestReferenceBlockNumber: 1500,
	}
	
	mockL2CC.On("GetTableUpdaterReferenceTimeAndBlock", ctx, tableUpdaterAddresses[1], taskBlockNumber).
		Return(latestRefTimeAndBlock, nil)
	
	operatorAddresses := []common.Address{
		common.HexToAddress("0xaaa1111111111111111111111111111111111111"),
	}
	operatorWeights := [][]*big.Int{
		{big.NewInt(100), big.NewInt(200)},
	}
	
	tableData := &contractCaller.OperatorTableData{
		Operators:                operatorAddresses,
		OperatorWeights:          operatorWeights,
		LatestReferenceTimestamp: 12345,
		LatestReferenceBlockNumber: 999,
	}
	
	mockL1CC.On("GetOperatorTableDataForOperatorSet", ctx, common.HexToAddress(cfg.AvsAddress), operatorSetId, cfg.L1ChainId, uint64(1500)).
		Return(tableData, nil)
	
	mockL1CC.On("GetOperatorSetCurveType", cfg.AvsAddress, operatorSetId).
		Return(config.CurveTypeECDSA, nil)
	
	operators := []*peering.OperatorPeerInfo{
		createTestOperatorPeerInfo("0xaaa1111111111111111111111111111111111111", []uint32{1}),
	}
	
	mockPDF.On("ListExecutorOperators", ctx, cfg.AvsAddress).Return(operators, nil)
	
	result, err := om.GetExecutorPeersAndWeightsForBlock(ctx, chainId, taskBlockNumber, operatorSetId)
	
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, taskBlockNumber, result.BlockNumber)
	assert.Equal(t, chainId, result.ChainId)
	assert.Equal(t, operatorSetId, result.OperatorSetId)
	assert.Equal(t, uint32(54321), result.RootReferenceTimestamp) // From L2's latest reference
	assert.Equal(t, config.CurveTypeECDSA, result.CurveType)
	assert.Len(t, result.Operators, 1)
	assert.Len(t, result.Weights, 1)
	
	mockL1CC.AssertExpectations(t)
	mockL2CC.AssertExpectations(t)
	mockPDF.AssertExpectations(t)
}

func TestGetExecutorPeersAndWeightsForBlock_Errors(t *testing.T) {
	tests := []struct {
		name          string
		chainId       config.ChainId
		setupMocks    func(*MockContractCaller, *MockContractCaller, *MockPeeringDataFetcher)
		errorContains string
	}{
		{
			name:    "no L1 contract caller",
			chainId: 1,
			setupMocks: func(mockL1CC, mockL2CC *MockContractCaller, mockPDF *MockPeeringDataFetcher) {
				// Don't setup any mocks - the L1 contract caller will be missing
			},
			errorContains: "no contract caller found",
		},
		{
			name:    "no L2 contract caller",
			chainId: 10,
			setupMocks: func(mockL1CC, mockL2CC *MockContractCaller, mockPDF *MockPeeringDataFetcher) {
				// Don't set up any mocks - the L2 contract caller will be missing
			},
			errorContains: "no contract caller found",
		},
		{
			name:    "error getting supported chains",
			chainId: 1,
			setupMocks: func(mockL1CC, mockL2CC *MockContractCaller, mockPDF *MockPeeringDataFetcher) {
				mockL1CC.On("GetSupportedChainsForMultichain", mock.Anything, mock.Anything).
					Return(nil, nil, errors.New("blockchain error"))
			},
			errorContains: "blockchain error",
		},
		{
			name:    "no table updater for chain",
			chainId: 10,
			setupMocks: func(mockL1CC, mockL2CC *MockContractCaller, mockPDF *MockPeeringDataFetcher) {
				mockL1CC.On("GetSupportedChainsForMultichain", mock.Anything, mock.Anything).
					Return([]*big.Int{big.NewInt(1)}, []common.Address{{}}, nil) // Chain 10 not in list
			},
			errorContains: "no table updater address found",
		},
		{
			name:    "error getting table updater reference",
			chainId: 10,
			setupMocks: func(mockL1CC, mockL2CC *MockContractCaller, mockPDF *MockPeeringDataFetcher) {
				mockL1CC.On("GetSupportedChainsForMultichain", mock.Anything, mock.Anything).
					Return([]*big.Int{big.NewInt(10)}, []common.Address{common.HexToAddress("0x1111111111111111111111111111111111111111")}, nil)
				mockL2CC.On("GetTableUpdaterReferenceTimeAndBlock", mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("reference error"))
			},
			errorContains: "reference error",
		},
		{
			name:    "error getting operator table data",
			chainId: 1,
			setupMocks: func(mockL1CC, mockL2CC *MockContractCaller, mockPDF *MockPeeringDataFetcher) {
				mockL1CC.On("GetSupportedChainsForMultichain", mock.Anything, mock.Anything).
					Return([]*big.Int{big.NewInt(1)}, []common.Address{common.HexToAddress("0x1111111111111111111111111111111111111111")}, nil)
				
				latestRefTimeAndBlock := &contractCaller.LatestReferenceTimeAndBlock{
					LatestReferenceTimestamp:   12345,
					LatestReferenceBlockNumber: 999,
				}
				mockL1CC.On("GetTableUpdaterReferenceTimeAndBlock", mock.Anything, mock.Anything, mock.Anything).
					Return(latestRefTimeAndBlock, nil)
				
				mockL1CC.On("GetOperatorTableDataForOperatorSet", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("table data error"))
			},
			errorContains: "table data error",
		},
		{
			name:    "error listing executor operators",
			chainId: 1,
			setupMocks: func(mockL1CC, mockL2CC *MockContractCaller, mockPDF *MockPeeringDataFetcher) {
				mockL1CC.On("GetSupportedChainsForMultichain", mock.Anything, mock.Anything).
					Return([]*big.Int{big.NewInt(1)}, []common.Address{common.HexToAddress("0x1111111111111111111111111111111111111111")}, nil)
				
				latestRefTimeAndBlock := &contractCaller.LatestReferenceTimeAndBlock{
					LatestReferenceTimestamp:   12345,
					LatestReferenceBlockNumber: 999,
				}
				mockL1CC.On("GetTableUpdaterReferenceTimeAndBlock", mock.Anything, mock.Anything, mock.Anything).
					Return(latestRefTimeAndBlock, nil)
				
				tableData := &contractCaller.OperatorTableData{
					Operators:       []common.Address{common.HexToAddress("0xaaa1111111111111111111111111111111111111")},
					OperatorWeights: [][]*big.Int{{big.NewInt(100)}},
				}
				mockL1CC.On("GetOperatorTableDataForOperatorSet", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(tableData, nil)
				
				mockPDF.On("ListExecutorOperators", mock.Anything, mock.Anything).
					Return(nil, errors.New("peering error"))
			},
			errorContains: "failed to list executor Operators",
		},
		{
			name:    "error getting curve type",
			chainId: 1,
			setupMocks: func(mockL1CC, mockL2CC *MockContractCaller, mockPDF *MockPeeringDataFetcher) {
				mockL1CC.On("GetSupportedChainsForMultichain", mock.Anything, mock.Anything).
					Return([]*big.Int{big.NewInt(1)}, []common.Address{common.HexToAddress("0x1111111111111111111111111111111111111111")}, nil)
				
				latestRefTimeAndBlock := &contractCaller.LatestReferenceTimeAndBlock{
					LatestReferenceTimestamp:   12345,
					LatestReferenceBlockNumber: 999,
				}
				mockL1CC.On("GetTableUpdaterReferenceTimeAndBlock", mock.Anything, mock.Anything, mock.Anything).
					Return(latestRefTimeAndBlock, nil)
				
				tableData := &contractCaller.OperatorTableData{
					Operators:       []common.Address{common.HexToAddress("0xaaa1111111111111111111111111111111111111")},
					OperatorWeights: [][]*big.Int{{big.NewInt(100)}},
				}
				mockL1CC.On("GetOperatorTableDataForOperatorSet", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(tableData, nil)
				
				mockPDF.On("ListExecutorOperators", mock.Anything, mock.Anything).
					Return([]*peering.OperatorPeerInfo{}, nil)
				
				mockL1CC.On("GetOperatorSetCurveType", mock.Anything, mock.Anything).
					Return(config.CurveTypeUnknown, errors.New("curve type error"))
			},
			errorContains: "failed to get operator set curve type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			cfg := createTestConfig()
			
			mockL1CC := new(MockContractCaller)
			mockL2CC := new(MockContractCaller)
			mockPDF := new(MockPeeringDataFetcher)
			
			tt.setupMocks(mockL1CC, mockL2CC, mockPDF)
			
			contractCallers := map[config.ChainId]contractCaller.IContractCaller{}
			if tt.name != "no L1 contract caller" {
				contractCallers[1] = mockL1CC
			}
			if tt.name != "no L2 contract caller" {
				contractCallers[10] = mockL2CC
			}
			
			om := NewOperatorManager(cfg, contractCallers, mockPDF, logger)
			
			ctx := context.Background()
			result, err := om.GetExecutorPeersAndWeightsForBlock(ctx, tt.chainId, 1000, 1)
			
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), tt.errorContains)
			
			mockL1CC.AssertExpectations(t)
			mockL2CC.AssertExpectations(t)
			mockPDF.AssertExpectations(t)
		})
	}
}

func TestGetContractCallerForChainId(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := createTestConfig()
	
	mockL1CC := new(MockContractCaller)
	mockL2CC := new(MockContractCaller)
	contractCallers := map[config.ChainId]contractCaller.IContractCaller{
		1:  mockL1CC,
		10: mockL2CC,
	}
	
	mockPDF := new(MockPeeringDataFetcher)
	om := NewOperatorManager(cfg, contractCallers, mockPDF, logger)
	
	t.Run("existing chain id", func(t *testing.T) {
		cc, err := om.getContractCallerForChainId(1)
		assert.NoError(t, err)
		assert.Equal(t, mockL1CC, cc)
		
		cc, err = om.getContractCallerForChainId(10)
		assert.NoError(t, err)
		assert.Equal(t, mockL2CC, cc)
	})
	
	t.Run("non-existing chain id", func(t *testing.T) {
		cc, err := om.getContractCallerForChainId(100)
		assert.Error(t, err)
		assert.Nil(t, cc)
		assert.Contains(t, err.Error(), "no contract caller found for chain ID 100")
	})
}

func TestOperatorFiltering(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := createTestConfig()
	
	mockL1CC := new(MockContractCaller)
	contractCallers := map[config.ChainId]contractCaller.IContractCaller{
		1: mockL1CC,
	}
	
	mockPDF := new(MockPeeringDataFetcher)
	om := NewOperatorManager(cfg, contractCallers, mockPDF, logger)
	
	ctx := context.Background()
	chainId := config.ChainId(1)
	taskBlockNumber := uint64(1000)
	operatorSetId := uint32(1)
	
	// Setup mock expectations
	supportedChains := []*big.Int{big.NewInt(1)}
	tableUpdaterAddresses := []common.Address{
		common.HexToAddress("0x1111111111111111111111111111111111111111"),
	}
	
	mockL1CC.On("GetSupportedChainsForMultichain", ctx, int64(taskBlockNumber)).
		Return(supportedChains, tableUpdaterAddresses, nil)
	
	// Add mock for GetTableUpdaterReferenceTimeAndBlock
	latestRefTimeAndBlock := &contractCaller.LatestReferenceTimeAndBlock{
		LatestReferenceTimestamp:   12345,
		LatestReferenceBlockNumber: 999,
	}
	mockL1CC.On("GetTableUpdaterReferenceTimeAndBlock", ctx, tableUpdaterAddresses[0], taskBlockNumber).
		Return(latestRefTimeAndBlock, nil)
	
	// Create operator addresses with mixed case to test case-insensitive matching
	operatorAddresses := []common.Address{
		common.HexToAddress("0xAAA1111111111111111111111111111111111111"), // Upper case
		common.HexToAddress("0xbbb2222222222222222222222222222222222222"), // Lower case
		common.HexToAddress("0xCcC3333333333333333333333333333333333333"), // Mixed case
	}
	
	operatorWeights := [][]*big.Int{
		{big.NewInt(100)},
		{big.NewInt(200)},
		{big.NewInt(300)},
	}
	
	tableData := &contractCaller.OperatorTableData{
		Operators:                operatorAddresses,
		OperatorWeights:          operatorWeights,
		LatestReferenceTimestamp: 12345,
	}
	
	mockL1CC.On("GetOperatorTableDataForOperatorSet", ctx, common.HexToAddress(cfg.AvsAddress), operatorSetId, cfg.L1ChainId, taskBlockNumber).
		Return(tableData, nil)
	
	mockL1CC.On("GetOperatorSetCurveType", cfg.AvsAddress, operatorSetId).
		Return(config.CurveTypeBN254, nil)
	
	// Create operators with different cases and operator sets
	operators := []*peering.OperatorPeerInfo{
		createTestOperatorPeerInfo("0xaaa1111111111111111111111111111111111111", []uint32{1}), // Matches first operator (case insensitive)
		createTestOperatorPeerInfo("0xBBB2222222222222222222222222222222222222", []uint32{1}), // Matches second operator (case insensitive)
		createTestOperatorPeerInfo("0xccc3333333333333333333333333333333333333", []uint32{2}), // Has weight but wrong operator set
		createTestOperatorPeerInfo("0xddd4444444444444444444444444444444444444", []uint32{1}), // In operator set but no weight
	}
	
	mockPDF.On("ListExecutorOperators", ctx, cfg.AvsAddress).Return(operators, nil)
	
	result, err := om.GetExecutorPeersAndWeightsForBlock(ctx, chainId, taskBlockNumber, operatorSetId)
	
	require.NoError(t, err)
	assert.NotNil(t, result)
	// Should only include operators that have weights AND are in the correct operator set
	assert.Len(t, result.Operators, 2)
	assert.Len(t, result.Weights, 3) // All operators with weights
	
	// Verify the filtered operators are the correct ones
	operatorAddrs := make([]string, len(result.Operators))
	for i, op := range result.Operators {
		operatorAddrs[i] = op.OperatorAddress
	}
	assert.Contains(t, operatorAddrs, "0xaaa1111111111111111111111111111111111111")
	assert.Contains(t, operatorAddrs, "0xBBB2222222222222222222222222222222222222")
	
	mockL1CC.AssertExpectations(t)
	mockPDF.AssertExpectations(t)
}
package web3Signer

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/Layr-Labs/crypto-libs/pkg/ecdsa"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer/inMemorySigner"
	"github.com/ethereum/go-ethereum/crypto"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/internal/testUtils"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/web3signer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWeb3Signer(t *testing.T) {
	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	require.NoError(t, err)

	cfg := web3signer.DefaultConfig()
	client, err := web3signer.NewClient(cfg, l)
	require.NoError(t, err)

	fromAddress := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")

	t.Run("successful creation with ECDSA", func(t *testing.T) {
		publicKey := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef12"
		signer, err := NewWeb3Signer(client, fromAddress, publicKey, config.CurveTypeECDSA, l)
		require.NoError(t, err)
		assert.NotNil(t, signer)

		web3Signer, ok := signer.(*Web3Signer)
		require.True(t, ok)
		assert.Equal(t, client, web3Signer.client)
		assert.Equal(t, fromAddress, web3Signer.fromAddress)
		assert.Equal(t, "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef12", web3Signer.publicKey)
		assert.Equal(t, config.CurveTypeECDSA, web3Signer.curveType)
		assert.Equal(t, l, web3Signer.logger)
	})

	t.Run("fails with BN254 curve type", func(t *testing.T) {
		publicKey := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef12"
		signer, err := NewWeb3Signer(client, fromAddress, publicKey, config.CurveTypeBN254, l)
		assert.Error(t, err)
		assert.Nil(t, signer)
		assert.Contains(t, err.Error(), "web3signer only supports ECDSA curve type")
	})

	t.Run("fails with nil client", func(t *testing.T) {
		publicKey := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef12"
		signer, err := NewWeb3Signer(nil, fromAddress, publicKey, config.CurveTypeECDSA, l)
		assert.Error(t, err)
		assert.Nil(t, signer)
		assert.Contains(t, err.Error(), "web3signer client cannot be nil")
	})

	t.Run("fails with nil logger", func(t *testing.T) {
		publicKey := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef12"
		signer, err := NewWeb3Signer(client, fromAddress, publicKey, config.CurveTypeECDSA, nil)
		assert.Error(t, err)
		assert.Nil(t, signer)
		assert.Contains(t, err.Error(), "logger cannot be nil")
	})

	t.Run("fails with empty public key", func(t *testing.T) {
		signer, err := NewWeb3Signer(client, fromAddress, "", config.CurveTypeECDSA, l)
		assert.Error(t, err)
		assert.Nil(t, signer)
		assert.Contains(t, err.Error(), "publicKey cannot be empty")
	})
}

func TestWeb3Signer_SignMessage(t *testing.T) {
	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	require.NoError(t, err)

	fromAddress := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	testData := []byte("Hello, Web3Signer!")
	expectedSignature := "0xb3baa751d0a9132cfe93e4e3d5ff9075111100e3789dca219ade5a24d27e19d16b3353149da1833e9b691bb38634e8dc04469be7032132906c927d7e1a49b414730612877bc6b2810c8f202daf793d1ab0d6b5cb21d52f9e52e883859887a5d9"

	t.Run("successful signing", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Equal(t, "/api/v1/eth1/sign/1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef12", r.URL.Path)

			var payload map[string]interface{}
			err := json.NewDecoder(r.Body).Decode(&payload)
			require.NoError(t, err)

			expectedDataHex := "0x" + hex.EncodeToString(testData)
			assert.Equal(t, expectedDataHex, payload["data"])

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(expectedSignature)
		}))
		defer server.Close()

		cfg := web3signer.DefaultConfig()
		cfg.BaseURL = server.URL
		client, err := web3signer.NewClient(cfg, l)
		require.NoError(t, err)

		signer, err := NewWeb3Signer(client, fromAddress, "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef12", config.CurveTypeECDSA, l)
		require.NoError(t, err)

		signature, err := signer.SignMessage(testData)
		require.NoError(t, err)

		expectedBytes, err := hex.DecodeString(expectedSignature[2:]) // Remove 0x prefix
		require.NoError(t, err)
		assert.Equal(t, expectedBytes, signature)
	})

	t.Run("fails with BN254 curve type", func(t *testing.T) {
		cfg := web3signer.DefaultConfig()
		client, err := web3signer.NewClient(cfg, l)
		require.NoError(t, err)

		// Create signer with BN254 (bypassing constructor validation for test)
		web3Signer := &Web3Signer{
			client:      client,
			fromAddress: fromAddress,
			publicKey:   "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef12",
			curveType:   config.CurveTypeBN254,
			logger:      l,
		}

		signature, err := web3Signer.SignMessage(testData)
		assert.Error(t, err)
		assert.Nil(t, signature)
		assert.Contains(t, err.Error(), "web3signer only supports ECDSA curve type")
	})

	t.Run("web3signer error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("Account not found"))
		}))
		defer server.Close()

		cfg := web3signer.DefaultConfig()
		cfg.BaseURL = server.URL
		client, err := web3signer.NewClient(cfg, l)
		require.NoError(t, err)

		signer, err := NewWeb3Signer(client, fromAddress, "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef12", config.CurveTypeECDSA, l)
		require.NoError(t, err)

		signature, err := signer.SignMessage(testData)
		assert.Error(t, err)
		assert.Nil(t, signature)
		assert.Contains(t, err.Error(), "failed to sign message with web3signer")

		var web3SignerErr *web3signer.Web3SignerError
		assert.ErrorAs(t, err, &web3SignerErr)
		assert.Equal(t, 400, web3SignerErr.Code)
	})

	t.Run("invalid signature response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode("invalid-hex-signature")
		}))
		defer server.Close()

		cfg := web3signer.DefaultConfig()
		cfg.BaseURL = server.URL
		client, err := web3signer.NewClient(cfg, l)
		require.NoError(t, err)

		signer, err := NewWeb3Signer(client, fromAddress, "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef12", config.CurveTypeECDSA, l)
		require.NoError(t, err)

		signature, err := signer.SignMessage(testData)
		assert.Error(t, err)
		assert.Nil(t, signature)
		assert.Contains(t, err.Error(), "failed to decode signature from web3signer")
	})
}

func TestWeb3Signer_SignMessageForSolidity(t *testing.T) {
	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	require.NoError(t, err)

	fromAddress := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	testData := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20}
	expectedSignature := "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890ab"

	t.Run("successful signing for solidity", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Equal(t, "/api/v1/eth1/sign/1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef12", r.URL.Path)

			var payload map[string]interface{}
			err := json.NewDecoder(r.Body).Decode(&payload)
			require.NoError(t, err)

			expectedDataHex := "0x" + hex.EncodeToString(testData[:])
			assert.Equal(t, expectedDataHex, payload["data"])

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(expectedSignature)
		}))
		defer server.Close()

		cfg := web3signer.DefaultConfig()
		cfg.BaseURL = server.URL
		client, err := web3signer.NewClient(cfg, l)
		require.NoError(t, err)

		signer, err := NewWeb3Signer(client, fromAddress, "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef12", config.CurveTypeECDSA, l)
		require.NoError(t, err)

		signature, err := signer.SignMessageForSolidity(testData)
		require.NoError(t, err)

		expectedBytes, err := hex.DecodeString(expectedSignature[2:]) // Remove 0x prefix
		require.NoError(t, err)
		assert.Equal(t, expectedBytes, signature)
	})

	t.Run("fails with BN254 curve type", func(t *testing.T) {
		cfg := web3signer.DefaultConfig()
		client, err := web3signer.NewClient(cfg, l)
		require.NoError(t, err)

		// Create signer with BN254 (bypassing constructor validation for test)
		web3Signer := &Web3Signer{
			client:      client,
			fromAddress: fromAddress,
			publicKey:   "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef12",
			curveType:   config.CurveTypeBN254,
			logger:      l,
		}

		signature, err := web3Signer.SignMessageForSolidity(testData)
		assert.Error(t, err)
		assert.Nil(t, signature)
		assert.Contains(t, err.Error(), "web3signer only supports ECDSA curve type")
	})
}

func TestWeb3Signer_GetFromAddress(t *testing.T) {
	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	require.NoError(t, err)

	cfg := web3signer.DefaultConfig()
	client, err := web3signer.NewClient(cfg, l)
	require.NoError(t, err)

	fromAddress := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")

	signer, err := NewWeb3Signer(client, fromAddress, "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef12", config.CurveTypeECDSA, l)
	require.NoError(t, err)

	web3Signer, ok := signer.(*Web3Signer)
	require.True(t, ok)

	assert.Equal(t, fromAddress, web3Signer.GetFromAddress())
}

func TestWeb3Signer_GetCurveType(t *testing.T) {
	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	require.NoError(t, err)

	cfg := web3signer.DefaultConfig()
	client, err := web3signer.NewClient(cfg, l)
	require.NoError(t, err)

	fromAddress := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")

	signer, err := NewWeb3Signer(client, fromAddress, "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef12", config.CurveTypeECDSA, l)
	require.NoError(t, err)

	web3Signer, ok := signer.(*Web3Signer)
	require.True(t, ok)

	assert.Equal(t, config.CurveTypeECDSA, web3Signer.GetCurveType())
}

func TestWeb3Signer_SupportsRemoteSigning(t *testing.T) {
	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	require.NoError(t, err)

	cfg := web3signer.DefaultConfig()
	client, err := web3signer.NewClient(cfg, l)
	require.NoError(t, err)

	fromAddress := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")

	signer, err := NewWeb3Signer(client, fromAddress, "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef12", config.CurveTypeECDSA, l)
	require.NoError(t, err)

	web3Signer, ok := signer.(*Web3Signer)
	require.True(t, ok)

	assert.True(t, web3Signer.SupportsRemoteSigning())
}

func TestWeb3Signer_Validate(t *testing.T) {
	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	require.NoError(t, err)

	fromAddress := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")

	t.Run("successful validation when account exists", func(t *testing.T) {
		accounts := []string{
			"0x9876543210fedcba9876543210fedcba98765432",
			fromAddress.Hex(),
			"0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req web3signer.JSONRPCRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			assert.Equal(t, "eth_accounts", req.Method)

			response := web3signer.JSONRPCResponse{
				Jsonrpc: "2.0",
				Result:  accounts,
				ID:      req.ID,
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		cfg := web3signer.DefaultConfig()
		cfg.BaseURL = server.URL
		client, err := web3signer.NewClient(cfg, l)
		require.NoError(t, err)

		signer, err := NewWeb3Signer(client, fromAddress, "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef12", config.CurveTypeECDSA, l)
		require.NoError(t, err)

		web3Signer, ok := signer.(*Web3Signer)
		require.True(t, ok)

		err = web3Signer.Validate()
		assert.NoError(t, err)
	})

	t.Run("fails when account not found", func(t *testing.T) {
		accounts := []string{
			"0x9876543210fedcba9876543210fedcba98765432",
			"0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req web3signer.JSONRPCRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			assert.Equal(t, "eth_accounts", req.Method)

			response := web3signer.JSONRPCResponse{
				Jsonrpc: "2.0",
				Result:  accounts,
				ID:      req.ID,
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		cfg := web3signer.DefaultConfig()
		cfg.BaseURL = server.URL
		client, err := web3signer.NewClient(cfg, l)
		require.NoError(t, err)

		signer, err := NewWeb3Signer(client, fromAddress, "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef12", config.CurveTypeECDSA, l)
		require.NoError(t, err)

		web3Signer, ok := signer.(*Web3Signer)
		require.True(t, ok)

		err = web3Signer.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), fmt.Sprintf("signing address %s not found in web3signer accounts", fromAddress.Hex()))
	})

	t.Run("fails when web3signer service error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req web3signer.JSONRPCRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)

			response := web3signer.JSONRPCResponse{
				Jsonrpc: "2.0",
				Error: &web3signer.JSONRPCError{
					Code:    -32000,
					Message: "Internal server error",
				},
				ID: req.ID,
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		cfg := web3signer.DefaultConfig()
		cfg.BaseURL = server.URL
		client, err := web3signer.NewClient(cfg, l)
		require.NoError(t, err)

		signer, err := NewWeb3Signer(client, fromAddress, "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef12", config.CurveTypeECDSA, l)
		require.NoError(t, err)

		web3Signer, ok := signer.(*Web3Signer)
		require.True(t, ok)

		err = web3Signer.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to validate web3signer connection")
	})

	t.Run("validates with case-insensitive address comparison", func(t *testing.T) {
		accounts := []string{
			"0x9876543210fedcba9876543210fedcba98765432",
			strings.ToUpper(fromAddress.Hex()), // Different case
			"0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req web3signer.JSONRPCRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			assert.Equal(t, "eth_accounts", req.Method)

			response := web3signer.JSONRPCResponse{
				Jsonrpc: "2.0",
				Result:  accounts,
				ID:      req.ID,
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		cfg := web3signer.DefaultConfig()
		cfg.BaseURL = server.URL
		client, err := web3signer.NewClient(cfg, l)
		require.NoError(t, err)

		signer, err := NewWeb3Signer(client, fromAddress, "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef12", config.CurveTypeECDSA, l)
		require.NoError(t, err)

		web3Signer, ok := signer.(*Web3Signer)
		require.True(t, ok)

		err = web3Signer.Validate()
		assert.NoError(t, err)
	})
}

func TestWeb3Signer_InterfaceCompliance(t *testing.T) {
	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: false})
	require.NoError(t, err)

	cfg := web3signer.DefaultConfig()
	client, err := web3signer.NewClient(cfg, l)
	require.NoError(t, err)

	fromAddress := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")

	signerImpl, err := NewWeb3Signer(client, fromAddress, "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef12", config.CurveTypeECDSA, l)
	require.NoError(t, err)

	// Verify that Web3Signer implements the ISigner interface
	var _ signer.ISigner = signerImpl
	assert.NotNil(t, signerImpl)
}

func TestWeb3Signer_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: true})
	require.NoError(t, err)

	// Get project root using testUtils function
	projectRoot := testUtils.GetProjectRootPath()

	// Read chain config to get test addresses
	chainConfig, err := testUtils.ReadChainConfig(projectRoot)
	require.NoError(t, err)

	// Configure web3signer client to connect to test server
	cfg := web3signer.DefaultConfig()
	cfg.BaseURL = testUtils.L1Web3SignerUrl // http://localhost:9100
	client, err := web3signer.NewClient(cfg, l)
	require.NoError(t, err)

	// Use the operator account address from test config
	fromAddress := common.HexToAddress(chainConfig.OperatorAccountAddress)
	publicKey := chainConfig.OperatorAccountPublicKey

	t.Run("test actual web3signer REST API signing", func(t *testing.T) {
		// Create Web3Signer instance
		signer, err := NewWeb3Signer(client, fromAddress, publicKey, config.CurveTypeECDSA, l)
		require.NoError(t, err)

		// Test data to sign
		testData := []byte("Hello from Web3Signer REST API test!")

		payloadHash := crypto.Keccak256Hash(testData)

		// Perform signing using REST API
		signature, err := signer.SignMessage(testData)

		// This might fail with 404 if the REST API endpoint is not correct
		// or if the server doesn't support the endpoint we're trying to use
		if err != nil {
			t.Logf("Web3Signer signing failed with error: %v", err)
			// Check if it's a 404 error which would indicate wrong endpoint
			if strings.Contains(err.Error(), "404") {
				t.Errorf("Got 404 error - this indicates the REST API endpoint '/api/v1/eth1/sign/%s' is not found on the server", fromAddress.Hex())
			}
			t.FailNow()
		}

		typedSig, err := ecdsa.NewSignatureFromBytes(signature)
		assert.Nil(t, err, "Failed to create ECDSA signature from bytes")

		verified, err := typedSig.VerifyWithAddress(payloadHash[:], common.HexToAddress(chainConfig.OperatorAccountAddress))
		assert.Nil(t, err, "Signature verification failed")
		assert.True(t, verified, "Signature should be valid for the given data and address")

		// Verify we got a signature
		assert.NotNil(t, signature)
		assert.Greater(t, len(signature), 0)
		t.Logf("Successfully signed data with Web3Signer REST API. Signature length: %d bytes", len(signature))
		t.Logf("Signature: 0x%x", signature)
	})

	t.Run("test SignMessageForSolidity", func(t *testing.T) {
		// Create Web3Signer instance
		signer, err := NewWeb3Signer(client, fromAddress, publicKey, config.CurveTypeECDSA, l)
		require.NoError(t, err)

		// Perform signing for Solidity
		signature, err := signer.SignMessageForSolidity([]byte("Solidity test data"))

		if err != nil {
			t.Logf("Web3Signer SignMessageForSolidity failed with error: %v", err)
			if strings.Contains(err.Error(), "404") {
				t.Errorf("Got 404 error - this indicates the REST API endpoint '/api/v1/eth1/sign/%s' is not found on the server", fromAddress.Hex())
			}
			t.FailNow()
		}

		// Verify we got a signature
		assert.NotNil(t, signature)
		assert.Greater(t, len(signature), 0)
		t.Logf("Successfully signed Solidity data with Web3Signer REST API. Signature length: %d bytes", len(signature))
	})

	t.Run("test account validation", func(t *testing.T) {
		// Create Web3Signer instance
		signer, err := NewWeb3Signer(client, fromAddress, publicKey, config.CurveTypeECDSA, l)
		require.NoError(t, err)

		web3Signer, ok := signer.(*Web3Signer)
		require.True(t, ok)

		// Test account validation (this uses JSON-RPC eth_accounts)
		err = web3Signer.Validate()
		if err != nil {
			t.Logf("Account validation failed: %v", err)
			// This could fail if the account is not loaded in web3signer
			assert.Contains(t, err.Error(), "not found in web3signer accounts")
		} else {
			t.Logf("Account validation successful for address: %s", fromAddress.Hex())
		}
	})
}
func Test_CompareWeb3SignerToInMemorySigner(t *testing.T) {
	l, err := logger.NewLogger(&logger.LoggerConfig{Debug: true})
	require.NoError(t, err)

	// Get project root using testUtils function
	projectRoot := testUtils.GetProjectRootPath()

	// Read chain config to get test addresses
	chainConfig, err := testUtils.ReadChainConfig(projectRoot)
	require.NoError(t, err)

	privateKey, err := ecdsa.NewPrivateKeyFromHexString(chainConfig.OperatorAccountPrivateKey)
	require.NoError(t, err)

	ims := inMemorySigner.NewInMemorySigner(privateKey, config.CurveTypeECDSA)

	cfg := web3signer.DefaultConfig()
	cfg.BaseURL = testUtils.L1Web3SignerUrl // http://localhost:9100
	client, err := web3signer.NewClient(cfg, l)
	require.NoError(t, err)

	web3Signer, err := NewWeb3Signer(client, common.HexToAddress(chainConfig.OperatorAccountAddress), chainConfig.OperatorAccountPublicKey, config.CurveTypeECDSA, l)
	require.NoError(t, err)

	payload := []byte("Hello from Web3Signer REST API test!")

	imsSignedHashed, err := ims.SignMessage(payload)
	require.NoError(t, err)

	w3SignedNoHash, err := web3Signer.SignMessage(payload)
	require.NoError(t, err)

	//t.Logf("InMemorySigner signed payload:        %x", imsSignedNoHash)
	t.Logf("InMemorySigner signed hashed payload: %x", imsSignedHashed)
	t.Logf("Web3Signer signed payload:            %x", w3SignedNoHash)

	assert.Equal(t, hex.EncodeToString(imsSignedHashed), hex.EncodeToString(w3SignedNoHash), "Web3Signer and InMemorySigner should produce the same signature for the same payload")
}

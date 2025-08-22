package client

import (
	"context"
	"encoding/json"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	t.Skip("Web3signer is currently experimental")
	l := logger.NewLogger(false)

	t.Run("with default config", func(t *testing.T) {
		client, err := NewWeb3SignerClient(DefaultWeb3SignerConfig(), l)
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.NotNil(t, client.config)
		assert.Equal(t, "http://localhost:9000", client.config.BaseURL)
		assert.Equal(t, 30*time.Second, client.config.Timeout)
	})

	t.Run("with custom config", func(t *testing.T) {
		cfg := &Config{
			BaseURL: "http://custom:8080",
			Timeout: 10 * time.Second,
		}
		client, err := NewWeb3SignerClient(cfg, l)
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, "http://custom:8080", client.config.BaseURL)
		assert.Equal(t, 10*time.Second, client.config.Timeout)
	})

	t.Run("with nil config", func(t *testing.T) {
		client, err := NewWeb3SignerClient(nil, l)
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "cfg cannot be nil")
	})

	t.Run("with nil logger", func(t *testing.T) {
		client, err := NewWeb3SignerClient(DefaultWeb3SignerConfig(), nil)
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "logger cannot be nil")
	})
}

func TestClient_EthAccounts(t *testing.T) {
	t.Skip("Web3signer is currently experimental")
	t.Run("successful accounts request", func(t *testing.T) {
		expectedAccounts := []string{
			"0x1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b",
			"0x2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c",
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			var req JSONRPCRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			assert.Equal(t, "2.0", req.Jsonrpc)
			assert.Equal(t, "eth_accounts", req.Method)
			assert.Nil(t, req.Params)

			response := JSONRPCResponse{
				Jsonrpc: "2.0",
				Result:  expectedAccounts,
				ID:      req.ID,
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		l := logger.NewLogger(false)

		cfg := DefaultWeb3SignerConfig()
		cfg.BaseURL = server.URL
		client, err := NewWeb3SignerClient(cfg, l)
		require.NoError(t, err)

		accounts, err := client.EthAccounts(context.Background())
		require.NoError(t, err)
		assert.Equal(t, expectedAccounts, accounts)
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Internal server error"))
		}))
		defer server.Close()

		l := logger.NewLogger(false)

		cfg := DefaultWeb3SignerConfig()
		cfg.BaseURL = server.URL
		client, err := NewWeb3SignerClient(cfg, l)
		require.NoError(t, err)

		_, err = client.EthAccounts(context.Background())
		require.Error(t, err)

		var web3SignerErr *Web3SignerError
		assert.ErrorAs(t, err, &web3SignerErr)
		assert.Equal(t, 500, web3SignerErr.Code)
	})
}

func TestClient_EthSign(t *testing.T) {
	t.Skip("Web3signer is currently experimental")
	t.Run("successful sign", func(t *testing.T) {
		expectedSignature := "0xb3baa751d0a9132cfe93e4e3d5ff9075111100e3789dca219ade5a24d27e19d16b3353149da1833e9b691bb38634e8dc04469be7032132906c927d7e1a49b414730612877bc6b2810c8f202daf793d1ab0d6b5cb21d52f9e52e883859887a5d9"

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			var req JSONRPCRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			assert.Equal(t, "2.0", req.Jsonrpc)
			assert.Equal(t, "eth_sign", req.Method)

			params, ok := req.Params.([]interface{})
			require.True(t, ok)
			assert.Equal(t, "0x1234567890abcdef", params[0])
			assert.Equal(t, "0x48656c6c6f2c20776f726c6421", params[1])

			response := JSONRPCResponse{
				Jsonrpc: "2.0",
				Result:  expectedSignature,
				ID:      req.ID,
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		l := logger.NewLogger(false)

		cfg := DefaultWeb3SignerConfig()
		cfg.BaseURL = server.URL
		client, err := NewWeb3SignerClient(cfg, l)
		require.NoError(t, err)

		signature, err := client.EthSign(context.Background(), "0x1234567890abcdef", "0x48656c6c6f2c20776f726c6421")
		require.NoError(t, err)
		assert.Equal(t, expectedSignature, signature)
	})

	t.Run("account not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req JSONRPCRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)

			response := JSONRPCResponse{
				Jsonrpc: "2.0",
				Error: &JSONRPCError{
					Code:    -32000,
					Message: "Account not found",
				},
				ID: req.ID,
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		l := logger.NewLogger(false)

		cfg := DefaultWeb3SignerConfig()
		cfg.BaseURL = server.URL
		client, err := NewWeb3SignerClient(cfg, l)
		require.NoError(t, err)

		_, err = client.EthSign(context.Background(), "0x1234567890abcdef", "0x48656c6c6f2c20776f726c6421")
		require.Error(t, err)

		var web3SignerErr *Web3SignerError
		assert.ErrorAs(t, err, &web3SignerErr)
		assert.Equal(t, -32000, web3SignerErr.Code)
	})
}

func TestClient_EthSignTransaction(t *testing.T) {
	t.Skip("Web3signer is currently experimental")
	t.Run("successful transaction sign", func(t *testing.T) {
		expectedSignature := "0xf86c808504a817c800825208943535353535353535353535353535353535353535880de0b6b3a76400008025a028ef61340bd939bc2195fe537567866003e1a15d3c71ff63e1590620aa636276a067cbe9d8997f761aecb703304b3800ccf555c9f3dc64214b297fb1966a3b6d83"

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			var req JSONRPCRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			assert.Equal(t, "2.0", req.Jsonrpc)
			assert.Equal(t, "eth_signTransaction", req.Method)

			params, ok := req.Params.([]interface{})
			require.True(t, ok)
			require.Len(t, params, 1)

			txData, ok := params[0].(map[string]interface{})
			require.True(t, ok)
			assert.Equal(t, "0x1234567890abcdef", txData["from"])
			assert.Equal(t, "0x742d35Cc6634C0532925a3b8D39E1b86D8a10f23", txData["to"])

			response := JSONRPCResponse{
				Jsonrpc: "2.0",
				Result:  expectedSignature,
				ID:      req.ID,
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		l := logger.NewLogger(false)

		cfg := DefaultWeb3SignerConfig()
		cfg.BaseURL = server.URL
		client, err := NewWeb3SignerClient(cfg, l)
		require.NoError(t, err)

		transaction := map[string]interface{}{
			"to":       "0x742d35Cc6634C0532925a3b8D39E1b86D8a10f23",
			"value":    "0x1",
			"gasPrice": "0x9184e72a000",
			"gas":      "0x5208",
			"nonce":    "0x0",
		}

		signature, err := client.EthSignTransaction(context.Background(), "0x1234567890abcdef", transaction)
		require.NoError(t, err)
		assert.Equal(t, expectedSignature, signature)
	})
}

func TestWeb3SignerError_Error(t *testing.T) {
	t.Skip("Web3signer is currently experimental")
	err := &Web3SignerError{
		Code:    -32000,
		Message: "Account not found",
	}

	expected := "Web3SignerClient error -32000: Account not found"
	assert.Equal(t, expected, err.Error())
}

func TestNewConfigWithTLS(t *testing.T) {
	t.Skip("Web3signer is currently experimental")

	t.Run("http url without tls", func(t *testing.T) {
		config := NewConfigWithTLS("http://localhost:9000", "ca-cert", "client-cert", "client-key")
		assert.Equal(t, "http://localhost:9000", config.BaseURL)
		assert.Nil(t, config.TLS, "TLS config should be nil for HTTP URLs")
	})

	t.Run("https url with tls config", func(t *testing.T) {
		config := NewConfigWithTLS("https://localhost:9000", "ca-cert", "client-cert", "client-key")
		assert.Equal(t, "https://localhost:9000", config.BaseURL)
		assert.NotNil(t, config.TLS, "TLS config should be set for HTTPS URLs with TLS parameters")
		assert.Equal(t, "ca-cert", config.TLS.CACert)
		assert.Equal(t, "client-cert", config.TLS.ClientCert)
		assert.Equal(t, "client-key", config.TLS.ClientKey)
	})

	t.Run("https url without tls config", func(t *testing.T) {
		config := NewConfigWithTLS("https://localhost:9000", "", "", "")
		assert.Equal(t, "https://localhost:9000", config.BaseURL)
		assert.Nil(t, config.TLS, "TLS config should be nil when no TLS parameters provided")
	})

	t.Run("https url with partial tls config", func(t *testing.T) {
		config := NewConfigWithTLS("https://localhost:9000", "ca-cert", "", "")
		assert.Equal(t, "https://localhost:9000", config.BaseURL)
		assert.NotNil(t, config.TLS, "TLS config should be set when any TLS parameter is provided")
		assert.Equal(t, "ca-cert", config.TLS.CACert)
		assert.Equal(t, "", config.TLS.ClientCert)
		assert.Equal(t, "", config.TLS.ClientKey)
	})
}

func TestClient_SignRaw(t *testing.T) {
	t.Skip("Web3signer is currently experimental")
	l := logger.NewLogger(false)

	t.Run("successful raw signing", func(t *testing.T) {
		expectedSignature := "0xb3baa751d0a9132cfe93e4e3d5ff9075111100e3789dca219ade5a24d27e19d16b3353149da1833e9b691bb38634e8dc04469be7032132906c927d7e1a49b414730612877bc6b2810c8f202daf793d1ab0d6b5cb21d52f9e52e883859887a5d9"
		testData := []byte("Hello, Web3SignerClient!")
		identifier := "0x1234567890abcdef"

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Equal(t, "/api/v1/eth1/sign/"+identifier, r.URL.Path)

			var payload map[string]interface{}
			err := json.NewDecoder(r.Body).Decode(&payload)
			require.NoError(t, err)

			expectedDataHex := "0x48656c6c6f2c20576562335369676e657221"
			assert.Equal(t, expectedDataHex, payload["data"])

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(expectedSignature)
		}))
		defer server.Close()

		cfg := DefaultWeb3SignerConfig()
		cfg.BaseURL = server.URL
		client, err := NewWeb3SignerClient(cfg, l)
		require.NoError(t, err)

		signature, err := client.SignRaw(context.Background(), identifier, testData)
		require.NoError(t, err)
		assert.Equal(t, expectedSignature, signature)
	})

	t.Run("server error", func(t *testing.T) {
		identifier := "0x1234567890abcdef"
		testData := []byte("test data")

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("Bad request"))
		}))
		defer server.Close()

		cfg := DefaultWeb3SignerConfig()
		cfg.BaseURL = server.URL
		client, err := NewWeb3SignerClient(cfg, l)
		require.NoError(t, err)

		_, err = client.SignRaw(context.Background(), identifier, testData)
		require.Error(t, err)

		var web3SignerErr *Web3SignerError
		assert.ErrorAs(t, err, &web3SignerErr)
		assert.Equal(t, 400, web3SignerErr.Code)
	})
}

func TestTLSClientCreation(t *testing.T) {
	t.Skip("Web3signer is currently experimental")
	l := logger.NewLogger(false)

	t.Run("http client creation", func(t *testing.T) {
		config := &Config{
			BaseURL: "http://localhost:9000",
			Timeout: 30 * time.Second,
		}
		client, err := NewWeb3SignerClient(config, l)
		require.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("https client creation without tls config", func(t *testing.T) {
		config := &Config{
			BaseURL: "https://localhost:9000",
			Timeout: 30 * time.Second,
		}
		client, err := NewWeb3SignerClient(config, l)
		require.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("https client creation with tls config", func(t *testing.T) {
		config := &Config{
			BaseURL: "https://localhost:9000",
			Timeout: 30 * time.Second,
			TLS: &TLSConfig{
				InsecureSkipVerify: true, // For testing - skip cert validation
			},
		}
		client, err := NewWeb3SignerClient(config, l)
		require.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("invalid client certificate", func(t *testing.T) {
		config := &Config{
			BaseURL: "https://localhost:9000",
			Timeout: 30 * time.Second,
			TLS: &TLSConfig{
				ClientCert: "invalid-cert",
				ClientKey:  "invalid-key",
			},
		}
		client, err := NewWeb3SignerClient(config, l)
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "failed to load client certificate and key")
	})
}

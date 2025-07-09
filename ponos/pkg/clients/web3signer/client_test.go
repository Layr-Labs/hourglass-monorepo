package web3signer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewClient(t *testing.T) {
	t.Run("with default config", func(t *testing.T) {
		client := NewClient(nil, nil)
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
		client := NewClient(cfg, zap.NewNop())
		assert.NotNil(t, client)
		assert.Equal(t, "http://custom:8080", client.config.BaseURL)
		assert.Equal(t, 10*time.Second, client.config.Timeout)
	})
}

func TestClient_Sign(t *testing.T) {
	t.Run("successful sign", func(t *testing.T) {
		expectedSignature := "0xb3baa751d0a9132cfe93e4e3d5ff9075111100e3789dca219ade5a24d27e19d16b3353149da1833e9b691bb38634e8dc04469be7032132906c927d7e1a49b414730612877bc6b2810c8f202daf793d1ab0d6b5cb21d52f9e52e883859887a5d9"

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/api/v1/eth1/sign/test-key", r.URL.Path)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			var req SignRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			assert.Equal(t, "0x48656c6c6f2c20776f726c6421", req.Data)

			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `"%s"`, expectedSignature)
		}))
		defer server.Close()

		client := NewClient(&Config{BaseURL: server.URL}, zap.NewNop())

		signature, err := client.Sign(context.Background(), "test-key", "0x48656c6c6f2c20776f726c6421")
		require.NoError(t, err)
		assert.Equal(t, expectedSignature, signature)
	})

	t.Run("key not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Public Key not found"))
		}))
		defer server.Close()

		client := NewClient(&Config{BaseURL: server.URL}, zap.NewNop())

		_, err := client.Sign(context.Background(), "non-existent-key", "0x48656c6c6f2c20776f726c6421")
		require.Error(t, err)

		var web3SignerErr *Web3SignerError
		assert.ErrorAs(t, err, &web3SignerErr)
		assert.Equal(t, 404, web3SignerErr.Code)
	})

	t.Run("bad request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Bad request format"))
		}))
		defer server.Close()

		client := NewClient(&Config{BaseURL: server.URL}, zap.NewNop())

		_, err := client.Sign(context.Background(), "test-key", "invalid-data")
		require.Error(t, err)

		var web3SignerErr *Web3SignerError
		assert.ErrorAs(t, err, &web3SignerErr)
		assert.Equal(t, 400, web3SignerErr.Code)
	})
}

func TestClient_ListPublicKeys(t *testing.T) {
	t.Run("successful list", func(t *testing.T) {
		expectedKeys := []string{
			"0x1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b",
			"0x2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c",
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/api/v1/eth1/publicKeys", r.URL.Path)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedKeys)
		}))
		defer server.Close()

		client := NewClient(&Config{BaseURL: server.URL}, zap.NewNop())

		keys, err := client.ListPublicKeys(context.Background())
		require.NoError(t, err)
		assert.Equal(t, expectedKeys, keys)
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal server error"))
		}))
		defer server.Close()

		client := NewClient(&Config{BaseURL: server.URL}, zap.NewNop())

		_, err := client.ListPublicKeys(context.Background())
		require.Error(t, err)

		var web3SignerErr *Web3SignerError
		assert.ErrorAs(t, err, &web3SignerErr)
		assert.Equal(t, 500, web3SignerErr.Code)
	})
}

func TestClient_Reload(t *testing.T) {
	t.Run("successful reload", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/reload", r.URL.Path)

			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewClient(&Config{BaseURL: server.URL}, zap.NewNop())

		err := client.Reload(context.Background())
		require.NoError(t, err)
	})

	t.Run("reload error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to reload"))
		}))
		defer server.Close()

		client := NewClient(&Config{BaseURL: server.URL}, zap.NewNop())

		err := client.Reload(context.Background())
		require.Error(t, err)
	})
}

func TestClient_Upcheck(t *testing.T) {
	t.Run("successful upcheck", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/upcheck", r.URL.Path)

			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`"OK"`))
		}))
		defer server.Close()

		client := NewClient(&Config{BaseURL: server.URL}, zap.NewNop())

		status, err := client.Upcheck(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "OK", status)
	})
}

func TestClient_HealthCheck(t *testing.T) {
	t.Run("healthy status", func(t *testing.T) {
		expectedHealthCheck := HealthCheck{
			Status: "UP",
			Checks: []StatusCheck{
				{ID: "disk-space", Status: "UP"},
				{ID: "memory", Status: "UP"},
			},
			Outcome: "UP",
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/healthcheck", r.URL.Path)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedHealthCheck)
		}))
		defer server.Close()

		client := NewClient(&Config{BaseURL: server.URL}, zap.NewNop())

		health, err := client.HealthCheck(context.Background())
		require.NoError(t, err)
		assert.Equal(t, expectedHealthCheck.Status, health.Status)
		assert.Equal(t, expectedHealthCheck.Outcome, health.Outcome)
		assert.Len(t, health.Checks, 2)
	})

	t.Run("unhealthy status", func(t *testing.T) {
		expectedHealthCheck := HealthCheck{
			Status: "DOWN",
			Checks: []StatusCheck{
				{ID: "disk-space", Status: "DOWN"},
			},
			Outcome: "DOWN",
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(expectedHealthCheck)
		}))
		defer server.Close()

		client := NewClient(&Config{BaseURL: server.URL}, zap.NewNop())

		_, err := client.HealthCheck(context.Background())
		require.Error(t, err)

		var web3SignerErr *Web3SignerError
		assert.ErrorAs(t, err, &web3SignerErr)
		assert.Equal(t, 503, web3SignerErr.Code)
	})
}

func TestClient_buildURL(t *testing.T) {
	client := NewClient(&Config{BaseURL: "http://localhost:9000"}, zap.NewNop())

	tests := []struct {
		name     string
		endpoint string
		expected string
	}{
		{
			name:     "endpoint with leading slash",
			endpoint: "/api/v1/eth1/publicKeys",
			expected: "http://localhost:9000/api/v1/eth1/publicKeys",
		},
		{
			name:     "endpoint without leading slash",
			endpoint: "api/v1/eth1/publicKeys",
			expected: "http://localhost:9000/api/v1/eth1/publicKeys",
		},
		{
			name:     "root endpoint",
			endpoint: "/",
			expected: "http://localhost:9000/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.buildURL(tt.endpoint)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWeb3SignerError_Error(t *testing.T) {
	err := &Web3SignerError{
		Code:    404,
		Message: "Public key not found",
	}

	expected := "Web3Signer error 404: Public key not found"
	assert.Equal(t, expected, err.Error())
}

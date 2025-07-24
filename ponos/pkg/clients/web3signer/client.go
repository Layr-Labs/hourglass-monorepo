// Package web3signer provides a client for interacting with Web3Signer services.
//
// Web3Signer is a remote signing service that provides a JSON-RPC API for signing
// Ethereum transactions and messages. This client package provides a Go
// interface for interacting with Web3Signer instances via JSON-RPC.
//
// The client supports the following operations:
//   - Signing transactions with specified keys
//   - Listing available public keys
//   - Reloading signer keys
//   - Health checking and status monitoring
//
// Example usage:
//
//	// Basic HTTP configuration
//	cfg := web3signer.DefaultConfig()
//	cfg.BaseURL = "http://localhost:9000"
//	cfg.Timeout = 30 * time.Second
//	client, err := web3signer.NewClient(cfg, logger)
//
//	// HTTPS with TLS configuration
//	tlsConfig := web3signer.NewConfigWithTLS(
//		"https://web3signer.example.com:9000",
//		caCert,    // PEM-encoded CA certificate
//		clientCert, // PEM-encoded client certificate
//		clientKey,  // PEM-encoded client private key
//	)
//	secureClient, err := web3signer.NewClient(tlsConfig, logger)
//
//	// Sign a transaction
//	signature, err := client.EthSignTransaction(ctx, "0x1234...", txData)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// List available accounts
//	accounts, err := client.EthAccounts(ctx)
//	if err != nil {
//		log.Fatal(err)
//	}
package web3signer

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// Client represents a Web3Signer JSON-RPC client that provides methods for
// interacting with a Web3Signer service instance.
type Client struct {
	// Logger is used for logging client operations and debugging
	Logger *zap.Logger
	// httpClient is the underlying HTTP client used for requests
	httpClient *http.Client
	// config contains the client configuration including base URL and timeout
	config *Config
	// requestID is used to generate unique request IDs for JSON-RPC calls
	requestID int64
}

// Config holds the configuration for the Web3Signer client.
type Config struct {
	// BaseURL is the base URL of the Web3Signer service (e.g., "http://localhost:9000")
	BaseURL string
	// Timeout is the maximum duration for HTTP requests
	Timeout time.Duration
	// TLS configuration for HTTPS connections
	TLS *TLSConfig
}

// TLSConfig holds TLS configuration for secure connections to Web3Signer.
// According to the Web3Signer TLS documentation, this supports:
// - Server certificate verification using custom CA certificates
// - Mutual TLS authentication using client certificates
// - Optional certificate verification skipping for testing
type TLSConfig struct {
	// CACert is the PEM-encoded CA certificate to verify the server's certificate
	CACert string
	// ClientCert is the PEM-encoded client certificate for mutual TLS authentication
	ClientCert string
	// ClientKey is the PEM-encoded client private key for mutual TLS authentication
	ClientKey string
	// InsecureSkipVerify skips server certificate verification (not recommended for production)
	InsecureSkipVerify bool
}

// DefaultConfig returns a default configuration for the Web3Signer client.
// The default configuration uses localhost:9000 as the base URL and a 30-second timeout.
func DefaultConfig() *Config {
	return &Config{
		BaseURL: "http://localhost:9000",
		Timeout: 30 * time.Second,
	}
}

// NewConfigWithTLS creates a Web3Signer Config with TLS configuration.
// This is a convenience function to create a config with TLS settings for secure connections.
//
// Parameters:
//   - baseURL: The Web3Signer service URL (e.g., "https://web3signer.example.com:9000")
//   - caCert: PEM-encoded CA certificate to verify the server's certificate (optional)
//   - clientCert: PEM-encoded client certificate for mutual TLS authentication (optional)
//   - clientKey: PEM-encoded client private key for mutual TLS authentication (optional)
//
// TLS configuration is only applied for HTTPS URLs. For HTTP URLs, TLS settings are ignored.
// If any TLS parameter is provided for an HTTPS URL, a TLS config will be created.
func NewConfigWithTLS(baseURL string, caCert, clientCert, clientKey string) *Config {
	config := DefaultConfig()
	config.BaseURL = baseURL

	// Only configure TLS if we have HTTPS and at least one TLS field
	if strings.HasPrefix(baseURL, "https://") && (caCert != "" || clientCert != "" || clientKey != "") {
		tlsConfig := &TLSConfig{
			CACert:     caCert,
			ClientCert: clientCert,
			ClientKey:  clientKey,
		}
		config.TLS = tlsConfig
	}

	return config
}

func NewWeb3SignerClientFromRemoteSignerConfig(cfg *config.RemoteSignerConfig, l *zap.Logger) (*Client, error) {
	var web3SignerConfig *Config
	if cfg != nil {
		web3SignerConfig = NewConfigWithTLS(
			cfg.Url,
			cfg.CACert,
			cfg.Cert,
			cfg.Key,
		)
	} else {
		web3SignerConfig = DefaultConfig()
	}
	return NewClient(web3SignerConfig, l)
}

// NewClient creates a new Web3Signer client with the given configuration and logger.
// Both cfg and logger must be non-nil. Use DefaultConfig() if you want default configuration.
func NewClient(cfg *Config, logger *zap.Logger) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("cfg cannot be nil")
	}

	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	httpClient, err := createHTTPClient(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	logger.Sugar().Debugw("Creating new Web3Signer client",
		zap.String("baseURL", cfg.BaseURL),
		zap.Duration("timeout", cfg.Timeout),
		zap.Bool("tlsEnabled", cfg.TLS != nil),
	)

	return &Client{
		Logger:     logger,
		httpClient: httpClient,
		config:     cfg,
		requestID:  0,
	}, nil
}

// createHTTPClient creates an HTTP client with appropriate TLS configuration
func createHTTPClient(cfg *Config, logger *zap.Logger) (*http.Client, error) {
	transport := &http.Transport{}

	// Only configure TLS for HTTPS URLs
	if strings.HasPrefix(cfg.BaseURL, "https://") && cfg.TLS != nil {
		tlsConfig, err := buildTLSConfig(cfg.TLS, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to build TLS config: %w", err)
		}
		transport.TLSClientConfig = tlsConfig
		logger.Sugar().Debugw("Configured TLS for Web3Signer client")
	} else if strings.HasPrefix(cfg.BaseURL, "https://") {
		// HTTPS without custom TLS config - use default TLS
		transport.TLSClientConfig = &tls.Config{}
		logger.Sugar().Debugw("Using default TLS configuration for HTTPS")
	} else {
		// HTTP connection - no TLS needed
		logger.Sugar().Debugw("Using HTTP connection (no TLS)")
	}

	return &http.Client{
		Timeout:   cfg.Timeout,
		Transport: transport,
	}, nil
}

// buildTLSConfig creates a TLS configuration from the provided TLS config
func buildTLSConfig(tlsConfig *TLSConfig, logger *zap.Logger) (*tls.Config, error) {
	config := &tls.Config{
		InsecureSkipVerify: tlsConfig.InsecureSkipVerify,
	}

	// Configure CA certificate for server verification
	if tlsConfig.CACert != "" {
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM([]byte(tlsConfig.CACert)) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
		config.RootCAs = caCertPool
		logger.Sugar().Debugw("Configured custom CA certificate")
	}

	// Configure client certificate for mutual TLS
	if tlsConfig.ClientCert != "" && tlsConfig.ClientKey != "" {
		clientCert, err := tls.X509KeyPair([]byte(tlsConfig.ClientCert), []byte(tlsConfig.ClientKey))
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate and key: %w", err)
		}
		config.Certificates = []tls.Certificate{clientCert}
		logger.Sugar().Debugw("Configured client certificate for mutual TLS")
	} else if tlsConfig.ClientCert != "" || tlsConfig.ClientKey != "" {
		return nil, fmt.Errorf("both client certificate and key must be provided for mutual TLS")
	}

	return config, nil
}

// SetHttpClient allows setting a custom HTTP client for the Web3Signer client.
// This is useful for testing or when custom HTTP client configuration is needed.
func (c *Client) SetHttpClient(client *http.Client) {
	c.httpClient = client
}

// EthAccounts returns a list of accounts available for signing.
// This corresponds to the eth_accounts JSON-RPC method.
func (c *Client) EthAccounts(ctx context.Context) ([]string, error) {
	var result []string
	err := c.makeJSONRPCRequest(ctx, "eth_accounts", nil, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}
	return result, nil
}

// EthSignTransaction signs a transaction and returns the signature.
// This corresponds to the eth_signTransaction JSON-RPC method.
func (c *Client) EthSignTransaction(ctx context.Context, from string, transaction map[string]interface{}) (string, error) {
	// Add the from field to the transaction object
	transaction["from"] = from
	params := []interface{}{transaction}
	var result string
	err := c.makeJSONRPCRequest(ctx, "eth_signTransaction", params, &result)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}
	return result, nil
}

// EthSign signs data with the specified account.
// This corresponds to the eth_sign JSON-RPC method.
func (c *Client) EthSign(ctx context.Context, account string, data string) (string, error) {
	params := []interface{}{account, data}
	var result string
	err := c.makeJSONRPCRequest(ctx, "eth_sign", params, &result)
	if err != nil {
		return "", fmt.Errorf("failed to sign data: %w", err)
	}
	return result, nil
}

// EthSignTypedData signs typed data with the specified account.
// This corresponds to the eth_signTypedData JSON-RPC method.
func (c *Client) EthSignTypedData(ctx context.Context, account string, typedData interface{}) (string, error) {
	params := []interface{}{account, typedData}
	var result string
	err := c.makeJSONRPCRequest(ctx, "eth_signTypedData", params, &result)
	if err != nil {
		return "", fmt.Errorf("failed to sign typed data: %w", err)
	}
	return result, nil
}

// ListPublicKeys retrieves all available public keys from the Web3Signer service.
// This is a convenience method that calls EthAccounts.
func (c *Client) ListPublicKeys(ctx context.Context) ([]string, error) {
	return c.EthAccounts(ctx)
}

// Sign signs data with the specified account using eth_sign.
// This is a convenience method that calls EthSign.
func (c *Client) Sign(ctx context.Context, account string, data string) (string, error) {
	return c.EthSign(ctx, account, data)
}

// SignRaw performs raw ECDSA signing using the REST API endpoint.
// This method signs raw data without Ethereum message prefixes, making it
// compatible with generic ECDSA libraries like crypto-libs.
// The identifier parameter is the signing key identifier (typically an address).
func (c *Client) SignRaw(ctx context.Context, identifier string, data []byte) (string, error) {
	// Convert data to hex format
	dataHex := "0x" + hex.EncodeToString(data)

	// Create the request payload
	payload := map[string]interface{}{
		"type": "MESSAGE",
		"data": dataHex,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request payload: %w", err)
	}

	// Build the REST API URL
	url := strings.TrimSuffix(c.config.BaseURL, "/") + "/api/v1/eth1/sign/" + identifier

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	c.Logger.Sugar().Debugw("Making Web3Signer REST API sign request",
		zap.String("identifier", identifier),
		zap.String("url", url),
		zap.Int("dataLength", len(data)),
		zap.String("dataHex", dataHex),
	)

	// Make HTTP request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("REST API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	c.Logger.Sugar().Debugw("Web3Signer REST API response received",
		zap.Int("status_code", resp.StatusCode),
		zap.String("response", string(responseData)),
	)

	// Check HTTP status
	if resp.StatusCode >= 400 {
		return "", c.handleHTTPError(resp.StatusCode, responseData)
	}

	// Parse the response - Web3Signer REST API returns just the signature as plain text
	signature := strings.TrimSpace(string(responseData))

	// Remove any quotes if present (some implementations might return quoted strings)
	signature = strings.Trim(signature, `"`)

	return signature, nil
}

// makeJSONRPCRequest performs a JSON-RPC request to the Web3Signer service.
func (c *Client) makeJSONRPCRequest(ctx context.Context, method string, params interface{}, result interface{}) error {
	// Generate unique request ID
	id := atomic.AddInt64(&c.requestID, 1)

	// Create JSON-RPC request
	request := JSONRPCRequest{
		Jsonrpc: "2.0",
		Method:  method,
		Params:  params,
		ID:      id,
	}

	// Marshal request to JSON
	requestData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON-RPC request: %w", err)
	}

	// Create HTTP request
	url := strings.TrimSuffix(c.config.BaseURL, "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(requestData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	c.Logger.Sugar().Debugw("Making Web3Signer JSON-RPC request",
		zap.String("method", method),
		zap.String("url", url),
		zap.Any("params", params),
	)

	// Make HTTP request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("JSON-RPC request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	c.Logger.Sugar().Debugw("Web3Signer JSON-RPC response received",
		zap.Int("status_code", resp.StatusCode),
		zap.String("response", string(responseData)),
	)

	// Check HTTP status
	if resp.StatusCode >= 400 {
		return c.handleHTTPError(resp.StatusCode, responseData)
	}

	// Parse JSON-RPC response
	var jsonRPCResponse JSONRPCResponse
	if err := json.Unmarshal(responseData, &jsonRPCResponse); err != nil {
		return fmt.Errorf("failed to unmarshal JSON-RPC response: %w", err)
	}

	// Check for JSON-RPC error
	if jsonRPCResponse.Error != nil {
		return &Web3SignerError{
			Code:    jsonRPCResponse.Error.Code,
			Message: jsonRPCResponse.Error.Message,
		}
	}

	// Unmarshal result if provided
	if result != nil && jsonRPCResponse.Result != nil {
		resultData, err := json.Marshal(jsonRPCResponse.Result)
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}

		if err := json.Unmarshal(resultData, result); err != nil {
			return fmt.Errorf("failed to unmarshal result: %w", err)
		}
	}

	return nil
}

// handleHTTPError converts HTTP error responses into appropriate Web3SignerError instances.
func (c *Client) handleHTTPError(statusCode int, responseData []byte) error {
	errorMsg := string(responseData)

	return &Web3SignerError{
		Code:    statusCode,
		Message: fmt.Sprintf("HTTP error %d: %s", statusCode, errorMsg),
	}
}

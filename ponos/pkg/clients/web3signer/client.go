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
//	cfg := &web3signer.Config{
//		BaseURL: "http://localhost:9000",
//		Timeout: 30 * time.Second,
//	}
//	client := web3signer.NewClient(cfg, logger)
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
	"encoding/json"
	"fmt"
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
}

// DefaultConfig returns a default configuration for the Web3Signer client.
// The default configuration uses localhost:9000 as the base URL and a 30-second timeout.
func DefaultConfig() *Config {
	return &Config{
		BaseURL: "http://localhost:9000",
		Timeout: 30 * time.Second,
	}
}

// NewClient creates a new Web3Signer client with the given configuration and logger.
// If cfg is nil, DefaultConfig() is used. If logger is nil, a no-op logger is used.
func NewClient(cfg *Config, logger *zap.Logger) *Client {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	httpClient := &http.Client{
		Timeout: cfg.Timeout,
	}

	logger.Sugar().Debugw("Creating new Web3Signer client", zap.Any("config", cfg))

	return &Client{
		Logger:     logger,
		httpClient: httpClient,
		config:     cfg,
		requestID:  0,
	}
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

// Reload instructs the Web3Signer service to reload its key configuration.
// This uses the REST API endpoint as JSON-RPC doesn't have a reload method.
func (c *Client) Reload(ctx context.Context) error {
	url := c.buildURL("/reload")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create reload request: %w", err)
	}

	c.Logger.Sugar().Debugw("Making Web3Signer reload request", zap.String("url", url))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("reload request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return &Web3SignerError{
			Code:    resp.StatusCode,
			Message: fmt.Sprintf("Reload failed: %s", string(body)),
		}
	}

	return nil
}

// HealthCheck performs a detailed health check on the Web3Signer service.
// This uses the REST API endpoint as JSON-RPC doesn't have a health check method.
func (c *Client) HealthCheck(ctx context.Context) (*HealthCheck, error) {
	url := c.buildURL("/healthcheck")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create health check request: %w", err)
	}

	c.Logger.Sugar().Debugw("Making Web3Signer health check request", zap.String("url", url))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("health check request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, &Web3SignerError{
			Code:    resp.StatusCode,
			Message: fmt.Sprintf("Health check failed: %s", string(body)),
		}
	}

	var healthCheck HealthCheck
	if err := json.NewDecoder(resp.Body).Decode(&healthCheck); err != nil {
		return nil, fmt.Errorf("failed to decode health check response: %w", err)
	}

	return &healthCheck, nil
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
	url := c.buildURL("")
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

// buildURL constructs the full URL for an API endpoint.
func (c *Client) buildURL(endpoint string) string {
	baseURL := strings.TrimSuffix(c.config.BaseURL, "/")

	// For JSON-RPC requests, we don't append anything to the base URL
	if endpoint == "" {
		return baseURL
	}

	// For REST endpoints like /reload and /healthcheck
	endpoint = strings.TrimPrefix(endpoint, "/")
	return fmt.Sprintf("%s/%s", baseURL, endpoint)
}

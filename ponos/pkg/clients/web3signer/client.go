// Package web3signer provides a client for interacting with Web3Signer services.
//
// Web3Signer is a remote signing service that provides a REST API for signing
// Ethereum transactions and messages. This client package provides a Go
// interface for interacting with Web3Signer instances.
//
// The client supports the following operations:
//   - Signing data with specified keys
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
//	// Sign some data
//	signature, err := client.Sign(ctx, "key-id", "0x48656c6c6f")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// List available keys
//	keys, err := client.ListPublicKeys(ctx)
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
	"time"

	"go.uber.org/zap"
)

// Client represents a Web3Signer HTTP client that provides methods for
// interacting with a Web3Signer service instance.
type Client struct {
	// Logger is used for logging client operations and debugging
	Logger     *zap.Logger
	// httpClient is the underlying HTTP client used for requests
	httpClient *http.Client
	// config contains the client configuration including base URL and timeout
	config     *Config
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
	}
}

// SetHttpClient allows setting a custom HTTP client for the Web3Signer client.
// This is useful for testing or when custom HTTP client configuration is needed.
func (c *Client) SetHttpClient(client *http.Client) {
	c.httpClient = client
}

// Sign requests a signature for the given data using the specified key identifier.
// The data should be hex-encoded. Returns the signature as a hex string.
//
// Parameters:
//   - ctx: Context for the request
//   - identifier: The key identifier/name to use for signing
//   - data: Hex-encoded data to sign
//
// Returns the signature as a hex string or an error if the operation fails.
func (c *Client) Sign(ctx context.Context, identifier, data string) (string, error) {
	endpoint := fmt.Sprintf("/api/v1/eth1/sign/%s", identifier)

	signRequest := SignRequest{
		Data: data,
	}

	var signature string
	err := c.makeRequest(ctx, http.MethodPost, endpoint, signRequest, &signature)
	if err != nil {
		return "", fmt.Errorf("failed to sign data: %w", err)
	}

	return signature, nil
}

// ListPublicKeys retrieves all available public keys from the Web3Signer service.
// Returns a slice of hex-encoded public key strings.
func (c *Client) ListPublicKeys(ctx context.Context) ([]string, error) {
	var publicKeys []string
	err := c.makeRequest(ctx, http.MethodGet, "/api/v1/eth1/publicKeys", nil, &publicKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to list public keys: %w", err)
	}

	return publicKeys, nil
}

// Reload instructs the Web3Signer service to reload its key configuration.
// This is useful when keys have been added or modified on the filesystem.
func (c *Client) Reload(ctx context.Context) error {
	var response interface{}
	err := c.makeRequest(ctx, http.MethodPost, "/reload", nil, &response)
	if err != nil {
		return fmt.Errorf("failed to reload signer keys: %w", err)
	}

	return nil
}

// Upcheck performs a basic health check on the Web3Signer service.
// Returns the status string (typically "OK") or an error if the service is down.
func (c *Client) Upcheck(ctx context.Context) (string, error) {
	var status string
	err := c.makeRequest(ctx, http.MethodGet, "/upcheck", nil, &status)
	if err != nil {
		return "", fmt.Errorf("failed to check server status: %w", err)
	}

	return status, nil
}

// HealthCheck performs a detailed health check on the Web3Signer service.
// Returns a HealthCheck struct with detailed status information about various
// service components (disk space, memory, etc.).
func (c *Client) HealthCheck(ctx context.Context) (*HealthCheck, error) {
	var healthCheck HealthCheck
	err := c.makeRequest(ctx, http.MethodGet, "/healthcheck", nil, &healthCheck)
	if err != nil {
		return nil, fmt.Errorf("failed to check server health: %w", err)
	}

	return &healthCheck, nil
}

// makeRequest performs an HTTP request to the Web3Signer service.
// This is a private method used internally by the client methods.
func (c *Client) makeRequest(ctx context.Context, method, endpoint string, requestBody interface{}, responseBody interface{}) error {
	url := c.buildURL(endpoint)

	var body io.Reader
	if requestBody != nil {
		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		body = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if requestBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	c.Logger.Sugar().Debugw("Making Web3Signer request",
		zap.String("method", method),
		zap.String("url", url),
	)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	c.Logger.Sugar().Debugw("Web3Signer response received",
		zap.Int("status_code", resp.StatusCode),
		zap.String("response", string(responseData)),
	)

	if resp.StatusCode >= 400 {
		return c.handleErrorResponse(resp.StatusCode, responseData)
	}

	if responseBody != nil && len(responseData) > 0 {
		if strings.HasPrefix(resp.Header.Get("Content-Type"), "text/plain") {
			if strPtr, ok := responseBody.(*string); ok {
				*strPtr = strings.Trim(string(responseData), "\"")
				return nil
			}
		}

		if err := json.Unmarshal(responseData, responseBody); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// handleErrorResponse converts HTTP error responses into appropriate Web3SignerError instances.
// This is a private method used internally by makeRequest.
func (c *Client) handleErrorResponse(statusCode int, responseData []byte) error {
	errorMsg := string(responseData)

	switch statusCode {
	case 400:
		return &Web3SignerError{Code: 400, Message: fmt.Sprintf("Bad request format: %s", errorMsg)}
	case 404:
		return &Web3SignerError{Code: 404, Message: fmt.Sprintf("Public key not found: %s", errorMsg)}
	case 500:
		return &Web3SignerError{Code: 500, Message: fmt.Sprintf("Internal Web3Signer server error: %s", errorMsg)}
	case 503:
		return &Web3SignerError{Code: 503, Message: fmt.Sprintf("Service unavailable: %s", errorMsg)}
	default:
		return &Web3SignerError{Code: statusCode, Message: fmt.Sprintf("HTTP error %d: %s", statusCode, errorMsg)}
	}
}

// buildURL constructs the full URL for an API endpoint.
// This is a private method used internally by makeRequest.
func (c *Client) buildURL(endpoint string) string {
	baseURL := strings.TrimSuffix(c.config.BaseURL, "/")
	endpoint = strings.TrimPrefix(endpoint, "/")
	return fmt.Sprintf("%s/%s", baseURL, endpoint)
}

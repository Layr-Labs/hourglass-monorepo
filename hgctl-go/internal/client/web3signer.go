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
//	cfg := web3signer.DefaultWeb3SignerConfig()
//	cfg.BaseURL = "http://localhost:9000"
//	cfg.Timeout = 30 * time.Second
//	client, err := web3signer.NewWeb3Signer(cfg, logger)
//
//	// HTTPS with TLS configuration
//	tlsConfig := web3signer.NewConfigWithTLS(
//		"https://web3signer.example.com:9000",
//		caCert,    // PEM-encoded CA certificate
//		clientCert, // PEM-encoded client certificate
//		clientKey,  // PEM-encoded client private key
//	)
//	secureClient, err := web3signer.NewWeb3Signer(tlsConfig, logger)
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
package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/signer/web3Signer"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/signer"

	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

const (
	web3SignerSignTransaction = "eth_signTransaction"
)

// JSONRPCRequest represents a JSON-RPC 2.0 request.
type JSONRPCRequest struct {
	// Jsonrpc specifies the JSON-RPC version (always "2.0")
	Jsonrpc string `json:"jsonrpc"`
	// Method is the JSON-RPC method name
	Method string `json:"method"`
	// Params contains the method parameters
	Params interface{} `json:"params,omitempty"`
	// ID is a unique identifier for the request
	ID int64 `json:"id"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response.
type JSONRPCResponse struct {
	// Jsonrpc specifies the JSON-RPC version (always "2.0")
	Jsonrpc string `json:"jsonrpc"`
	// Result contains the method result (present on success)
	Result interface{} `json:"result,omitempty"`
	// Error contains error information (present on error)
	Error *JSONRPCError `json:"error,omitempty"`
	// ID is the request identifier
	ID int64 `json:"id"`
}

// JSONRPCError represents a JSON-RPC 2.0 error.
type JSONRPCError struct {
	// Code is the error code
	Code int `json:"code"`
	// Message is the error message
	Message string `json:"message"`
	// Data contains additional error data
	Data interface{} `json:"data,omitempty"`
}

// EthSignTransactionRequest represents the parameters for eth_signTransaction.
type EthSignTransactionRequest struct {
	// From is the account to sign with
	From string `json:"from"`
	// To is the destination address
	To string `json:"to,omitempty"`
	// Gas is the gas limit
	Gas string `json:"gas,omitempty"`
	// GasPrice is the gas price
	GasPrice string `json:"gasPrice,omitempty"`
	// Value is the value to send
	Value string `json:"value,omitempty"`
	// Data is the transaction data
	Data string `json:"data,omitempty"`
	// Nonce is the transaction nonce
	Nonce string `json:"nonce,omitempty"`
	// ChainID is the chain ID
	ChainID string `json:"chainId,omitempty"`
}

// SignRequest represents a request to sign data sent to the Web3Signer service.
type SignRequest struct {
	// Data is the hex-encoded data to be signed
	Data string `json:"data"`
}

// SignResponse represents the response from a signing operation.
type SignResponse struct {
	// Signature is the hex-encoded signature returned by the service
	Signature string `json:"signature"`
}

// HealthCheck represents the detailed health status of the Web3Signer service.
type HealthCheck struct {
	// Status is the overall status of the service ("UP" or "DOWN")
	Status string `json:"status"`
	// Checks contains detailed status information for individual components
	Checks []StatusCheck `json:"checks"`
	// Outcome is the final health determination ("UP" or "DOWN")
	Outcome string `json:"outcome"`
}

// StatusCheck represents the status of an individual component within the health check.
type StatusCheck struct {
	// ID is the identifier of the component being checked (e.g., "disk-space", "memory")
	ID string `json:"id"`
	// Status is the status of this component ("UP" or "DOWN")
	Status string `json:"status"`
}

// Web3SignerError represents an error response from the Web3Signer service.
type Web3SignerError struct {
	// Code is the HTTP status code associated with the error
	Code int `json:"code"`
	// Message is the error message describing what went wrong
	Message string `json:"message"`
}

// Error implements the error interface for Web3SignerError.
func (e *Web3SignerError) Error() string {
	return fmt.Sprintf("Web3Signer error %d: %s", e.Code, e.Message)
}

// Web3SignerResponse represents a generic response structure from the Web3Signer service.
type Web3SignerResponse struct {
	// Status indicates the response status
	Status string `json:"status,omitempty"`
	// Data contains the response payload
	Data interface{} `json:"data,omitempty"`
	// Error contains error information if the request failed
	Error *Web3SignerError `json:"error,omitempty"`
}

// IsError returns true if the response contains an error.
func (r *Web3SignerResponse) IsError() bool {
	return r.Error != nil
}

// PublicKeysResponse represents a list of public keys returned by the Web3Signer service.
type PublicKeysResponse []string

// MarshalJSON implements the json.Marshaler interface for PublicKeysResponse.
func (p PublicKeysResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal([]string(p))
}

// UnmarshalJSON implements the json.Unmarshaler interface for PublicKeysResponse.
func (p *PublicKeysResponse) UnmarshalJSON(data []byte) error {
	var keys []string
	if err := json.Unmarshal(data, &keys); err != nil {
		return err
	}
	*p = PublicKeysResponse(keys)
	return nil
}

// Web3Signer represents a Web3Signer JSON-RPC client that provides methods for
// interacting with a Web3Signer service instance.
type Web3Signer struct {
	// Logger is used for logging client operations and debugging
	Logger logger.Logger
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

// DefaultWeb3SignerConfig returns a default configuration for the Web3Signer client.
// The default configuration uses localhost:9000 as the base URL and a 30-second timeout.
func DefaultWeb3SignerConfig() *Config {
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
	config := DefaultWeb3SignerConfig()
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

func NewWeb3SignerClientFromRemoteSignerConfig(cfg *config.RemoteSignerConfig, l logger.Logger) (*Web3Signer, error) {
	var web3SignerConfig *Config
	if cfg != nil {
		web3SignerConfig = NewConfigWithTLS(
			cfg.Url,
			cfg.CACert,
			cfg.Cert,
			cfg.Key,
		)
	} else {
		web3SignerConfig = DefaultWeb3SignerConfig()
	}
	return NewWeb3Signer(web3SignerConfig, l)
}

// NewWeb3Signer creates a new Web3Signer client with the given configuration and logger.
// Both cfg and logger must be non-nil. Use DefaultWeb3SignerConfig() if you want default configuration.
func NewWeb3Signer(cfg *Config, logger logger.Logger) (*Web3Signer, error) {
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

	return &Web3Signer{
		Logger:     logger,
		httpClient: httpClient,
		config:     cfg,
		requestID:  0,
	}, nil
}

// NewSigner creates a new Web3Signer that implements the ISigner interface.
// It only supports ECDSA curve type - attempting to use BN254 will result in errors.
// The publicKey parameter should be the hex-encoded public key (with or without 0x prefix)
// that corresponds to the fromAddress.
func NewSigner(client *Web3Signer, fromAddress common.Address, publicKey string, curveType config.CurveType, logger logger.Logger) (signer.ISigner, error) {
	if curveType != config.CurveTypeECDSA {
		return nil, fmt.Errorf("web3signer only supports ECDSA curve type, got %s", curveType)
	}

	if client == nil {
		return nil, fmt.Errorf("web3signer client cannot be nil")
	}

	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	if publicKey == "" {
		return nil, fmt.Errorf("publicKey cannot be empty")
	}

	// Clean up public key format - remove 0x prefix if present
	cleanPublicKey := strings.TrimPrefix(publicKey, "0x")

	logger.Sugar().Debugw("Creating new Web3Signer",
		"fromAddress", fromAddress.Hex(),
		"publicKey", cleanPublicKey,
		"curveType", curveType,
	)

	return web3Signer.NewWeb3Signer(client, fromAddress, cleanPublicKey, curveType, logger)
}

// createHTTPClient creates an HTTP client with appropriate TLS configuration
func createHTTPClient(cfg *Config, logger logger.Logger) (*http.Client, error) {
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
func buildTLSConfig(tlsConfig *TLSConfig, logger logger.Logger) (*tls.Config, error) {
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
		logger.Debug("Configured custom CA certificate")
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
func (c *Web3Signer) SetHttpClient(client *http.Client) {
	c.httpClient = client
}

// EthAccounts returns a list of accounts available for signing.
// This corresponds to the eth_accounts JSON-RPC method.
func (c *Web3Signer) EthAccounts(ctx context.Context) ([]string, error) {
	var result []string
	err := c.makeJSONRPCRequest(ctx, "eth_accounts", nil, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}
	return result, nil
}

// EthSignTransaction signs a transaction and returns the signature.
// This corresponds to the eth_signTransaction JSON-RPC method.
func (c *Web3Signer) EthSignTransaction(ctx context.Context, from string, transaction map[string]interface{}) (string, error) {
	// Add the from field to the transaction object
	transaction["from"] = from
	params := []interface{}{transaction}
	var result string
	err := c.makeJSONRPCRequest(ctx, web3SignerSignTransaction, params, &result)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}
	return result, nil
}

// EthSign signs data with the specified account.
// This corresponds to the eth_sign JSON-RPC method.
func (c *Web3Signer) EthSign(ctx context.Context, account string, data string) (string, error) {
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
func (c *Web3Signer) EthSignTypedData(ctx context.Context, account string, typedData interface{}) (string, error) {
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
func (c *Web3Signer) ListPublicKeys(ctx context.Context) ([]string, error) {
	return c.EthAccounts(ctx)
}

// Sign signs data with the specified account using eth_sign.
// This is a convenience method that calls EthSign.
func (c *Web3Signer) Sign(ctx context.Context, account string, data string) (string, error) {
	return c.EthSign(ctx, account, data)
}

// SignRaw performs raw ECDSA signing using the REST API endpoint.
// This method signs raw data without Ethereum message prefixes, making it
// compatible with generic ECDSA libraries like crypto-libs.
// The identifier parameter is the signing key identifier (typically an address).
func (c *Web3Signer) SignRaw(ctx context.Context, identifier string, data []byte) (string, error) {
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
func (c *Web3Signer) makeJSONRPCRequest(ctx context.Context, method string, params interface{}, result interface{}) error {
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
func (c *Web3Signer) handleHTTPError(statusCode int, responseData []byte) error {
	errorMsg := string(responseData)

	return &Web3SignerError{
		Code:    statusCode,
		Message: fmt.Sprintf("HTTP error %d: %s", statusCode, errorMsg),
	}
}

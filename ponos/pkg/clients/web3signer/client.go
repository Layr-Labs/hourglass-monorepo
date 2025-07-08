package web3signer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"go.uber.org/zap"
)

type Client struct {
	Logger     *zap.Logger
	httpClient *http.Client
	config     *Config
}

type Config struct {
	BaseURL string
	Timeout time.Duration
}

func DefaultConfig() *Config {
	return &Config{
		BaseURL: "http://localhost:9000",
		Timeout: 30 * time.Second,
	}
}

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

func (c *Client) SetHttpClient(client *http.Client) {
	c.httpClient = client
}

func (c *Client) Sign(ctx context.Context, identifier, data string) (string, error) {
	endpoint := path.Join("/api/v1/eth1/sign", identifier)
	
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

func (c *Client) ListPublicKeys(ctx context.Context) ([]string, error) {
	var publicKeys []string
	err := c.makeRequest(ctx, http.MethodGet, "/api/v1/eth1/publicKeys", nil, &publicKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to list public keys: %w", err)
	}

	return publicKeys, nil
}

func (c *Client) Reload(ctx context.Context) error {
	var response interface{}
	err := c.makeRequest(ctx, http.MethodPost, "/reload", nil, &response)
	if err != nil {
		return fmt.Errorf("failed to reload signer keys: %w", err)
	}

	return nil
}

func (c *Client) Upcheck(ctx context.Context) (string, error) {
	var status string
	err := c.makeRequest(ctx, http.MethodGet, "/upcheck", nil, &status)
	if err != nil {
		return "", fmt.Errorf("failed to check server status: %w", err)
	}

	return status, nil
}

func (c *Client) HealthCheck(ctx context.Context) (*HealthCheck, error) {
	var healthCheck HealthCheck
	err := c.makeRequest(ctx, http.MethodGet, "/healthcheck", nil, &healthCheck)
	if err != nil {
		return nil, fmt.Errorf("failed to check server health: %w", err)
	}

	return &healthCheck, nil
}

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

	if responseBody != nil {
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

func (c *Client) buildURL(endpoint string) string {
	baseURL := strings.TrimSuffix(c.config.BaseURL, "/")
	endpoint = strings.TrimPrefix(endpoint, "/")
	return fmt.Sprintf("%s/%s", baseURL, endpoint)
}
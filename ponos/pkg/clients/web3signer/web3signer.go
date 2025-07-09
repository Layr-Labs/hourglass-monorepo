package web3signer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type SignRequest struct {
	Data string `json:"data"`
}

type HealthCheckStatus struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type HealthCheck struct {
	Status  string              `json:"status"`
	Checks  []HealthCheckStatus `json:"checks"`
	Outcome string              `json:"outcome"`
}

type ClientConfig struct {
	BaseURL string
	Timeout time.Duration
}

type Client struct {
	config     *ClientConfig
	httpClient *http.Client
	logger     *zap.Logger
}

func NewClient(config *ClientConfig, logger *zap.Logger) *Client {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		logger: logger,
	}
}

func (c *Client) SignETH1(ctx context.Context, identifier string, data string) (string, error) {
	url := fmt.Sprintf("%s/api/v1/eth1/sign/%s", c.config.BaseURL, identifier)
	
	request := SignRequest{
		Data: data,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal sign request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/plain; charset=utf-8")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return string(body), nil
	case http.StatusNotFound:
		return "", fmt.Errorf("public key not found for identifier: %s", identifier)
	case http.StatusBadRequest:
		return "", fmt.Errorf("bad request format: %s", string(body))
	case http.StatusInternalServerError:
		return "", fmt.Errorf("internal Web3Signer server error: %s", string(body))
	default:
		return "", fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}
}

func (c *Client) ListPublicKeys(ctx context.Context) ([]string, error) {
	url := fmt.Sprintf("%s/api/v1/eth1/publicKeys", c.config.BaseURL)
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		var publicKeys []string
		if err := json.Unmarshal(body, &publicKeys); err != nil {
			return nil, fmt.Errorf("failed to unmarshal public keys response: %w", err)
		}
		return publicKeys, nil
	case http.StatusBadRequest:
		return nil, fmt.Errorf("bad request format: %s", string(body))
	case http.StatusInternalServerError:
		return nil, fmt.Errorf("internal Web3Signer server error: %s", string(body))
	default:
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}
}

func (c *Client) Reload(ctx context.Context) error {
	url := fmt.Sprintf("%s/reload", c.config.BaseURL)
	
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusInternalServerError:
		return fmt.Errorf("internal Web3Signer server error: %s", string(body))
	default:
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}
}

func (c *Client) Upcheck(ctx context.Context) (string, error) {
	url := fmt.Sprintf("%s/upcheck", c.config.BaseURL)
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Accept", "text/plain; charset=utf-8")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return string(body), nil
	case http.StatusInternalServerError:
		return "", fmt.Errorf("internal Web3Signer server error: %s", string(body))
	default:
		return "", fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}
}

func (c *Client) HealthCheck(ctx context.Context) (*HealthCheck, error) {
	url := fmt.Sprintf("%s/healthcheck", c.config.BaseURL)
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var healthCheck HealthCheck
	if err := json.Unmarshal(body, &healthCheck); err != nil {
		return nil, fmt.Errorf("failed to unmarshal health check response: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return &healthCheck, nil
	case http.StatusServiceUnavailable:
		return &healthCheck, fmt.Errorf("at least one procedure is unhealthy")
	default:
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}
}
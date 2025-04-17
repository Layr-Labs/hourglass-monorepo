package avsPerformerClient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/performer"
	"io"
	"net/http"
	"time"
)

type AvsPerformerClient struct {
	FullUrl    string
	httpClient *http.Client
}

func DefaultHttpClient() *http.Client {
	return &http.Client{
		Timeout: 5 * time.Second,
	}
}

func NewAvsPerformerClient(fullUrl string, httpClient *http.Client) *AvsPerformerClient {
	if httpClient == nil {
		httpClient = DefaultHttpClient()
	}
	return &AvsPerformerClient{
		FullUrl:    fullUrl,
		httpClient: DefaultHttpClient(),
	}
}

func (apc *AvsPerformerClient) makeRequest(ctx context.Context, method string, path string, body []byte) ([]byte, error) {
	fullUrl := fmt.Sprintf("%s%s", apc.FullUrl, path)

	ctx, cancel := context.WithTimeout(ctx, apc.httpClient.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, fullUrl, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := apc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status code %d: %s", res.StatusCode, string(responseBody))
	}

	return responseBody, nil
}

type HealthResponse struct {
	Status string `json:"status"`
}

func (apc *AvsPerformerClient) GetHealth(ctx context.Context) (*HealthResponse, error) {
	path := "/health"

	resData, err := apc.makeRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var data *HealthResponse
	if err := json.Unmarshal(resData, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return data, nil
}

func (apc *AvsPerformerClient) SendTask(ctx context.Context, t *performer.Task) (*performer.TaskResult, error) {
	path := "/tasks"

	taskBytes, err := json.Marshal(t)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal task: %w", err)
	}

	resData, err := apc.makeRequest(ctx, http.MethodPost, path, taskBytes)
	if err != nil {
		return nil, err
	}

	var data *performer.TaskResult
	if err := json.Unmarshal(resData, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return data, nil
}

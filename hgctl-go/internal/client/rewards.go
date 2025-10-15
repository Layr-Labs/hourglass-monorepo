package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/logger"
	"go.uber.org/zap"
)

type RewardsClient struct {
	baseURL    string
	httpClient *http.Client
	logger     logger.Logger
}

type TokenReward struct {
	Token      string `json:"token"`
	Earned     string `json:"earned"`
	Claimed    string `json:"claimed"`
	Claimable  string `json:"claimable"`
	TokenName  string `json:"tokenName,omitempty"`
}

type RewardsSummary struct {
	Earner string         `json:"earner"`
	Tokens []*TokenReward `json:"tokens"`
}

type ClaimProof struct {
	RootIndex       uint32          `json:"rootIndex"`
	EarnerIndex     uint32          `json:"earnerIndex"`
	EarnerTreeProof string          `json:"earnerTreeProof"`
	EarnerLeaf      EarnerLeaf      `json:"earnerLeaf"`
	TokenIndices    []uint32        `json:"tokenIndices"`
	TokenTreeProofs []string        `json:"tokenTreeProofs"`
	TokenLeaves     []TokenLeaf     `json:"tokenLeaves"`
}

type EarnerLeaf struct {
	Earner          string `json:"earner"`
	EarnerTokenRoot string `json:"earnerTokenRoot"`
}

type TokenLeaf struct {
	Token              string `json:"token"`
	CumulativeEarnings string `json:"cumulativeEarnings"`
}

func NewRewardsClient(baseURL string, log logger.Logger) (*RewardsClient, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("baseURL cannot be empty")
	}

	return &RewardsClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: log,
	}, nil
}

func (c *RewardsClient) GetSummarizedRewards(ctx context.Context, earnerAddress string) (*RewardsSummary, error) {
	if earnerAddress == "" {
		return nil, fmt.Errorf("earner address cannot be empty")
	}

	url := fmt.Sprintf("%s/api/v1/rewards/earner/%s/summary", c.baseURL, earnerAddress)

	c.logger.Info("Fetching rewards summary",
		zap.String("url", url),
		zap.String("earner", earnerAddress))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch rewards: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("sidecar returned status %d: %s", resp.StatusCode, string(body))
	}

	var summary RewardsSummary
	if err := json.NewDecoder(resp.Body).Decode(&summary); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	c.logger.Info("Successfully fetched rewards",
		zap.Int("tokenCount", len(summary.Tokens)))

	return &summary, nil
}

func (c *RewardsClient) GetClaimProof(ctx context.Context, earnerAddress string) (*ClaimProof, error) {
	if earnerAddress == "" {
		return nil, fmt.Errorf("earner address cannot be empty")
	}

	url := fmt.Sprintf("%s/api/v1/rewards/earner/%s/proof", c.baseURL, earnerAddress)

	c.logger.Info("Fetching claim proof",
		zap.String("url", url),
		zap.String("earner", earnerAddress))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch claim proof: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("sidecar returned status %d: %s", resp.StatusCode, string(body))
	}

	var proof ClaimProof
	if err := json.NewDecoder(resp.Body).Decode(&proof); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	c.logger.Info("Successfully fetched claim proof",
		zap.Uint32("rootIndex", proof.RootIndex),
		zap.Int("tokenCount", len(proof.TokenLeaves)))

	return &proof, nil
}

func (c *RewardsClient) Close() {
	c.logger.Debug("Rewards client closed")
}
package auth

import (
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"github.com/google/uuid"
)

// ChallengeTokenEntry represents a single challenge token with its metadata
type ChallengeTokenEntry struct {
	Token     string
	CreatedAt time.Time
	ExpiresAt time.Time
	Used      bool
}

// ChallengeTokenManager handles challenge token generation and validation
type ChallengeTokenManager struct {
	mu               sync.RWMutex
	tokens           map[string]*ChallengeTokenEntry
	authorizedEntity string // Can be operator address or aggregator address
	expiration       time.Duration
}

// NewChallengeTokenManager creates a new challenge token manager
func NewChallengeTokenManager(authorizedEntity string, expiration time.Duration) *ChallengeTokenManager {
	ctm := &ChallengeTokenManager{
		tokens:           make(map[string]*ChallengeTokenEntry),
		authorizedEntity: strings.ToLower(authorizedEntity),
		expiration:       expiration,
	}
	// Start cleanup goroutine
	go ctm.cleanupExpiredTokens()
	return ctm
}

// GenerateChallengeToken creates a new challenge token for the given entity
func (ctm *ChallengeTokenManager) GenerateChallengeToken(entity string) (*ChallengeTokenEntry, error) {
	ctm.mu.Lock()
	defer ctm.mu.Unlock()

	// Verify entity matches
	if !strings.EqualFold(entity, ctm.authorizedEntity) {
		return nil, fmt.Errorf("entity mismatch: expected %s, got %s", ctm.authorizedEntity, entity)
	}

	// Generate UUID and hash it
	uuidStr := uuid.New().String()
	hash := util.GetKeccak256Digest([]byte(uuidStr))
	token := hex.EncodeToString(hash[:])

	now := time.Now()
	entry := &ChallengeTokenEntry{
		Token:     token,
		CreatedAt: now,
		ExpiresAt: now.Add(ctm.expiration),
		Used:      false,
	}

	ctm.tokens[token] = entry
	return entry, nil
}

// UseChallengeToken validates and marks a challenge token as used
func (ctm *ChallengeTokenManager) UseChallengeToken(token string) error {
	ctm.mu.Lock()
	defer ctm.mu.Unlock()

	entry, exists := ctm.tokens[token]
	if !exists {
		return fmt.Errorf("challenge token not found")
	}

	if entry.Used {
		return fmt.Errorf("challenge token already used")
	}

	if time.Now().After(entry.ExpiresAt) {
		return fmt.Errorf("challenge token expired")
	}

	// Mark as used (keep in map for tracking)
	entry.Used = true
	return nil
}

// cleanupExpiredTokens periodically removes expired challenge tokens
func (ctm *ChallengeTokenManager) cleanupExpiredTokens() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		ctm.mu.Lock()
		now := time.Now()
		for token, entry := range ctm.tokens {
			if now.After(entry.ExpiresAt) {
				delete(ctm.tokens, token)
			}
		}
		ctm.mu.Unlock()
	}
}

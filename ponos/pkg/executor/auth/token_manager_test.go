package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChallengeTokenManager(t *testing.T) {
	t.Run("GenerateChallengeToken", func(t *testing.T) {
		ctm := NewChallengeTokenManager("0x123", 5*time.Minute)

		entry, err := ctm.GenerateChallengeToken("0x123")
		require.NoError(t, err)
		assert.NotEmpty(t, entry.Token)
		assert.Len(t, entry.Token, 64) // Keccak256 hash in hex
		assert.True(t, entry.ExpiresAt.After(time.Now()))
	})

	t.Run("GenerateChallengeToken_WrongOperator", func(t *testing.T) {
		ctm := NewChallengeTokenManager("0x123", 5*time.Minute)

		_, err := ctm.GenerateChallengeToken("0x456")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "operator address mismatch")
	})

	t.Run("UseChallengeToken_Success", func(t *testing.T) {
		ctm := NewChallengeTokenManager("0x123", 5*time.Minute)

		entry, err := ctm.GenerateChallengeToken("0x123")
		require.NoError(t, err)

		err = ctm.UseChallengeToken(entry.Token)
		assert.NoError(t, err)
	})

	t.Run("UseChallengeToken_AlreadyUsed", func(t *testing.T) {
		ctm := NewChallengeTokenManager("0x123", 5*time.Minute)

		entry, err := ctm.GenerateChallengeToken("0x123")
		require.NoError(t, err)

		// Use token once
		err = ctm.UseChallengeToken(entry.Token)
		require.NoError(t, err)

		// Try to use again
		err = ctm.UseChallengeToken(entry.Token)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "challenge token already used")
	})

	t.Run("UseChallengeToken_NotFound", func(t *testing.T) {
		ctm := NewChallengeTokenManager("0x123", 5*time.Minute)

		err := ctm.UseChallengeToken("non-existent-token")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "challenge token not found")
	})

	t.Run("UseChallengeToken_Expired", func(t *testing.T) {
		ctm := NewChallengeTokenManager("0x123", 1*time.Millisecond)

		entry, err := ctm.GenerateChallengeToken("0x123")
		require.NoError(t, err)

		// Wait for expiration
		time.Sleep(2 * time.Millisecond)

		err = ctm.UseChallengeToken(entry.Token)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "challenge token expired")
	})
}

func TestConstructSignedMessage(t *testing.T) {
	challengeToken := "test-token"
	methodName := "DeployArtifact"
	requestPayload := []byte("test-payload")

	message := ConstructSignedMessage(challengeToken, methodName, requestPayload)

	// Verify it's a hash (32 bytes)
	assert.Len(t, message, 32)

	// Verify deterministic
	message2 := ConstructSignedMessage(challengeToken, methodName, requestPayload)
	assert.Equal(t, message, message2)

	// Verify different inputs produce different hashes
	message3 := ConstructSignedMessage("different-token", methodName, requestPayload)
	assert.NotEqual(t, message, message3)
}

package aggregator

import (
	"testing"

	commonV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/common"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/logger"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestVerifyAuth tests the verifyAuth method behavior
func TestVerifyAuth(t *testing.T) {
	l, _ := logger.NewLogger(&logger.LoggerConfig{Debug: false})

	t.Run("NoAuthVerifier_NoAuthProvided", func(t *testing.T) {
		// Aggregator without auth verifier
		agg := &Aggregator{
			logger:       l,
			authVerifier: nil,
		}

		// No auth provided - should succeed
		err := agg.verifyAuth(nil)
		assert.NoError(t, err)
	})

	t.Run("NoAuthVerifier_AuthProvided", func(t *testing.T) {
		// Aggregator without auth verifier
		agg := &Aggregator{
			logger:       l,
			authVerifier: nil,
		}

		// Auth provided when not enabled - should return Unimplemented
		auth := &commonV1.AuthSignature{
			ChallengeToken: "test-token",
			Signature:      []byte("test-sig"),
		}

		err := agg.verifyAuth(auth)
		assert.Error(t, err)

		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unimplemented, statusErr.Code())
		assert.Contains(t, statusErr.Message(), "authentication is not enabled")
	})
}

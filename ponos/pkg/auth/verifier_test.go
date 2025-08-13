package auth

import (
	"testing"
	"time"

	cryptoLibsEcdsa "github.com/Layr-Labs/crypto-libs/pkg/ecdsa"
	commonV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/common"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer/inMemorySigner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAuthVerifier(t *testing.T) {
	entityAddress := "0xTestEntity123"
	tokenManager := NewChallengeTokenManager(entityAddress, 5*time.Minute)

	// Create a test signer with a proper ECDSA private key
	// This is a test private key for testing purposes only
	testPrivateKeyHex := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	testPrivateKey, err := cryptoLibsEcdsa.NewPrivateKeyFromHexString(testPrivateKeyHex)
	require.NoError(t, err)

	testSigner := inMemorySigner.NewInMemorySigner(testPrivateKey, config.CurveTypeECDSA)
	verifier := NewVerifier(tokenManager, testSigner)

	t.Run("VerifyAuthentication_Success", func(t *testing.T) {
		// Generate a token
		entry, err := verifier.GenerateChallengeToken(entityAddress)
		require.NoError(t, err)

		// Create auth signature (simplified - only sign the token)
		signedMessage := ConstructSignedMessage(entry.Token)
		signature, err := testSigner.SignMessage(signedMessage)
		require.NoError(t, err)

		authSig := &commonV1.AuthSignature{
			ChallengeToken: entry.Token,
			Signature:      signature,
		}

		// Verify
		err = verifier.VerifyAuthentication(authSig)
		assert.NoError(t, err)
	})

	t.Run("VerifyAuthentication_MissingAuth", func(t *testing.T) {
		err := verifier.VerifyAuthentication(nil)
		assert.Error(t, err)
		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, statusErr.Code())
	})

	t.Run("VerifyAuthentication_InvalidToken", func(t *testing.T) {
		authSig := &commonV1.AuthSignature{
			ChallengeToken: "invalid-token",
			Signature:      []byte("signature"),
		}

		err := verifier.VerifyAuthentication(authSig)
		assert.Error(t, err)
		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, statusErr.Code())
	})

	t.Run("VerifyAuthentication_WrongSignature", func(t *testing.T) {
		// Generate a token
		entry, err := verifier.GenerateChallengeToken(entityAddress)
		require.NoError(t, err)

		authSig := &commonV1.AuthSignature{
			ChallengeToken: entry.Token,
			Signature:      []byte("wrong-signature"),
		}

		err = verifier.VerifyAuthentication(authSig)
		assert.Error(t, err)
		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, statusErr.Code())
	})

	t.Run("VerifyAuthentication_DifferentSigner", func(t *testing.T) {
		// Create a different signer
		differentKeyHex := "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
		differentKey, err := cryptoLibsEcdsa.NewPrivateKeyFromHexString(differentKeyHex)
		require.NoError(t, err)
		differentSigner := inMemorySigner.NewInMemorySigner(differentKey, config.CurveTypeECDSA)

		// Generate a token
		entry, err := verifier.GenerateChallengeToken(entityAddress)
		require.NoError(t, err)

		// Sign with different signer
		signedMessage := ConstructSignedMessage(entry.Token)
		signature, err := differentSigner.SignMessage(signedMessage)
		require.NoError(t, err)

		authSig := &commonV1.AuthSignature{
			ChallengeToken: entry.Token,
			Signature:      signature,
		}

		// Should fail verification
		err = verifier.VerifyAuthentication(authSig)
		assert.Error(t, err)
		statusErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, statusErr.Code())
	})
}
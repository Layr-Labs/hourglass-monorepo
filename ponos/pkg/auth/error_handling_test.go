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

// MockFailingSigner is a signer that always fails with an error
type MockFailingSigner struct{}

func (m *MockFailingSigner) SignMessage(message []byte) ([]byte, error) {
	return nil, assert.AnError
}

func (m *MockFailingSigner) SignMessageForSolidity(message []byte) ([]byte, error) {
	return nil, assert.AnError
}

func (m *MockFailingSigner) GetPublicKey() []byte {
	return []byte("mock-public-key")
}

func (m *MockFailingSigner) GetPublicKeyString() string {
	return "mock-public-key"
}

func (m *MockFailingSigner) GetEthereumAddress() string {
	return "0xMockAddress"
}

func (m *MockFailingSigner) GetCurveType() config.CurveType {
	return config.CurveTypeECDSA
}

// TestVerifyAuthenticationErrorCodes tests that VerifyAuthentication returns the correct error codes
func TestVerifyAuthenticationErrorCodes(t *testing.T) {
	entityAddress := "0xTestEntity"
	tokenManager := NewChallengeTokenManager(entityAddress, 5*time.Minute)

	t.Run("Returns_Unauthenticated_For_Missing_Auth", func(t *testing.T) {
		verifier := NewVerifier(tokenManager, &MockFailingSigner{})
		
		err := verifier.VerifyAuthentication(nil)
		require.Error(t, err)
		
		statusErr, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, statusErr.Code())
		assert.Contains(t, statusErr.Message(), "missing authentication")
	})

	t.Run("Returns_Unauthenticated_For_Invalid_Token", func(t *testing.T) {
		verifier := NewVerifier(tokenManager, &MockFailingSigner{})
		
		authSig := &commonV1.AuthSignature{
			ChallengeToken: "invalid-token",
			Signature:      []byte("signature"),
		}
		
		err := verifier.VerifyAuthentication(authSig)
		require.Error(t, err)
		
		statusErr, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, statusErr.Code())
		assert.Contains(t, statusErr.Message(), "invalid challenge token")
	})

	t.Run("Returns_Internal_For_Signer_Failure", func(t *testing.T) {
		// Use a signer that will fail
		failingSigner := &MockFailingSigner{}
		verifier := NewVerifier(tokenManager, failingSigner)
		
		// Generate a valid token
		entry, err := tokenManager.GenerateChallengeToken(entityAddress)
		require.NoError(t, err)
		
		authSig := &commonV1.AuthSignature{
			ChallengeToken: entry.Token,
			Signature:      []byte("some-signature"),
		}
		
		// This should fail with Internal error because the signer fails
		err = verifier.VerifyAuthentication(authSig)
		require.Error(t, err)
		
		statusErr, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Internal, statusErr.Code())
		assert.Contains(t, statusErr.Message(), "failed to generate expected signature")
	})

	t.Run("Returns_Unauthenticated_For_Wrong_Signature", func(t *testing.T) {
		// Create a real signer for this test
		testPrivateKeyHex := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
		testPrivateKey, err := cryptoLibsEcdsa.NewPrivateKeyFromHexString(testPrivateKeyHex)
		require.NoError(t, err)
		testSigner := inMemorySigner.NewInMemorySigner(testPrivateKey, config.CurveTypeECDSA)
		
		verifier := NewVerifier(tokenManager, testSigner)
		
		// Generate a valid token
		entry, err := tokenManager.GenerateChallengeToken(entityAddress)
		require.NoError(t, err)
		
		authSig := &commonV1.AuthSignature{
			ChallengeToken: entry.Token,
			Signature:      []byte("wrong-signature"),
		}
		
		// This should fail with Unauthenticated because signature doesn't match
		err = verifier.VerifyAuthentication(authSig)
		require.Error(t, err)
		
		statusErr, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, statusErr.Code())
		assert.Contains(t, statusErr.Message(), "invalid signature")
	})

	t.Run("Returns_Unauthenticated_For_Expired_Token", func(t *testing.T) {
		// Create a token manager with very short expiration
		shortExpiryManager := NewChallengeTokenManager(entityAddress, 1*time.Millisecond)
		
		testPrivateKeyHex := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
		testPrivateKey, err := cryptoLibsEcdsa.NewPrivateKeyFromHexString(testPrivateKeyHex)
		require.NoError(t, err)
		testSigner := inMemorySigner.NewInMemorySigner(testPrivateKey, config.CurveTypeECDSA)
		
		verifier := NewVerifier(shortExpiryManager, testSigner)
		
		// Generate a token
		entry, err := shortExpiryManager.GenerateChallengeToken(entityAddress)
		require.NoError(t, err)
		
		// Wait for token to expire
		time.Sleep(2 * time.Millisecond)
		
		// Create valid signature
		signedMessage := ConstructSignedMessage(entry.Token)
		signature, err := testSigner.SignMessage(signedMessage)
		require.NoError(t, err)
		
		authSig := &commonV1.AuthSignature{
			ChallengeToken: entry.Token,
			Signature:      signature,
		}
		
		// This should fail with Unauthenticated because token is expired
		err = verifier.VerifyAuthentication(authSig)
		require.Error(t, err)
		
		statusErr, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, statusErr.Code())
		assert.Contains(t, statusErr.Message(), "invalid challenge token")
		// The underlying error should mention "expired"
	})

	t.Run("Returns_Unauthenticated_For_Reused_Token", func(t *testing.T) {
		testPrivateKeyHex := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
		testPrivateKey, err := cryptoLibsEcdsa.NewPrivateKeyFromHexString(testPrivateKeyHex)
		require.NoError(t, err)
		testSigner := inMemorySigner.NewInMemorySigner(testPrivateKey, config.CurveTypeECDSA)
		
		verifier := NewVerifier(tokenManager, testSigner)
		
		// Generate a token
		entry, err := tokenManager.GenerateChallengeToken(entityAddress)
		require.NoError(t, err)
		
		// Create valid signature
		signedMessage := ConstructSignedMessage(entry.Token)
		signature, err := testSigner.SignMessage(signedMessage)
		require.NoError(t, err)
		
		authSig := &commonV1.AuthSignature{
			ChallengeToken: entry.Token,
			Signature:      signature,
		}
		
		// First use should succeed
		err = verifier.VerifyAuthentication(authSig)
		assert.NoError(t, err)
		
		// Second use should fail with Unauthenticated
		err = verifier.VerifyAuthentication(authSig)
		require.Error(t, err)
		
		statusErr, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, statusErr.Code())
		assert.Contains(t, statusErr.Message(), "invalid challenge token")
		// The underlying error should mention "already used"
	})
}
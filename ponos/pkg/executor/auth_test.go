package executor

import (
	"testing"

	executorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// mockSigner implements signer.ISigner for testing
type mockSigner struct{}

func (m *mockSigner) SignMessage(message []byte) ([]byte, error) {
	// Return a predictable signature for testing
	return append([]byte("signature:"), message[:8]...), nil
}

func (m *mockSigner) SignMessageForSolidity(message []byte) ([]byte, error) {
	return m.SignMessage(message)
}

func (m *mockSigner) GetAddress() string {
	return "0x123"
}

func TestVerifyAuthentication(t *testing.T) {
	tokenManager := auth.NewChallengeTokenManager("0x123", 5*60)
	verifier := auth.NewVerifier(tokenManager, &mockSigner{})

	t.Run("Success", func(t *testing.T) {
		// Generate a challenge token
		entry, err := tokenManager.GenerateChallengeToken("0x123")
		require.NoError(t, err)

		// Create request payload
		requestPayload := []byte("test-request")
		
		// Create expected signature
		signedMessage := auth.ConstructSignedMessage(entry.Token, "TestMethod", requestPayload)
		expectedSig, err := (&mockSigner{}).SignMessage(signedMessage)
		require.NoError(t, err)

		// Create auth with correct signature
		authSig := &executorV1.AuthSignature{
			ChallengeToken: entry.Token,
			Signature:      expectedSig,
		}

		// Verify authentication
		err = verifier.VerifyAuthentication(authSig, "TestMethod", requestPayload)
		assert.NoError(t, err)
	})

	t.Run("InvalidSignature", func(t *testing.T) {
		// Generate a challenge token
		entry, err := tokenManager.GenerateChallengeToken("0x123")
		require.NoError(t, err)

		// Create auth with wrong signature
		authSig := &executorV1.AuthSignature{
			ChallengeToken: entry.Token,
			Signature:      []byte("wrong-signature"),
		}

		// Verify authentication
		err = verifier.VerifyAuthentication(authSig, "TestMethod", []byte("test-request"))
		assert.Error(t, err)
		
		// Check it's an unauthenticated error
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, st.Code())
		assert.Contains(t, st.Message(), "invalid signature")
	})

	t.Run("MissingAuth", func(t *testing.T) {
		err := verifier.VerifyAuthentication(nil, "TestMethod", []byte("test-request"))
		assert.Error(t, err)
		
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, st.Code())
		assert.Contains(t, st.Message(), "missing authentication")
	})

	t.Run("InvalidChallengeToken", func(t *testing.T) {
		authSig := &executorV1.AuthSignature{
			ChallengeToken: "invalid-token",
			Signature:      []byte("signature"),
		}

		err := verifier.VerifyAuthentication(authSig, "TestMethod", []byte("test-request"))
		assert.Error(t, err)
		
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, st.Code())
		assert.Contains(t, st.Message(), "invalid challenge token")
	})
}

func TestGetRequestWithoutAuth(t *testing.T) {
	t.Run("DeployArtifactRequest", func(t *testing.T) {
		req := &executorV1.DeployArtifactRequest{
			AvsAddress:  "0xAVS",
			Digest:      "sha256:abc",
			RegistryUrl: "registry.example.com",
			Auth: &executorV1.AuthSignature{
				ChallengeToken: "test-token",
				Signature:      []byte("test-sig"),
			},
		}

		requestBytes, err := auth.GetRequestWithoutAuth(req)
		require.NoError(t, err)

		// Unmarshal and verify auth is nil
		var unmarshaled executorV1.DeployArtifactRequest
		err = proto.Unmarshal(requestBytes, &unmarshaled)
		require.NoError(t, err)
		assert.Nil(t, unmarshaled.Auth)
		assert.Equal(t, req.AvsAddress, unmarshaled.AvsAddress)
	})

	t.Run("ListPerformersRequest", func(t *testing.T) {
		req := &executorV1.ListPerformersRequest{
			AvsAddress: "0xAVS",
			Auth: &executorV1.AuthSignature{
				ChallengeToken: "test-token",
				Signature:      []byte("test-sig"),
			},
		}

		requestBytes, err := auth.GetRequestWithoutAuth(req)
		require.NoError(t, err)

		// Unmarshal and verify auth is nil
		var unmarshaled executorV1.ListPerformersRequest
		err = proto.Unmarshal(requestBytes, &unmarshaled)
		require.NoError(t, err)
		assert.Nil(t, unmarshaled.Auth)
		assert.Equal(t, req.AvsAddress, unmarshaled.AvsAddress)
	})

	t.Run("RemovePerformerRequest", func(t *testing.T) {
		req := &executorV1.RemovePerformerRequest{
			PerformerId: "performer-123",
			Auth: &executorV1.AuthSignature{
				ChallengeToken: "test-token",
				Signature:      []byte("test-sig"),
			},
		}

		requestBytes, err := auth.GetRequestWithoutAuth(req)
		require.NoError(t, err)

		// Unmarshal and verify auth is nil
		var unmarshaled executorV1.RemovePerformerRequest
		err = proto.Unmarshal(requestBytes, &unmarshaled)
		require.NoError(t, err)
		assert.Nil(t, unmarshaled.Auth)
		assert.Equal(t, req.PerformerId, unmarshaled.PerformerId)
	})
}
package auth

import (
	"fmt"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthSignature represents the authentication signature for a request
type AuthSignature interface {
	GetChallengeToken() string
	GetSignature() []byte
}

// Verifier handles authentication verification for requests
type Verifier struct {
	tokenManager *ChallengeTokenManager
	signer       signer.ISigner
}

// NewVerifier creates a new authentication verifier
func NewVerifier(tokenManager *ChallengeTokenManager, signer signer.ISigner) *Verifier {
	return &Verifier{
		tokenManager: tokenManager,
		signer:       signer,
	}
}

// GenerateChallengeToken generates a new challenge token for the given entity
func (v *Verifier) GenerateChallengeToken(entity string) (*ChallengeTokenEntry, error) {
	return v.tokenManager.GenerateChallengeToken(entity)
}

// VerifyAuthentication verifies the authentication signature for a request
func (v *Verifier) VerifyAuthentication(auth AuthSignature, methodName string, requestPayload []byte) error {
	if auth == nil {
		return status.Error(codes.Unauthenticated, "missing authentication")
	}

	// Use the challenge token (this also validates it)
	if err := v.tokenManager.UseChallengeToken(auth.GetChallengeToken()); err != nil {
		return status.Errorf(codes.Unauthenticated, "invalid challenge token: %v", err)
	}

	// Construct the message that was signed
	signedMessage := ConstructSignedMessage(auth.GetChallengeToken(), methodName, requestPayload)

	// Verify the signature matches our entity's signature
	expectedSig, err := v.signer.SignMessage(signedMessage)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to generate expected signature: %v", err)
	}

	// Compare signatures
	if !bytesEqual(auth.GetSignature(), expectedSig) {
		return status.Error(codes.Unauthenticated, "invalid signature")
	}

	return nil
}

// ConstructSignedMessage creates the message to be signed
func ConstructSignedMessage(challengeToken, methodName string, requestPayload []byte) []byte {
	// Create a deterministic message to sign
	message := fmt.Sprintf("%s:%s:", challengeToken, methodName)
	messageBytes := append([]byte(message), requestPayload...)

	// Return the hash of the message
	digest := util.GetKeccak256Digest(messageBytes)
	return digest[:]
}

// bytesEqual compares two byte slices for equality
func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
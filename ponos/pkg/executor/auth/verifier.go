package auth

import (
	"fmt"

	executorV1 "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

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

// GenerateChallengeToken generates a new challenge token for the given operator
func (v *Verifier) GenerateChallengeToken(operatorAddress string) (*ChallengeTokenEntry, error) {
	return v.tokenManager.GenerateChallengeToken(operatorAddress)
}

// VerifyAuthentication verifies the authentication signature for a request
func (v *Verifier) VerifyAuthentication(auth *executorV1.AuthSignature, methodName string, requestPayload []byte) error {
	if auth == nil {
		return status.Error(codes.Unauthenticated, "missing authentication")
	}

	// Use the challenge token (this also validates it)
	if err := v.tokenManager.UseChallengeToken(auth.ChallengeToken); err != nil {
		return status.Errorf(codes.Unauthenticated, "invalid challenge token: %v", err)
	}

	// Construct the message that was signed
	signedMessage := ConstructSignedMessage(auth.ChallengeToken, methodName, requestPayload)

	// Verify the signature matches our operator's signature
	expectedSig, err := v.signer.SignMessage(signedMessage)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to generate expected signature: %v", err)
	}

	// Compare signatures
	if !bytesEqual(auth.Signature, expectedSig) {
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

// GetRequestWithoutAuth returns a copy of the request with the auth field removed
func GetRequestWithoutAuth[T proto.Message](req T) ([]byte, error) {
	// Clone the request
	cloned := proto.Clone(req)

	// Use reflection to set the auth field to nil
	switch v := any(cloned).(type) {
	case *executorV1.DeployArtifactRequest:
		v.Auth = nil
	case *executorV1.ListPerformersRequest:
		v.Auth = nil
	case *executorV1.RemovePerformerRequest:
		v.Auth = nil
	default:
		return nil, fmt.Errorf("unsupported request type")
	}

	// Marshal the request without auth
	return proto.Marshal(cloned)
}

package auth

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestHandleAuthError(t *testing.T) {
	t.Run("NilError", func(t *testing.T) {
		err := HandleAuthError(nil)
		assert.NoError(t, err)
	})

	t.Run("PreservesExistingStatusCode", func(t *testing.T) {
		// Test with PermissionDenied
		originalErr := status.Error(codes.PermissionDenied, "access denied")
		handledErr := HandleAuthError(originalErr)

		statusErr, ok := status.FromError(handledErr)
		assert.True(t, ok)
		assert.Equal(t, codes.PermissionDenied, statusErr.Code())
		assert.Equal(t, "access denied", statusErr.Message())
	})

	t.Run("PreservesUnimplemented", func(t *testing.T) {
		// Test with Unimplemented
		originalErr := status.Error(codes.Unimplemented, "not implemented")
		handledErr := HandleAuthError(originalErr)

		statusErr, ok := status.FromError(handledErr)
		assert.True(t, ok)
		assert.Equal(t, codes.Unimplemented, statusErr.Code())
		assert.Equal(t, "not implemented", statusErr.Message())
	})

	t.Run("ConvertsRegularErrorToUnauthenticated", func(t *testing.T) {
		// Regular Go error should become Unauthenticated
		originalErr := errors.New("regular error")
		handledErr := HandleAuthError(originalErr)

		statusErr, ok := status.FromError(handledErr)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, statusErr.Code())
		assert.Equal(t, "regular error", statusErr.Message())
	})
}

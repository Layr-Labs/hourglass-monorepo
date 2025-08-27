package auth

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HandleAuthError converts authentication errors to appropriate gRPC status codes.
// This provides consistent error handling across all services that require authentication.
//
// If the error is nil, returns nil.
// If the error already has a gRPC status code, preserves it.
// Otherwise, returns an Unauthenticated status error.
func HandleAuthError(err error) error {
	if err == nil {
		return nil
	}

	// Preserve original status code if it's already a status error
	if s, ok := status.FromError(err); ok {
		return status.Error(s.Code(), s.Message())
	}

	// Default to Unauthenticated for non-status errors
	return status.Error(codes.Unauthenticated, err.Error())
}

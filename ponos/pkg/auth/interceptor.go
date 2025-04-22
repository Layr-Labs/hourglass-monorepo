package auth

import (
	"context"
	executorpb "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TODO: make generic/configurable for Executor & Aggregator types.
func EthereumAuthInterceptor(allowedAddresses map[string]bool) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if info.FullMethod != "/eigenlayer.hourglass.v1.ExecutorService/SubmitTask" {
			return handler(ctx, req)
		}

		submission, ok := req.(*executorpb.TaskSubmission)
		if !ok {
			return nil, status.Error(codes.InvalidArgument, "invalid request type")
		}

		addr, err := VerifyEthereumSubmission(
			submission.GetPayload(),
			submission.GetSignature(),
			"", // TODO: pass the public key here
		)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "signature verification failed: %v", err)
		}

		if !allowedAddresses[addr] {
			return nil, status.Error(codes.PermissionDenied, "unauthorized sender")
		}

		return handler(ctx, req)
	}
}

func VerifyEthereumSubmission(payload []byte, sig []byte, pubKeyHex string) (string, error) {
	// TODO: verify inputs and return valid address.
	return "valid-address", nil
}

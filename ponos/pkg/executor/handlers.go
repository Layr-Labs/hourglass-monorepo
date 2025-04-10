package executor

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/executor/connectedAggregator"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"time"
)

func (e *Executor) Handshake(ctx context.Context, request *executor.HandshakeRequest) (*executor.HandshakeResponse, error) {
	aggregatorAddress := request.GetAggregatorAddress()
	if aggregatorAddress == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Aggregator address is required")
	}

	avsAddress := request.GetAvsAddress()
	if avsAddress == "" {
		return nil, status.Errorf(codes.InvalidArgument, "AVS address is required")
	}

	nonce := request.GetNonce()
	if nonce == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Nonce is required")
	}

	nonceSig := request.GetAggregatorNonceSig()
	if len(nonceSig) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Nonce signature is required")
	}

	handshake, err := e.PerformAggregatorHandshake(
		ctx,
		avsAddress,
		aggregatorAddress,
		nonce,
		nonceSig,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to perform handshake with aggregator: %v", err)
	}
	return &executor.HandshakeResponse{
		OperatorSignedNonce: handshake.NonceSignature,
		AuthToken:           handshake.AuthToken,
		AuthTokenSignature:  handshake.AuthTokenSignature,
	}, nil
}

func (e *Executor) WorkStream(stream grpc.BidiStreamingServer[executor.WorkStreamRequest, executor.WorkStreamResponse]) error {
	waitc := make(chan struct{})

	go func() {
		for {
			req, err := stream.Recv()
			if err == io.EOF {
				close(waitc)
				return
			}

			if err != nil {
				e.logger.Sugar().Errorw("Failed to receive request from stream",
					zap.Error(err),
				)
				close(waitc)
				continue
			}
			res, err := e.handleReceivedWorkStreamMessage(req)
			if err != nil {
				e.logger.Sugar().Errorw("Failed to handle received work stream message",
					zap.Error(err),
				)
				continue
			}
			if res != nil {
				if err := stream.Send(&executor.WorkStreamResponse{Payload: res}); err != nil {
					e.logger.Sugar().Errorw("Failed to send response to stream",
						zap.Any("response", res),
						zap.Error(err),
					)
				}
			}
		}
	}()
	<-waitc
	e.logger.Sugar().Info("Work stream closed")
	return nil
}

func (e *Executor) handleReceivedWorkStreamMessage(req *executor.WorkStreamRequest) (*executor.ResponsePayload, error) {
	fmt.Printf("Received work stream message: %v\n", req)
	if req.Payload == nil {
		e.logger.Sugar().Errorw("Received empty payload in work stream message")
		return nil, nil
	}

	switch req.Payload.Payload.(type) {
	case *executor.RequestPayload_Ping:
		return &executor.ResponsePayload{
			Payload: &executor.ResponsePayload_Pong{
				Pong: &executor.Pong{
					CurrentTime: uint64(time.Now().Unix()),
				},
			},
		}, nil
	default:
		return nil, nil
	}
}

func (e *Executor) PerformAggregatorHandshake(
	ctx context.Context,
	avsAddress string,
	aggregatorAddress string,
	nonce string,
	nonceSig []byte,
) (*connectedAggregator.HandshakeResponse, error) {
	aggId := connectedAggregator.BuildIdFromAvsAndAggregatorAddress(avsAddress, aggregatorAddress)

	agg, ok := e.aggregators[aggId]
	if ok {
		e.logger.Sugar().Infow("existing connection detected; terminating",
			zap.String("avsAddress", avsAddress),
			zap.String("aggregatorAddress", aggregatorAddress),
		)
		if err := agg.Terminate(); err != nil {
			e.logger.Sugar().Errorw("Failed to terminate existing connection",
				zap.String("avsAddress", avsAddress),
				zap.String("aggregatorAddress", aggregatorAddress),
				zap.Error(err),
			)
			return nil, fmt.Errorf("failed to terminate existing connection: %v", err)
		}
	}

	// create a new ConnectedAggregator instance
	agg = connectedAggregator.NewConnectedAggregator(avsAddress, aggregatorAddress, e.logger, e.signer)

	handshakeRes, err := agg.Handshake(nonce, nonceSig)
	if err != nil {
		e.logger.Sugar().Errorw("Failed to perform handshake with aggregator",
			zap.String("avsAddress", avsAddress),
			zap.String("aggregatorAddress", aggregatorAddress),
			zap.Error(err),
		)
		return nil, err
	}

	e.aggregators[aggId] = agg
	return handshakeRes, nil
}

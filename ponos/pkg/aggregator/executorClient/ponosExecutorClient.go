package executorClient

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"google.golang.org/grpc/credentials/insecure"

	executorpb "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"

	"github.com/ethereum/go-ethereum/crypto"
	"google.golang.org/grpc"
)

type PonosExecutorClient struct {
	client  executorpb.ExecutorServiceClient
	conn    *grpc.ClientConn
	privKey *ecdsa.PrivateKey
}

func NewPonosExecutorClient(grpcAddr string, privateKey *ecdsa.PrivateKey) (*PonosExecutorClient, error) {
	conn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to executor: %w", err)
	}

	client := executorpb.NewExecutorServiceClient(conn)

	return &PonosExecutorClient{
		client:  client,
		conn:    conn,
		privKey: privateKey,
	}, nil
}

func (pec *PonosExecutorClient) SubmitTask(ctx context.Context, task *types.Task) error {
	sig, err := signTask(task, pec.privKey)
	if err != nil {
		return fmt.Errorf("failed to sign task: %w", err)
	}

	submission := &executorpb.TaskSubmission{
		TaskId:            task.TaskId,
		AggregatorAddress: task.CallbackAddr,
		Payload:           task.Payload,
		Signature:         sig,
	}

	ack, err := pec.client.SubmitTask(ctx, submission)
	if err != nil {
		return fmt.Errorf("grpc submit error: %w", err)
	}

	if !ack.Success {
		return fmt.Errorf("executor returned failure: %s", ack.Message)
	}

	return nil
}

func signTask(task *types.Task, privKey *ecdsa.PrivateKey) ([]byte, error) {
	// Very simple message hash using task ID (for demo purposes only!)
	hash := crypto.Keccak256([]byte(task.TaskId))
	sig, err := crypto.Sign(hash, privKey)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

func (pec *PonosExecutorClient) Close() error {
	return pec.conn.Close()
}

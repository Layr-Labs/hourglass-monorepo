package executorClient

import (
	"context"
	"fmt"
	executorpb "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/signer"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"

	"google.golang.org/grpc"
)

type PonosExecutorClient struct {
	client executorpb.ExecutorServiceClient
	conn   *grpc.ClientConn
	signer signer.ISigner
}

func NewPonosExecutorClient(
	executorClient executorpb.ExecutorServiceClient,
	signer signer.ISigner,
) *PonosExecutorClient {

	return &PonosExecutorClient{
		client: executorClient,
		signer: signer,
	}
}

func NewPonosExecutorClientWithConn(
	conn *grpc.ClientConn,
	signer signer.ISigner,
) *PonosExecutorClient {
	return &PonosExecutorClient{
		client: executorpb.NewExecutorServiceClient(conn),
		conn:   conn,
		signer: signer,
	}
}

func (pec *PonosExecutorClient) SubmitTask(ctx context.Context, task *types.Task) error {
	sig, err := pec.signer.SignMessage([]byte(task.TaskId))
	if err != nil {
		return fmt.Errorf("failed to sign task: %w", err)
	}

	submission := &executorpb.TaskSubmission{
		TaskId:            task.TaskId,
		AvsAddress:        task.AVSAddress,
		AggregatorAddress: "0xaggregatorOperatoraddress",
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

func (pec *PonosExecutorClient) Close() error {
	return pec.conn.Close()
}

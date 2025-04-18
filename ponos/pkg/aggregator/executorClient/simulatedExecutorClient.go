package executorClient

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	executorpb "github.com/Layr-Labs/hourglass-monorepo/ponos/gen/protos/eigenlayer/hourglass/v1/executor"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
	"google.golang.org/grpc/status"
)

type SimulatedExecutorClient struct {
	config    *SimulatedExecutorClientConfig
	validKeys map[string]bool
}

type SimulatedExecutorClientConfig struct {
	ResponseDelay   time.Duration
	TimeoutDuration time.Duration
	TimeoutRate     float64
	AuthFailureRate float64
}

func NewSimulatedExecutorClient(cfg *SimulatedExecutorClientConfig, validKeys []string) *SimulatedExecutorClient {
	keyMap := make(map[string]bool)
	for _, k := range validKeys {
		keyMap[k] = true
	}
	return &SimulatedExecutorClient{
		config:    cfg,
		validKeys: keyMap,
	}
}

func (sec *SimulatedExecutorClient) SubmitTask(task *types.Task) error {
	if rand.Float64() < sec.config.TimeoutRate {
		time.Sleep(sec.config.TimeoutDuration)
		return errors.New("simulated timeout")
	}

	time.Sleep(sec.config.ResponseDelay)

	if rand.Float64() < sec.config.AuthFailureRate || !sec.validKeys[task.CallbackAddr] {
		return status.Error(403, "simulated authentication failure")
	}

	submission := &executorpb.TaskSubmission{
		TaskId:            task.TaskId,
		AggregatorAddress: task.CallbackAddr,
		Payload:           task.Payload,
		PublicKey:         task.CallbackAddr,
		Signature:         []byte("signature"),
	}

	fmt.Printf("[SIMULATED EXECUTOR] Task submitted: %s\n", submission.TaskId)
	return nil
}

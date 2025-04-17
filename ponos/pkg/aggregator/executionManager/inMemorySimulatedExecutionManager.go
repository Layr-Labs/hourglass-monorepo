package executionManager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
)

// TODO: implement ICapacityManager to inform coordinator of available capacity.
// TODO: inject IConnectionManager to negotiate connections and communication.
type InMemorySimulatedExecutionManager struct {
	mu      sync.Mutex
	results []*types.TaskResult
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

func NewInMemorySimulatedExecutionManager() *InMemorySimulatedExecutionManager {
	return &InMemorySimulatedExecutionManager{
		results: make([]*types.TaskResult, 0),
	}
}

func (em *InMemorySimulatedExecutionManager) Start(ctx context.Context) error {
	em.ctx, em.cancel = context.WithCancel(ctx)
	return nil
}

func (em *InMemorySimulatedExecutionManager) Close() error {
	if em.cancel != nil {
		em.cancel()
	}
	em.wg.Wait()
	return nil
}

func (em *InMemorySimulatedExecutionManager) ExecuteTask(task *types.Task) error {
	em.wg.Add(1)
	go func() {
		defer em.wg.Done()
		em.simulateExecution(task)
	}()
	return nil
}

func (em *InMemorySimulatedExecutionManager) simulateExecution(task *types.Task) {
	time.Sleep(50 * time.Millisecond)

	result := &types.TaskResult{
		TaskId:        task.TaskId,
		AvsAddress:    task.AVSAddress,
		CallbackAddr:  task.CallbackAddr,
		OperatorSetId: task.OperatorSetId,
		Output:        []byte(fmt.Sprintf("Output for: %s", task.TaskId)),
		ChainId:       task.ChainId,
		BlockNumber:   task.BlockNumber,
		BlockHash:     task.BlockHash,
	}

	em.mu.Lock()
	em.results = append(em.results, result)
	em.mu.Unlock()
}

func (em *InMemorySimulatedExecutionManager) LoadResults() []*types.TaskResult {
	em.mu.Lock()
	defer em.mu.Unlock()

	results := em.results
	em.results = nil
	return results
}

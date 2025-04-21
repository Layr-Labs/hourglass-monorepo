package executionManager

import (
	"context"
	"sync"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/executorClient"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/types"
)

type PonosExecutionManager struct {
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	mu         sync.Mutex
	results    []*types.TaskResult
	execClient executorClient.IExecutorClient
}

func NewPonosExecutionManager(execClient executorClient.IExecutorClient) *PonosExecutionManager {
	return &PonosExecutionManager{
		results:    make([]*types.TaskResult, 0),
		execClient: execClient,
	}
}

func (em *PonosExecutionManager) Start(ctx context.Context) error {
	em.ctx, em.cancel = context.WithCancel(ctx)
	return nil
}

func (em *PonosExecutionManager) Close() error {
	if em.cancel != nil {
		em.cancel()
	}
	em.wg.Wait()
	return nil
}

func (em *PonosExecutionManager) ExecuteTask(task *types.Task) error {
	em.wg.Add(1)
	go func() {
		defer em.wg.Done()
		err := em.execClient.SubmitTask(task)
		if err != nil {
			return
		}
	}()
	return nil
}

func (em *PonosExecutionManager) LoadResults() []*types.TaskResult {
	em.mu.Lock()
	defer em.mu.Unlock()

	results := em.results
	em.results = nil
	return results
}

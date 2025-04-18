package workQueue

import (
	"container/list"
	"fmt"
	"sync"

	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/aggregator/metrics"
)

const (
	queueLengthMetricFormat = "%sQueueLength"
)

type WorkQueue[T any] struct {
	mu             sync.Mutex
	queue          *list.List
	cond           *sync.Cond
	metricsContext metrics.MetricsContext
	queueName      string
}

func NewWorkQueue[T any]() *WorkQueue[T] {
	m := &WorkQueue[T]{
		queue: list.New(),
	}
	m.cond = sync.NewCond(&m.mu)
	m.metricsContext = metrics.NewStdOutMetricsContext()
	return m
}

func (m *WorkQueue[T]) Enqueue(item *T) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.queue.PushBack(item)
	m.cond.Signal()
	m.metricsContext.Emit(fmt.Sprintf(queueLengthMetricFormat, m.queueName), m.queue.Len())
	return nil
}

func (m *WorkQueue[T]) Dequeue() *T {
	m.mu.Lock()
	defer m.mu.Unlock()
	for m.queue.Len() == 0 {
		m.cond.Wait()
	}
	elem := m.queue.Front()
	m.queue.Remove(elem)
	return elem.Value.(*T)
}

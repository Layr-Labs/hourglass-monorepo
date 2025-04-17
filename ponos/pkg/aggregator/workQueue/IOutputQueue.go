package workQueue

type IOutputQueue[T any] interface {
	Dequeue() *T
}

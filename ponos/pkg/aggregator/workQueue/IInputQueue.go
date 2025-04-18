package workQueue

type IInputQueue[T any] interface {
	Enqueue(input *T) error
}

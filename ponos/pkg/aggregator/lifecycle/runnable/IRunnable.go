package runnable

import "context"

type IRunnable interface {
	Start(ctx context.Context) error
}

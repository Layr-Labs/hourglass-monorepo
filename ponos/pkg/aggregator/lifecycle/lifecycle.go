package lifecycle

import "context"

type Lifecycle interface {
	Start(ctx context.Context) error
	Close() error
}

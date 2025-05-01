package chainPoller

import "context"

type IChainPoller interface {
	Start(ctx context.Context) error
}

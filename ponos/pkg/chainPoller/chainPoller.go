package chainPoller

import (
	"context"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
)

type IChainPoller interface {
	Start(ctx context.Context) error
}

type LogWithBlock struct {
	Log   *ethereum.EthereumEventLog
	Block *ethereum.EthereumBlock
}

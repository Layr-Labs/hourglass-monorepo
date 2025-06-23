package chainPoller

import (
	"context"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/clients/ethereum"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionLogParser/log"
)

type IChainPoller interface {
	Start(ctx context.Context) error
}

type LogWithBlock struct {
	Log    *log.DecodedLog
	RawLog *ethereum.EthereumEventLog
	Block  *ethereum.EthereumBlock
}

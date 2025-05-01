package logSequencer

import (
	"context"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/chainPoller"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionLogParser"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/transactionLogParser/log"
	"go.uber.org/zap"
)

type DistributeLogFunc func(*chainPoller.LogWithBlock, *log.DecodedLog) error

type LogSequencer struct {
	sequencerChannel chan *chainPoller.LogWithBlock
	logger           *zap.Logger

	transactionLogParser *transactionLogParser.TransactionLogParser

	distributeLogFunc DistributeLogFunc
}

func NewLogSequencer(
	transactionLogParser *transactionLogParser.TransactionLogParser,
	dlf DistributeLogFunc,
	logger *zap.Logger,
) *LogSequencer {
	return &LogSequencer{
		sequencerChannel:     make(chan *chainPoller.LogWithBlock, 10000),
		logger:               logger,
		transactionLogParser: transactionLogParser,
		distributeLogFunc:    dlf,
	}
}

func (ls *LogSequencer) GetChannel() chan *chainPoller.LogWithBlock {
	return ls.sequencerChannel
}

func (ls *LogSequencer) ProcessLogs(ctx context.Context) error {

	for {
		select {
		case logWithBlock := <-ls.sequencerChannel:
			// Process the log
			if err := ls.processLog(logWithBlock); err != nil {
				ls.logger.Error("Error processing log", zap.Error(err))
				return err
			}
		case <-ctx.Done():
			ls.logger.Info("Log sequencer context done, stopping processing logs")
			return nil
		}
	}
}

func (ls *LogSequencer) processLog(lwb *chainPoller.LogWithBlock) error {
	decodedLog, err := ls.transactionLogParser.DecodeLog(nil, lwb.Log)
	if err != nil {
		ls.logger.Error("Error decoding log", zap.Error(err))
		return err
	}

	if ls.distributeLogFunc != nil {
		if err := ls.distributeLogFunc(lwb, decodedLog); err != nil {
			ls.logger.Error("Error distributing log", zap.Error(err))
			return err
		}
	}
	return nil
}

package chainWriter

import (
	"context"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
)

type Event struct {
	EventId    string         `json:"eventId"`
	ChainId    config.ChainId `json:"chainId"`
	AvsAddress string         `json:"avsAddress"`
}

// IChainWriter is an interface whose implementation writes to results on the
// target chain.
type IChainWriter interface {
	ListenForInboxEvents(ctx context.Context, queue chan *Event, inboxAddress string) error
}

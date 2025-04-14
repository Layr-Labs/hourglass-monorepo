package chainListener

import (
	"context"
	"github.com/Layr-Labs/hourglass-monorepo/ponos/pkg/config"
)

type Event struct {
	EventId    string         `json:"eventId"`
	ChainId    config.ChainID `json:"chainId"`
	AvsAddress string         `json:"avsAddress"`
}

// IChainListener is an interface whose implementation listens for events on the
// target chain, decodes the events and publishes them to the provided channel
type IChainListener interface {
	ListenForInboxEvents(ctx context.Context, queue chan *Event, inboxAddress string) error
}

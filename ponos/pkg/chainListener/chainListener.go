package chainListener

import "context"

type Event struct {
}

// IChainListener is an interface whose implementation listens for events on the
// target chain, decodes the events and publishes them to the provided channel
type IChainListener interface {
	ListenForInboxEvents(ctx context.Context, queue chan *Event, inboxAddress string) error
}

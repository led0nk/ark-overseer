package interfaces

import (
	"context"

	"github.com/led0nk/ark-overseer/pkg/events"
)

type Notification interface {
	Connect(context.Context) error
	Send(context.Context, string) error
	HandleEvent(context.Context, events.EventMessage)
	Disconnect() error
}

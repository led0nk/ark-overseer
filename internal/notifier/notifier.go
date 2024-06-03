package notifier

import (
	"context"

	"github.com/google/uuid"
	"github.com/led0nk/ark-overseer/internal"
	"github.com/led0nk/ark-overseer/internal/model"
	"github.com/led0nk/ark-overseer/pkg/events"
)

type Notifier struct {
	sStore internal.ServerStore
	em     *events.EventManager
}

func NewNotifier(s internal.ServerStore, eventManager *events.EventManager) *Notifier {
	return &Notifier{
		sStore: s,
		em:     eventManager,
	}
}

func (n *Notifier) Create(ctx context.Context, srv *model.Server) (*model.Server, error) {
	newServer, err := n.sStore.Create(ctx, srv)
	n.em.Publish(events.EventMessage{Type: "server.added", Payload: newServer})
	return newServer, err
}

func (n *Notifier) Delete(ctx context.Context, id uuid.UUID) error {
	n.em.Publish(events.EventMessage{Type: "server.deleted", Payload: id})
	err := n.sStore.Delete(ctx, id)
	return err
}

func (n *Notifier) GetByID(ctx context.Context, id uuid.UUID) (*model.Server, error) {
	//n.notify("get by id")
	return n.sStore.GetByID(ctx, id)
}

func (n *Notifier) GetByName(ctx context.Context, name string) (*model.Server, error) {
	//n.notify("get by name")
	return n.sStore.GetByName(ctx, name)
}

func (n *Notifier) List(ctx context.Context) ([]*model.Server, error) {
	//n.notify("list")
	return n.sStore.List(ctx)
}

func (n *Notifier) Update(ctx context.Context, srv *model.Server) error {
	//(n.notify("update")
	return n.sStore.Update(ctx, srv)
}

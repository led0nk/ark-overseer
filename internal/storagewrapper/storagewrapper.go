package storagewrapper

import (
	"context"

	"github.com/google/uuid"
	"github.com/led0nk/ark-overseer/internal/model"
	"github.com/led0nk/ark-overseer/internal/storage"
	"github.com/led0nk/ark-overseer/pkg/events"
)

type StorageWrapper struct {
	store storage.Database
	em    *events.EventManager
}

func NewStorageWrapper(
	s storage.Database,
	eventManager *events.EventManager,
) *StorageWrapper {
	return &StorageWrapper{
		store: s,
		em:    eventManager,
	}
}

func (n *StorageWrapper) Create(ctx context.Context, srv *model.Server) (*model.Server, error) {
	newServer, err := n.store.Create(ctx, srv)
	n.em.Publish(events.EventMessage{Type: "server.added", Payload: newServer})
	return newServer, err
}

func (n *StorageWrapper) Delete(ctx context.Context, id uuid.UUID) error {
	n.em.Publish(events.EventMessage{Type: "server.deleted", Payload: id})
	err := n.store.Delete(ctx, id)
	return err
}

func (n *StorageWrapper) GetByID(ctx context.Context, id uuid.UUID) (*model.Server, error) {
	return n.store.GetByID(ctx, id)
}

func (n *StorageWrapper) GetByName(ctx context.Context, name string) (*model.Server, error) {
	return n.store.GetByName(ctx, name)
}

func (n *StorageWrapper) List(ctx context.Context) ([]*model.Server, error) {
	return n.store.List(ctx)
}

func (n *StorageWrapper) Update(ctx context.Context, srv *model.Server) error {
	return n.store.Update(ctx, srv)
}

func (n *StorageWrapper) Save() error {
	return n.store.Save()
}

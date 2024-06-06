package internal

import (
	"context"

	"github.com/google/uuid"
	"github.com/led0nk/ark-overseer/internal/model"
	"github.com/led0nk/ark-overseer/pkg/events"
)

type Database interface {
	Create(context.Context, *model.Server) (*model.Server, error)
	List(context.Context) ([]*model.Server, error)
	GetByName(context.Context, string) (*model.Server, error)
	GetByID(context.Context, uuid.UUID) (*model.Server, error)
	Delete(context.Context, uuid.UUID) error
	Update(context.Context, *model.Server) error
	Save() error
}

type Blacklist interface {
	Create(context.Context, *model.BlacklistPlayers) (*model.BlacklistPlayers, error)
	List(context.Context) []*model.BlacklistPlayers
	Delete(context.Context, uuid.UUID) error
}

type Notification interface {
	Connect(context.Context) error
	Send(context.Context, string) error
	HandleEvent(context.Context, events.EventMessage)
	Disconnect() error
}

package internal

import (
	"context"

	"github.com/google/uuid"
	"github.com/led0nk/ark-clusterinfo/internal/events"
	"github.com/led0nk/ark-clusterinfo/internal/model"
)

type ServerStore interface {
	Create(context.Context, *model.Server) (*model.Server, error)
	List(context.Context) ([]*model.Server, error)
	GetByName(context.Context, string) (*model.Server, error)
	GetByID(context.Context, uuid.UUID) (*model.Server, error)
	Delete(context.Context, uuid.UUID) error
	Update(context.Context, *model.Server) error
}

type Observer interface {
	ReadEndpoint(*model.Server) error
	DataScraper(context.Context, *model.Server)
	SpawnScraper(context.Context)
	AddScraper(context.Context, *model.Server) error
	KillScraper(uuid.UUID) error
	HandleEvent(context.Context, events.EventMessage)
}

type Blacklist interface {
	Create(context.Context, *model.BlacklistPlayers) (*model.BlacklistPlayers, error)
	List(context.Context) []*model.BlacklistPlayers
	Delete(context.Context, uuid.UUID) error
}

type Overseer interface {
	ReadEndpoint(*model.Server) error
	Scanner(context.Context, *model.Server)
	SpawnScanner(context.Context)
	AddScanner(context.Context, *model.Server) error
	KillScanner(uuid.UUID) error
	HandleEvent(context.Context, events.EventMessage)
}

type Notification interface {
	Connect(context.Context) error
	Send(context.Context, string) error
	HandleEvent(context.Context, events.EventMessage)
}

type Configuration interface {
	Read() error
	Write() error
	Update(string, string, interface{}) error
}

package internal

import (
	"context"

	"github.com/google/uuid"
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
	ManageScraper(context.Context)
	AddScraper(context.Context, *model.Server) error
	KillScraper(uuid.UUID) error
}

type Blacklist interface {
	Create(context.Context, *model.Players) (*model.Players, error)
	List(context.Context) []*model.Players
	Delete(context.Context, uuid.UUID) error
}

type Overseer interface {
	ReadEndpoint(*model.Server) error
	Scanner(context.Context, *model.Server)
	ManageScanner(context.Context)
	AddScanner(context.Context, *model.Server) error
	KillScanner(uuid.UUID) error
}

type Notification interface {
	Send() error
}

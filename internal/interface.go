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

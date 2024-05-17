package internal

import (
	"context"

	"github.com/google/uuid"
	"github.com/led0nk/ark-clusterinfo/internal/model"
	"github.com/led0nk/ark-clusterinfo/internal/parser"
)

type ServerStore interface {
	Create(context.Context, *model.Server) (string, error)
	List(context.Context) ([]*model.Server, error)
	GetByName(context.Context, string) (*model.Server, error)
	GetByID(context.Context, uuid.UUID) (*model.Server, error)
	Delete(context.Context, uuid.UUID) error
}

type Observer interface {
	ReadEndpoint(*parser.Target) error
	DataScraper(context.Context, *parser.Target)
	ManageScraper(context.Context)
	AddScraper(context.Context, *parser.Target) error
	KillScraper(uuid.UUID) error
}

type Parser interface {
	Create(context.Context, *parser.Target) (*parser.Target, error)
	Delete(context.Context, uuid.UUID) error
	List(context.Context) ([]*parser.Target, error)
}

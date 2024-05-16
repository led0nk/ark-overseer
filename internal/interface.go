package internal

import (
	"context"

	"github.com/google/uuid"
	"github.com/led0nk/ark-clusterinfo/internal/model"
	"github.com/led0nk/ark-clusterinfo/internal/parser"
)

type ServerStore interface {
	CreateOrUpdateServer(*model.Server) (string, error)
	ListServer() ([]*model.Server, error)
	GetServerByName(string) (*model.Server, error)
	GetServerByID(uuid.UUID) (*model.Server, error)
	DeleteServer(uuid.UUID) error
}

type Observer interface {
	ReadEndpoint(*parser.Target) error
	DataScraper(context.Context, *parser.Target)
	InitScraper(context.Context, []*parser.Target)
	AddScraper(context.Context, *parser.Target) error
	KillScraper(uuid.UUID) error
}

type Parser interface {
	CreateTarget(*parser.Target) (*parser.Target, error)
	DeleteTarget(uuid.UUID) error
	ListTargets() ([]*parser.Target, error)
}

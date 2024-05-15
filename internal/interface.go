package internal

import (
	"github.com/google/uuid"
	"github.com/led0nk/ark-clusterinfo/internal/model"
)

type ClusterStore interface {
	CreateServer(*model.Server) (string, error)
	GetServerByName(string) (*model.Server, error)
	GetServerByID(uuid.UUID) (*model.Server, error)
	DeleteServer(uuid.UUID) error
	GetServerInfo() ([]*model.Server, error)
}

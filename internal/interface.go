package internal

import "github.com/led0nk/ark-clusterinfo/internal/model"

type ClusterStore interface {
	CreateServer(*model.Server) (string, error)
	GetServerByName(string) (*model.Server, error)
	DeleteServer(string) error
	GetServerInfo() ([]*model.Server, error)
}

package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/led0nk/ark-overseer/internal/model"
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

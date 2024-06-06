package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/led0nk/ark-overseer/internal/model"
)

type Blacklist interface {
	Create(context.Context, *model.BlacklistPlayers) (*model.BlacklistPlayers, error)
	List(context.Context) []*model.BlacklistPlayers
	Delete(context.Context, uuid.UUID) error
}

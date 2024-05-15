package model

import (
	"github.com/FlowingSPDG/go-steam"
	"github.com/google/uuid"
)

type Server struct {
	ID          uuid.UUID                  `json:"id" form:"-"`
	Name        string                     `json:"name" form:"-"`
	Addr        string                     `json:"addr" form:"-"`
	ServerInfo  *steam.InfoResponse        `json:"serverinfo" form:"-"`
	PlayersInfo *steam.PlayersInfoResponse `json:"playersinfo" form:"-"`
}

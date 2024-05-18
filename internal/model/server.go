package model

import (
	"time"

	"github.com/FlowingSPDG/go-steam"
	"github.com/google/uuid"
)

type Server struct {
	ID          uuid.UUID    `json:"id" form:"-"`
	Name        string       `json:"name" form:"-"`
	Addr        string       `json:"addr" form:"-"`
	Status      bool         `json:"status" form:"-"`
	ServerInfo  *ServerInfo  `json:"serverinfo" form:"-"`
	PlayersInfo *PlayersInfo `json:"playersinfo" form:"-"`
}

type ServerInfo struct {
	Protocol     int               `json:"protocol" form:"-"`
	Name         string            `json:"name" form:"-"`
	Map          string            `json:"map" form:"-"`
	Folder       string            `json:"folder" form:"-"`
	Game         string            `json:"game" form:"-"`
	ID           int               `json:"id" form:"-"`
	Players      int               `json:"players" form:"-"`
	MaxPlayers   int               `json:"maxplayers" form:"-"`
	Bots         int               `json:"bots" form:"-"`
	ServerType   steam.ServerType  `json:"servertype" form:"-"`
	Environment  steam.Environment `json:"environment" form:"-"`
	Visibility   steam.Visibility  `json:"visibility" form:"-"`
	VAC          steam.VAC         `json:"vac" form:"-"`
	Version      string            `json:"version" form:"-"`
	Port         int               `json:"port" form:"-"`
	SteamID      int64             `json:"steamid" form:"-"`
	SourceTVPort int               `json:"sourcetvport" form:"-"`
	SourceTVName string            `json:"sourcetvname" form:"-"`
	Keywords     string            `json:"keywords" form:"-"`
	GameID       int64             `json:"gameid" form:"-"`
}

type PlayersInfo struct {
	Players []*Players
}

type Players struct {
	Name     string        `json:"name" form:"-"`
	Score    int           `json:"score" form:"-"`
	Duration time.Duration `json:"duration" form:"-"`
}

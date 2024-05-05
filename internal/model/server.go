package model

import "github.com/FlowingSPDG/go-steam"

type Server struct {
	Name        string
	Addr        string
	ServerInfo  *steam.InfoResponse
	PlayersInfo *steam.PlayersInfoResponse
}

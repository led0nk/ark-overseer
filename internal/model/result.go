package model

import "github.com/FlowingSPDG/go-steam"

// only for testing purpose -> should be switched to unmarshalling data into *model.Server
type Result struct {
	ServerInfo *steam.InfoResponse
	PlayerInfo *steam.PlayersInfoResponse
}

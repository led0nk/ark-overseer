package model

import (
	"math"
	"time"

	"github.com/FlowingSPDG/go-steam"
)

func ToServerInfo(infoResponse *steam.InfoResponse) *ServerInfo {
	serverInfo := &ServerInfo{
		Protocol:     infoResponse.Protocol,
		Name:         infoResponse.Name,
		Map:          infoResponse.Map,
		Folder:       infoResponse.Folder,
		Game:         infoResponse.Game,
		ID:           infoResponse.ID,
		Players:      infoResponse.Players,
		MaxPlayers:   infoResponse.MaxPlayers,
		Bots:         infoResponse.Bots,
		ServerType:   infoResponse.ServerType,
		Environment:  infoResponse.Environment,
		Visibility:   infoResponse.Visibility,
		VAC:          infoResponse.VAC,
		Version:      infoResponse.Version,
		Port:         infoResponse.Port,
		SteamID:      infoResponse.SteamID,
		SourceTVPort: infoResponse.SourceTVPort,
		SourceTVName: infoResponse.SourceTVName,
		Keywords:     infoResponse.Keywords,
		GameID:       infoResponse.GameID,
	}
	return serverInfo
}

func ToPlayerInfo(playersInfoResponse *steam.PlayersInfoResponse) *PlayersInfo {
	playersInfo := &PlayersInfo{
		make([]Players, 0, len(playersInfoResponse.Players)),
	}
	for _, player := range playersInfoResponse.Players {
		newPlayer := Players{
			Name:     player.Name,
			Score:    player.Score,
			Duration: time.Duration(math.Round(player.Duration) * 1e9),
		}
		playersInfo.Players = append(playersInfo.Players, newPlayer)
	}
	return playersInfo
}

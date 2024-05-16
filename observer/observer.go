package observer

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/FlowingSPDG/go-steam"
	"github.com/google/uuid"
	"github.com/led0nk/ark-clusterinfo/internal"
	"github.com/led0nk/ark-clusterinfo/internal/model"
	"github.com/led0nk/ark-clusterinfo/internal/parser"
)

type Observer struct {
	endpoints   map[uuid.UUID]*parser.Target
	serverStore internal.ServerStore
	parser      internal.Parser
	logger      *slog.Logger
	mu          sync.Mutex
}

func NewObserver(sStore internal.ServerStore, prs internal.Parser) (*Observer, error) {
	observer := &Observer{
		endpoints:   make(map[uuid.UUID]*parser.Target),
		serverStore: sStore,
		parser:      prs,
		logger:      slog.Default(),
	}
	return observer, nil
}

func (o *Observer) ReadEndpoint(target *parser.Target) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if target.ID == uuid.Nil {
		target.ID = uuid.New()
	}

	o.endpoints[target.ID] = target
	return nil
}

func (o *Observer) DataScraper(ctx context.Context, id uuid.UUID, c chan<- *model.Server) {
	for _, endpoint := range o.endpoints {
		if endpoint.ID == id {
			for {
				helpSrv, err := steam.Connect(endpoint.Addr)
				if err != nil {
					o.logger.ErrorContext(ctx, "error connecting to endpoint", "error", err)
					continue
				}

				infoResponse, err := helpSrv.Info()
				if err != nil {
					o.logger.ErrorContext(ctx, "error fetching ServerInfo", "error", err)
					continue
				}

				playerResponse, err := helpSrv.PlayersInfo()
				if err != nil {
					o.logger.ErrorContext(ctx, "error fetching PlayersInfo", "error", err)
					continue
				}

				ping, err := helpSrv.Ping()
				if err != nil {
					o.logger.ErrorContext(ctx, "failed to ping server", "error", err)
					continue
				}

				server, err := Unmarshal(infoResponse)
				if err != nil {
					o.logger.ErrorContext(ctx, "failed to unmarshal server info", "error", err)
				}

				player, err := Unmarshal(playerResponse)
				if err != nil {
					o.logger.ErrorContext(ctx, "failed to unmarshal server info", "error", err)
				}
				if ping < time.Duration(5*time.Second) {
					server.Status = true
				}
				server.ID = id
				server.PlayersInfo = player.PlayersInfo
				c <- server
			}
		}
	}
}

func (o *Observer) InitScraper(ctx context.Context) {
	resultCh := make(chan *model.Server)

	targets, err := o.parser.ListTargets()
	if err != nil {
		o.logger.ErrorContext(ctx, "failed to list targets", "error", err)
		return
	}

	for _, target := range targets {
		err := o.ReadEndpoint(target)
		if err != nil {
			o.logger.ErrorContext(ctx, "failed to read endpoints", "error", err)
			return
		}
		go o.DataScraper(ctx, target.ID, resultCh)
	}

	for result := range resultCh {
		_, err = o.serverStore.CreateOrUpdateServer(result)
		if err != nil {
			o.logger.ErrorContext(ctx, "failed to update server info", "error", err)
		}
	}
}

func Unmarshal(info interface{}) (*model.Server, error) {
	server := &model.Server{}

	switch v := info.(type) {
	case *steam.InfoResponse:
		server.ServerInfo = &model.ServerInfo{
			Protocol:     v.Protocol,
			Name:         strings.Trim(v.Name, "\u0000"),
			Map:          strings.Trim(v.Map, "\u0000"),
			Folder:       strings.Trim(v.Folder, "\u0000"),
			Game:         strings.Trim(v.Game, "\u0000"),
			ID:           v.ID,
			Players:      v.Players,
			MaxPlayers:   v.MaxPlayers,
			Bots:         v.Bots,
			ServerType:   v.ServerType,
			Environment:  v.Environment,
			Visibility:   v.Visibility,
			VAC:          v.VAC,
			Version:      strings.Trim(v.Version, "\u0000"),
			Port:         v.Port,
			SteamID:      v.SteamID,
			SourceTVPort: v.SourceTVPort,
			SourceTVName: strings.Trim(v.SourceTVName, "\u0000"),
			Keywords:     strings.Trim(v.Keywords, "\u0000"),
			GameID:       v.GameID,
		}
	case *steam.PlayersInfoResponse:
		var players []model.Players
		for _, p := range v.Players {
			player := model.Players{
				Name:     strings.Trim(p.Name, "\u0000"),
				Score:    p.Score,
				Duration: time.Duration(math.Round(p.Duration) * 1e9),
			}
			players = append(players, player)
		}
		server.PlayersInfo = &model.PlayersInfo{Players: players}
	default:
		return nil, errors.New("unsupported type")
	}
	return server, nil
}

package observer

import (
	"context"
	"errors"
	"log/slog"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/FlowingSPDG/go-steam"
	"github.com/google/uuid"
	"github.com/led0nk/ark-clusterinfo/internal"
	"github.com/led0nk/ark-clusterinfo/internal/events"
	"github.com/led0nk/ark-clusterinfo/internal/model"
)

type Observer struct {
	endpoints   map[uuid.UUID]*model.Server
	cancelFuncs map[uuid.UUID]context.CancelFunc
	serverStore internal.ServerStore
	logger      *slog.Logger
	mu          sync.Mutex
	resultCh    chan *model.Server
}

func NewObserver(sStore internal.ServerStore) (*Observer, error) {
	observer := &Observer{
		endpoints:   make(map[uuid.UUID]*model.Server),
		cancelFuncs: make(map[uuid.UUID]context.CancelFunc),
		serverStore: sStore,
		logger:      slog.Default().WithGroup("observer"),
		resultCh:    make(chan *model.Server),
	}
	go observer.processResults()
	return observer, nil
}

func (o *Observer) ReadEndpoint(target *model.Server) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if target.ID == uuid.Nil {
		target.ID = uuid.New()
	}

	o.endpoints[target.ID] = target
	return nil
}

func (o *Observer) DataScraper(ctx context.Context, s *model.Server) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			helpSrv, err := steam.Connect(s.Addr)
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

			var status bool
			if ping < time.Duration(5*time.Second) {
				status = true
			}

			server := &model.Server{
				Name:        s.Name,
				Addr:        s.Addr,
				ID:          s.ID,
				Status:      status,
				ServerInfo:  model.ToServerInfo(infoResponse),
				PlayersInfo: model.ToPlayerInfo(playerResponse),
			}
			ReplaceNullCharsInStruct(server)
			server = correctPlayerNum(server)
			o.resultCh <- server
		}
	}
}

func (o *Observer) SpawnScraper(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	default:
		serverList, err := o.serverStore.List(ctx)
		if err != nil {
			o.logger.ErrorContext(ctx, "failed to list targets", "error", err)
		}

		for _, server := range serverList {
			err := o.AddScraper(ctx, server)
			if err != nil {
				o.logger.ErrorContext(ctx, "failed to spawn scraper", "error", err)
				return
			}
		}
	}
}

func (o *Observer) processResults() {
	for result := range o.resultCh {
		if result == nil {
			continue
		}
		err := o.serverStore.Update(context.Background(), result)
		if err != nil {
			o.logger.Error("failed to update server info", "error", err)
		}
	}
}

func (o *Observer) AddScraper(ctx context.Context, target *model.Server) error {
	err := o.ReadEndpoint(target)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(ctx)
	o.cancelFuncs[target.ID] = cancel
	go o.DataScraper(ctx, target)
	return nil
}

func (o *Observer) KillScraper(targetID uuid.UUID) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if cancel, exists := o.cancelFuncs[targetID]; exists {
		cancel()
		delete(o.cancelFuncs, targetID)
		delete(o.endpoints, targetID)
		return nil
	}
	return errors.New("Scraper with ID not found")
}

func (o *Observer) HandleEvent(ctx context.Context, event events.EventMessage) {
	switch event.Type {
	case "addedServer":
		server, ok := event.Payload.(*model.Server)
		if !ok {
			o.logger.ErrorContext(ctx, "invalid payload type for addedServer event", "error", errors.New("payload not of type *model.Server"))
			return
		}
		err := o.AddScraper(ctx, server)
		if err != nil {
			o.logger.ErrorContext(ctx, "failed to add scraper", "error", err)
			return
		}
	case "deletedServer":
		id, ok := event.Payload.(uuid.UUID)
		if !ok {
			o.logger.ErrorContext(ctx, "invalid payload type for deletedServer event", "error", errors.New("Payload not of type uuid.UUID"))
			return
		}
		err := o.KillScraper(id)
		if err != nil {
			o.logger.ErrorContext(ctx, "failed to add scraper", "error", err)
			return
		}
	default:
		return
	}
}

//NOTE: help-funcs for data-transfer

func ReplaceNullCharsInStruct(s any) {
	v := reflect.ValueOf(s)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return
	}
	replaceNullChars(v.Elem())
}

func replaceNullChars(v reflect.Value) {
	switch v.Kind() {
	case reflect.String:
		str := v.Interface().(string)
		str = strings.Trim(str, "\u0000")
		v.SetString(str)
	case reflect.Ptr:
		if v.IsNil() {
			return
		}
		replaceNullChars(v.Elem())
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			replaceNullChars(v.Field(i))
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			replaceNullChars(v.Index(i))
		}
	}
}

func correctPlayerNum(srv *model.Server) *model.Server {
	var playerList []*model.Players
	for _, player := range srv.PlayersInfo.Players {
		if player.Name != "" {
			playerList = append(playerList, player)
		}
	}

	srv.PlayersInfo.Players = playerList
	srv.ServerInfo.Players = len(playerList)

	return srv
}

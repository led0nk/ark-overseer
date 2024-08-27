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
	"github.com/led0nk/ark-overseer/internal/interfaces"
	"github.com/led0nk/ark-overseer/internal/model"
	"github.com/led0nk/ark-overseer/pkg/events"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

var meter = otel.GetMeterProvider().Meter("github.com/led0nk/ark-overseer/internal/observer")

type Overseer interface {
	HandleEvent(context.Context, events.EventMessage)
}

type Observer struct {
	endpoints   map[uuid.UUID]*model.Server
	cancelFuncs map[uuid.UUID]context.CancelFunc
	serverStore interfaces.Database
	blacklist   interfaces.Blacklist
	em          *events.EventManager
	logger      *slog.Logger
	mu          sync.Mutex
	resultCh    map[uuid.UUID]chan *model.Server
}

type NotificationStatus struct {
	isActive       bool
	joinedNotified bool
	leftNotified   bool
}

func NewObserver(
	ctx context.Context,
	sStore interfaces.Database,
	blacklist interfaces.Blacklist,
	eventManager *events.EventManager,
) *Observer {
	observer := &Observer{
		endpoints:   make(map[uuid.UUID]*model.Server),
		cancelFuncs: make(map[uuid.UUID]context.CancelFunc),
		serverStore: sStore,
		blacklist:   blacklist,
		em:          eventManager,
		logger:      slog.Default().WithGroup("observer"),
		resultCh:    make(map[uuid.UUID]chan *model.Server),
	}
	go observer.processResults(ctx)
	return observer
}

func (o *Observer) readEndpoint(target *model.Server) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if target.ID == uuid.Nil {
		target.ID = uuid.New()
	}

	o.endpoints[target.ID] = target
	return nil
}

func (o *Observer) dataScraper(ctx context.Context, target *model.Server) chan *model.Server {
	scrapesCtr, err := meter.Int64UpDownCounter(
		"scrapeCtr",
		metric.WithDescription("number of data scrapes from steam server"),
		metric.WithUnit("{InfoResponse}"),
	)
	if err != nil {
		return nil
	}

	failedScrapesCtr, err := meter.Int64UpDownCounter(
		"failedScrapeCtr",
		metric.WithDescription("number of data scrapes from steam server"),
		metric.WithUnit("{InfoResponse}"),
	)
	if err != nil {
		return nil
	}

	out := make(chan *model.Server)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				helpSrv, err := steam.Connect(target.Addr)
				if err != nil {
					o.logger.ErrorContext(ctx, "error connecting to endpoint", "error", err)
					continue
				}

				infoResponse, err := helpSrv.Info()
				if err != nil {
					o.logger.ErrorContext(ctx, "error fetching ServerInfo", "error", err)
					failedScrapesCtr.Add(ctx, 1)
					continue
				}
				scrapesCtr.Add(ctx, 1)

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
					Name:        target.Name,
					Addr:        target.Addr,
					ID:          target.ID,
					Status:      status,
					ServerInfo:  model.ToServerInfo(infoResponse),
					PlayersInfo: model.ToPlayerInfo(playerResponse),
				}
				replaceNullCharsInStruct(server)
				server = correctPlayerNum(server)
				select {
				case out <- server:
				default:
				}
			}
		}
	}()
	return out
}

func (o *Observer) scanner(ctx context.Context, in chan *model.Server) chan *model.Server {
	scanCtr, err := meter.Int64UpDownCounter(
		"scanCtr",
		metric.WithDescription("number of scans happened"),
	)
	if err != nil {
		return nil
	}

	out := make(chan *model.Server)
	go func() {
		defer close(out)
		previousPlayers := make(map[string]*NotificationStatus)
		for {
			select {
			case <-ctx.Done():
				return
			case server, ok := <-in:
				if !ok {
					return
				}
				blacklist := o.blacklist.List(ctx)
				if server.PlayersInfo == nil {
					continue
				}
				previousPlayers = o.scan(blacklist, server, previousPlayers)
				select {
				case out <- server:
					scanCtr.Add(ctx, 1)
				default:
				}
			}
		}
	}()
	return out
}

func (o *Observer) scan(blacklist []*model.BlacklistPlayers, server *model.Server, previousPlayers map[string]*NotificationStatus) map[string]*NotificationStatus {

	blacklistMap := make(map[string]bool)
	for _, blacklistedPlayer := range blacklist {
		blacklistMap[blacklistedPlayer.Name] = true
	}

	for _, status := range previousPlayers {
		status.isActive = false
	}

	for _, player := range server.PlayersInfo.Players {
		status, exists := previousPlayers[player.Name]
		if !exists {
			status = &NotificationStatus{}
			previousPlayers[player.Name] = status
		}
		status.isActive = true

		if blacklistMap[player.Name] {
			if !status.joinedNotified {
				o.em.Publish(events.EventMessage{Type: "player.joined", Payload: player.Name + " joined the server " + server.Name})
				status.joinedNotified = true
				status.leftNotified = false
			}
		}
	}

	for playerName, status := range previousPlayers {
		if blacklistMap[playerName] && !status.isActive && !status.leftNotified {
			o.em.Publish(events.EventMessage{Type: "player.left", Payload: playerName + " left the server " + server.Name})
			status.leftNotified = true
			status.joinedNotified = false
		}
	}

	return previousPlayers
}

func (o *Observer) spawnScraper(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	default:
		serverList, err := o.serverStore.List(ctx)
		if err != nil {
			o.logger.ErrorContext(ctx, "failed to list targets", "error", err)
		}

		for _, server := range serverList {
			err := o.addScraper(ctx, server)
			if err != nil {
				o.logger.ErrorContext(ctx, "failed to spawn scraper", "error", err)
				return
			}
		}
	}
}

func (o *Observer) processResults(ctx context.Context) {
	processCtr, err := meter.Int64Counter(
		"processCtr",
		metric.WithDescription("number of processed server data"),
	)
	if err != nil {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			for _, ch := range o.resultCh {
				select {
				case <-ctx.Done():
					return
				case result := <-ch:
					if result == nil {
						continue
					}
					err := o.serverStore.Update(ctx, result)
					if err != nil {
						o.logger.Error("failed to update server info", "error", err)
					}
					processCtr.Add(ctx, 1)
				}
			}

		}
	}
}

func (o *Observer) addScraper(ctx context.Context, target *model.Server) error {
	err := o.readEndpoint(target)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(ctx)
	o.cancelFuncs[target.ID] = cancel
	pipeCh := o.dataScraper(ctx, target)
	processCh := o.scanner(ctx, pipeCh)
	o.resultCh[target.ID] = processCh
	return nil
}

func (o *Observer) killScraper(targetID uuid.UUID) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if cancel, exists := o.cancelFuncs[targetID]; exists {
		cancel()
		delete(o.cancelFuncs, targetID)
		delete(o.endpoints, targetID)
		close(o.resultCh[targetID])
		delete(o.resultCh, targetID)
		return nil
	}
	return errors.New("Scraper with ID not found")
}

func (o *Observer) HandleEvent(ctx context.Context, event events.EventMessage) {
	switch event.Type {
	case "init":
		o.spawnScraper(ctx)
	case "server.added":
		server, ok := event.Payload.(*model.Server)
		if !ok {
			o.logger.ErrorContext(ctx, "invalid payload type", "error", event.Type)
			return
		}
		err := o.addScraper(ctx, server)
		if err != nil {
			o.logger.ErrorContext(ctx, "failed to add scraper", "error", err)
			return
		}
	case "server.deleted":
		id, ok := event.Payload.(uuid.UUID)
		if !ok {
			o.logger.ErrorContext(ctx, "invalid payload type", "error", event.Type)
			return
		}
		err := o.killScraper(id)
		if err != nil {
			o.logger.ErrorContext(ctx, "failed to add scraper", "error", err)
			return
		}
	default:
		return
	}
}

//NOTE: help-funcs for data-transfer

func replaceNullCharsInStruct(s any) {
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

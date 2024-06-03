package overseer

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/google/uuid"
	"github.com/led0nk/ark-clusterinfo/internal"
	"github.com/led0nk/ark-clusterinfo/internal/model"
	"github.com/led0nk/ark-clusterinfo/pkg/events"
)

type Overseer struct {
	endpoints   map[uuid.UUID]*model.Server
	cancelFuncs map[uuid.UUID]context.CancelFunc
	blacklist   internal.Blacklist
	serverStore internal.ServerStore
	em          *events.EventManager
	logger      *slog.Logger
	mu          sync.Mutex
	resultCh    chan map[string]*NotificationStatus
}

type NotificationStatus struct {
	isActive       bool
	joinedNotified bool
	leftNotified   bool
}

func NewOverseer(
	ctx context.Context,
	sStore internal.ServerStore,
	blacklist internal.Blacklist,
	eventManager *events.EventManager,
) (*Overseer, error) {
	overseer := &Overseer{
		endpoints:   make(map[uuid.UUID]*model.Server),
		cancelFuncs: make(map[uuid.UUID]context.CancelFunc),
		blacklist:   blacklist,
		serverStore: sStore,
		em:          eventManager,
		logger:      slog.Default().WithGroup("overseer"),
		resultCh:    make(chan map[string]*NotificationStatus),
	}
	return overseer, nil
}

func (o *Overseer) ReadEndpoint(target *model.Server) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if target.ID == uuid.Nil {
		target.ID = uuid.New()
	}
	o.endpoints[target.ID] = target
	return nil
}

func (o *Overseer) SpawnScanner(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	default:
		serverList, err := o.serverStore.List(ctx)
		if err != nil {
			o.logger.ErrorContext(ctx, "failed to list targets", "error", err)
		}

		for _, server := range serverList {
			err := o.AddScanner(ctx, server)
			if err != nil {
				o.logger.ErrorContext(ctx, "failed to add scanner", "error", err)
				return
			}
		}
	}
}

func (o *Overseer) Scanner(ctx context.Context, target *model.Server) {
	previousPlayers := make(map[string]*NotificationStatus)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			blacklist := o.blacklist.List(ctx)

			server, err := o.serverStore.GetByID(ctx, target.ID)
			if err != nil {
				o.logger.ErrorContext(ctx, "failed to get server", "error", err)
				continue
			}

			if server.PlayersInfo == nil {
				continue
			}

			previousPlayers = o.Scan(ctx, blacklist, server, previousPlayers)
		}
	}
}

func (o *Overseer) AddScanner(ctx context.Context, target *model.Server) error {
	err := o.ReadEndpoint(target)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(ctx)
	o.cancelFuncs[target.ID] = cancel
	go o.Scanner(ctx, target)
	return nil
}

func (o *Overseer) KillScanner(targetID uuid.UUID) error {
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

func (o *Overseer) Scan(ctx context.Context, blacklist []*model.BlacklistPlayers, server *model.Server, previousPlayers map[string]*NotificationStatus) map[string]*NotificationStatus {

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

func (o *Overseer) HandleEvent(ctx context.Context, event events.EventMessage) {
	switch event.Type {
	case "init":
		o.SpawnScanner(ctx)
	case "server.added":
		server, ok := event.Payload.(*model.Server)
		if !ok {
			o.logger.ErrorContext(ctx, "invalid payload type for addedServer event", "error", errors.New("Payload not of type *model.Server"))
			return
		}
		err := o.AddScanner(ctx, server)
		if err != nil {
			o.logger.ErrorContext(ctx, "failed to add scraper", "error", err)
			return
		}
	case "server.deleted":
		id, ok := event.Payload.(uuid.UUID)
		if !ok {
			o.logger.ErrorContext(ctx, "invalid payload type for deletedServer event", "error", errors.New("Payload not of type uuid.UUID"))
			return
		}
		err := o.KillScanner(id)
		if err != nil {
			o.logger.ErrorContext(ctx, "failed to add scraper", "error", err)
			return
		}
	default:
		return
	}
}

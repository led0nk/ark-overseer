package overseer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/google/uuid"
	"github.com/led0nk/ark-clusterinfo/internal"
	"github.com/led0nk/ark-clusterinfo/internal/model"
)

type Overseer struct {
	endpoints   map[uuid.UUID]*model.Server
	cancelFuncs map[uuid.UUID]context.CancelFunc
	blacklist   internal.Blacklist
	serverStore internal.ServerStore
	messaging   internal.Notification
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
	messaging internal.Notification,
) (*Overseer, error) {
	overseer := &Overseer{
		endpoints:   make(map[uuid.UUID]*model.Server),
		cancelFuncs: make(map[uuid.UUID]context.CancelFunc),
		blacklist:   blacklist,
		serverStore: sStore,
		messaging:   messaging,
		logger:      slog.Default().WithGroup("overseer"),
		resultCh:    make(chan map[string]*NotificationStatus),
	}
	err := messaging.Connect(ctx)
	if err != nil {
		overseer.logger.ErrorContext(ctx, "failed to connect messaging service", "error", err)
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

func (o *Overseer) ManageScanner(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	default:
		serverList, err := o.serverStore.List(ctx)
		if err != nil {
			o.logger.ErrorContext(ctx, "failed to list targets", "error", err)
		}

		newServers := make(map[uuid.UUID]*model.Server)

		for _, server := range serverList {
			newServers[server.ID] = server
		}

		for id := range o.endpoints {
			if _, exists := newServers[id]; !exists {
				err := o.KillScanner(id)
				if err != nil {
					o.logger.ErrorContext(ctx, "failed to kill scraper", "error", err)
					continue
				}
			}
		}

		for id, server := range newServers {
			if _, exists := o.endpoints[id]; !exists {
				err := o.AddScanner(ctx, server)
				if err != nil {
					o.logger.ErrorContext(ctx, "failed to read endpoints", "error", err)
					return
				}
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

func (o *Overseer) Scan(ctx context.Context, blacklist []*model.Players, server *model.Server, previousPlayers map[string]*NotificationStatus) map[string]*NotificationStatus {

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
				//TODO: call function for notification services, context
				err := o.messaging.Send(ctx, "1204937103750725634", player.Name+" joined the server "+server.Name)
				if err != nil {
					o.logger.Error("failed to send message", "error", err)
					continue
				}
				fmt.Println(player.Name + " joined the server " + server.Name)
				status.joinedNotified = true
				status.leftNotified = false
			}
		}
	}

	for playerName, status := range previousPlayers {
		if blacklistMap[playerName] && !status.isActive && !status.leftNotified {
			//TODO: call function for notification services, context
			err := o.messaging.Send(ctx, "1204937103750725634", playerName+" left the server "+server.Name)
			if err != nil {
				o.logger.Error("failed to send message", "error", err)
				continue
			}
			fmt.Println(playerName + " left the server " + server.Name)
			status.leftNotified = true
			status.joinedNotified = false
		}
	}

	return previousPlayers
}

//TODO: figure out a processor func to process the messaging

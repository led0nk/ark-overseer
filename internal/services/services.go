package services

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/led0nk/ark-overseer/internal/interfaces"
	"github.com/led0nk/ark-overseer/internal/services/discord"
	"github.com/led0nk/ark-overseer/pkg/config"
	"github.com/led0nk/ark-overseer/pkg/events"
)

type Services interface {
	HandleEvent(context.Context, events.EventMessage)
}

type ServiceManager struct {
	services   map[string]interfaces.Notification
	cancelFunc map[string]context.CancelFunc
	mu         sync.Mutex
	initWg     *sync.WaitGroup
	logger     *slog.Logger
	em         *events.EventManager
}

func NewServiceManager(
	em *events.EventManager,
	initWg *sync.WaitGroup,
) *ServiceManager {
	return &ServiceManager{
		services:   make(map[string]interfaces.Notification),
		cancelFunc: make(map[string]context.CancelFunc),
		logger:     slog.Default().WithGroup("serviceManager"),
		em:         em,
		initWg:     initWg,
	}
}

func (sm *ServiceManager) HandleEvent(ctx context.Context, event events.EventMessage) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	switch event.Type {
	case "init.services":
		defer sm.initWg.Done()

		cfg, ok := event.Payload.(*config.Config)
		if !ok {
			sm.logger.ErrorContext(ctx, "invalid payload type", "error", event.Type)
			return
		}

		nService, err := cfg.GetSection("notification-service")
		if err != nil {
			sm.logger.Error("failed to get section from config", "error", err)
			return
		}
		for key, value := range nService {
			switch key {
			case "discord":
				err := sm.createDiscordService(ctx, value)
				if err != nil {
					sm.logger.ErrorContext(ctx, "failed to create discord service", "error", err)
					continue
				}
			}
		}
		sm.createServices()

	case "config.changed":
		sectionMap, ok := event.Payload.(map[interface{}]interface{})
		if !ok {
			sm.logger.ErrorContext(ctx, "invalid payload type", "error", event.Type)
			return
		}

		for serviceName, service := range sm.services {
			err := service.Disconnect()
			if err != nil {
				sm.logger.ErrorContext(ctx, "failed to disconnect notification service", "error", serviceName)
				continue
			}
		}
		sm.deleteServices()

		//NOTE: range over sectionMap for notification services
		for k, v := range sectionMap {
			switch k {
			case "discord":
				err := sm.createDiscordService(ctx, v)
				if err != nil {
					sm.logger.ErrorContext(ctx, "failed to create discord service", "error", err)
					continue
				}
			}
		}
		sm.createServices()
	}
}

func (sm *ServiceManager) createDiscordService(ctx context.Context, v interface{}) error {

	newConfig, ok := v.(map[interface{}]interface{})
	if !ok {
		return fmt.Errorf("invalid payload type")
	}

	token, ok := newConfig["token"].(string)
	if !ok {
		return fmt.Errorf("invalid token type")
	}

	channelID, ok := newConfig["channelID"].(string)
	if !ok {
		return fmt.Errorf("invalid channelID type")
	}

	var err error
	newDiscord, err := discord.NewDiscordNotifier(ctx, token, channelID)
	if err != nil {
		return fmt.Errorf("failed to create discord notifier: %w", err)
	}
	sm.services["discord"] = newDiscord
	return nil
}

func (sm *ServiceManager) createServices() {
	for serviceName, service := range sm.services {
		ctx, cancel := context.WithCancel(context.Background())
		sm.cancelFunc[serviceName] = cancel
		go sm.em.StartListening(ctx, service, serviceName, func() {})
	}
}

func (sm *ServiceManager) deleteServices() {
	for serviceName, cancel := range sm.cancelFunc {
		cancel()
		delete(sm.cancelFunc, serviceName)
		delete(sm.services, serviceName)
	}
}

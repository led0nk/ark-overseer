package services

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/led0nk/ark-overseer/internal"
	"github.com/led0nk/ark-overseer/internal/services/discord"
	"github.com/led0nk/ark-overseer/pkg/config"
	"github.com/led0nk/ark-overseer/pkg/events"
)

type ServiceManager struct {
	services   map[string]internal.Notification
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
		services:   make(map[string]internal.Notification),
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
				discordConfig, ok := value.(map[interface{}]interface{})
				if !ok {
					sm.logger.ErrorContext(ctx, "discord section has wrong type", "error", "discord")
					continue
				}

				token, ok := discordConfig["token"].(string)
				if !ok {
					sm.logger.ErrorContext(ctx, "token was not found or has wrong type", "error", "discord")
					continue
				}

				channelID, ok := discordConfig["channelID"].(string)
				if !ok {
					sm.logger.ErrorContext(ctx, "channelID was not found or has wrong type", "error", "discord")
					continue
				}

				sm.services["discord"], err = discord.NewDiscordNotifier(ctx, token, channelID)
				if err != nil {
					sm.logger.ErrorContext(ctx, "failed to create notification service", "error", err)
					continue
				}
			}
		}
		for serviceName, service := range sm.services {
			ctx, cancel := context.WithCancel(context.Background())
			sm.cancelFunc[serviceName] = cancel
			go sm.em.StartListening(ctx, service, serviceName)
		}
		sm.initWg.Done()

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

		for serviceName, cancel := range sm.cancelFunc {
			cancel()
			delete(sm.cancelFunc, serviceName)
			delete(sm.services, serviceName)
		}

		//NOTE: range over sectionMap for notification services
		for k, v := range sectionMap {
			switch k {
			case "discord":
				fmt.Println("case discord")
				newConfig, ok := v.(map[interface{}]interface{})
				if !ok {
					sm.logger.Error("invalid payload type", "error", event.Type)
					continue
				}

				token, ok := newConfig["token"].(string)
				if !ok {
					sm.logger.Error("invalid payload type", "error", event.Type)
					continue
				}

				channelID, ok := newConfig["channelID"].(string)
				if !ok {
					sm.logger.Error("invalid payload type", "error", event.Type)
					continue
				}

				var err error
				sm.services["discord"], err = discord.NewDiscordNotifier(ctx, token, channelID)
				if err != nil {
					sm.logger.ErrorContext(ctx, "failed to create discord notifier", "error ", err)
					continue
				}

			}
		}
		for serviceName, service := range sm.services {
			ctx, cancel := context.WithCancel(context.Background())
			sm.cancelFunc[serviceName] = cancel
			go sm.em.StartListening(ctx, service, serviceName)
		}
	}

}

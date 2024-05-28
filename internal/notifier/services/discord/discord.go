package discord

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/led0nk/ark-clusterinfo/internal/events"
)

type DiscordNotifier struct {
	logger    *slog.Logger
	session   *discordgo.Session
	token     string
	channelID string
}

func NewDiscordNotifier(token string, channelID string) (*DiscordNotifier, error) {
	discord := &DiscordNotifier{
		logger:    slog.Default().WithGroup("discord"),
		token:     token,
		channelID: channelID,
	}
	return discord, nil
}

func (dn *DiscordNotifier) HandleEvent(ctx context.Context, event events.EventMessage) {
	switch event.Type {
	case "playerJoined":
		fmt.Println(event.Payload, "joined the server")
		err := dn.Send(ctx, " joined the server ")
		if err != nil {
			dn.logger.ErrorContext(ctx, "failed to send message", "error", err)
		}
	case "playerLeft":
		fmt.Println(event.Payload, "left the server")
		err := dn.Send(ctx, " left the server ")
		if err != nil {
			dn.logger.ErrorContext(ctx, "failed to send message", "error", err)
		}
	default:
	}
}

func (dn *DiscordNotifier) StartListening(ctx context.Context, em *events.EventManager) {
	id, ch := em.Subscribe()
	if id == uuid.Nil {
		return
	}

	go func() {
		for event := range ch {
			dn.HandleEvent(ctx, event)
		}
	}()
}

func (dn *DiscordNotifier) Connect(ctx context.Context) error {
	session, err := discordgo.New(dn.token)
	if err != nil {
		dn.logger.ErrorContext(ctx, "failed to create session", "error", err)
		return err
	}

	dn.session = session
	err = dn.session.Open()
	if err != nil {
		dn.logger.ErrorContext(ctx, "failed to open session", "error", err)
		return err
	}

	return nil
}

func (dn *DiscordNotifier) Send(ctx context.Context, message string) error {
	_, err := dn.session.ChannelMessageSend(dn.channelID, message)
	if err != nil {
		dn.logger.ErrorContext(ctx, "failed to send discord message", "error", err)
		return err
	}
	return nil
}

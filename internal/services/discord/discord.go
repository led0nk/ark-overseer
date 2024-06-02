package discord

import (
	"context"
	"errors"
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/led0nk/ark-clusterinfo/internal/events"
)

type DiscordNotifier struct {
	logger    *slog.Logger
	session   *discordgo.Session
	token     string
	channelID string
}

func NewDiscordNotifier(ctx context.Context, token string, channelID string) (*DiscordNotifier, error) {
	discord := &DiscordNotifier{
		logger:    slog.Default().WithGroup("discord"),
		token:     token,
		channelID: channelID,
	}
	err := discord.Connect(ctx)
	if err != nil {
		discord.logger.ErrorContext(ctx, "failed to connect discord notification service", "error", err)
		return nil, err
	}
	return discord, nil
}

func (dn *DiscordNotifier) HandleEvent(ctx context.Context, event events.EventMessage) {
	switch event.Type {
	case "player.joined":
		msg, ok := event.Payload.(string)
		if !ok {
			dn.logger.ErrorContext(ctx, "invalid payload type for playerJoined event", "error", errors.New("payload not of type string"))
			return
		}
		err := dn.Send(ctx, msg)
		if err != nil {
			dn.logger.ErrorContext(ctx, "failed to send message", "error", err)
		}
	case "player.left":
		msg, ok := event.Payload.(string)
		if !ok {
			dn.logger.ErrorContext(ctx, "invalid payload type for playerLeft event", "error", errors.New("payload not of type string"))
			return
		}
		err := dn.Send(ctx, msg)
		if err != nil {
			dn.logger.ErrorContext(ctx, "failed to send message", "error", err)
		}
	default:
		return
	}
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

func (dn *DiscordNotifier) Setup(ctx context.Context, newDN *DiscordNotifier) error {
	dn = newDN

	err := dn.Connect(ctx)
	if err != nil {
		dn.logger.ErrorContext(ctx, "failed to setup discord service", "error", err)
		return err
	}
	return nil
}

func (dn *DiscordNotifier) Disconnect() error {

	dn.channelID = ""
	dn.token = ""

	return dn.session.Close()
}

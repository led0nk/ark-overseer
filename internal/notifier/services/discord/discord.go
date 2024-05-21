package discord

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

type DiscordNotifier struct {
	logger   *slog.Logger
	client   bot.Client
	notifyCh chan os.Signal
}

func NewDiscordNotifier(token string, opts ...bot.ConfigOpt) *DiscordNotifier {
	discord := &DiscordNotifier{
		logger:   slog.Default().WithGroup("discord"),
		client:   disgo.New(token, opts...),
		notifyCh: make(chan os.Signal, 1),
	}
	return discord
}

func (d *DiscordNotifier) Connect(ctx context.Context) {
	err := d.client.OpenGateway(ctx)
	if err != nil {
		d.logger.ErrorContext(ctx, "failed to open gateway", "error", err)
	}

	signal.Notify(d.notifyCh, syscall.SIGINT, syscall.SIGTERM)
	<-d.notifyCh
}

func (d *DiscordNotifier) Message(event *events.MessageCreate, message string) error {
	if message == "" {
		return errors.New("message cannot be empty")
	}

	event.Client().Rest().CreateMessage(event.ChannelID, discord.NewMessageCreateBuilder().SetContent(message).Build())
	return nil
}

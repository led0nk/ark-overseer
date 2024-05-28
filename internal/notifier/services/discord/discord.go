package discord

import (
	"context"
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

type DiscordNotifier struct {
	logger  *slog.Logger
	session *discordgo.Session
	channel *string
}

func NewDiscordNotifier(token string) (*DiscordNotifier, error) {
	session, err := discordgo.New(token)
	if err != nil {
		return nil, err
	}
	discord := &DiscordNotifier{
		logger:  slog.Default().WithGroup("discord"),
		session: session,
	}
	return discord, nil
}

func (d *DiscordNotifier) Connect(ctx context.Context) error {
	err := d.session.Open()
	if err != nil {
		d.logger.ErrorContext(ctx, "failed to open session", "error", err)
		return err
	}
	return nil
}

func (d *DiscordNotifier) Send(ctx context.Context, channelID string, message string) error {
	_, err := d.session.ChannelMessageSend(channelID, message)
	if err != nil {
		d.logger.ErrorContext(ctx, "failed to send discord message", "error", err)
		return err
	}
	return nil
}

func (d *DiscordNotifier) Setup(channelID string) {

}

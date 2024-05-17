package notifier

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/led0nk/ark-clusterinfo/internal"
	"github.com/led0nk/ark-clusterinfo/internal/parser"
)

type Notifier struct {
	parser internal.Parser
	ch     chan string
	logger *slog.Logger
}

func NewNotifier(p internal.Parser) *Notifier {
	return &Notifier{
		parser: p,
		ch:     make(chan string),
		logger: slog.Default().WithGroup("notifier"),
	}
}

func (n *Notifier) Create(ctx context.Context, trg *parser.Target) (*parser.Target, error) {
	n.ch <- "create"
	return n.parser.Create(ctx, trg)
}

func (n *Notifier) Delete(ctx context.Context, id uuid.UUID) error {
	n.ch <- "delete"
	return n.parser.Delete(ctx, id)
}

func (n *Notifier) List(ctx context.Context) ([]*parser.Target, error) {
	n.ch <- "list"
	return n.parser.List(ctx)
}

func (n *Notifier) Signal() <-chan string {
	return n.ch
}

func (n *Notifier) Run(obs internal.Observer, ctx context.Context) {
	for {
		notification := <-n.Signal()
		n.logger.InfoContext(ctx, "targets were updated", "type", notification)
		go obs.ManageScraper(ctx)
	}
}

package notifier

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/led0nk/ark-clusterinfo/internal"
	"github.com/led0nk/ark-clusterinfo/internal/parser"
)

type Notifier struct {
	parser internal.Parser
	ch     chan struct{}
}

func NewNotifier(p internal.Parser) *Notifier {
	return &Notifier{
		parser: p,
		ch:     make(chan struct{}),
	}
}

func (n *Notifier) Create(ctx context.Context, trg *parser.Target) (*parser.Target, error) {
	n.ch <- struct{}{}
	return n.parser.Create(ctx, trg)
}

func (n *Notifier) Delete(ctx context.Context, id uuid.UUID) error {
	n.ch <- struct{}{}
	return n.parser.Delete(ctx, id)
}

func (n *Notifier) List(ctx context.Context) ([]*parser.Target, error) {
	n.ch <- struct{}{}
	return n.parser.List(ctx)
}

func (n *Notifier) Signal() <-chan struct{} {
	return n.ch
}

func (n *Notifier) Run(obs internal.Observer, ctx context.Context) {
	for {
		<-n.Signal()
		fmt.Println("targets updated")
		go obs.ManageScraper(ctx)
	}
}

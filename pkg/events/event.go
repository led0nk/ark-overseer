package events

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/google/uuid"
)

type EventHandler interface {
	HandleEvent(context.Context, EventMessage)
}

type EventManager struct {
	logger     *slog.Logger
	subscriber map[uuid.UUID]chan EventMessage
	mu         sync.RWMutex
}

type EventMessage struct {
	Type    string
	Payload interface{}
}

func NewEventManager() *EventManager {
	return &EventManager{
		logger:     slog.Default().WithGroup("event"),
		subscriber: make(map[uuid.UUID]chan EventMessage),
	}
}

func (e *EventManager) Subscribe(name string) (uuid.UUID, <-chan EventMessage) {
	e.mu.Lock()
	defer e.mu.Unlock()

	id, err := uuid.NewUUID()
	if err != nil {
		return uuid.Nil, nil
	}

	ch := make(chan EventMessage)
	e.subscriber[id] = ch

	e.logger.Info("service subscribed to eventManager", "service id", id, "service name", name)
	return id, ch
}

func (e *EventManager) Unsubscribe(id uuid.UUID, name string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	ch := e.subscriber[id]
	close(ch)
	delete(e.subscriber, id)
	e.logger.Info("service unsubscribed to eventManager", "service id", id, "service name", name)
}

func (e *EventManager) Publish(emsg EventMessage) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for subscriber, ch := range e.subscriber {
		e.logger.Debug("publish eventMessage", "debug", "publish", subscriber.String(), fmt.Sprintf("%s", emsg))
		ch <- emsg
	}
}

func (e *EventManager) StartListening(ctx context.Context, handler EventHandler, serviceName string) {
	id, ch := e.Subscribe(serviceName)
	if id == uuid.Nil {
		return
	}
	defer e.Unsubscribe(id, serviceName)

	for {
		select {
		case <-ctx.Done():
			return
		case event := <-ch:
			handler.HandleEvent(ctx, event)
		}
	}
}

package events

import (
	"log/slog"
	"sync"

	"github.com/google/uuid"
)

type eventManager struct {
	logger     *slog.Logger
	subscriber map[uuid.UUID]chan string
	mu         sync.RWMutex
}

type eventMessage struct {
	eventType string
}

func NewEventManager() *eventManager {
	return &eventManager{
		logger:     slog.Default().WithGroup("event"),
		subscriber: make(map[uuid.UUID]chan string),
	}
}

func (e *eventManager) Subscribe() (uuid.UUID, <-chan string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	id, err := uuid.NewUUID()
	if err != nil {
		return uuid.Nil, nil
	}

	ch := make(chan string)
	e.subscriber[id] = ch

	e.logger.Info("service subscribed", "service id", id)
	return id, ch
}

func (e *eventManager) Unsubscribe(id uuid.UUID) {
	e.mu.Lock()
	defer e.mu.Unlock()

	ch := e.subscriber[id]
	close(ch)
	delete(e.subscriber, id)
	e.logger.Info("service unsubscribed", "service id", id)
}

package events

import (
	"log/slog"
	"sync"

	"github.com/google/uuid"
)

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

func (e *EventManager) Subscribe() (uuid.UUID, <-chan EventMessage) {
	e.mu.Lock()
	defer e.mu.Unlock()

	id, err := uuid.NewUUID()
	if err != nil {
		return uuid.Nil, nil
	}

	ch := make(chan EventMessage)
	e.subscriber[id] = ch

	e.logger.Info("service subscribed", "service id", id)
	return id, ch
}

func (e *EventManager) Unsubscribe(id uuid.UUID) {
	e.mu.Lock()
	defer e.mu.Unlock()

	ch := e.subscriber[id]
	close(ch)
	delete(e.subscriber, id)
	e.logger.Info("service unsubscribed", "service id", id)
}

func (e *EventManager) Publish(emsg EventMessage) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, ch := range e.subscriber {
		ch <- emsg
	}
}

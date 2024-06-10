package events

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type MockHandler struct {
	handledEvents []EventMessage
	mu            sync.Mutex
}

func (m *MockHandler) HandleEvent(ctx context.Context, event EventMessage) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handledEvents = append(m.handledEvents, event)
}

func TestMultipleSubscribers(t *testing.T) {
	em := NewEventManager()
	_, ch1 := em.Subscribe("service-1")
	_, ch2 := em.Subscribe("service-2")

	go em.Publish(EventMessage{Type: "test-event", Payload: "test-payload"})

	select {
	case event := <-ch1:
		assert.Equal(t, "test-event", event.Type)
		assert.Equal(t, "test-payload", event.Payload)
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for event for service-1")
	}

	select {
	case event := <-ch2:
		assert.Equal(t, "test-event", event.Type)
		assert.Equal(t, "test-payload", event.Payload)
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for event for service-2")
	}
}

func TestStartListeningWithTimeout(t *testing.T) {
	em := NewEventManager()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	mockHandler := &MockHandler{}
	subscribed := false

	go em.StartListening(ctx, mockHandler, "test-service", func() {
		subscribed = true
	})

	time.Sleep(100 * time.Millisecond)
	assert.True(t, subscribed, "subscribed should have been called")

	em.Publish(EventMessage{Type: "test-event", Payload: "test-payload"})

	time.Sleep(100 * time.Millisecond)

	mockHandler.mu.Lock()
	assert.Len(t, mockHandler.handledEvents, 1)
	assert.Equal(t, "test-event", mockHandler.handledEvents[0].Type)
	assert.Equal(t, "test-payload", mockHandler.handledEvents[0].Payload)
	mockHandler.mu.Unlock()

	<-ctx.Done()
	assert.Equal(t, context.DeadlineExceeded, ctx.Err(), "context should have timed out")

	em.Publish(EventMessage{Type: "after-ctx-done", Payload: "payload-after-ctx-done"})

	time.Sleep(100 * time.Millisecond)

	mockHandler.mu.Lock()
	assert.Len(t, mockHandler.handledEvents, 1, "no more events handled after ctx-done")
	mockHandler.mu.Unlock()
}

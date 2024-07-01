package storage

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/led0nk/ark-overseer/internal/model"
	"go.opentelemetry.io/otel"
)

var tracer = otel.GetTracerProvider().Tracer("github.com/led0nk/ark-overseer/internal/storage")

type Database interface {
	Create(context.Context, *model.Server) (*model.Server, error)
	List(context.Context) ([]*model.Server, error)
	GetByName(context.Context, string) (*model.Server, error)
	GetByID(context.Context, uuid.UUID) (*model.Server, error)
	Delete(context.Context, uuid.UUID) error
	Update(context.Context, *model.Server) error
	Save() error
}

type ServerStorage struct {
	filename string
	server   map[uuid.UUID]*model.Server
	mu       sync.Mutex
}

func NewServerStorage(ctx context.Context, filename string) (*ServerStorage, error) {
	store := &ServerStorage{
		filename: filename,
		server:   make(map[uuid.UUID]*model.Server),
	}
	if err := store.load(); err != nil {
		return nil, err
	}

	go store.autoSave(ctx)

	return store, nil
}

func (s *ServerStorage) Save() error {
	as_json, err := json.MarshalIndent(s.server, "", "\t")
	if err != nil {
		return err
	}

	err = os.WriteFile(s.filename, as_json, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (s *ServerStorage) autoSave(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := s.Save()
			if err != nil {
				return
			}
		}
	}
}

func (s *ServerStorage) load() error {
	if _, err := os.Stat(s.filename); os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(s.filename), 0777)
		if err != nil {
			return err
		}
		err = s.Save()
		if err != nil {
			return err
		}
	}
	data, err := os.ReadFile(s.filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &s.server)
}

func (s *ServerStorage) Create(ctx context.Context, server *model.Server) (*model.Server, error) {
	_, span := tracer.Start(ctx, "Create")
	defer span.End()

	span.AddEvent("Lock")
	s.mu.Lock()

	defer span.AddEvent("Unlock")
	defer s.mu.Unlock()

	if server.ID == uuid.Nil {
		server.ID = uuid.New()
	}

	s.server[server.ID] = server
	if err := s.Save(); err != nil {
		return nil, err
	}

	return server, nil
}

func (s *ServerStorage) Update(ctx context.Context, server *model.Server) error {
	_, span := tracer.Start(ctx, "Update")
	defer span.End()

	span.AddEvent("Lock")
	s.mu.Lock()

	defer span.AddEvent("Unlock")
	defer s.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		s.server[server.ID] = server
		return nil
	}
}

func (s *ServerStorage) GetByName(ctx context.Context, name string) (*model.Server, error) {
	_, span := tracer.Start(ctx, "GetByName")
	defer span.End()

	span.AddEvent("Lock")
	s.mu.Lock()

	span.AddEvent("Unlock")
	defer s.mu.Unlock()

	if name == "" {
		return nil, errors.New("empty name")
	}

	fetchedServer := &model.Server{}
	for _, server := range s.server {
		if server.Name == name {
			fetchedServer = server
		}
	}
	return fetchedServer, nil
}

func (s *ServerStorage) GetByID(ctx context.Context, id uuid.UUID) (*model.Server, error) {
	_, span := tracer.Start(ctx, "GetByID")
	defer span.End()

	span.AddEvent("Lock")
	s.mu.Lock()

	defer span.AddEvent("Unlock")
	defer s.mu.Unlock()

	if id == uuid.Nil {
		return nil, errors.New("empty uuid")
	}

	for _, server := range s.server {
		if server.ID == id {
			return server, nil
		}
	}
	return nil, errors.New("server not found")
}

func (s *ServerStorage) Delete(ctx context.Context, ID uuid.UUID) error {
	_, span := tracer.Start(ctx, "Delete")
	defer span.End()

	span.AddEvent("Lock")
	s.mu.Lock()

	defer span.AddEvent("Unlock")
	defer s.mu.Unlock()

	if ID == uuid.Nil {
		return errors.New("requires server ID")
	}

	if _, exists := s.server[ID]; !exists {
		return errors.New("server doesn't exist")
	}

	delete(s.server, ID)

	return nil
}

func (s *ServerStorage) List(ctx context.Context) ([]*model.Server, error) {
	_, span := tracer.Start(ctx, "List")
	defer span.End()

	span.AddEvent("Lock")
	s.mu.Lock()

	defer span.AddEvent("Unlock")
	defer s.mu.Unlock()

	serverlist := make([]*model.Server, 0, len(s.server))
	for _, server := range s.server {
		serverlist = append(serverlist, server)
	}

	sort.Slice(serverlist, func(i, j int) bool { return serverlist[i].Name < serverlist[j].Name })
	return serverlist, nil
}

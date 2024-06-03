package jsondb

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sort"
	"sync"

	"github.com/google/uuid"
	"github.com/led0nk/ark-overseer/internal/model"
)

type ServerStorage struct {
	filename string
	server   map[uuid.UUID]*model.Server
	mu       sync.Mutex
}

func NewServerStorage(ctx context.Context, filename string) (*ServerStorage, error) {
	cluster := &ServerStorage{
		filename: filename,
		server:   make(map[uuid.UUID]*model.Server),
	}
	if err := cluster.readJSON(ctx); err != nil {
		return nil, err
	}
	return cluster, nil
}

func (s *ServerStorage) writeJSON(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
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
}

// read JSON data from file = filename
func (s *ServerStorage) readJSON(ctx context.Context) error {
	if _, err := os.Stat(s.filename); os.IsNotExist(err) {
		err = s.writeJSON(ctx)
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
	s.mu.Lock()
	defer s.mu.Unlock()

	if server.ID == uuid.Nil {
		server.ID = uuid.New()
	}

	s.server[server.ID] = server
	if err := s.writeJSON(ctx); err != nil {
		return nil, err
	}

	return server, nil
}

func (s *ServerStorage) Update(ctx context.Context, server *model.Server) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		s.server[server.ID] = server
		return s.writeJSON(ctx)
	}
}

func (s *ServerStorage) GetByName(ctx context.Context, name string) (*model.Server, error) {
	s.mu.Lock()
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
	s.mu.Lock()
	defer s.mu.Unlock()

	if id == uuid.Nil {
		return nil, errors.New("empty uuid")
	}

	fetchedServer := &model.Server{}
	for _, server := range s.server {
		if server.ID == id {
			fetchedServer = server
		}
	}
	return fetchedServer, nil
}

func (s *ServerStorage) Delete(ctx context.Context, ID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if ID == uuid.Nil {
		return errors.New("requires server ID")
	}

	if _, exists := s.server[ID]; !exists {
		return errors.New("server doesn't exist")
	}

	delete(s.server, ID)

	if err := s.writeJSON(ctx); err != nil {
		return err
	}
	return nil
}

func (s *ServerStorage) List(ctx context.Context) ([]*model.Server, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	serverlist := make([]*model.Server, 0, len(s.server))
	for _, server := range s.server {
		serverlist = append(serverlist, server)
	}

	sort.Slice(serverlist, func(i, j int) bool { return serverlist[i].Name < serverlist[j].Name })
	return serverlist, nil
}

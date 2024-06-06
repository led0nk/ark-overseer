package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/led0nk/ark-overseer/internal/model"
	"github.com/stretchr/testify/assert"
)

func createTempDir(t *testing.T) string {
	dir, err := os.MkdirTemp("", "storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %s", err)
	}
	return dir
}

func cleanupTempDir(t *testing.T, dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		t.Fatalf("Failed to remove temp dir: %s", err)
	}
}

func TestNewServerStorage(t *testing.T) {
	ctx := context.Background()
	dir := createTempDir(t)
	defer cleanupTempDir(t, dir)

	storage, err := NewServerStorage(ctx, filepath.Join(dir, "cluster.json"))
	assert.NoError(t, err)
	assert.NotNil(t, storage)
}

func TestServerStorageCRUD(t *testing.T) {
	ctx := context.Background()
	dir := createTempDir(t)
	defer cleanupTempDir(t, dir)

	storage, err := NewServerStorage(ctx, filepath.Join(dir, "cluster.json"))
	assert.NoError(t, err)

	tests := []struct {
		name      string
		server    *model.Server
		update    *model.Server
		expectErr bool
	}{
		{
			name: "create and get server",
			server: &model.Server{
				ID:   uuid.MustParse("64dfb157-37b8-41de-b24d-14f304e15402"),
				Name: "test server",
				Addr: "127.0.0.1:27015"},
			expectErr: false,
		},
		{
			name: "update server",
			server: &model.Server{
				ID:   uuid.MustParse("047d119f-fa67-4e1f-ae1a-85ca77f67a74"),
				Name: "test server",
				Addr: "127.0.0.1:27015"},
			update: &model.Server{
				ID:   uuid.MustParse("047d119f-fa67-4e1f-ae1a-85ca77f67a74"),
				Name: "updated server",
				Addr: "127.0.0.1:27015"},
			expectErr: false,
		},
		{
			name: "delete server",
			server: &model.Server{
				ID:   uuid.MustParse("fd1c54a0-d2c4-4407-8501-9085ffe20902"),
				Name: "test server",
				Addr: "127.0.0.1:27015"},
			expectErr: false,
		},
		{
			name: "get non-existent server",
			server: &model.Server{
				ID:   uuid.MustParse("3854e285-6cde-4cfa-9924-10620fdcce5c"),
				Name: "non-existent server",
				Addr: "127.0.0.1:27016"},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.expectErr {
				_, err = storage.Create(ctx, tt.server)
				assert.NoError(t, err)

				retrieved, err := storage.GetByID(ctx, tt.server.ID)
				assert.NoError(t, err)
				assert.Equal(t, tt.server.Name, retrieved.Name)
			}

			if tt.update != nil {
				err = storage.Update(ctx, tt.update)
				assert.NoError(t, err)

				retrieved, err := storage.GetByID(ctx, tt.update.ID)
				assert.NoError(t, err)
				assert.Equal(t, tt.update.Name, retrieved.Name)
			}

			if !tt.expectErr {
				err = storage.Delete(ctx, tt.server.ID)
				assert.NoError(t, err)

				retrieved, err := storage.GetByID(ctx, tt.server.ID)
				assert.Error(t, err)
				assert.Nil(t, retrieved)
			}
		})
	}
}

package blacklist

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
	dir, err := os.MkdirTemp("", "blacklist_test")
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

func TestNewBlacklist(t *testing.T) {
	dir := createTempDir(t)
	defer cleanupTempDir(t, dir)

	bl, err := NewBlacklist(filepath.Join(dir, "blacklist.json"))
	assert.NoError(t, err)
	assert.NotNil(t, bl)
}

func TestBlacklistCRUD(t *testing.T) {
	dir := createTempDir(t)
	defer cleanupTempDir(t, dir)

	bl, err := NewBlacklist(filepath.Join(dir, "blacklist.json"))
	assert.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name      string
		player    *model.BlacklistPlayers
		expectErr bool
	}{
		{
			name: "Create and List Player",
			player: &model.BlacklistPlayers{
				ID:   uuid.MustParse("d8e92b5e-4d1d-4f38-bdbc-d1d3f1d2e3b7"),
				Name: "Test Player",
			},
			expectErr: false,
		},
		{
			name: "Delete Player",
			player: &model.BlacklistPlayers{
				ID:   uuid.MustParse("b2d4f7d6-5a27-4b8a-bef4-c1a0d3f8f3b7"),
				Name: "Test Player",
			},
			expectErr: false,
		},
		{
			name: "List Players",
			player: &model.BlacklistPlayers{
				ID:   uuid.MustParse("c2e4e7d6-6a27-4c8a-bef4-c1a0d3f8f3b7"),
				Name: "Another Test Player",
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createdPlayer, err := bl.Create(ctx, tt.player)
			if tt.expectErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.player.Name, createdPlayer.Name)

			players := bl.List(ctx)
			assert.NotEmpty(t, players)

			if err := bl.Delete(ctx, tt.player.ID); !tt.expectErr {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}

			players = bl.List(ctx)
			for _, player := range players {
				assert.NotEqual(t, tt.player.ID, player.ID)
			}
		})
	}
}

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/led0nk/ark-overseer/pkg/events"
	"github.com/stretchr/testify/assert"
)

func createTempDir(t *testing.T) string {
	dir, err := os.MkdirTemp("", "config_test")
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

func TestNewConfiguration(t *testing.T) {
	dir := createTempDir(t)
	defer cleanupTempDir(t, dir)

	em := events.NewEventManager()
	cfg, err := NewConfiguration(filepath.Join(dir, "config.json"), em)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
}

func TestConfigCRUD(t *testing.T) {
	dir := createTempDir(t)
	defer cleanupTempDir(t, dir)

	em := events.NewEventManager()
	cfg, err := NewConfiguration(filepath.Join(dir, "config.json"), em)
	assert.NoError(t, err)

	tests := []struct {
		name      string
		section   string
		key       string
		value     interface{}
		expectErr bool
	}{
		{
			name:      "update and get section",
			section:   "notification-service",
			key:       "discord",
			value:     map[interface{}]interface{}{"token": "123456", "channelID": "abcdef"},
			expectErr: false,
		},
		{
			name:      "get non-existent section",
			section:   "non-existent-section",
			key:       "key",
			value:     nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.expectErr {
				err := cfg.Update(tt.section, tt.key, tt.value)
				assert.NoError(t, err)

			}

			section, err := cfg.GetSection(tt.section)
			if tt.expectErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.value, section[tt.key])
		})
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	dir := createTempDir(t)
	defer cleanupTempDir(t, dir)

	em := events.NewEventManager()
	cfg, err := NewConfiguration(filepath.Join(dir, "config.yaml"), em)
	assert.NoError(t, err)

	section := "notification-service"
	key := "discord"
	value := map[interface{}]interface{}{"token": "123456", "channelID": "abcdef"}
	err = cfg.Update(section, key, value)
	assert.NoError(t, err)

	err = cfg.Save()
	assert.NoError(t, err)

	cfg, err = NewConfiguration(filepath.Join(dir, "config.yaml"), em)
	assert.NoError(t, err)

	loadedSection, err := cfg.GetSection(section)
	assert.NoError(t, err)
	assert.Equal(t, value, loadedSection[key])
}

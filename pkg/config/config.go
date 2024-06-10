package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/led0nk/ark-overseer/pkg/events"
	"gopkg.in/yaml.v2"
)

type Configuration interface {
	Load() error
	Save() error
	Update(string, string, interface{}) error
}

type Config struct {
	filename string
	mu       sync.Mutex
	config   map[interface{}]interface{}
	em       *events.EventManager
}

func NewConfiguration(
	filename string,
	em *events.EventManager,
) (*Config, error) {
	cfg := &Config{
		filename: filename,
		config:   make(map[interface{}]interface{}),
		em:       em,
	}

	err := cfg.Load()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Load() error {
	_, err := os.Stat(c.filename)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(filepath.Dir(c.filename), 0777)
			if err != nil {
				return err
			}
			c.config["notification-service"] = nil
			err = c.Save()
			if err != nil {
				return err
			}
		}
	}

	data, err := os.ReadFile(c.filename)
	if err != nil {
		return err
	}

	var config map[interface{}]interface{}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return err
	}
	c.config = config

	return nil
}

func (c *Config) Save() error {
	data, err := yaml.Marshal(c.config)
	if err != nil {
		return err
	}

	err = os.WriteFile(c.filename, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) Update(section string, key string, value interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	sectionMap, ok := c.config[section].(map[interface{}]interface{})
	if !ok {
		sectionMap = make(map[interface{}]interface{})
		c.config[section] = sectionMap
	}
	sectionMap[key] = value

	err := c.Save()
	if err != nil {
		return err
	}
	c.em.Publish(events.EventMessage{Type: "config.changed", Payload: sectionMap})

	return nil
}

func (c *Config) GetSection(section string) (map[interface{}]interface{}, error) {
	sectionMap, exists := c.config[section]
	if !exists {
		return nil, fmt.Errorf("section %s not found", section)
	}

	sectionData, ok := sectionMap.(map[interface{}]interface{})
	if !ok {
		return nil, fmt.Errorf("section %s not a valid type", section)
	}

	return sectionData, nil
}

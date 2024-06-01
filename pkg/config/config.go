package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/led0nk/ark-clusterinfo/internal/events"
	"gopkg.in/yaml.v2"
)

type Configuration struct {
	filename string
	mu       sync.Mutex
	config   map[string]interface{}
	em       *events.EventManager
}

func NewConfiguration(filename string) (*Configuration, error) {
	cfg := &Configuration{
		filename: filename,
	}

	err := cfg.Read()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Configuration) Read() error {
	if _, err := os.Stat(c.filename); os.IsNotExist(err) {
		err = c.Write()
		if err != nil {
			return err
		}
	}

	data, err := os.ReadFile(c.filename)
	if err != nil {
		return err
	}

	var config map[string]interface{}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return err
	}
	c.config = config

	return nil
}

func (c *Configuration) Write() error {
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

func (c *Configuration) Update(section string, key string, value interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	sectionMap, ok := c.config[section].(map[interface{}]interface{})
	if !ok {
		sectionMap = make(map[interface{}]interface{})
		c.config[section] = sectionMap
	}
	sectionMap[key] = value

	err := c.Write()
	if err != nil {
		return err
	}

	c.em.Publish(events.EventMessage{Type: "configChanged", Payload: sectionMap})

	return nil
}

func (c *Configuration) GetSection(section string) (map[interface{}]interface{}, error) {
	sectionMap, ok := c.config[section].(map[interface{}]interface{})
	if !ok {
		return nil, fmt.Errorf("section %s not found", section)
	}
	return sectionMap, nil
}

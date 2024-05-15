package cluster

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"sort"
	"sync"

	"github.com/FlowingSPDG/go-steam"
	"github.com/google/uuid"
	"github.com/led0nk/ark-clusterinfo/internal/model"
)

type Cluster struct {
	filename string
	server   map[uuid.UUID]*model.Server
	mu       sync.Mutex
}

func NewCluster(filename string) (*Cluster, error) {
	cluster := &Cluster{
		filename: filename,
		server:   make(map[uuid.UUID]*model.Server),
	}
	if err := cluster.readJSON(); err != nil {
		return nil, err
	}
	return cluster, nil
}

func (c *Cluster) writeJSON() error {

	as_json, err := json.MarshalIndent(c.server, "", "\t")
	if err != nil {
		return err
	}

	err = os.WriteFile(c.filename, as_json, 0644)
	if err != nil {
		return err
	}
	return nil
}

// read JSON data from file = filename
func (c *Cluster) readJSON() error {

	if _, err := os.Stat(c.filename); os.IsNotExist(err) {
		return errors.New("file does not exist")
	}
	data, err := os.ReadFile(c.filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &c.server)
}

func (c *Cluster) CreateServer(server *model.Server) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if server.ID == uuid.Nil {
		server.ID = uuid.New()
	}

	c.server[server.ID] = server
	if err := c.writeJSON(); err != nil {
		return "", err
	}

	return server.Name, nil
}

func (c *Cluster) GetServerByName(name string) (*model.Server, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if name == "" {
		return nil, errors.New("empty name")
	}

	fetchedServer := &model.Server{}
	for _, server := range c.server {
		if server.Name == name {
			fetchedServer = server
		}
	}
	return fetchedServer, nil
}

func (c *Cluster) DeleteServer(ID uuid.UUID) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ID == uuid.Nil {
		return errors.New("requires server ID")
	}

	if _, exists := c.server[ID]; !exists {
		return errors.New("server doesn't exist")
	}

	delete(c.server, ID)

	if err := c.writeJSON(); err != nil {
		return err
	}
	return nil
}

func (c *Cluster) GetServerInfo() ([]*model.Server, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	serverlist := make([]*model.Server, 0, len(c.server))
	for _, server := range c.server {
		var err error
		helpServer, _ := steam.Connect(server.Addr)
		server.ServerInfo, _ = helpServer.Info()
		server.PlayersInfo, err = helpServer.PlayersInfo()
		if err != nil {
			log.Println(err)
		}
		var playerList []*steam.Player

		for _, player := range server.PlayersInfo.Players {
			if player != nil && len(player.Name) > 2 {
				playerList = append(playerList, player)
			}
		}
		server.PlayersInfo.Players = playerList
		server.ServerInfo.Players = len(playerList)
		serverlist = append(serverlist, server)
	}
	sort.Slice(serverlist, func(i, j int) bool { return serverlist[i].Name < serverlist[j].Name })
	return serverlist, nil
}

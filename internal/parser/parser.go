package parser

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sort"
	"sync"

	"github.com/google/uuid"
)

type Parser struct {
	filename string
	targets  map[uuid.UUID]*Target
	mu       sync.Mutex
}

type Target struct {
	ID   uuid.UUID
	Name string
	Addr string
}

func NewParserWithTargets(filename string) (*Parser, error) {
	parser := &Parser{
		filename: filename,
		targets:  make(map[uuid.UUID]*Target),
	}
	if err := parser.readJSON(); err != nil {
		return nil, err
	}
	return parser, nil
}

func (p *Parser) writeJSON() error {
	as_json, err := json.MarshalIndent(p.targets, "", "\t")
	if err != nil {
		return err
	}

	err = os.WriteFile(p.filename, as_json, 0644)
	if err != nil {
		return err
	}
	return nil
}

// read JSON data from file = filename
func (p *Parser) readJSON() error {
	if _, err := os.Stat(p.filename); os.IsNotExist(err) {
		err = p.writeJSON()
		if err != nil {
			return err
		}
	}

	data, err := os.ReadFile(p.filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &p.targets)
}

func (p *Parser) List(ctx context.Context) ([]*Target, error) {
	trgtlist := make([]*Target, 0, len(p.targets))
	for _, target := range p.targets {
		trgtlist = append(trgtlist, target)
	}

	sort.Slice(trgtlist, func(i, j int) bool { return trgtlist[i].Name < trgtlist[j].Name })
	return trgtlist, nil
}

func (p *Parser) Create(ctx context.Context, trg *Target) (*Target, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if trg.ID == uuid.Nil {
		trg.ID = uuid.New()
	}
	if trg.Name == "" || trg.Addr == "" {
		return nil, errors.New("empty values are not allowed")
	}
	p.targets[trg.ID] = trg
	err := p.writeJSON()
	if err != nil {
		return nil, err
	}
	return trg, nil
}

func (p *Parser) Delete(ctx context.Context, ID uuid.UUID) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if ID == uuid.Nil {
		return errors.New("requires server ID")
	}

	if _, exists := p.targets[ID]; !exists {
		return errors.New("target doesn't exist")
	}

	delete(p.targets, ID)

	if err := p.writeJSON(); err != nil {
		return err
	}
	return nil
}

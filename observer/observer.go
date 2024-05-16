package observer

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"sync"

	"github.com/FlowingSPDG/go-steam"
	"github.com/google/uuid"
	"github.com/led0nk/ark-clusterinfo/internal"
	"github.com/led0nk/ark-clusterinfo/internal/model"
	"github.com/led0nk/ark-clusterinfo/internal/parser"
)

type Observer struct {
	endpoints   map[uuid.UUID]*parser.Target
	serverStore internal.ServerStore
	parser      internal.Parser
	logger      *slog.Logger
	mu          sync.Mutex
}

func NewObserver(sStore internal.ServerStore, prs internal.Parser) (*Observer, error) {
	observer := &Observer{
		endpoints:   make(map[uuid.UUID]*parser.Target),
		serverStore: sStore,
		parser:      prs,
		logger:      slog.Default(),
	}
	return observer, nil
}

func (o *Observer) ReadEndpoint(target *parser.Target) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if target.ID == uuid.Nil {
		target.ID = uuid.New()
	}

	o.endpoints[target.ID] = target
	return nil
}

// only for testing purpose -> should be switched to unmarshalling data into *model.Server
type Result struct {
	ServerInfo *steam.InfoResponse
	PlayerInfo *steam.PlayersInfoResponse
}

func (o *Observer) DataScraper(ctx context.Context, id uuid.UUID, c chan<- any) {
	for _, endpoint := range o.endpoints {
		if endpoint.ID == id {
			for {
				helpSrv, err := steam.Connect(endpoint.Addr)
				if err != nil {
					log.Println("error connecting to endpoint", err)
					continue
				}

				infoResponse, err := helpSrv.Info()
				if err != nil {
					log.Println("error fetching ServerInfo", err)
					continue
				}

				playerResponse, err := helpSrv.PlayersInfo()
				if err != nil {
					log.Println("error fetching PlayersInfo", err)
					continue
				}

				result := Result{
					ServerInfo: infoResponse,
					PlayerInfo: playerResponse,
				}
				c <- result
			}
		}
	}
}

func (o *Observer) InitScraper() {
	resultCh := make(chan any)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	targets, err := o.parser.ListTargets()
	if err != nil {
		o.logger.ErrorContext(ctx, "failed to list targets", "error", err)
		return
	}

	for _, target := range targets {
		err := o.ReadEndpoint(target)
		if err != nil {
			o.logger.ErrorContext(ctx, "failed to read endpoints", "error", err)
			return
		}
		go o.DataScraper(ctx, target.ID, resultCh)
	}

	for result := range resultCh {
		fmt.Println(result)
	}
}

func Unmarshal(data any) (*model.Server, error) {
	srv := &model.Server{}
	//structType := reflect.TypeOf(data)

	return srv, nil
}

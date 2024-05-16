package observer

import (
	"context"
	"log/slog"
	"reflect"
	"strings"
	"sync"
	"time"

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

func (o *Observer) DataScraper(ctx context.Context, t *parser.Target, c chan<- *model.Server) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			helpSrv, err := steam.Connect(t.Addr)
			if err != nil {
				o.logger.ErrorContext(ctx, "error connecting to endpoint", "error", err)
				continue
			}

			infoResponse, err := helpSrv.Info()
			if err != nil {
				o.logger.ErrorContext(ctx, "error fetching ServerInfo", "error", err)
				continue
			}

			playerResponse, err := helpSrv.PlayersInfo()
			if err != nil {
				o.logger.ErrorContext(ctx, "error fetching PlayersInfo", "error", err)
				continue
			}

			ping, err := helpSrv.Ping()
			if err != nil {
				o.logger.ErrorContext(ctx, "failed to ping server", "error", err)
				continue
			}

			var status bool
			if ping < time.Duration(5*time.Second) {
				status = true
			}

			server := &model.Server{
				Name:        t.Name,
				Addr:        t.Addr,
				ID:          t.ID,
				Status:      status,
				ServerInfo:  model.ToServerInfo(infoResponse),
				PlayersInfo: model.ToPlayerInfo(playerResponse),
			}
			ReplaceNullCharsInStruct(server)

			c <- server
		}
	}
}

func (o *Observer) InitScraper(ctx context.Context, targets []*parser.Target) {
	resultCh := make(chan *model.Server)

	for _, target := range targets {
		err := o.ReadEndpoint(target)
		if err != nil {
			o.logger.ErrorContext(ctx, "failed to read endpoints", "error", err)
			return
		}
		go o.DataScraper(ctx, target, resultCh)
	}

	for result := range resultCh {
		_, err := o.serverStore.CreateOrUpdateServer(result)
		if err != nil {
			o.logger.ErrorContext(ctx, "failed to update server info", "error", err)
		}
	}
}

func ReplaceNullCharsInStruct(s any) {
	v := reflect.ValueOf(s)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return
	}
	replaceNullChars(v.Elem())
}

func replaceNullChars(v reflect.Value) {
	switch v.Kind() {
	case reflect.String:
		str := v.Interface().(string)
		str = strings.Trim(str, "\u0000")
		v.SetString(str)
	case reflect.Ptr:
		if v.IsNil() {
			return
		}
		replaceNullChars(v.Elem())
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			replaceNullChars(v.Field(i))
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			replaceNullChars(v.Index(i))
		}
	}
}

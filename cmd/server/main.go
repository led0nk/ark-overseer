package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	v1 "github.com/led0nk/ark-clusterinfo/api/v1"
	"github.com/led0nk/ark-clusterinfo/internal"
	"github.com/led0nk/ark-clusterinfo/internal/jsondb"
	"github.com/led0nk/ark-clusterinfo/internal/model"
	"github.com/led0nk/ark-clusterinfo/internal/model/templates"
	"github.com/led0nk/ark-clusterinfo/internal/notifier"
	"github.com/led0nk/ark-clusterinfo/observer"
)

func main() {

	var (
		addr = flag.String("addr", "localhost:8080", "server port")
		//grpcaddr    = flag.String("grpcaddr", "", "grpc address, e.g. localhost:4317")
		//dbase       = flag.String("db", "file://testdata", "path to database")
		domain      = flag.String("domain", "127.0.0.1", "given domain for cookies/mail")
		logLevelStr = flag.String("loglevel", "INFO", "define the level for logs")
		sStore      internal.ServerStore
		obs         internal.Observer
		logLevel    slog.Level
	)
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := logLevel.UnmarshalText([]byte(*logLevelStr))
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	if err != nil {
		logger.ErrorContext(ctx, "error parsing loglevel", "loglevel", *logLevelStr, "error", err)
	}
	slog.SetDefault(logger)

	logger.Info("server address", "addr", *addr)
	//	logger.Info("otlp/grpc", "gprcaddr", *grpcaddr)

	sStore, err = jsondb.NewServerStorage("testdata/cluster.json")
	if err != nil {
		logger.ErrorContext(ctx, "failed to create new cluster", "error", err)
	}

	//initTargets(ctx, sStore)

	notify := notifier.NewNotifier(sStore)
	sStore = notify

	obs, err = observer.NewObserver(sStore)
	if err != nil {
		logger.ErrorContext(ctx, "failed to create endpoint storage", "error", err)
	}

	go notify.Run(obs.ManageScraper, ctx)
	go obs.ManageScraper(ctx)

	templates := templates.NewTemplateHandler()
	server := v1.NewServer(*addr, *domain, templates, logger, sStore)
	server.ServeHTTP()
}

func initTargets(ctx context.Context, sStore internal.ServerStore) error {
	sStore.Create(ctx, &model.Server{
		Name: "Ragnarok",
		Addr: "51.195.60.114:27019",
	})

	sStore.Create(ctx, &model.Server{
		Name: "LostIsland",
		Addr: "51.195.60.114:27020",
	})

	sStore.Create(ctx, &model.Server{
		Name: "Aberration",
		Addr: "51.195.60.114:27018",
	})

	sStore.Create(ctx, &model.Server{
		Name: "TheIsland",
		Addr: "51.195.60.114:27016",
	})
	return nil
}

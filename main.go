package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	v1 "github.com/led0nk/ark-clusterinfo/api/v1"
	"github.com/led0nk/ark-clusterinfo/internal"
	"github.com/led0nk/ark-clusterinfo/internal/jsondb"
	"github.com/led0nk/ark-clusterinfo/internal/model/templates"
	"github.com/led0nk/ark-clusterinfo/internal/parser"
	"github.com/led0nk/ark-clusterinfo/observer"
)

func main() {

	var (
		addr = flag.String("addr", "localhost:8080", "server port")
		//grpcaddr    = flag.String("grpcaddr", "", "grpc address, e.g. localhost:4317")
		//dbase       = flag.String("db", "file://testdata", "path to database")
		domain      = flag.String("domain", "127.0.0.1", "given domain for cookies/mail")
		logLevelStr = flag.String("loglevel", "INFO", "define the level for logs")
		parse       internal.Parser
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
		logger.Error("error parsing loglevel", "loglevel", *logLevelStr, "error", err)
	}
	slog.SetDefault(logger)

	logger.Info("server address", "addr", *addr)
	//	logger.Info("otlp/grpc", "gprcaddr", *grpcaddr)

	parse, err = parser.NewParserWithTargets("testdata/targets.json")
	if err != nil {
		logger.Error("failed to create parser", "error", err)
	}

	sStore, err = jsondb.NewServerStorage("testdata/cluster.json")
	if err != nil {
		logger.Error("failed to create new cluster", "error", err)
	}

	obs, err = observer.NewObserver(sStore, parse)
	if err != nil {
		logger.Error("failed to create endpoint storage", "error", err)
	}

	targets, err := parse.ListTargets()
	if err != nil {
		logger.Error("failed to list targets", "error", err)
	}
	go obs.InitScraper(ctx, targets)

	templates := templates.NewTemplateHandler()
	server := v1.NewServer(*addr, *domain, templates, logger, sStore, parse, obs)
	server.ServeHTTP()
}

package main

import (
	"flag"
	"log/slog"
	"os"

	v1 "github.com/led0nk/ark-clusterinfo/api/v1"
	"github.com/led0nk/ark-clusterinfo/internal"
	"github.com/led0nk/ark-clusterinfo/internal/cluster"
	"github.com/led0nk/ark-clusterinfo/internal/model/templates"
)

func main() {

	var (
		addr = flag.String("addr", "localhost:8080", "server port")
		//grpcaddr    = flag.String("grpcaddr", "", "grpc address, e.g. localhost:4317")
		//dbase       = flag.String("db", "file://testdata", "path to database")
		domain      = flag.String("domain", "127.0.0.1", "given domain for cookies/mail")
		logLevelStr = flag.String("loglevel", "INFO", "define the level for logs")
		cStore      internal.ClusterStore
		logLevel    slog.Level
	)
	flag.Parse()

	err := logLevel.UnmarshalText([]byte(*logLevelStr))
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	if err != nil {
		logger.Error("error parsing loglevel", "loglevel", *logLevelStr, "error", err)
	}
	slog.SetDefault(logger)

	logger.Info("server address", "addr", *addr)
	//	logger.Info("otlp/grpc", "gprcaddr", *grpcaddr)

	cStore, err = cluster.NewCluster("testdata/cluster.json")
	if err != nil {
		logger.Error("failed to create new cluster", "error", err)
	}

	templates := templates.NewTemplateHandler()

	server := v1.NewServer(*addr, *domain, templates, logger, cStore)
	server.ServeHTTP()

}

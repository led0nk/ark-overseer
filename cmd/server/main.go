package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	v1 "github.com/led0nk/ark-clusterinfo/api/v1"
	"github.com/led0nk/ark-clusterinfo/internal"
	blist "github.com/led0nk/ark-clusterinfo/internal/blacklist"
	"github.com/led0nk/ark-clusterinfo/internal/events"
	"github.com/led0nk/ark-clusterinfo/internal/jsondb"
	"github.com/led0nk/ark-clusterinfo/internal/notifier"
	"github.com/led0nk/ark-clusterinfo/internal/overseer"
	"github.com/led0nk/ark-clusterinfo/internal/services"
	"github.com/led0nk/ark-clusterinfo/observer"
	"github.com/led0nk/ark-clusterinfo/pkg/config"
)

func main() {

	var (
		addr   = flag.String("addr", "localhost:8080", "server port")
		db     = flag.String("db", "testdata", "path to the database")
		blpath = flag.String("blacklist", "testdata", "path to the blacklist")
		//grpcaddr    = flag.String("grpcaddr", "", "grpc address, e.g. localhost:4317")
		domain      = flag.String("domain", "127.0.0.1", "given domain for cookies/mail")
		logLevelStr = flag.String("loglevel", "INFO", "define the level for logs")
		sStore      internal.ServerStore
		obs         internal.Observer
		ovs         internal.Overseer
		blacklist   internal.Blacklist
		logLevel    slog.Level
		wg          sync.WaitGroup
		initWg      sync.WaitGroup
	)
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := logLevel.UnmarshalText([]byte(*logLevelStr))
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	if err != nil {
		logger.ErrorContext(ctx, "error parsing loglevel", "loglevel", *logLevelStr, "error", err)
		os.Exit(1)
	}
	slog.SetDefault(logger)

	logger.Info("server address", "addr", *addr)

	cfg, err := config.NewConfiguration("testdata/config.yaml")
	if err != nil {
		logger.Error("failed to create new config", "error", err)
	}

	sStore, err = jsondb.NewServerStorage(ctx, *db+"/cluster.json")
	if err != nil {
		logger.ErrorContext(ctx, "failed to create new cluster", "error", err)
		os.Exit(1)
	}

	em := events.NewEventManager()
	sm := services.NewServiceManager(em)

	notify := notifier.NewNotifier(sStore, em)
	sStore = notify

	obs, err = observer.NewObserver(ctx, sStore)
	if err != nil {
		logger.ErrorContext(ctx, "failed to create endpoint storage", "error", err)
		os.Exit(1)
	}

	blacklist, err = blist.NewBlacklist(*blpath + "/blacklist.json")
	if err != nil {
		logger.ErrorContext(ctx, "failed to create blacklist", "error", err)
		os.Exit(1)
	}

	ovs, err = overseer.NewOverseer(ctx, sStore, blacklist, em)
	if err != nil {
		logger.ErrorContext(ctx, "failed to create overseer", "error", err)
		os.Exit(1)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		logger.InfoContext(ctx, "received signal", "signal", sig)
		cancel()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		em.StartListening(ctx, sm, "serviceManager")
	}()

	//TODO: Wait group for initialization
	initWg.Add(1)
	go func() {
		defer initWg.Done()
		defer fmt.Println("initWG done")
		em.Publish(events.EventMessage{Type: "init.services", Payload: cfg})
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		em.StartListening(ctx, obs, "observer")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		em.StartListening(ctx, ovs, "overseer")
	}()

	initWg.Wait()
	go em.Publish(events.EventMessage{Type: "init"})

	server := v1.NewServer(*addr, *domain, logger, sStore, blacklist)

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := server.ServeHTTP(ctx)
		if err != nil {
			logger.ErrorContext(ctx, "failed to server http server", "error", err)
			return
		}
	}()

	wg.Wait()
	logger.InfoContext(ctx, "application stopped gracefully", "info", "shutdown")
}

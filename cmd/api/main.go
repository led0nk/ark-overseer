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

	"github.com/led0nk/ark-overseer/internal"
	blist "github.com/led0nk/ark-overseer/internal/blacklist"
	"github.com/led0nk/ark-overseer/internal/jsondb"
	"github.com/led0nk/ark-overseer/internal/notifier"
	"github.com/led0nk/ark-overseer/internal/observer"
	v1 "github.com/led0nk/ark-overseer/internal/server"
	"github.com/led0nk/ark-overseer/internal/services"
	"github.com/led0nk/ark-overseer/pkg/config"
	"github.com/led0nk/ark-overseer/pkg/events"
)

func main() {

	var (
		addr   = flag.String("addr", "localhost:8080", "server port")
		dbpath = flag.String("db", "testdata", "path to the database")
		blpath = flag.String("blacklist", "testdata", "path to the blacklist")
		//grpcaddr    = flag.String("grpcaddr", "", "grpc address, e.g. localhost:4317")
		domain      = flag.String("domain", "127.0.0.1", "given domain for cookies/mail")
		logLevelStr = flag.String("loglevel", "INFO", "define the level for logs")
		configPath  = flag.String("config", "config", "path to config-file")
		logLevel    slog.Level
		shutdownWg  sync.WaitGroup
		initWg      sync.WaitGroup
		listenerWg  sync.WaitGroup
	)
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger, err := setupLogger(logLevelStr, logLevel)
	if err != nil {
		logger.ErrorContext(ctx, "failed to setup logger", "error", err)
		os.Exit(1)
	}

	logger.Info("server address", "addr", *addr)

	eventManager := events.NewEventManager()
	serviceManager := services.NewServiceManager(eventManager, &initWg)

	database, blacklist, obs, cfg, err := initServices(ctx, dbpath, blpath, configPath, eventManager)
	if err != nil {
		logger.ErrorContext(ctx, "failed to initialize services", "error", err)
		os.Exit(1)
	}

	listenerWg.Add(2)
	startEventListeners(ctx, eventManager, &listenerWg, &shutdownWg, serviceManager, obs)
	listenerWg.Wait()

	initWg.Add(2)
	go func(config.Configuration) {
		defer initWg.Done()
		eventManager.Publish(events.EventMessage{Type: "init.services", Payload: cfg})
	}(cfg)
	initWg.Wait()

	initWg.Add(1)
	go func() {
		defer initWg.Done()
		eventManager.Publish(events.EventMessage{Type: "init"})
	}()

	server := v1.NewServer(*addr, *domain, database, blacklist, cfg)
	startHTTPServer(ctx, server, &shutdownWg)

	handleShutdown(ctx, cancel, &initWg, &shutdownWg, database)
}

func initServices(
	ctx context.Context,
	dbpath *string,
	blpath *string,
	configPath *string,
	eventManager *events.EventManager,
) (
	internal.Database,
	internal.Blacklist,
	observer.Overseer,
	config.Configuration,
	error) {
	var (
		database  internal.Database
		blacklist internal.Blacklist
		obs       observer.Overseer
		cfg       config.Configuration
	)

	database, err := jsondb.NewServerStorage(ctx, *dbpath+"/cluster.json")
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to create new server storage: %w", err)
	}

	notify := notifier.NewNotifier(database, eventManager)
	database = notify

	blacklist, err = blist.NewBlacklist(*blpath + "/blacklist.json")
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to create blacklist: %w", err)
	}

	obs, err = observer.NewObserver(ctx, database, blacklist, eventManager)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to create observer: %w", err)
	}

	cfg, err = config.NewConfiguration(*configPath+"/config.yaml", eventManager)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to create config: %w", err)
	}

	return database, blacklist, obs, cfg, nil
}

func startHTTPServer(
	ctx context.Context,
	server *v1.Server,
	shutdownWg *sync.WaitGroup,
) {
	shutdownWg.Add(1)
	go func() {
		defer shutdownWg.Done()
		server.ServeHTTP(ctx)
	}()
}

func startEventListeners(
	ctx context.Context,
	em *events.EventManager,
	listenerWg, shutdownWg *sync.WaitGroup,
	sm *services.ServiceManager,
	obs observer.Overseer,
) {
	shutdownWg.Add(1)
	go func() {
		defer shutdownWg.Done()
		em.StartListening(ctx, sm, "serviceManager", func() { listenerWg.Done() })
		fmt.Println("after listening")
	}()

	shutdownWg.Add(1)
	go func() {
		defer shutdownWg.Done()
		em.StartListening(ctx, obs, "observer", func() { listenerWg.Done() })
	}()
}

func handleShutdown(
	ctx context.Context,
	cancel context.CancelFunc,
	initWg, shutdownWg *sync.WaitGroup,
	database internal.Database,
) {
	logger := slog.Default()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		logger.InfoContext(ctx, "received signal", "signal", sig)
		initWg.Wait()
		cancel()
	}()

	shutdownWg.Wait()
	shutdownWg.Add(1)

	logger.InfoContext(ctx, "finally saving server storage", "info", "shutdown")
	err := database.Save()
	if err != nil {
		logger.ErrorContext(ctx, "failed to save server storage", "error", err)
		return
	}

	logger.InfoContext(ctx, "application stopped gracefully", "info", "shutdown")
}

func setupLogger(logLevelStr *string, logLevel slog.Level) (*slog.Logger, error) {
	err := logLevel.UnmarshalText([]byte(*logLevelStr))
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	if err != nil {
		return nil, fmt.Errorf("error parsing logLevel: %w", err)
	}
	slog.SetDefault(logger)

	return logger, nil
}

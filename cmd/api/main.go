package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/led0nk/ark-overseer/internal/blacklist"
	"github.com/led0nk/ark-overseer/internal/interfaces"
	"github.com/led0nk/ark-overseer/internal/observer"
	"github.com/led0nk/ark-overseer/internal/server"
	"github.com/led0nk/ark-overseer/internal/services"
	"github.com/led0nk/ark-overseer/internal/storage"
	"github.com/led0nk/ark-overseer/internal/storagewrapper"
	"github.com/led0nk/ark-overseer/pkg/config"
	"github.com/led0nk/ark-overseer/pkg/events"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {

	var (
		addr        = flag.String("addr", "localhost:8080", "server port")
		grpcAddr    = flag.String("grpc", "", "grpc address, e.g. localhost:4317")
		dbPath      = flag.String("db", "testdata", "path to the database")
		blPath      = flag.String("blacklist", "testdata", "path to the blacklist")
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
	logger.Info("grpc address", "grpcaddr", *grpcAddr)
	logger.Info("level for logging", "loglevel", *logLevelStr)
	logger.Info("path to database", "db", *dbPath)
	logger.Info("path to config", "config", *configPath)
	logger.Info("path to blacklist", "blacklist", *blPath)

	conn, err := setupOTEL(ctx, *grpcAddr)
	if err != nil {
		logger.ErrorContext(ctx, "failed to setup OTEL", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	eventManager := events.NewEventManager()
	serviceManager := services.NewServiceManager(eventManager, &initWg)

	database, blackList, obs, cfg, err := initServices(ctx, dbPath, blPath, configPath, eventManager)
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

	srv := server.NewServer(*addr, *domain, database, blackList, cfg)
	startHTTPServer(ctx, srv, &shutdownWg)

	handleShutdown(ctx, cancel, &initWg, &shutdownWg, database)
}

func initServices(
	ctx context.Context,
	dbpath *string,
	blpath *string,
	configPath *string,
	eventManager *events.EventManager,
) (
	interfaces.Database,
	interfaces.Blacklist,
	observer.Overseer,
	config.Configuration,
	error) {
	var (
		database  interfaces.Database
		blackList interfaces.Blacklist
		obs       observer.Overseer
		cfg       config.Configuration
	)

	database, err := storage.NewServerStorage(ctx, filepath.Join(*dbpath, "cluster.json"))
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to create new server storage: %w", err)
	}

	storageWrapper := storagewrapper.NewStorageWrapper(database, eventManager)
	database = storageWrapper

	blackList, err = blacklist.NewBlacklist(filepath.Join(*blpath, "blacklist.json"))
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to create blacklist: %w", err)
	}

	obs, err = observer.NewObserver(ctx, database, blackList, eventManager)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to create observer: %w", err)
	}

	cfg, err = config.NewConfiguration(filepath.Join(*configPath, "config.yaml"), eventManager)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to create config: %w", err)
	}

	return database, blackList, obs, cfg, nil
}

func startHTTPServer(
	ctx context.Context,
	server *server.Server,
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
	database interfaces.Database,
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

func setupOTEL(ctx context.Context, grpcaddr string) (conn *grpc.ClientConn, err error) {
	if grpcaddr != "" {
		grpcOptions := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
		conn, err := grpc.NewClient(grpcaddr, grpcOptions...)
		if err != nil {
			return nil, fmt.Errorf("failed to create grpc client: %w", err)
		}

		oteltraceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
		if err != nil {
			return nil, fmt.Errorf("failed to create otlp trace exporter: %w", err)
		}
		tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(oteltraceExporter))
		otel.SetTracerProvider(tp)

		otelmetricsExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(conn))
		if err != nil {
			return nil, fmt.Errorf("failed to create otlp metrics exporter: %w", err)
		}
		mp := metric.NewMeterProvider(metric.WithReader(metric.NewPeriodicReader(otelmetricsExporter)))
		otel.SetMeterProvider(mp)
	}
	return conn, nil
}

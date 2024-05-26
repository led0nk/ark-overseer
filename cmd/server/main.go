package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	v1 "github.com/led0nk/ark-clusterinfo/api/v1"
	"github.com/led0nk/ark-clusterinfo/internal"
	blist "github.com/led0nk/ark-clusterinfo/internal/blacklist"
	"github.com/led0nk/ark-clusterinfo/internal/jsondb"
	"github.com/led0nk/ark-clusterinfo/internal/notifier"
	"github.com/led0nk/ark-clusterinfo/internal/notifier/services/discord"
	"github.com/led0nk/ark-clusterinfo/internal/overseer"
	"github.com/led0nk/ark-clusterinfo/internal/utilities"
	"github.com/led0nk/ark-clusterinfo/observer"
)

func main() {

	var (
		addr   = flag.String("addr", "localhost:8080", "server port")
		db     = flag.String("db", "testdata", "path to the database")
		blpath = flag.String("blacklist", "testdata", "path to the blacklist")
		//grpcaddr    = flag.String("grpcaddr", "", "grpc address, e.g. localhost:4317")
		envStr      = flag.String("env", "testdata/.env", "path to .env file")
		domain      = flag.String("domain", "127.0.0.1", "given domain for cookies/mail")
		logLevelStr = flag.String("loglevel", "INFO", "define the level for logs")
		sStore      internal.ServerStore
		obs         internal.Observer
		ovs         internal.Overseer
		blacklist   internal.Blacklist
		messaging   internal.Notification
		logLevel    slog.Level
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
	//	logger.Info("otlp/grpc", "gprcaddr", *grpcaddr)

	envmap, err := utilities.LoadEnv(logger, *envStr)
	if err != nil {
		logger.Error("failed to load .env variables", "error", err)
	}

	sStore, err = jsondb.NewServerStorage(*db + "/cluster.json")
	if err != nil {
		logger.ErrorContext(ctx, "failed to create new cluster", "error", err)
		os.Exit(1)
	}

	notify := notifier.NewNotifier(sStore)
	sStore = notify

	obs, err = observer.NewObserver(sStore)
	if err != nil {
		logger.ErrorContext(ctx, "failed to create endpoint storage", "error", err)
		os.Exit(1)
	}

	blacklist, err = blist.NewBlacklist(*blpath + "/blacklist.json")
	if err != nil {
		logger.ErrorContext(ctx, "failed to create blacklist", "error", err)
		os.Exit(1)
	}

	//initBlacklist(ctx, blacklist, logger)

	messaging, err = discord.NewDiscordNotifier(envmap["DCTOKEN"])
	if err != nil {
		logger.ErrorContext(ctx, "failed to create notification service", "error", err)
		os.Exit(1)
	}

	ovs, err = overseer.NewOverseer(ctx, sStore, blacklist, messaging)
	if err != nil {
		logger.ErrorContext(ctx, "failed to create overseer", "error", err)
		os.Exit(1)
	}

	go notify.Run(obs.ManageScraper, ovs.ManageScanner, ctx)
	go obs.ManageScraper(ctx)
	go ovs.ManageScanner(ctx)

	server := v1.NewServer(*addr, *domain, logger, sStore)
	server.ServeHTTP()
}

// NOTE: just to initialize first Targets
//func initTargets(ctx context.Context, sStore internal.ServerStore, logger *slog.Logger) error {
//	sStore.Create(ctx, &model.Server{
//		Name: "Ragnarok",
//		Addr: "51.195.60.114:27019",
//	})
//
//	sStore.Create(ctx, &model.Server{
//		Name: "LostIsland",
//		Addr: "51.195.60.114:27020",
//	})
//
//	sStore.Create(ctx, &model.Server{
//		Name: "Aberration",
//		Addr: "51.195.60.114:27018",
//	})
//
//	sStore.Create(ctx, &model.Server{
//		Name: "TheIsland",
//		Addr: "51.195.60.114:27016",
//	})
//	return nil
//}
//
//func initBlacklist(ctx context.Context, blacklist internal.Blacklist, logger *slog.Logger) error {
//	_, err := blacklist.Create(ctx, &model.Players{
//		Name: "Fadem",
//	})
//	if err != nil {
//		logger.ErrorContext(ctx, "failed to create blacklist entry", "error", err)
//		return err
//	}
//	_, err = blacklist.Create(ctx, &model.Players{
//		Name: "FisherSpider",
//	})
//	if err != nil {
//		logger.ErrorContext(ctx, "failed to create blacklist entry", "error", err)
//		return err
//	}
//	_, err = blacklist.Create(ctx, &model.Players{
//		Name: "Hermes Headquart...",
//	})
//	if err != nil {
//		logger.ErrorContext(ctx, "failed to create blacklist entry", "error", err)
//		return err
//	}
//	return nil
//}

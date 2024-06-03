package v1

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/led0nk/ark-overseer/internal"
	"github.com/led0nk/ark-overseer/pkg/config"
	sloghttp "github.com/samber/slog-http"
)

type Server struct {
	addr      string
	domain    string
	logger    *slog.Logger
	sStore    internal.ServerStore
	blacklist internal.Blacklist
	config    config.Configuration
}

func NewServer(
	address string,
	domain string,
	logger *slog.Logger,
	sStore internal.ServerStore,
	blacklist internal.Blacklist,
	config config.Configuration,
) *Server {
	return &Server{
		addr:      address,
		domain:    domain,
		logger:    slog.Default().WithGroup("http"),
		sStore:    sStore,
		blacklist: blacklist,
		config:    config,
	}
}

func (s *Server) ServeHTTP(ctx context.Context) error {
	r := http.NewServeMux()

	slogmw := sloghttp.NewWithConfig(
		s.logger, sloghttp.Config{
			DefaultLevel:     slog.LevelInfo,
			ClientErrorLevel: slog.LevelWarn,
			ServerErrorLevel: slog.LevelError,
			WithUserAgent:    true,
		},
	)

	r.Handle("GET /", http.HandlerFunc(s.mainPage))
	r.Handle("POST /", http.HandlerFunc(s.showServerInput))
	r.Handle("PUT /", http.HandlerFunc(s.addServer))
	r.Handle("POST /{ID}", http.HandlerFunc(s.showPlayers))
	r.Handle("DELETE /{ID}", http.HandlerFunc(s.deleteServer))
	r.Handle("GET /serverdata/{ID}", http.HandlerFunc(s.sseServerUpdate))
	r.Handle("GET /serverdata/{ID}/players", http.HandlerFunc(s.updatePlayerInfo))
	r.Handle("GET /settings", http.HandlerFunc(s.setupPage))
	r.Handle("POST /settings", http.HandlerFunc(s.saveChanges))
	r.Handle("GET /blacklist", http.HandlerFunc(s.blacklistPage))
	r.Handle("POST /blacklist", http.HandlerFunc(s.blacklistAdd))
	r.Handle("DELETE /blacklist/{ID}", http.HandlerFunc(s.blacklistDelete))

	s.logger.Info("listen and serve", "addr", s.addr)

	srv := http.Server{
		Addr:    s.addr,
		Handler: slogmw(r),
	}

	go func() {
		select {
		case <-ctx.Done():
			shutdownCtx, cancel := context.WithCancel(ctx)
			defer cancel()
			err := srv.Shutdown(shutdownCtx)
			if err != nil {
				s.logger.ErrorContext(ctx, "failed to shutdown http server", "error", err)
				return
			}
		default:
			err := srv.ListenAndServe()
			if err != nil {
				s.logger.Error("error during listen and serve", "error", err)
			}
		}
	}()

	<-ctx.Done()
	s.logger.InfoContext(ctx, "server shutdown completed", "info", "shutdown")
	return nil
}

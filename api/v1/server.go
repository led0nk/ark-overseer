package v1

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/led0nk/ark-clusterinfo/internal"
	"github.com/led0nk/ark-clusterinfo/internal/model/templates"
	sloghttp "github.com/samber/slog-http"
)

type Server struct {
	addr      string
	domain    string
	templates *templates.TemplateHandler
	logger    *slog.Logger
	cStore    internal.ClusterStore
}

func NewServer(
	address string,
	domain string,
	templates *templates.TemplateHandler,
	logger *slog.Logger,
	cStore internal.ClusterStore,
) *Server {
	return &Server{
		addr:      address,
		domain:    domain,
		templates: templates,
		logger:    slog.Default().WithGroup("http"),
		cStore:    cStore,
	}
}

func (s *Server) ServeHTTP() {
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
	r.Handle("POST /{ID}", http.HandlerFunc(s.showPlayers))
	r.Handle("DELETE /{ID}", http.HandlerFunc(s.deleteServer))
	r.Handle("GET /{ID}", http.HandlerFunc(s.updatePlayers))

	s.logger.Info("listen and serve", "addr", s.addr)

	srv := http.Server{
		Addr:    s.addr,
		Handler: slogmw(r),
	}

	err := srv.ListenAndServe()
	if err != nil {
		s.logger.Error("error during listen and serve", "error", err)
		os.Exit(1)
	}
}

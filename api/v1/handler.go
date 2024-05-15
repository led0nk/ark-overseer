package v1

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/led0nk/ark-clusterinfo/internal/model"
)

func (s *Server) mainPage(w http.ResponseWriter, r *http.Request) {
	var (
		serverList []*model.Server
		err        error
	)

	ctx := r.Context()

	serverList, err = s.cStore.GetServerInfo()
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get server info", "error", err)
	}

	err = s.templates.TmplHome.Execute(w, serverList)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to execute template", "error", err)
		return
	}
}

func (s *Server) showPlayers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := uuid.Parse(r.PathValue("ID"))
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to parse uuid", "error", err)
		return
	}
	server, err := s.cStore.GetServerByID(id)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get server", "error", err)
		return
	}
	err = s.templates.TmplPlayer.ExecuteTemplate(w, "player", server.PlayersInfo)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to execute template", "error", err)
		return
	}
}

func (s *Server) deleteServer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("ID")
	fmt.Println(id)
}

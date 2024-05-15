package v1

import (
	"net/http"

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
	}
}

func (s *Server) showPlayers(w http.ResponseWriter, r *http.Request) {

}

package v1

import (
	"context"
	"fmt"
	"html"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/led0nk/ark-clusterinfo/internal/model"
)

func (s *Server) mainPage(w http.ResponseWriter, r *http.Request) {
	var (
		serverList []*model.Server
		err        error
	)

	ctx := r.Context()

	serverList, err = s.sStore.List(ctx)
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
	server, err := s.sStore.GetByID(ctx, id)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get server", "error", err)
		return
	}
	err = s.templates.TmplBlocks.ExecuteTemplate(w, "player", server)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to execute template", "error", err)
		return
	}
}

func (s *Server) updatePlayerCounter(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	ctx := r.Context()
	dataCh := make(chan any)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func(ctx context.Context) {
		for data := range dataCh {
			select {
			case <-ctx.Done():
				return
			default:
				fmt.Fprintf(w, "data: %s\n\n", data)
				w.(http.Flusher).Flush()
			}
		}
	}(ctx)

	for {
		srv, err := s.sStore.GetByID(ctx, uuid.MustParse(r.PathValue("ID")))
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to get server", "error", err)
		}
		playerInfo := strconv.Itoa(srv.ServerInfo.Players) + "/" + strconv.Itoa(srv.ServerInfo.MaxPlayers)
		dataCh <- playerInfo
		time.Sleep(5 * time.Second)
	}
}

func (s *Server) deleteServer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := uuid.Parse(r.PathValue("ID"))
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to parse uuid", "error", err)
		return
	}

	err = s.sStore.Delete(ctx, id)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to delete target", "error", err)
		return
	}
}

func (s *Server) showServerInput(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := s.templates.TmplBlocks.ExecuteTemplate(w, "server", nil)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to execute template", "error", err)
		return
	}
}

func (s *Server) addServer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := r.ParseForm()
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to parse form", "error", err)
		return
	}

	newServer := &model.Server{
		Name: html.EscapeString(r.FormValue("servername")),
		Addr: html.EscapeString(r.FormValue("address")),
	}
	_, err = s.sStore.Create(ctx, newServer)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create server", "error", err)
	}

	time.Sleep(1 * time.Second)

	server, err := s.sStore.GetByID(ctx, newServer.ID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get server", "error", err)
		return
	}

	err = s.templates.TmplBlocks.ExecuteTemplate(w, "newServer", server)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to execute template", "error", err)
		return
	}
}

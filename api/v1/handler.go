package v1

import (
	"context"
	"fmt"
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

	serverList, err = s.sStore.ListServer()
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
	server, err := s.sStore.GetServerByID(id)
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

	_, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		for data := range dataCh {
			fmt.Fprintf(w, "data: %s\n\n", data)
			w.(http.Flusher).Flush()
		}
	}()

	for {
		srv, err := s.sStore.GetServerByID(uuid.MustParse(r.PathValue("ID")))
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to get server", "error", err)
		}
		playerInfo := strconv.Itoa(srv.ServerInfo.Players) + "/" + strconv.Itoa(srv.ServerInfo.MaxPlayers)
		dataCh <- playerInfo
		time.Sleep(5 * time.Second)
	}
}

func (s *Server) deleteServer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("ID")
	fmt.Println(id)
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

	newServer := &model.Server{
		Name: r.FormValue("servername"),
		Addr: r.FormValue("address"),
	}
	_, err := s.sStore.CreateOrUpdateServer(newServer)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create server", "error", err)
	}
}

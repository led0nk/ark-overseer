package v1

import (
	"bytes"
	"context"
	"fmt"
	"html"
	"net/http"
	"strconv"
	"time"

	"github.com/a-h/templ"
	"github.com/google/uuid"
	"github.com/led0nk/ark-clusterinfo/internal/model"
	"github.com/led0nk/ark-clusterinfo/internal/model/templates/layout"
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

	templ.Handler(layout.Table(serverList))

	err = layout.Render(ctx, w, layout.Main(serverList))
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to render templ", "error", err)
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
	err = layout.Render(ctx, w, layout.PlayerTable(server))
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to render templ", "error", err)
		return
	}
}

func (s *Server) updatePlayerCounter(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	ctx := r.Context()
	dataCh := make(chan string)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case data := <-dataCh:
				fmt.Fprintf(w, "data: %s\n\n", data)
				w.(http.Flusher).Flush()
			}
		}
	}(ctx)

	go func() {
		defer close(dataCh)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				srv, err := s.sStore.GetByID(ctx, uuid.MustParse(r.PathValue("ID")))
				if err != nil {
					s.logger.ErrorContext(ctx, "failed to get server", "error", err)
				}
				playerInfo := strconv.Itoa(srv.ServerInfo.Players) + "/" + strconv.Itoa(srv.ServerInfo.MaxPlayers)
				dataCh <- playerInfo
				time.Sleep(5 * time.Second)
			}
		}
	}()
	<-ctx.Done()
}

func (s *Server) updatePlayerInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	ctx := r.Context()
	dataCh := make(chan *model.Server)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func(ctx context.Context) {
		defer close(dataCh)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				srv, err := s.sStore.GetByID(ctx, uuid.MustParse(r.PathValue("ID")))
				if err != nil {
					s.logger.ErrorContext(ctx, "failed to get server", "error", err)
				}
				dataCh <- srv
				time.Sleep(1 * time.Second)
			}
		}
	}(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case data := <-dataCh:
				var buffer bytes.Buffer
				for _, player := range data.PlayersInfo.Players {
					playerRow := fmt.Sprintf(`<tr class="hover:bg-gray-50"><td class="px-6 py-4"><div class="font-medium text-gray-700">%s</div></td><td class="px-6 py-4"><div class="font-medium text-gray-700">%s</div></td></tr>`, player.Name, player.Duration)
					buffer.WriteString(playerRow)
				}
				fmt.Fprintf(w, "data: %s\n\n", buffer.String())
				w.(http.Flusher).Flush()
			}
		}
	}()
	<-ctx.Done()
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

	err := layout.Render(ctx, w, layout.NewServerInput())
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to render templ", "error", err)
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

	_, err = s.sStore.GetByID(ctx, newServer.ID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get server", "error", err)
		return
	}
}

func (s *Server) blacklistPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	blacklist := s.blacklist.List(ctx)
	err := layout.Render(ctx, w, layout.Blacklist(blacklist))
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to render templ", "error", err)
		return
	}
}

func (s *Server) setupPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := layout.Render(ctx, w, layout.Setup())
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to render templ", "error", err)
		return
	}
}

func (s *Server) blacklistAdd(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := r.ParseForm()
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to parse form", "error", err)
		return
	}
	_, err = s.blacklist.Create(ctx, &model.BlacklistPlayers{
		Name: r.FormValue("blacklistPlayer"),
	})

	if err != nil {
		s.logger.ErrorContext(ctx, "failed to add player to blacklist", "error", err)
		return
	}

	http.Redirect(w, r, "/blacklist", http.StatusFound)
}

func (s *Server) blacklistDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	fmt.Println(r.PathValue("ID"))

	id, err := uuid.Parse(r.PathValue("ID"))
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to parse uuid", "error", err)
		return
	}
	err = s.blacklist.Delete(ctx, id)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to delete from blacklist", "error", err)
		return
	}
}

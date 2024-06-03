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
	"github.com/led0nk/ark-overseer/cmd/web"
	"github.com/led0nk/ark-overseer/internal/model"
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

	templ.Handler(web.Table(serverList))

	err = web.Render(ctx, w, web.Main(serverList))
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
	err = web.Render(ctx, w, web.PlayerTable(server))
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to render templ", "error", err)
		return
	}
}

func (s *Server) sseServerUpdate(w http.ResponseWriter, r *http.Request) {
	type event struct {
		Type string
		Data any
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	ctx := r.Context()
	dataCh := make(chan event)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-dataCh:
				fmt.Fprintf(w, "event: %s\n", event.Type)
				fmt.Fprintf(w, "data: %s\n\n", event.Data)
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
				status := `<span class="inline-flex items-center gap-1 rounded-full dark:bg-[#0D1117] bg-green-50 px-2 py-1 text-xs font-semibold text-green-600"><span class="h-1.5 w-1.5 rounded-full bg-green-600"></span>online</span>`
				if !srv.Status {
					srv.ServerInfo.Players = 0
					status = `<span class="inline-flex items-center gap-1 rounded-full dark:bg-[#0D1117] bg-red-50 px-2 py-1 text-xs font-semibold text-red-600"><span class="h-1.5 w-1.5 rounded-full bg-red-600"></span>offline</span>`
				}
				playerInfo := strconv.Itoa(srv.ServerInfo.Players) + "/" + strconv.Itoa(srv.ServerInfo.MaxPlayers)
				dataCh <- event{Type: "PlayerCounter", Data: playerInfo}
				dataCh <- event{Type: "ServerStatus", Data: status}
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
					playerRow := fmt.Sprintf(`<tr class="hover:bg-gray-50 dark:hover:bg-[#21262d]/50"><td class="px-6 py-4"><div class="font-medium text-gray-700 dark:text-gray-200">%s</div></td><td class="px-6 py-4"><div class="font-medium text-gray-700 dark:text-gray-200">%s</div></td></tr>`, player.Name, player.Duration)
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

	err := web.Render(ctx, w, web.NewServerInput())
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
	err := web.Render(ctx, w, web.Blacklist(blacklist))
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to render templ", "error", err)
		return
	}
}

func (s *Server) setupPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := web.Render(ctx, w, web.Setup())
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to render templ", "error", err)
		return
	}
}

// NOTE: possible implementation of other services
func (s *Server) saveChanges(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	err := r.ParseForm()
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to parse form", "error", err)
		return
	}

	sectionMap := make(map[interface{}]interface{})
	sectionMap["token"] = r.FormValue("token")
	sectionMap["channelID"] = r.FormValue("channelID")

	err = s.config.Update("notification-service", "discord", sectionMap)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to update config", "error", err)
		return
	}

	http.Redirect(w, r, "/settings", http.StatusFound)
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

	newBlacklist := s.blacklist.List(ctx)

	err = web.Render(ctx, w, web.BlacklistTable(newBlacklist))
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to render templ", "error", err)
		return
	}
}

func (s *Server) blacklistDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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

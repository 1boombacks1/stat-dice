package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/1boombacks1/stat_dice/appctx"
	"github.com/1boombacks1/stat_dice/resources"
	"github.com/1boombacks1/stat_dice/server/handlers"
	"github.com/1boombacks1/stat_dice/server/middlewares"
	"github.com/go-chi/chi/v5"
)

type HTTPServer struct {
	ctx    *appctx.AppCtx
	router chi.Router
	http.Server
}

type HTTPHandler func(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request)

func NewHTTPServer(ctx *appctx.AppCtx) *HTTPServer {
	mux := newRouter(ctx)
	server := &HTTPServer{
		router: mux,
		Server: http.Server{
			Addr:    ctx.Config().Host + ":" + strconv.FormatUint(uint64(ctx.Config().Port), 10),
			Handler: mux,
			BaseContext: func(net.Listener) context.Context {
				return ctx
			},
		},
		ctx: ctx,
	}

	server.initRoutes()

	return server
}

func (s *HTTPServer) initRoutes() {
	s.router.Get("/css/*", http.FileServer(http.FS(resources.CSS)).ServeHTTP)
	s.router.Get("/fonts/*", http.FileServer(http.FS(resources.Fonts)).ServeHTTP)
	s.router.Get("/images/*", http.FileServer(http.FS(resources.Images)).ServeHTTP)
	s.router.Get("/js/*", http.FileServer(http.FS(resources.JS)).ServeHTTP)

	{
		s.DefineRoute(s.router, "GET", "/", handlers.AuthPage)
		s.DefineRoute(s.router, "GET", "/login", handlers.AuthPage)

		s.router.Route("/counter", func(appR chi.Router) {
			appR.Use(middlewares.Auth)
			s.DefineRoute(appR, "GET", "/", handlers.Index)
			s.DefineRoute(appR, "GET", "/logout", handlers.Logout)

			{
				appR.Route("/lobby/{id}", func(lobbyR chi.Router) {
					s.DefineRoute(lobbyR, "GET", "/players", handlers.GetLobbyPlayers)
					s.DefineRoute(lobbyR, "GET", "/status", handlers.GetLobbyStatus)

					lobbyR.Group(func(playersR chi.Router) {
						playersR.Use(middlewares.CheckAccessLobby)
						s.DefineRoute(playersR, "GET", "/", handlers.LobbyPage)
						s.DefineRoute(playersR, "POST", "/leave", handlers.LeaveLobby)
						s.DefineRoute(playersR, "POST", "/win", handlers.WinMatch)
						s.DefineRoute(playersR, "POST", "/lose", handlers.LoseMatch)
					})

					lobbyR.Group(func(hostR chi.Router) {
						hostR.Use(middlewares.CheckIsHost)
						s.DefineRoute(hostR, "POST", "/start", handlers.StartLobby)
						s.DefineRoute(hostR, "POST", "/stop", handlers.StopLobby)
						s.DefineRoute(hostR, "DELETE", "/", handlers.CancelLobby)
					})
				})
			}

			s.DefineRoute(appR, "GET", "/create-lobby", handlers.CreateLobbyContent)
			s.DefineRoute(appR, "GET", "/find-lobbies", handlers.FindLobbiesContent)
			s.DefineRoute(appR, "GET", "/leaderboard", handlers.LeaderboardContent)
			s.DefineRoute(appR, "GET", "/completed-lobbies", handlers.GetCompletedLobbies)

			s.DefineRoute(appR, "POST", "/create-lobby", handlers.CreateLobby)
			s.DefineRoute(appR, "GET", "/open-lobbies", handlers.GetOpenLobbies)
			s.DefineRoute(appR, "GET", "/{id}/join", handlers.JoinLobby)
			s.DefineRoute(appR, "GET", "/get-players-win-stats", handlers.GetWinStats)
			s.DefineRoute(appR, "GET", "/get-players-lose-stats", handlers.GetLoseStats)
			s.DefineRoute(appR, "GET", "/get-players-total-stats", handlers.GetTotalStats)
		})
	}

	{
		s.router.Route("/auth", func(r chi.Router) {
			s.DefineRoute(r, "POST", "/login", handlers.Login)
			s.DefineRoute(r, "POST", "/register", handlers.Registration)
		})
	}
}

func (s *HTTPServer) DefineRoute(r chi.Router, method, path string, callback HTTPHandler) {
	r.MethodFunc(method, path, func(w http.ResponseWriter, r *http.Request) {
		callback(s.ctx, w, r)
	})
}

func (hs *HTTPServer) Start() error {
	err := hs.Server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (hs *HTTPServer) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := hs.Server.Shutdown(ctx); err != nil {
		hs.ctx.Warn().AnErr("failed to gracefully shutdown http server", err)

		err := hs.Server.Close()
		if err != nil {
			hs.ctx.Warn().AnErr("failed to shutdown http server", err)
		}
	}
}

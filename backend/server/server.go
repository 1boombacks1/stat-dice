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
			Addr:    ctx.Config().Address + ":" + strconv.FormatUint(uint64(ctx.Config().Port), 10),
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
		s.DefineRoute(s.router, "GET", "/", handlers.Auth)
		s.DefineRoute(s.router, "GET", "/login", handlers.Auth)
	}

	{
		s.router.Route("/auth", func(r chi.Router) {
			s.DefineRoute(r, "POST", "/login", handlers.Login)
			s.DefineRoute(r, "POST", "/register", handlers.Registration)
		})

		s.router.Route("/api", func(r chi.Router) {
			r.Use(middlewares.Auth)
		})
	}
}

func (s *HTTPServer) DefineRoute(rt chi.Router, method, path string, callback HTTPHandler) {
	rt.MethodFunc(method, path, func(w http.ResponseWriter, r *http.Request) {
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

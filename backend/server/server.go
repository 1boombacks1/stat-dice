package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/1boombacks1/stat_dice/appctx"
	"github.com/1boombacks1/stat_dice/server/handlers"
	"github.com/go-chi/chi/v5"
)

type HTTPServer struct {
	ctx    *appctx.AppCtx
	router chi.Router
	http.Server
}

// type HTTPHandler func(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request)

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
	s.router.Route("/auth", func(r chi.Router) {
		r.Post("/login", handlers.Login)
	})

	s.router.Route("/api", func(r chi.Router) {
		r.Use()
	})
}

// func (s *HTTPServer) DefineRoute(rt chi.Router, method, path string, callback HTTPHandler) {
// 	rt.MethodFunc(method, path, func(w http.ResponseWriter, r *http.Request) {
// 		callback(s.ctx, w, r)
// 	})
// }

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

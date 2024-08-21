package server

import (
	"errors"
	"net/http"
	"time"

	"github.com/1boombacks1/stat_dice/appctx"
	httpErrors "github.com/1boombacks1/stat_dice/server/http_errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

func newRouter(ctx *appctx.AppCtx) *chi.Mux {
	r := chi.NewRouter()
	r.Use(
		middleware.RequestID,
		middleware.RealIP,
		middleware.CleanPath,
	)

	if ctx.Config().LogRequests {
		r.Use(newLogger(ctx))
	}

	// r.Use(recoverer(ctx))

	setDefaultHandlers(r)
	return r
}

func newLogger(ctx *appctx.AppCtx) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			t1 := time.Now()

			defer func() {
				ctx.Log().
					Str("method", r.Method).
					Str("proto", r.Proto).
					Str("path", r.URL.Host).
					Str("host", r.Host).
					Str("addr", r.RemoteAddr).
					Str("id", middleware.GetReqID(ctx)).
					Dur("time", time.Since(t1)).
					Int("status", ww.Status()).
					Int("sizeze", ww.BytesWritten())
			}()

			next.ServeHTTP(ww, r)
		}

		return http.HandlerFunc(fn)
	}
}

// func recoverer(ctx *appctx.AppCtx) func(next http.Handler) http.Handler {
// 	return func(next http.Handler) http.Handler {
// 		fn := func(w http.ResponseWriter, r *http.Request) {
// 			defer func() {
// 				if e := recover(); e != nil {
// 					if e == http.ErrAbortHandler {
// 						panic(e)
// 					}

// 					switch e2 := e.(type) {
// 					case error:
// 						ctx.Warn().AnErr("internal error", e2)
// 					case string:
// 						ctx.Warn().Str("internal error", e2)
// 					default:
// 						ctx.Warn().Msg("internal error")
// 					}

// 					if r.Context().Err() == nil {
// 						httpErr.HTTPInternalServerError(r.Context().Err()).Render(w, r)
// 					}
// 				}
// 			}()

// 			next.ServeHTTP(w, r)
// 		}

//			return http.HandlerFunc(fn)
//		}
//	}
func setDefaultHandlers(r *chi.Mux) {
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		render.Render(w, r, httpErrors.HTTPNotFound(errors.New("route not found")))
	})

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		render.Render(w, r, httpErrors.HTTPMethodNotAllowed(errors.New("method not allowed")))
	})
}

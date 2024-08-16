package middlewares

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/1boombacks1/stat_dice/appctx"
	"github.com/1boombacks1/stat_dice/models"
	httpErrors "github.com/1boombacks1/stat_dice/server/http_errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func CheckAccessLobby(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := appctx.FromContext(r.Context())

		lobbyID := chi.URLParam(r, "id")
		user := models.GetUserFromContext(ctx)

		if user.Match == nil || user.Match.Lobby.GetID() != lobbyID {
			render.Render(w, r, httpErrors.ErrUnauthorized(errors.New("ты в сделку не входил. Отказано в доступе к лобби")))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func CheckIsHost(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := appctx.FromContext(r.Context())
		user := models.GetUserFromContext(ctx)

		if !user.Match.IsHost {
			render.Render(w, r,
				httpErrors.ErrUnauthorized(fmt.Errorf("user '%s' not host", user.Name)).
					WithExplanation("Ты не хост, отказано в доступе."))
			// httpErrors.ErrBadRequest(fmt.Errorf("user '%s' not host", user.Name)).SetTitle("Start Lobby Error").
			// 	WithExplanation("You are not host, only host can start lobby").
			// 	Execute(w, httpErrors.AppErrTmplName, ctx.Error())
			return
		}

		next.ServeHTTP(w, r)
	})
}

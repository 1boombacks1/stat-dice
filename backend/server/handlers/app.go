package handlers

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/1boombacks1/stat_dice/appctx"
	"github.com/1boombacks1/stat_dice/models"
	httpErrors "github.com/1boombacks1/stat_dice/server/http_errors"
	"github.com/1boombacks1/stat_dice/server/templates"
	"github.com/go-chi/render"
)

var appTmpl *template.Template

func MainPage(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	games, err := models.GetGames(ctx)
	if err != nil {
		render.Render(w, r, httpErrors.ErrInternalServer(fmt.Errorf("getting games: %w", err)))
	}

	if len(games) > 0 {
		http.SetCookie(w, &http.Cookie{
			Name:  "game-id",
			Value: games[0].GetID(),
			// Secure:   true, при HTTPS - включить
			HttpOnly: true,
		})
	}

	user := models.GetUserFromContext(appctx.FromContext(r.Context()))

	if err := appTmpl.ExecuteTemplate(w, "main-page",
		struct {
			WindowName string
			Games      []models.Game
			Username   string
		}{
			WindowName: "Find Lobby",
			Games:      games,

			Username: user.Name,
		},
	); err != nil {
		panic(fmt.Errorf("rendering main page: %w", err))
	}
}

func Logout(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Path:   "/counter",
		Name:   "token",
		Value:  "",
		MaxAge: -1,
	})
	http.SetCookie(w, &http.Cookie{
		Path:   "/",
		Name:   "game-id",
		Value:  "",
		MaxAge: -1,
	})

	redirectTo(w, "/")
}

func redirectTo(w http.ResponseWriter, path string) {
	w.Header().Set("HX-Redirect", path)
}

func redirectToMainPage(w http.ResponseWriter) {
	w.Header().Set("HX-Redirect", "/counter")
}

func refreshPage(w http.ResponseWriter) {
	w.Header().Set("HX-Refresh", "true")
}

func FindLobbies(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	if err := appTmpl.ExecuteTemplate(w, "find-lobbies", nil); err != nil {
		panic(fmt.Errorf("rendering main page: %w", err))
	}
}

func init() {
	appTmpl = template.Must(template.ParseFS(templates.Main,
		"main/root/*.html",
		"main/pages/*.html",
		"main/sections/*.html",
		"main/components/*.html",
	))
}

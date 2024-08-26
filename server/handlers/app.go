package handlers

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"

	"github.com/1boombacks1/stat_dice/appctx"
	"github.com/1boombacks1/stat_dice/models"
	httpErrors "github.com/1boombacks1/stat_dice/server/http_errors"
	"github.com/1boombacks1/stat_dice/server/templates"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

var baseTmpl *template.Template

type pageInfo struct {
	AppName         string
	CurrentGameName string
	WindowName      string
	SidebarInfo     sidebarInfo
}
type sidebarInfo struct {
	Username      string
	CurrentGameID string
	Games         []models.Game
}

func Index(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	games, err := models.GetGames(ctx)
	if err != nil {
		render.Render(w, r, httpErrors.ErrInternalServer(fmt.Errorf("getting games: %w", err)))
	}

	var currentGame models.Game
	gameID, err := getGameIDFromCookie(r)
	if err != nil {
		if len(games) > 0 {
			currentGame = games[0]
			http.SetCookie(w, &http.Cookie{
				Name:     "game-id",
				Value:    currentGame.GetID(),
				HttpOnly: true,
			})
		} else {
			render.Render(w, r, httpErrors.ErrInternalServer(errors.New("админ забыл добавить игры. Напишите сюда t.me/boombacks")))
			return
		}
	} else {
		currentGame, err = games.GetByID(*gameID)
		if err != nil {
			render.Render(w, r, httpErrors.ErrInternalServer(fmt.Errorf("not found game '%s'", gameID)))
			ctx.Error().Str("game-id", gameID.String()).Msg("not found game")
			return
		}
	}

	index, err := prepareIndexTemplate(ctx.Config().StartPage)
	if err != nil {
		render.Render(w, r, httpErrors.ErrInternalServer(fmt.Errorf("preparing index template: %w", err)).WithLog(ctx.Error()))
		return
	}

	user := models.GetUserFromContext(appctx.FromContext(r.Context()))

	if err := index.ExecuteTemplate(w, "index",
		struct {
			PageInfo pageInfo
			Match    *models.Match
		}{
			PageInfo: pageInfo{
				AppName:         ctx.Config().AppName,
				WindowName:      ctx.Config().AppName,
				CurrentGameName: currentGame.Name,
				SidebarInfo: sidebarInfo{
					Username:      user.Name,
					CurrentGameID: currentGame.GetID(),
					Games:         games,
				},
			},
			Match: user.Match,
		},
	); err != nil {
		render.Render(w, r, httpErrors.ErrInternalServer(fmt.Errorf("rendering main page: %w", err)).WithLog(ctx.Error()))
	}
}

func prepareIndexTemplate(startPage templates.PageContent) (*template.Template, error) {
	startTmpl, err := startPage.GetTemplate(nil)
	if err != nil {
		return nil, fmt.Errorf("getting template: %w", err)
	}

	tmpl, err := baseTmpl.Clone()
	if err != nil {
		return nil, fmt.Errorf("cloning template: %w", err)
	}

	_, err = tmpl.AddParseTree("content", startTmpl.Lookup("content").Tree)
	if err != nil {
		return nil, fmt.Errorf("adding tree: %w", err)
	}

	return tmpl, nil
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

func getGameIDFromCookie(r *http.Request) (*uuid.UUID, error) {
	cookie, err := r.Cookie("game-id")
	if err != nil {
		return nil, fmt.Errorf("getting game id cookie: %w", err)
	}
	gameID, err := uuid.Parse(cookie.Value)
	if err != nil {
		return nil, fmt.Errorf("parsing game id cookie: %w", err)
	}
	return &gameID, nil
}

func init() {
	baseTmpl = template.Must(template.ParseFS(templates.Main,
		"main/base.html",
		"main/root/*.html",
		"main/components/*.html",
		// "main/pages/*.html",
		// "main/sections/*.html",
	))
}

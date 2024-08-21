package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/1boombacks1/stat_dice/appctx"
	"github.com/1boombacks1/stat_dice/models"
	httpErrors "github.com/1boombacks1/stat_dice/server/http_errors"
	"github.com/1boombacks1/stat_dice/server/templates"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

func CreateLobbyContent(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	content, err := templates.CREATE_LOBBY_CONTENT.GetTemplate()
	if err != nil {
		render.Render(w, r,
			httpErrors.ErrInternalServer(fmt.Errorf("getting create lobby page content: %w", err)).
				WithLog(ctx.Error()))
		return
	}

	if err := content.Execute(w, nil); err != nil {
		render.Render(w, r, httpErrors.ErrInternalServer(fmt.Errorf("rendering create lobby page: %w", err)).WithLog(ctx.Error()))
	}
}

func CreateLobby(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
		httpErrors.ErrUnsupportedMediaType().SetTitle("Content-Type Error").Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		return
	}

	user := models.GetUserFromContext(appctx.FromContext(r.Context()))
	if user.Match != nil {
		httpErrors.ErrBadRequest(errors.New("user already in lobby")).
			WithExplanation("You already in lobby. You can't create lobby").SetTitle("Create Lobby Error").
			Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		return
	}

	isCompetitive := r.FormValue("competitive") == "on"
	lobbyName := r.FormValue("lobby-name")
	if lobbyName == "" {
		httpErrors.ErrBadRequest(errors.New("lobby name not specified")).SetTitle("Create Lobby Error").
			WithExplanation("Set lobby name").
			Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		return
	}

	cookie, err := r.Cookie("game-id")
	if err != nil {
		httpErrors.ErrBadRequest(errors.New("cookie with game-id not found")).SetTitle("Cookie Error").
			Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		return
	}
	gameID := cookie.Value

	lobby := &models.Lobby{
		Name:          lobbyName,
		Status:        models.LOBBY_STATUS_OPEN,
		GameID:        uuid.MustParse(gameID),
		IsCompetitive: isCompetitive,
	}

	lobbyID, err := user.CreateLobby(ctx, lobby)
	if err != nil {
		httpErrors.ErrInternalServer(fmt.Errorf("creating lobby: %w", err)).WithLog(ctx.Error()).
			SetTitle("DB Error").
			Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		return
	}

	redirectTo(w, "/counter/lobby/"+lobbyID.String())
}

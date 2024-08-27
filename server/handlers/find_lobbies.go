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
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

var findLobbiesTmpl *template.Template

func FindLobbiesContent(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	if err := findLobbiesTmpl.ExecuteTemplate(w, "content", nil); err != nil {
		httpErrors.ErrInternalServer(fmt.Errorf("rendering find-lobbies content: %w", err)).SetTitle("Template Error").
			WithLog(ctx.Error()).Execute(w, httpErrors.AppErrTmplName, ctx.Error())
	}
}

func GetOpenLobbies(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	gameID, err := getGameIDFromCookie(r)
	if err != nil {
		httpErrors.ErrInternalServer(fmt.Errorf("getting game id from cookie: %w", err)).SetTitle("Cookie Error").
			Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		return
	}

	lobbies, err := models.GetOpenLobbies(ctx, gameID)
	if err != nil {
		httpErrors.ErrInternalServer(fmt.Errorf("getting open lobbies: %w", err)).WithLog(ctx.Error()).
			SetTitle("DB Error").Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		return
	}

	type LobbyInfo struct {
		Context      *appctx.AppCtx
		ID           string
		Name         string
		CreatedAt    string
		GameID       string
		PlayersCount string
		Players      []*models.User
	}

	lobbiesInfo := make([]LobbyInfo, 0, len(lobbies))
	for _, lobby := range lobbies {
		lobbiesInfo = append(lobbiesInfo, LobbyInfo{
			Context:      ctx,
			ID:           lobby.GetID(),
			Name:         lobby.Name,
			CreatedAt:    lobby.GetCreatedAt(),
			GameID:       lobby.GameID.String(),
			PlayersCount: lobby.GetPlayerCount(),
			Players:      lobby.Players,
		})
	}

	if err := findLobbiesTmpl.ExecuteTemplate(w, "lobbies-list", lobbiesInfo); err != nil {
		httpErrors.ErrInternalServer(fmt.Errorf("rendering lobbies-list: %w", err)).WithLog(ctx.Error()).
			SetTitle("Template Error").Execute(w, httpErrors.AppErrTmplName, ctx.Error())
	}
}

func JoinLobby(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	lobbyIDParam := chi.URLParam(r, "id")
	user := models.GetUserFromContext(appctx.FromContext(r.Context()))
	if user.Match != nil {
		httpErrors.ErrBadRequest(errors.New("user already in match")).SetTitle("Bad Request").
			WithExplanation("You already in match. You can't join another match").Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		return
	}

	lobbyID, err := uuid.Parse(lobbyIDParam)
	if err != nil {
		httpErrors.ErrBadRequest(fmt.Errorf("parsing lobby id: %w", err)).WithLog(ctx.Error()).
			SetTitle("Bad Request").WithExplanation("Lobby id is invalid").
			Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		return
	}

	match := models.Match{
		LobbyID: lobbyID,
		UserID:  user.ID,
		Result:  models.RESULT_STATUS_PLAYING,
		IsHost:  false,
	}

	if err := match.Create(ctx); err != nil {
		httpErrors.ErrInternalServer(fmt.Errorf("creating match: %w", err)).SetTitle("DB Error").Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		ctx.Error().EmbedObject(user).Err(fmt.Errorf("creating match: %w", err)).Msg("user can't join to lobby. failed to create match")
		return
	}

	redirectTo(w, "/counter/lobby/"+lobbyID.String())
}

func init() {
	var err error
	findLobbiesTmpl, err = templates.FIND_LOBBY_CONTENT.GetTemplate(nil)
	if err != nil {
		panic(fmt.Errorf("failed to get FIND_LOBBY template: %w", err))
	}
}

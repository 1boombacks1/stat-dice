package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/1boombacks1/stat_dice/appctx"
	"github.com/1boombacks1/stat_dice/models"
	httpErrors "github.com/1boombacks1/stat_dice/server/http_errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

func FindLobbiesContent(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	if err := appTmpl.ExecuteTemplate(w, "find-lobbies", nil); err != nil {
		err = fmt.Errorf("rendering find-lobbies page: %w", err)
		render.Render(w, r, httpErrors.ErrInternalServer(err))
		ctx.Error().Err(err).Send()
	}
}

func GetOpenLobbies(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	lobbies, err := models.GetOpenLobbies(ctx)
	if err != nil {
		err = fmt.Errorf("getting open lobbies: %w", err)
		httpErrors.ErrInternalServer(err).SetTitle("DB Error").Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		ctx.Error().Err(err).Send()
		return
	}

	type LobbyInfo struct {
		Context      *appctx.AppCtx
		ID           string
		Name         string
		CreatedAt    string
		PlayersCount string
		Players      []*models.User
	}

	lobbiesInfo := make([]LobbyInfo, 0, len(lobbies))
	for _, lobby := range lobbies {
		// players, err := lobby.GetPlayersWithMatchInfo(ctx)
		// if err != nil {
		// 	err = fmt.Errorf("getting players for match: %w", err)
		// 	httpErrors.ErrInternalServer(err).SetTitle("DB Error").Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		// 	ctx.Error().Err(err).Send()
		// 	return
		// }
		lobbiesInfo = append(lobbiesInfo, LobbyInfo{
			Context:      ctx,
			ID:           lobby.GetID(),
			Name:         lobby.Name,
			CreatedAt:    lobby.GetCreatedAt(),
			PlayersCount: lobby.GetPlayerCount(),
			Players:      lobby.Players,
		})
	}

	if err := appTmpl.ExecuteTemplate(w, "lobbies-list", lobbiesInfo); err != nil {
		err = fmt.Errorf("rendering main page: %w", err)
		httpErrors.ErrInternalServer(err).SetTitle("Template Error").Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		ctx.Error().Err(err).Send()
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
		err = fmt.Errorf("parsing lobby id: %w", err)
		httpErrors.ErrBadRequest(err).SetTitle("Bad Request").
			WithExplanation("Lobby id is invalid").Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		ctx.Error().EmbedObject(user).Err(err).Msg("user can't join to lobby. failed to parse lobby id")
		return
	}

	match := models.Match{
		LobbyID: lobbyID,
		UserID:  user.ID,
		Result:  models.RESULT_STATUS_PLAYING,
		IsHost:  false,
	}

	if err := match.Create(ctx); err != nil {
		err = fmt.Errorf("creating match: %w", err)
		httpErrors.ErrInternalServer(err).SetTitle("DB Error").Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		ctx.Error().EmbedObject(user).Err(err).Msg("user can't join to lobby. failed to create match")
		return
	}

	redirectTo(w, "/counter/lobby/"+lobbyID.String())
}

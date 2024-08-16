package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/1boombacks1/stat_dice/appctx"
	"github.com/1boombacks1/stat_dice/models"
	httpErrors "github.com/1boombacks1/stat_dice/server/http_errors"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type listInfo struct {
	Players         []*models.User
	CurrentPlayerID string
	LobbyStatus     models.LobbyStatus
}

func LobbyPage(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	games, err := models.GetGames(ctx)
	if err != nil {
		render.Render(w, r, httpErrors.ErrInternalServer(fmt.Errorf("getting games: %w", err)))
		return
	}

	user := models.GetUserFromContext(appctx.FromContext(r.Context()))

	players, err := user.Match.GetPlayers(ctx.DB().DB)
	if err != nil {
		render.Render(w, r, httpErrors.ErrInternalServer(fmt.Errorf("getting players: %w", err)))
		return
	}

	type LobbyInfo struct {
		ID        string
		Name      string
		CreatedAt string
		Status    models.LobbyStatus
	}

	if err := appTmpl.ExecuteTemplate(w, "lobby-page",
		struct {
			WindowName string
			Games      []models.Game
			Username   string

			IsHost    bool
			Match     *models.Match
			LobbyInfo LobbyInfo
			ListInfo  listInfo
		}{
			WindowName: "Lobby",
			Games:      games,
			Username:   user.Name,

			Match:  user.Match,
			IsHost: user.Match.IsHost,
			LobbyInfo: LobbyInfo{
				ID:        user.Match.Lobby.GetID(),
				Name:      user.Match.Lobby.Name,
				CreatedAt: user.Match.Lobby.GetCreatedAt(),
				Status:    user.Match.Lobby.Status,
			},
			ListInfo: listInfo{
				Players:         players,
				CurrentPlayerID: user.GetID(),
				LobbyStatus:     user.Match.Lobby.Status,
			},
		},
	); err != nil {
		panic("failed to execute template: " + err.Error())
	}
}

func CreateLobbyContent(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	if err := appTmpl.ExecuteTemplate(w, "create-lobby", nil); err != nil {
		panic(fmt.Errorf("rendering create lobby page: %w", err))
	}
}

func GetMatchPlayers(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	user := models.GetUserFromContext(appctx.FromContext(r.Context()))

	players, err := user.Match.GetPlayers(ctx.DB().DB)
	if err != nil {
		httpErrors.ErrInternalServer(fmt.Errorf("getting players: %w", err)).SetTitle("DB Error").
			Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		return
	}

	data := listInfo{
		Players:         players,
		CurrentPlayerID: user.GetID(),
		LobbyStatus:     user.Match.Lobby.Status,
	}

	if err := appTmpl.ExecuteTemplate(w, "lobby-list", data); err != nil {
		httpErrors.ErrInternalServer(fmt.Errorf("rendering players list: %w", err)).SetTitle("Template Error").
			Execute(w, httpErrors.AppErrTmplName, ctx.Error())
	}
}

func GetLobbyStatus(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	user := models.GetUserFromContext(appctx.FromContext(r.Context()))

	if user.Match.Lobby.Status == models.LOBBY_STATUS_RESULT {
		w.Header().Set("HX-Reswap", "innerHTML")
		if err := appTmpl.ExecuteTemplate(w, "lobby-player-btns", nil); err != nil {
			httpErrors.ErrInternalServer(fmt.Errorf("rendering lobby status: %w", err)).SetTitle("Template Error").
				Execute(w, httpErrors.AppErrTmplName, ctx.Error())
			return
		}
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
		httpErrors.ErrInternalServer(fmt.Errorf("creating lobby: %w", err)).SetTitle("DB Error").
			Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		return
	}

	redirectTo(w, "/counter/lobby/"+lobbyID.String())
}

func StartLobby(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	user := models.GetUserFromContext(appctx.FromContext(r.Context()))

	if err := user.Match.Lobby.Start(ctx.DB().DB); err != nil {
		httpErrors.ErrInternalServer(fmt.Errorf("starting lobby: %w", err)).SetTitle("DB Error").
			Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		return
	}

	refreshPage(w)
}

func StopLobby(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	user := models.GetUserFromContext(appctx.FromContext(r.Context()))

	if err := user.Match.Lobby.Stop(ctx.DB().DB); err != nil {
		httpErrors.ErrInternalServer(fmt.Errorf("stopping lobby: %w", err)).SetTitle("DB Error").
			Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		return
	}

	refreshPage(w)
}

func CancelLobby(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	user := models.GetUserFromContext(appctx.FromContext(r.Context()))

	if err := user.Match.Lobby.Delete(ctx); err != nil {
		httpErrors.ErrInternalServer(fmt.Errorf("deleting lobby: %w", err)).SetTitle("DB Error").
			Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		return
	}

	redirectToMainPage(w)
}

func LeaveLobby(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	user := models.GetUserFromContext(appctx.FromContext(r.Context()))

	if !user.Match.IsHost {
		if err := user.LeaveFromMatch(ctx); err != nil {
			httpErrors.ErrInternalServer(fmt.Errorf("leaving from lobby: %w", err)).SetTitle("DB Error").
				Execute(w, httpErrors.AppErrTmplName, ctx.Error())
			return
		}
		redirectToMainPage(w)
		return
	}

	players, err := user.Match.GetPlayers(ctx.DB().DB)
	if err != nil {
		httpErrors.ErrInternalServer(fmt.Errorf("getting players: %w", err)).SetTitle("DB Error").
			Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		return
	}

	if len(players) == 1 || allLeaved(players) {
		if err := user.Match.Lobby.Delete(ctx); err != nil {
			httpErrors.ErrInternalServer(fmt.Errorf("deleting lobby: %w", err)).SetTitle("DB Error").
				Execute(w, httpErrors.AppErrTmplName, ctx.Error())
			return
		}

		ctx.Log().Str("user", user.GetID()).Str("lobby", user.Match.LobbyID.String()).Msg("user leaved and deleted lobby beacuse he was alone")
		redirectToMainPage(w)
		return
	}

	if err := user.Match.SwapHost(ctx, players[1]); err != nil {
		httpErrors.ErrInternalServer(fmt.Errorf("swapping host: %w", err)).SetTitle("DB Error").
			Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		return
	}

	ctx.Log().Str("user", user.GetID()).
		Str("lobby", user.Match.LobbyID.String()).
		Str("new_host", players[1].GetID()).Msg("success swapped host")

	redirectToMainPage(w)
}

func allLeaved(players []*models.User) bool {
	count := 0
	for _, player := range players {
		if player.Match.Result != models.RESULT_STATUS_LEAVE {
			count++
		}
	}
	return count == 1
}

func WinMatch(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	user := models.GetUserFromContext(appctx.FromContext(r.Context()))
	user.Match.Result = models.RESULT_STATUS_WIN

	if err := user.Match.Update(ctx, []string{"result"}); err != nil {
		httpErrors.ErrInternalServer(fmt.Errorf("updating match result: %w", err)).SetTitle("DB Error").
			Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func LoseMatch(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	user := models.GetUserFromContext(appctx.FromContext(r.Context()))
	user.Match.Result = models.RESULT_STATUS_LOSE

	if err := user.Match.Update(ctx, []string{"result"}); err != nil {
		httpErrors.ErrInternalServer(fmt.Errorf("updating match result: %w", err)).SetTitle("DB Error").
			Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

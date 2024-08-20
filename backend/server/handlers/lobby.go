package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"text/template"

	"github.com/1boombacks1/stat_dice/appctx"
	"github.com/1boombacks1/stat_dice/models"
	httpErrors "github.com/1boombacks1/stat_dice/server/http_errors"
	"github.com/1boombacks1/stat_dice/server/templates"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

var lobbyTmpl *template.Template

func LobbyPage(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	games, err := models.GetGames(ctx)
	if err != nil {
		render.Render(w, r, httpErrors.ErrInternalServer(fmt.Errorf("getting games: %w", err)).WithLog(ctx.Error()))
		return
	}

	user := models.GetUserFromContext(appctx.FromContext(r.Context()))
	lobby := user.Match.Lobby

	type LobbyInfo struct {
		ID        string
		Name      string
		CreatedAt string
		Status    models.LobbyStatus
	}

	if err := lobbyTmpl.ExecuteTemplate(w, "index",
		struct {
			AppName    string
			WindowName string
			Username   string
			Games      []models.Game

			IsHost    bool
			Match     *models.Match
			LobbyInfo LobbyInfo
		}{
			AppName:    ctx.Config().AppName,
			WindowName: "Lobby " + lobby.Name,
			Username:   user.Name,
			Games:      games,

			Match:  user.Match,
			IsHost: user.Match.IsHost,
			LobbyInfo: LobbyInfo{
				ID:        lobby.GetID(),
				Name:      lobby.Name,
				CreatedAt: lobby.GetCreatedAt(),
				Status:    lobby.Status,
			},
		},
	); err != nil {
		render.Render(w, r, httpErrors.ErrInternalServer(fmt.Errorf("rendering lobby page: %w", err)).WithLog(ctx.Error()))
	}
}

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

func GetLobbyPlayers(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	user := models.GetUserFromContext(appctx.FromContext(r.Context()))

	lobbyIDParam := chi.URLParam(r, "id")
	lobbyID, err := uuid.Parse(lobbyIDParam)
	if err != nil {
		httpErrors.ErrBadRequest(errors.New("invalid lobby id")).SetTitle("Invalid Lobby ID").Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		return
	}

	lobby, err := models.GetLobbyByID(ctx, lobbyID)
	if err != nil {
		httpErrors.ErrInternalServer(fmt.Errorf("getting lobby: %w", err)).SetTitle("DB Error").
			Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		return
	}

	players, err := lobby.GetPlayersWithMatch(ctx)
	if err != nil {
		httpErrors.ErrInternalServer(fmt.Errorf("getting players: %w", err)).SetTitle("DB Error").
			Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		return
	}

	type ListInfo struct {
		Players         []*models.User
		CurrentPlayerID string
		LobbyStatus     models.LobbyStatus
	}

	data := ListInfo{
		Players:         players,
		CurrentPlayerID: user.GetID(),
		LobbyStatus:     lobby.Status,
	}

	if err := lobbyTmpl.ExecuteTemplate(w, "lobby-players-list", data); err != nil {
		httpErrors.ErrInternalServer(fmt.Errorf("rendering players list: %w", err)).SetTitle("Template Error").
			Execute(w, httpErrors.AppErrTmplName, ctx.Error())
	}
}

func GetLobbyStatus(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	user := models.GetUserFromContext(appctx.FromContext(r.Context()))

	if user.Match.Lobby.Status == models.LOBBY_STATUS_RESULT {
		w.Header().Set("HX-Reswap", "innerHTML")
		if err := lobbyTmpl.ExecuteTemplate(w, "lobby-player-btns", user.Match.Lobby.GetID()); err != nil {
			httpErrors.ErrInternalServer(fmt.Errorf("rendering lobby status: %w", err)).WithLog(ctx.Error()).
				SetTitle("Template Error").
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
		httpErrors.ErrInternalServer(fmt.Errorf("creating lobby: %w", err)).WithLog(ctx.Error()).
			SetTitle("DB Error").
			Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		return
	}

	redirectTo(w, "/counter/lobby/"+lobbyID.String())
}

func StartLobby(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	user := models.GetUserFromContext(appctx.FromContext(r.Context()))

	if err := user.Match.Lobby.Start(ctx.DB().DB); err != nil {
		httpErrors.ErrInternalServer(fmt.Errorf("starting lobby: %w", err)).WithLog(ctx.Error()).
			SetTitle("DB Error").
			Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		return
	}

	refreshPage(w)
}

func StopLobby(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	user := models.GetUserFromContext(appctx.FromContext(r.Context()))

	if err := user.Match.Lobby.Stop(ctx.DB().DB); err != nil {
		httpErrors.ErrInternalServer(fmt.Errorf("stopping lobby: %w", err)).WithLog(ctx.Error()).
			SetTitle("DB Error").
			Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		return
	}

	refreshPage(w)
}

func CancelLobby(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	user := models.GetUserFromContext(appctx.FromContext(r.Context()))

	if err := user.Match.Lobby.Delete(ctx); err != nil {
		httpErrors.ErrInternalServer(fmt.Errorf("deleting lobby: %w", err)).WithLog(ctx.Error()).
			SetTitle("DB Error").
			Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		return
	}

	redirectToMainPage(w)
}

// Deprecated
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

	players, err := user.Match.Lobby.GetPlayersWithMatch(ctx)
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
		httpErrors.ErrInternalServer(fmt.Errorf("updating match result: %w", err)).WithLog(ctx.Error()).
			SetTitle("DB Error").
			Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func LoseMatch(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	user := models.GetUserFromContext(appctx.FromContext(r.Context()))
	user.Match.Result = models.RESULT_STATUS_LOSE

	if err := user.Match.Update(ctx, []string{"result"}); err != nil {
		httpErrors.ErrInternalServer(fmt.Errorf("updating match result: %w", err)).WithLog(ctx.Error()).
			SetTitle("DB Error").
			Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func init() {
	var err error
	lobbyTmpl, err = template.ParseFS(templates.Main,
		"main/base.html",
		"main/sections/lobby.html",
		"main/components/*.html",
		"main/root/*.html",
	)
	if err != nil {
		panic(fmt.Errorf("failed to get LOBBY template: %w", err))
	}
}

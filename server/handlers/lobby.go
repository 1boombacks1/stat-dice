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

func GetLobbyPlayers(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
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
		Players     []*models.User
		LobbyStatus models.LobbyStatus
	}

	data := ListInfo{
		Players:     players,
		LobbyStatus: lobby.Status,
	}

	if err := lobbyTmpl.ExecuteTemplate(w, "lobby-players-list", data); err != nil {
		httpErrors.ErrInternalServer(fmt.Errorf("rendering players list: %w", err)).SetTitle("Template Error").
			Execute(w, httpErrors.AppErrTmplName, ctx.Error())
	}
}

func GetLobbyStatus(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	user := models.GetUserFromContext(appctx.FromContext(r.Context()))

	if user.Match.IsHost {
		w.Header().Set("HX-Reswap", "outerHTML")
		if err := lobbyTmpl.ExecuteTemplate(w, "lobby-host-end-btns", user.Match.Lobby.GetID()); err != nil {
			httpErrors.ErrInternalServer(fmt.Errorf("rendering lobby status: %w", err)).WithLog(ctx.Error()).
				SetTitle("Template Error").
				Execute(w, httpErrors.AppErrTmplName, ctx.Error())
			return
		}
		return
	}

	if user.Match.Lobby.Status == models.LOBBY_STATUS_RESULT {
		w.Header().Set("HX-Reswap", "outerHTML")
		if err := lobbyTmpl.ExecuteTemplate(w, "lobby-player-btns", user.Match.Lobby.GetID()); err != nil {
			httpErrors.ErrInternalServer(fmt.Errorf("rendering lobby status: %w", err)).WithLog(ctx.Error()).
				SetTitle("Template Error").
				Execute(w, httpErrors.AppErrTmplName, ctx.Error())
			return
		}
	}
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

	if err := user.Match.Lobby.Delete(ctx.DB().DB); err != nil {
		httpErrors.ErrInternalServer(fmt.Errorf("deleting lobby: %w", err)).WithLog(ctx.Error()).
			SetTitle("DB Error").
			Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		return
	}

	redirectToMainPage(w)
}

func LeaveLobby(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	user := models.GetUserFromContext(appctx.FromContext(r.Context()))

	if !user.Match.IsHost {
		if err := user.Match.Delete(ctx); err != nil {
			httpErrors.ErrInternalServer(fmt.Errorf("deleting match: %w", err)).WithLog(ctx.Error()).
				SetTitle("DB Error").
				Execute(w, httpErrors.AppErrTmplName, ctx.Error())
			return
		}
		ctx.Log().EmbedObject(user.Match).Msg("user leaved from match")
		redirectTo(w, "/counter")
		return
	}

	players := user.Match.Lobby.Players
	for _, player := range players {
		if player.ID != user.ID {
			if err := user.Match.SwapHost(ctx, player); err != nil {
				httpErrors.ErrInternalServer(fmt.Errorf("swapping host: %w", err)).WithLog(ctx.Error()).
					SetTitle("DB Error").Execute(w, httpErrors.AppErrTmplName, ctx.Error())
				return
			}
			ctx.Log().Str("lobby-id", user.Match.Lobby.GetID()).
				Str("user-id", user.GetID()).Str("user-name", user.Name).
				Str("new-host-id", player.GetID()).Str("new-host-name", player.Name).
				Msg("swapped host")
			break
		}
	}

	if err := user.Match.Delete(ctx); err != nil {
		httpErrors.ErrInternalServer(fmt.Errorf("deleting match: %w", err)).WithLog(ctx.Error()).
			SetTitle("DB Error").
			Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		return
	}

	ctx.Log().EmbedObject(user.Match).Msg("host leaved from match")
	redirectTo(w, "/counter")
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

	if err := renderBackBtn(w); err != nil {
		httpErrors.ErrInternalServer(fmt.Errorf("executing template: %w", err)).WithLog(ctx.Error()).
			SetTitle("Template Error").
			Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		return
	}
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

	if err := renderBackBtn(w); err != nil {
		httpErrors.ErrInternalServer(fmt.Errorf("executing template: %w", err)).WithLog(ctx.Error()).
			SetTitle("Template Error").
			Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		return
	}
}

func renderBackBtn(w http.ResponseWriter) error {
	tmpl, err := template.New("home").Parse(`<a href="/counter">Back to main</a>`)
	if err != nil {
		return err
	}
	if err := tmpl.Execute(w, nil); err != nil {
		return err
	}
	return nil
}

func init() {
	var err error
	lobbyTmpl, err = template.New("lobby-template").Funcs(template.FuncMap{
		"renderPlayerResult": func(status models.ResultStatus) template.HTML {
			var color string
			switch status {
			case models.RESULT_STATUS_WIN:
				color = "green"
			case models.RESULT_STATUS_LOSE:
				color = "red"
			case models.RESULT_STATUS_LEAVE:
				color = "grey"
			default:
				color = "black"
			}
			return template.HTML(fmt.Sprintf(`<p style="color: %s;">%s</p>`, color, status))
		},
	}).ParseFS(templates.Main,
		"main/base.html",
		"main/sections/lobby.html",
		"main/components/*.html",
		"main/root/*.html",
	)
	if err != nil {
		panic(fmt.Errorf("failed to get LOBBY template: %w", err))
	}
}

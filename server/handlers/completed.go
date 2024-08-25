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

var completedTmpl *template.Template

func GetCompletedLobbies(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	lobbies, err := models.GetCompletedLobbies(ctx)
	if err != nil {
		render.Render(w, r, httpErrors.ErrInternalServer(fmt.Errorf("getting completed lobbies: %w", err)))
		return
	}

	for i, lobby := range lobbies {
		players, err := lobby.GetPlayersWithMatch(ctx)
		if err != nil {
			render.Render(w, r, httpErrors.ErrInternalServer(fmt.Errorf("getting players with match: %w", err)))
			return
		}
		lobbies[i].Players = players
	}

	if err := completedTmpl.Execute(w, lobbies); err != nil {
		render.Render(w, r, httpErrors.ErrInternalServer(fmt.Errorf("rendering completed lobbies: %w", err)))
	}
}

func init() {
	funcs := &template.FuncMap{
		"RenderPlayerResult": func(status models.ResultStatus) template.HTML {
			color, text := "", ""
			switch status {
			case models.RESULT_STATUS_LOSE:
				color, text = "red", "L"
			case models.RESULT_STATUS_WIN:
				color, text = "green", "W"
			case models.RESULT_STATUS_LEAVE:
				color, text = "var(--color-grey)", "E"
			default:
				color, text = "black", "?"
			}

			return template.HTML(fmt.Sprintf(`<div class="result" style="color: %s;">%s</div>`, color, text))
		},
		"IsEven": func(n int) bool {
			return n%2 == 0
		},
	}
	var err error
	completedTmpl, err = templates.COMPLETED_LOBBIES_CONTENT.GetTemplate(funcs)
	if err != nil {
		panic(fmt.Errorf("failed to get COMPLETED template: %w", err))
	}
}

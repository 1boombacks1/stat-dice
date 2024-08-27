package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/1boombacks1/stat_dice/appctx"
	"github.com/1boombacks1/stat_dice/models"
	httpErrors "github.com/1boombacks1/stat_dice/server/http_errors"
	"github.com/1boombacks1/stat_dice/server/templates"
	"github.com/go-chi/chi/v5"
)

var leaderboardTmpl *template.Template

func LeaderboardContent(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	if err := leaderboardTmpl.ExecuteTemplate(w, "content", time.Now().Format("02 Jan 2006")); err != nil {
		httpErrors.ErrInternalServer(fmt.Errorf("executing leaderboard content template: %w", err)).SetTitle("Template Error").
			Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		return
	}
}

const (
	competitiveMode = "competitive"
	unratedMode     = "unrated"
)

type filterData struct {
	TemplateName string
	Stats        []models.PlayerStat
	Status       models.ResultStatus
	MaxFilter    int
}

func GetWinStats(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	gameID, err := getGameIDFromCookie(r)
	if err != nil {
		httpErrors.ErrInternalServer(fmt.Errorf("getting game id from cookie: %w", err)).SetTitle("Cookie Error").
			Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		return
	}

	isCompetitive := competitiveMode == chi.URLParam(r, "mode")
	stats, err := models.GetFilterStats(ctx, *gameID, isCompetitive, "win desc")
	if err != nil {
		httpErrors.ErrInternalServer(fmt.Errorf("getting win stats: %w", err)).SetTitle("Database Error").
			Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		return
	}

	var maxWins int
	if len(stats) > 0 {
		maxWins = stats[0].Win
	}

	if err := executeFilterTemplate(w, filterData{
		TemplateName: "win-leaderboard",
		Stats:        stats,
		Status:       models.RESULT_STATUS_WIN,
		MaxFilter:    maxWins,
	}); err != nil {
		httpErrors.ErrInternalServer(fmt.Errorf("executing win-filter template: %w", err)).SetTitle("Template Error").
			Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		return
	}
}

func GetLoseStats(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	gameID, err := getGameIDFromCookie(r)
	if err != nil {
		httpErrors.ErrInternalServer(fmt.Errorf("getting game id from cookie: %w", err)).SetTitle("Cookie Error").
			Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		return
	}

	isCompetitive := competitiveMode == chi.URLParam(r, "mode")
	stats, err := models.GetFilterStats(ctx, *gameID, isCompetitive, "lose desc")
	if err != nil {
		httpErrors.ErrInternalServer(fmt.Errorf("getting lose stats: %w", err)).SetTitle("Database Error").
			Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		return
	}

	var maxLose int
	if len(stats) > 0 {
		maxLose = stats[0].Lose
	}

	if err := executeFilterTemplate(w, filterData{
		TemplateName: "lose-leaderboard",
		Stats:        stats,
		Status:       models.RESULT_STATUS_LOSE,
		MaxFilter:    maxLose,
	}); err != nil {
		httpErrors.ErrInternalServer(fmt.Errorf("executing lose-filter template: %w", err)).SetTitle("Template Error").
			Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		return
	}
}

func GetTotalStats(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	gameID, err := getGameIDFromCookie(r)
	if err != nil {
		httpErrors.ErrInternalServer(fmt.Errorf("getting game id from cookie: %w", err)).SetTitle("Cookie Error").
			Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		return
	}

	isCompetitive := competitiveMode == chi.URLParam(r, "mode")
	stats, err := models.GetFilterStats(ctx, *gameID, isCompetitive, "total desc")
	if err != nil {
		httpErrors.ErrInternalServer(fmt.Errorf("getting total stats: %w", err)).SetTitle("Database Error").
			Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		return
	}

	var maxTotal int
	if len(stats) > 0 {
		maxTotal = stats[0].Total
	}

	if err := executeFilterTemplate(w, filterData{
		TemplateName: "total-leaderboard",
		Stats:        stats,
		Status:       models.ResultStatus(5),
		MaxFilter:    maxTotal,
	}); err != nil {
		httpErrors.ErrInternalServer(fmt.Errorf("executing total-filter template: %w", err)).SetTitle("Template Error").
			Execute(w, httpErrors.AppErrTmplName, ctx.Error())
		return
	}
}

func executeFilterTemplate(w http.ResponseWriter, data filterData) error {
	if err := leaderboardTmpl.ExecuteTemplate(w, data.TemplateName, data); err != nil {
		return fmt.Errorf("executing filter template: %w", err)
	}
	return nil
}

func init() {
	var err error
	leaderboardTmpl, err = templates.LEADERBOARD_CONTENT.GetTemplate(nil)
	if err != nil {
		panic(fmt.Errorf("failed to get LEADERBOARD template: %w", err))
	}
}

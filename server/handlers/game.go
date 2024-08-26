package handlers

import (
	"net/http"

	"github.com/1boombacks1/stat_dice/appctx"
	"github.com/go-chi/chi/v5"
)

func SetGame(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	gameID := chi.URLParam(r, "id")

	http.SetCookie(w, &http.Cookie{
		Name:     "game-id",
		Path:     "/",
		Value:    gameID,
		HttpOnly: true,
	})
	refreshPage(w)
}

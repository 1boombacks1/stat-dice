package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/1boombacks1/stat_dice/appctx"
	"github.com/1boombacks1/stat_dice/models"
	"github.com/1boombacks1/stat_dice/server/httpErr"
	"github.com/go-chi/render"
)

func Registration(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	type Request struct {
		Login    string `json:"login"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}

	if r.Header.Get("Content-Type") != "application/json" {
		render.Render(w, r, httpErr.HTTPUnsupportedMediaType(errors.New("unsupported media type")))
		return
	}

	var request Request
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		render.Render(w, r, httpErr.HTTPInternalServerError(fmt.Errorf("failed to decode request: %w", err)))
		return
	}

	user := models.User{
		Login:    request.Login,
		Password: request.Password,
		Name:     request.Name,
	}

	if err := user.Create(ctx); err != nil {
		render.Render(w, r, httpErr.HTTPInternalServerError(fmt.Errorf("failed to create user: %w", err)))
		return
	}

	render.DefaultResponder(w, r, render.M{
		"ok":      true,
		"user_id": user.GetID(),
	})
}

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

func Login(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	type Request struct {
		Login    string `json:"login"`
		Password string `json:"password"`
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

	user, err := models.GetUserByCredentials(ctx, request.Login, request.Password)
	if err != nil {
		render.Render(w, r, httpErr.HTTPInternalServerError(fmt.Errorf("failed to get user: %w", err)))
		return
	}
	if user == nil {
		render.Render(w, r, httpErr.HTTPUnauthorized(errors.New("invalid credentials")))
		return
	}

	token, err := user.GenerateJWT(ctx)
	if err != nil {
		render.Render(w, r, httpErr.HTTPInternalServerError(fmt.Errorf("failed to generate token: %w", err)))
		return
	}

	render.DefaultResponder(w, r, render.M{
		"token": token,
	})
}

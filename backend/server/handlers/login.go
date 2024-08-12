package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/1boombacks1/stat_dice/appctx"
	"github.com/1boombacks1/stat_dice/models"
	"github.com/1boombacks1/stat_dice/server/templates"
	"github.com/go-chi/render"
)

const signInErrElement = "err-sign-in"

func Login(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("HX-Request") != "true" {
		templates.WriteAuthError(w, signInErrElement, http.StatusBadRequest, errors.New("invalid request: not HTPX request"))
		return
	}
	if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
		templates.WriteAuthError(w, signInErrElement, http.StatusUnsupportedMediaType, errors.New("unsupport media type"))
		return
	}

	login := r.FormValue("login")
	password := r.FormValue("password")

	user, err := models.GetUserByCredentials(ctx, login, password)
	if err != nil {
		templates.WriteAuthError(w, signInErrElement, http.StatusInternalServerError, fmt.Errorf("failed to get user: %w", err))
		return
	}
	if user == nil {
		templates.WriteAuthError(w, signInErrElement, http.StatusUnauthorized, errors.New("invalid credentials"))
		return
	}

	token, err := user.GenerateJWT(ctx)
	if err != nil {
		templates.WriteAuthError(w, signInErrElement, http.StatusInternalServerError, fmt.Errorf("failed to generate token: %w", err))
		return
	}

	// TODO: Set cookie with token
	// TODO: Redirect to home page

	render.DefaultResponder(w, r, render.M{
		"token": token,
	})
}

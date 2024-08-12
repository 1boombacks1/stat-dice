package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/1boombacks1/stat_dice/appctx"
	"github.com/1boombacks1/stat_dice/models"
	"github.com/1boombacks1/stat_dice/server/templates"
)

const signUpErrElement = "err-sign-up"

func Registration(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("HX-Request") != "true" {
		templates.WriteAuthError(w, signUpErrElement, http.StatusBadRequest, errors.New("invalid request: not HTPX request"))
		return
	}
	if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
		templates.WriteAuthError(w, signUpErrElement, http.StatusUnsupportedMediaType, errors.New("unsupport media type"))
		return
	}

	nickname := r.FormValue("nickname")
	login := r.FormValue("login")
	password := r.FormValue("password")

	user := models.User{
		Name:     nickname,
		Login:    login,
		Password: password,
	}

	if err := user.Create(ctx); err != nil {
		templates.WriteAuthError(w, signUpErrElement, http.StatusInternalServerError, fmt.Errorf("failed to create user: %w", err))
		return
	}

	fmt.Fprint(w, `<span id="success-label" class="success-text">☑️ Registration success!</span>`)
}

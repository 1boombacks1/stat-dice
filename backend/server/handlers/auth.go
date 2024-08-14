package handlers

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"

	"github.com/1boombacks1/stat_dice/appctx"
	"github.com/1boombacks1/stat_dice/models"
	"github.com/1boombacks1/stat_dice/server/templates"
)

var authTmpl *template.Template

const (
	signInErrElement = "err-sign-in"
	signUpErrElement = "err-sign-up"
)

func AuthPage(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	if err := authTmpl.ExecuteTemplate(w, "auth", nil); err != nil {
		panic(fmt.Errorf("rendering template: %w", err))
	}
}

func Login(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	if r.Header.Get("HX-Request") != "true" {
		writeAuthError(w, signInErrElement, http.StatusBadRequest, errors.New("invalid request: not HTPX request"))
		return
	}
	if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
		writeAuthError(w, signInErrElement, http.StatusUnsupportedMediaType, errors.New("unsupport media type"))
		return
	}

	// login := r.FormValue("login")
	// password := r.FormValue("password")

	// user, err := models.GetUserByCredentials(ctx, login, password)
	// if err != nil {
	// 	writeAuthError(w, signInErrElement, http.StatusInternalServerError, fmt.Errorf("failed to get user: %w", err))
	// 	return
	// }
	// if user == nil {
	// 	writeAuthError(w, signInErrElement, http.StatusUnauthorized, errors.New("invalid credentials"))
	// 	return
	// }

	// token, err := user.GenerateJWT(ctx)
	// if err != nil {
	// 	writeAuthError(w, signInErrElement, http.StatusInternalServerError, fmt.Errorf("failed to generate token: %w", err))
	// 	return
	// }

	// http.SetCookie(w, &http.Cookie{
	// 	Name:  "token",
	// 	Value: token,
	// 	// Secure:   true,
	// 	HttpOnly: true,
	// })

	w.Header().Set("HX-Redirect", "/counter")
}

func Registration(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("HX-Request") != "true" {
		writeAuthError(w, signUpErrElement, http.StatusBadRequest, errors.New("invalid request: not HTPX request"))
		return
	}
	if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
		writeAuthError(w, signUpErrElement, http.StatusUnsupportedMediaType, errors.New("unsupport media type"))
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
		writeAuthError(w, signUpErrElement, http.StatusInternalServerError, fmt.Errorf("failed to create user: %w", err))
		return
	}

	fmt.Fprint(w, `<span id="success-label" class="success-text">☑️ Registration success!</span>`)
}

func writeAuthError(w http.ResponseWriter, elementID string, status int, err error) {
	w.WriteHeader(status)

	if err := authTmpl.ExecuteTemplate(w, "auth-err",
		struct {
			ElementID string
			Error     string
		}{
			ElementID: elementID,
			Error:     err.Error(),
		}); err != nil {
		panic("failed to EXECUTE auth-err-template: " + err.Error())
	}
}

func init() {
	authTmpl = template.Must(template.ParseFS(templates.Auth, "auth/*.html"))
}

package handlers

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/1boombacks1/stat_dice/appctx"
	"github.com/1boombacks1/stat_dice/models"
	httpErrors "github.com/1boombacks1/stat_dice/server/http_errors"
	"github.com/1boombacks1/stat_dice/server/templates"
	"github.com/go-chi/render"
)

var authTmpl *template.Template

const (
	signInErrElement = "err-sign-in"
	signUpErrElement = "err-sign-up"
)

func AuthPage(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("token")
	if err != nil {
		if err := authTmpl.Execute(w, nil); err != nil {
			panic(fmt.Errorf("rendering template: %w", err))
		}
	} else {
		http.Redirect(w, r, "/counter", http.StatusTemporaryRedirect)
	}
}

func Login(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
		httpErrors.ErrUnsupportedMediaType().SetElementID(signInErrElement).Execute(w, httpErrors.AuthErrTmplName, ctx.Error())
		return
	}

	login := r.FormValue("login")
	password := r.FormValue("password")

	user, err := models.GetUserByCredentials(ctx, login, password)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			httpErrors.ErrUnauthorized(errors.New("invalid credentials")).
				SetElementID(signInErrElement).Execute(w, httpErrors.AuthErrTmplName, ctx.Error())
			return
		}

		httpErrors.ErrInternalServer(fmt.Errorf("failed to get user: %w", err)).
			SetElementID(signInErrElement).Execute(w, httpErrors.AuthErrTmplName, ctx.Error())
		return
	}
	if user == nil {
		httpErrors.ErrUnauthorized(errors.New("invalid credentials")).
			SetElementID(signInErrElement).Execute(w, httpErrors.AuthErrTmplName, ctx.Error())
		return
	}

	token, err := user.GenerateJWT(ctx)
	if err != nil {
		httpErrors.ErrInternalServer(fmt.Errorf("failed to generate token: %w", err)).
			SetElementID(signInErrElement).Execute(w, httpErrors.AuthErrTmplName, ctx.Error())
		return
	}

	http.SetCookie(w, &http.Cookie{
		Path:    "/",
		Name:    "token",
		Value:   token,
		Expires: time.Now().Add(ctx.Config().JWTDuration),
		// Secure:   true, при HTTPS - включить
		HttpOnly: true,
	})

	w.Header().Set("Content-Type", "application/json")
	render.DefaultResponder(w, r, render.M{
		"token": token,
	})
}

func Registration(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("HX-Request") != "true" {
		httpErrors.ErrBadRequest(errors.New("invalid request: not HTPX request")).
			SetElementID(signUpErrElement).Execute(w, httpErrors.AuthErrTmplName, ctx.Error())
		return
	}
	if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
		httpErrors.ErrUnsupportedMediaType().SetElementID(signUpErrElement).Execute(w, httpErrors.AuthErrTmplName, ctx.Error())
		return
	}

	nickname := r.FormValue("nickname")
	login := r.FormValue("login")
	password := r.FormValue("password")
	password2 := r.FormValue("repeat-password")
	if password != password2 {
		httpErrors.ErrBadRequest(errors.New("passwords don't match")).WithExplanation("Пароли не совпадают").
			SetElementID(signUpErrElement).Execute(w, httpErrors.AuthErrTmplName, ctx.Error())
		return
	}

	user := models.User{
		Name:     nickname,
		Login:    login,
		Password: password,
	}

	if err := user.Create(ctx); err != nil {
		httpErrors.ErrInternalServer(fmt.Errorf("failed to create user: %w", err)).
			SetElementID(signUpErrElement).Execute(w, httpErrors.AuthErrTmplName, ctx.Error())
		return
	}

	// fmt.Fprint(w, `<span id="success-label" class="success-text">☑️ Registration success!</span>`)
	token, err := user.GenerateJWT(ctx)
	if err != nil {
		httpErrors.ErrInternalServer(fmt.Errorf("failed to generate token: %w", err)).
			SetElementID(signUpErrElement).Execute(w, httpErrors.AuthErrTmplName, ctx.Error())
		return
	}

	http.SetCookie(w, &http.Cookie{
		Path:    "/",
		Name:    "token",
		Value:   token,
		Expires: time.Now().Add(ctx.Config().JWTDuration),
		// Secure:   true, при HTTPS - включить
		HttpOnly: true,
	})

	redirectTo(w, "/counter")
}

func init() {
	authTmpl = template.Must(template.ParseFS(templates.Auth, "auth/*.html"))
}

package handlers

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/1boombacks1/stat_dice/appctx"
	"github.com/1boombacks1/stat_dice/server/templates"
)

var appTmpl *template.Template

func MainPage(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	if err := appTmpl.ExecuteTemplate(w, "main", nil); err != nil {
		panic(fmt.Errorf("rendering main page: %w", err))
	}
}

func FindLobbies(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.RemoteAddr)
	fmt.Println(r.URL.Path)

	if err := appTmpl.ExecuteTemplate(w, "find-lobbies", nil); err != nil {
		panic(fmt.Errorf("rendering main page: %w", err))
	}
}

func CreateLobbyPage(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.RemoteAddr)
	fmt.Println(r.URL.Path)
	if err := appTmpl.ExecuteTemplate(w, "create-lobby", nil); err != nil {
		panic(fmt.Errorf("rendering create lobby page: %w", err))
	}
}

func init() {
	appTmpl = template.Must(template.ParseFS(templates.Main, "main/*.html"))
}

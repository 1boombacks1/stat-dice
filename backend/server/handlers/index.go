package handlers

import (
	"fmt"
	"net/http"

	"github.com/1boombacks1/stat_dice/appctx"
	"github.com/1boombacks1/stat_dice/server/templates"
)

func Auth(ctx *appctx.AppCtx, w http.ResponseWriter, r *http.Request) {
	if err := templates.ExecuteAuth(w); err != nil {
		panic(fmt.Errorf("rendering template: %w", err))
	}
}

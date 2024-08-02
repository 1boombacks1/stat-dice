package cmd

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/1boombacks1/stat_dice/app"
	"github.com/1boombacks1/stat_dice/appctx"
	"github.com/1boombacks1/stat_dice/config"
	"github.com/1boombacks1/stat_dice/server"
	"github.com/alecthomas/kong"
)

type CLI struct {
	config.Config
}

func Execute() {
	cli := &CLI{}
	ctx := kong.Parse(cli, kong.Name("stat-dice"), kong.Description("web-service for work application 'stat-dice'"))
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}

func (c *CLI) Run() error {
	return app.WithApp(&c.Config, func(ctx *appctx.AppCtx) error {
		ctx.Log().Str("addr", ctx.Config().Address).Uint16("port", ctx.Config().Port).Msg("serving app")

		srv := server.NewHTTPServer(ctx)
		defer srv.Stop()

		go func() {
			if err := srv.Start(); err != nil {
				ctx.Log().Err(err).Msg("failed to start server")
			}
		}()

		<-ctx.Done()

		sctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := srv.Shutdown(sctx)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			ctx.Error().Err(err).Msg("closing http server")
		}

		return nil
	})
}

func (c *CLI) Validate() error {
	//TODO
	return nil
}

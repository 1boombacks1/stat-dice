package cmd

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/1boombacks1/stat_dice/app"
	"github.com/1boombacks1/stat_dice/appctx"
	"github.com/1boombacks1/stat_dice/config"
	"github.com/1boombacks1/stat_dice/server"
	"github.com/alecthomas/kong"
)

type CLI struct {
	config.Config

	Game  GameCMD   `cmd:""`
	Serve ServerCMD `cmd:"" default:"withargs" help:"serve http-server"`
}

func Execute() {
	cli := &CLI{}

	ctx := kong.Parse(cli, kong.Name("stat-dice"), kong.Description("web-service for work application 'stat-dice'"))
	err := ctx.Run(&cli.Config)
	ctx.FatalIfErrorf(err)
}

type ServerCMD struct {
	Path string `default:"config.yaml" help:"path to yaml config"`
}

func (c *ServerCMD) Run(cfg *config.Config) error {
	return app.WithApp(cfg, func(ctx *appctx.AppCtx) error {
		// todo: без must
		config.MustParseYAML(c.Path, cfg)

		ctx.Log().Str("addr", ctx.Config().Host).Uint16("port", ctx.Config().Port).Msg("serving app")

		srv := server.NewHTTPServer(ctx)
		defer srv.Stop()

		wg := sync.WaitGroup{}
		wg.Add(1)

		go func() {
			defer wg.Done()
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

		wg.Wait()
		return nil
	})
}

func (c *CLI) Validate() error {
	// TODO
	return nil
}

package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/1boombacks1/stat_dice/appctx"
	"github.com/1boombacks1/stat_dice/config"
	"github.com/1boombacks1/stat_dice/db"
	"github.com/rs/zerolog"
)

type App struct {
	ctx appctx.AppCtx
}

func newApp(ctx context.Context, cfg *config.Config) (*App, error) {
	logger := makeLogger(cfg.Debug)

	db, err := db.NewDB(ctx, logger, cfg)
	if err != nil {
		return nil, fmt.Errorf("initializing db: %w", err)
	}

	version, err := db.GetVersion()
	if err != nil {
		logger.Error().Err(err).Msg("db: failed to get version")
	} else {
		logger.Debug().Str("version", version).Msg("db: running version")
	}

	app := App{
		ctx: appctx.NewAppCtx(ctx, cfg, logger, db),
	}

	err = db.SetMaxConnections(10)
	if err != nil {
		return nil, fmt.Errorf("setting db max connections: %w", err)
	}

	app.ctx.Log().Msg("app: created instance")
	return &app, nil
}

func (app *App) Shutdown(err error) {
	app.ctx.Log().Err(err).Msg("shutting down")
	err = app.ctx.DB().Shutdown()
	if err != nil {
		app.ctx.Log().Err(err).Msg("error while shutting down")
	}
}

func WithApp(cfg *config.Config, callback func(ctx *appctx.AppCtx) error) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	a, err := newApp(ctx, cfg)
	if err != nil {
		return fmt.Errorf("initializing app: %w", err)
	}

	defer func() {
		a.Shutdown(err)
	}()

	err = callback(&a.ctx)
	return err
}

func makeLogger(debug bool) *zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339Nano

	logger := zerolog.New(os.Stdout)
	if debug {
		logger = logger.Level(zerolog.DebugLevel).Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "02.01.2006 15:04:05.000000",
		})
	} else {
		logger = logger.Level(zerolog.InfoLevel)
	}

	logger = logger.With().Timestamp().Logger()
	logger.Debug().Msg("logger: initialized")

	return &logger
}

package appctx

import (
	"context"
	"time"

	"github.com/1boombacks1/stat_dice/config"
	"github.com/1boombacks1/stat_dice/db"
	"github.com/rs/zerolog"
)

type AppCtx struct {
	context.Context
}

type appCtxKey struct{}

type appCtxContents struct {
	config *config.Config
	logger *zerolog.Logger
	db     *db.DB
}

func NewAppCtx(ctx context.Context, config *config.Config, logger *zerolog.Logger, db *db.DB) AppCtx {
	ac := AppCtx{
		context.WithValue(ctx, appCtxKey{}, &appCtxContents{
			config: config,
			logger: logger,
			db:     db,
		}),
	}

	return ac
}

func FromContext(ctx context.Context) *AppCtx {
	ac, ok := ctx.(*AppCtx)
	if !ok || ac == nil {
		return &AppCtx{
			ctx,
		}
	} else {
		return ac
	}
}

func (ac *AppCtx) Logger() *zerolog.Logger {
	return contents(ac).logger
}

func (ac *AppCtx) Log() *zerolog.Event {
	return contents(ac).logger.Info()
}

func (ac *AppCtx) Debug() *zerolog.Event {
	return contents(ac).logger.Debug()
}

func (ac *AppCtx) Error() *zerolog.Event {
	return contents(ac).logger.Error()
}

func (ac *AppCtx) Warn() *zerolog.Event {
	return contents(ac).logger.Warn()
}

func (ac *AppCtx) DB() *db.DB {
	return &db.DB{
		DB: contents(ac).db.WithContext(ac),
	}
}

func (ac *AppCtx) Config() *config.Config {
	return contents(ac).config
}

func (ac *AppCtx) WithValue(key, val any) *AppCtx {
	return &AppCtx{
		Context: context.WithValue(ac, key, val),
	}
}

func (ac *AppCtx) WithTimeout(d time.Duration) (*AppCtx, func()) {
	ctx, cancel := context.WithTimeout(ac, d)
	return &AppCtx{
		Context: ctx,
	}, cancel
}

func contents(ctx context.Context) *appCtxContents {
	c, ok := ctx.Value(appCtxKey{}).(*appCtxContents)
	if !ok || c == nil {
		panic("missing contents from app context")
	}

	return c
}

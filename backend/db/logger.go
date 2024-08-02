package db

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/gorm/logger"
)

type DBLogger struct {
	logger *zerolog.Logger
	trace  bool
}

func NewDBLogger(logger *zerolog.Logger, trace bool) *DBLogger {
	return &DBLogger{
		logger: logger,
		trace:  trace,
	}
}

func (l DBLogger) LogMode(logger.LogLevel) logger.Interface {
	return l
}

func (l DBLogger) Error(ctx context.Context, msg string, opts ...any) {
	l.logger.Error().Msg(fmt.Sprintf(msg, opts...))
}

func (l DBLogger) Warn(ctx context.Context, msg string, opts ...any) {
	l.logger.Warn().Msg(fmt.Sprintf(msg, opts...))
}

func (l DBLogger) Info(ctx context.Context, msg string, opts ...any) {
	l.logger.Info().Msg(fmt.Sprintf(msg, opts...))
}

func (l DBLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if !l.trace {
		return
	}

	event := l.logger.Info()

	event = event.Dur("elapsed", time.Since(begin))

	sql, rows := fc()
	if sql != "" {
		event = event.Str("sql", sql)
	}
	if rows > -1 {
		event = event.Int64("rows", rows)
	}

	event.Send()
}

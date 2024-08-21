package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/1boombacks1/stat_dice/config"
	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
}

func NewDB(ctx context.Context, logger *zerolog.Logger, cfg *config.Config) (*DB, error) {
	db := DB{}

	var err error
	db.DB, err = gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: NewDBLogger(logger, cfg.TraceSQL),
	})
	if err != nil {
		return nil, fmt.Errorf("initializing db: %w", err)
	}

	sqlDB, err := db.sqlDB()
	if err != nil {
		return nil, fmt.Errorf("getting SQL DB: %w", err)
	}

	sqlDB.SetConnMaxLifetime(time.Minute * 5)

	return &db, nil
}

func (db *DB) SetMaxConnections(maxConnections int) error {
	sqlDB, err := db.sqlDB()
	if err != nil {
		return fmt.Errorf("getting SQL DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(min(4, maxConnections))
	sqlDB.SetMaxOpenConns(maxConnections)
	return nil
}

func (db *DB) Shutdown() error {
	sqlDB, err := db.sqlDB()
	if err != nil {
		return fmt.Errorf("getting SQL DB: %w", err)
	}

	err = sqlDB.Close()
	if err != nil {
		return fmt.Errorf("closing DB: %w", err)
	}

	db.DB = nil
	return nil
}

func (db *DB) sqlDB() (*sql.DB, error) {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return nil, err
	}

	return sqlDB, nil
}

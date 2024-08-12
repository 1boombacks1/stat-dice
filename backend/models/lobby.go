package models

import (
	"fmt"
	"time"

	"github.com/1boombacks1/stat_dice/appctx"
	"github.com/google/uuid"
	"gorm.io/gorm/clause"
)

type Lobby struct {
	Base

	Name     string        `gorm:"not null"`
	Status   LobbyStatus   `gorm:"not null"`
	Duration time.Duration `gorm:"not null"`

	GameID uuid.UUID `gorm:"not null"`
	Game   Game
}

func (l *Lobby) Create(ctx *appctx.AppCtx) error {
	if err := ctx.DB().Create(l).Error; err != nil {
		return fmt.Errorf("creating lobby: %w", err)
	}
	return nil
}

func (l *Lobby) Update(ctx *appctx.AppCtx, fields []string) error {
	if err := ctx.DB().Clauses(clause.Returning{}).Select(fields).Updates(l).Error; err != nil {
		return fmt.Errorf("updating lobby: %w", err)
	}
	return nil
}

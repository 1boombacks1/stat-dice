package models

import (
	"fmt"
	"time"

	"github.com/1boombacks1/stat_dice/appctx"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Lobby struct {
	Base

	Name          string      `gorm:"not null"`
	Status        LobbyStatus `gorm:"not null"`
	StartedAt     *time.Time
	EndedAt       *time.Time
	IsCompetitive bool `gorm:"not null;default:false"`

	GameID uuid.UUID `gorm:"type:uuid;not null"`
	Game   Game
}

func (l *Lobby) GetCreatedAt() string {
	return l.CreatedAt.Format("02-01-2006 15:04")
}

func (l *Lobby) GetCurrentDuration() string {
	if l.StartedAt == nil {
		return "draft"
	}
	if l.EndedAt == nil {
		return time.Since(*l.StartedAt).Truncate(time.Minute).String()
	}
	return l.EndedAt.Sub(*l.StartedAt).Truncate(time.Minute).String()
}

func (l *Lobby) Start(db *gorm.DB) error {
	now := time.Now()
	l.StartedAt = &now
	l.Status = LOBBY_STATUS_PROCESSING

	return l.Update(db, []string{"Status", "StartedAt"})
}

func (l *Lobby) Stop(db *gorm.DB) error {
	now := time.Now()
	l.EndedAt = &now
	l.Status = LOBBY_STATUS_RESULT

	return l.Update(db, []string{"Status", "EndedAt"})
}

func (l *Lobby) Close(db *gorm.DB) error {
	l.Status = LOBBY_STATUS_CLOSED
	return l.Update(db, []string{"Status"})
}

func (l *Lobby) Create(ctx *appctx.AppCtx) error {
	if err := ctx.DB().Create(l).Error; err != nil {
		return fmt.Errorf("creating lobby: %w", err)
	}
	return nil
}

func (l *Lobby) Update(db *gorm.DB, fields []string) error {
	if err := db.Clauses(clause.Returning{}).Select(fields).Updates(l).Error; err != nil {
		return fmt.Errorf("updating lobby: %w", err)
	}
	return nil
}

func (l *Lobby) Delete(ctx *appctx.AppCtx) error {
	if err := ctx.DB().Delete(l).Error; err != nil {
		return fmt.Errorf("deleting lobby: %w", err)
	}
	return nil
}

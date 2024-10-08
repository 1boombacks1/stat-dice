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

	Players []*User `gorm:"many2many:matches;"`

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
	duration := time.Since(*l.StartedAt)
	if l.EndedAt != nil {
		duration = l.EndedAt.Sub(*l.StartedAt)
	}
	duration = duration.Truncate(time.Minute)

	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %02dm", hours, minutes)
	}

	return fmt.Sprintf("%02dm", minutes)
}

func (l *Lobby) GetPlayerCount() string {
	return fmt.Sprintf("%02d", len(l.Players))
}

func GetLobbyByID(ctx *appctx.AppCtx, id uuid.UUID) (*Lobby, error) {
	var lobby Lobby
	if err := ctx.DB().Preload("Players").Preload("Game").First(&lobby, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("getting lobby: %w", err)
	}
	return &lobby, nil
}

func GetOpenLobbies(ctx *appctx.AppCtx, gameID *uuid.UUID) ([]*Lobby, error) {
	var lobbies []*Lobby
	err := ctx.DB().Preload("Players").Where("status = ? AND game_id = ?", LOBBY_STATUS_OPEN, gameID).Find(&lobbies).Error
	if err != nil {
		return nil, fmt.Errorf("getting lobbies: %w", err)
	}
	return lobbies, nil
}

func (l *Lobby) GetPlayersWithMatch(ctx *appctx.AppCtx) ([]*User, error) {
	var players []*User
	if err := ctx.DB().Model(&User{}).Preload("Match", "lobby_id = ?", l.ID).
		Joins("JOIN matches on matches.user_id = users.id").
		Joins("JOIN lobbies on matches.lobby_id = lobbies.id").
		Where("lobbies.id = ?", l.ID).
		Find(&players).Error; err != nil {
		return nil, fmt.Errorf("getting players: %w", err)
	}

	return players, nil
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

func (l *Lobby) Delete(db *gorm.DB) error {
	if err := db.Delete(l).Error; err != nil {
		return fmt.Errorf("deleting lobby: %w", err)
	}
	return nil
}

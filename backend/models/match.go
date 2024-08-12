package models

import (
	"fmt"

	"github.com/1boombacks1/stat_dice/appctx"
	"github.com/google/uuid"
)

type Match struct {
	LobbyID uuid.UUID    `gorm:"primaryKey"`
	UserID  uuid.UUID    `gorm:"primaryKey"`
	Result  ResultStatus `gorm:"not null"`
	IsHost  bool         `gorm:"not null;default:false"`
}

func GetActiveMatchCreatedByUser(ctx *appctx.AppCtx, user *User) (*Match, error) {
	var match Match
	if err := ctx.DB().Where("user_id = ? AND result = ?", user.ID, RESULT_STATUS_PLAYING).First(&match).Error; err != nil {
		return nil, fmt.Errorf("getting active match: %w", err)
	}
	return &match, nil
}

func (m *Match) Create(ctx *appctx.AppCtx) error {
	if err := ctx.DB().Create(m).Error; err != nil {
		return fmt.Errorf("creating match: %w", err)
	}
	return nil
}

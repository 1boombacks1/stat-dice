package models

import (
	"fmt"
	"time"

	"github.com/1boombacks1/stat_dice/appctx"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type Base struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()"`
	CreatedAt  time.Time `gorm:"autoCreateTime;not null"`
	ModifiedAt time.Time `gorm:"autoUpdateTime;not null"`
}

func (b *Base) GetID() string {
	return b.ID.String()
}

func (b Base) MarshalZerologObject(e *zerolog.Event) {
	e.Str("id", b.GetID()).Time("created_at", b.CreatedAt).Time("modified_at", b.ModifiedAt)
}

func AutoMigrateModels(ctx *appctx.AppCtx) error {
	models := []interface{}{
		User{},
		Game{},
		Lobby{},
		Match{},
	}

	for _, model := range models {
		if err := ctx.DB().AutoMigrate(model); err != nil {
			return fmt.Errorf("migrating table: %T: %w", model, err)
		}
	}
	return nil
}

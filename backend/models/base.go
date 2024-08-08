package models

import (
	"time"

	"github.com/google/uuid"
)

type Base struct {
	ID         uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt  time.Time `gorm:"autoCreateTime;not null"`
	ModifiedAt time.Time `gorm:"autoUpdateTime;not null"`
}

func (b *Base) GetID() string {
	return b.ID.String()
}

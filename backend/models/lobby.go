package models

import (
	"time"

	"github.com/google/uuid"
)

type Lobby struct {
	Base

	Name     string        `gorm:"not null"`
	Status   LobbyStatus   `gorm:"not null"`
	Duration time.Duration `gorm:"not null"`

	GameID uuid.UUID `gorm:"not null"`
	Game   Game
}

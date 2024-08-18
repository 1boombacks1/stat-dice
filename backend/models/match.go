package models

import (
	"fmt"

	"github.com/1boombacks1/stat_dice/appctx"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Match struct {
	LobbyID uuid.UUID `gorm:"type:uuid;primaryKey"`
	Lobby   *Lobby    `gorm:"foreignKey:LobbyID;references:ID;constraint:OnDelete:CASCADE"`

	UserID uuid.UUID `gorm:"type:uuid;primaryKey"`
	User   *User     `gorm:"foreignKey:UserID;references:ID"`

	Result ResultStatus `gorm:"not null"`
	IsHost bool         `gorm:"not null;default:false"`

	PlayerCount int `gorm:"-:all"`
}

func (m Match) MarshalZerologObject(e *zerolog.Event) {
	e.EmbedObject(m.Lobby).EmbedObject(m.User).Str("user_result_status", m.Result.String()).Bool("is_host", m.IsHost)
}

func (m Match) GetPlayerCount() string {
	return fmt.Sprintf("%02d", m.PlayerCount)
}

// Get open matches. Match preloaded Lobby field.
// func GetOpenMatches(ctx *appctx.AppCtx) ([]*Match, error) {
// 	var matches []*Match
// 	err := ctx.DB().Model(&Match{}).Preload("Lobby").
// 		Joins("JOIN lobbies ON matches.lobby_id = lobbies.id").
// 		Where("lobbies.status = ?", LOBBY_STATUS_OPEN).
// 		Distinct("lobby_id").Find(&matches).Error

// 	return matches, err
// }

// Get players from match. Player preloaded Match field.
// func (m *Match) GetPlayers(db *gorm.DB) ([]*User, error) {
// 	var players []*User
// 	if err := db.Model(&User{}).Preload("Match").
// 		Joins("JOIN matches on matches.user_id = users.id").
// 		Joins("JOIN lobbies on matches.lobby_id = lobbies.id").
// 		Where("lobbies.status IN ?", []LobbyStatus{LOBBY_STATUS_OPEN, LOBBY_STATUS_PROCESSING, LOBBY_STATUS_RESULT}).
// 		Find(&players).Error; err != nil {
// 		return nil, fmt.Errorf("getting players: %w", err)
// 	}
// 	if len(players) == 0 {
// 		return nil, ErrPlayersNotFound
// 	}

// 	m.PlayerCount = len(players)
// 	return players, nil
// }

func (m *Match) SwapHost(ctx *appctx.AppCtx, newHost *User) error {
	err := ctx.DB().Transaction(func(tx *gorm.DB) error {
		err := tx.Model(m).Select("IsHost", "Result").Updates(&Match{Result: RESULT_STATUS_LEAVE, IsHost: false}).Error
		if err != nil {
			return err
		}

		err = tx.Model(&Match{LobbyID: m.LobbyID, UserID: newHost.ID}).Update("is_host", true).Error
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("swapping host: %w", err)
	}

	return nil
}

func (m *Match) Create(ctx *appctx.AppCtx) error {
	if err := ctx.DB().Create(m).Error; err != nil {
		return fmt.Errorf("creating match: %w", err)
	}
	return nil
}

func (m *Match) Update(ctx *appctx.AppCtx, fields []string) error {
	if err := ctx.DB().Clauses(clause.Returning{}).Select(fields).Updates(m).Error; err != nil {
		return fmt.Errorf("updating match: %w", err)
	}
	return nil
}

func (m *Match) AfterUpdate(tx *gorm.DB) error {
	var players []*User
	err := tx.Model(&User{}).Preload("Match").
		Joins("JOIN matches on matches.user_id = users.id").
		Joins("JOIN lobbies on matches.lobby_id = lobbies.id").
		Where("lobbies.status IN ?", []LobbyStatus{LOBBY_STATUS_OPEN, LOBBY_STATUS_PROCESSING, LOBBY_STATUS_RESULT}).
		Find(&players).Error
	if err != nil {
		return fmt.Errorf("getting players: %w", err)
	}

	for _, player := range players {
		if player.Match.Result == RESULT_STATUS_PLAYING {
			return nil
		}
	}

	if err := m.Lobby.Close(tx); err != nil {
		return fmt.Errorf("closing lobby: %w", err)
	}

	fmt.Printf("lobby '%v' are closed\n", m.Lobby.GetID())
	return nil
}

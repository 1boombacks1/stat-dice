package models

import (
	"errors"
	"fmt"

	"github.com/1boombacks1/stat_dice/appctx"
	"github.com/google/uuid"
)

type Game struct {
	Base

	Name string `gorm:"unique;not null"`
}

type Games []Game

func (g Games) GetByID(id uuid.UUID) (Game, error) {
	for _, game := range g {
		if game.ID == id {
			return game, nil
		}
	}
	return Game{}, errors.New("game not found")
}

func GetGames(ctx *appctx.AppCtx) (Games, error) {
	var games []Game
	if err := ctx.DB().Find(&games).Error; err != nil {
		return nil, fmt.Errorf("getting game list: %w", err)
	}
	return games, nil
}

func GetGameByName(ctx *appctx.AppCtx, name string) (*Game, error) {
	var game *Game
	if err := ctx.DB().Where("name = ?", name).First(&game).Error; err != nil {
		return nil, fmt.Errorf("getting game by name: %w", err)
	}
	return game, nil
}

func (g *Game) Create(ctx *appctx.AppCtx) error {
	if err := ctx.DB().Create(g).Error; err != nil {
		return fmt.Errorf("creating game: %w", err)
	}
	return nil
}

func (g *Game) Delete(ctx *appctx.AppCtx) error {
	if err := ctx.DB().Delete(g).Error; err != nil {
		return fmt.Errorf("deleting game: %w", err)
	}
	return nil
}

package models

import (
	"fmt"

	"github.com/1boombacks1/stat_dice/appctx"
)

type Game struct {
	Base

	Name string `gorm:"unique;not null"`
}

func GetGames(ctx *appctx.AppCtx) ([]Game, error) {
	var games []Game
	if err := ctx.DB().Find(&games).Error; err != nil {
		return nil, fmt.Errorf("getting game list: %w", err)
	}
	return games, nil
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

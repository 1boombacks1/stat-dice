package cmd

import (
	"errors"
	"fmt"

	"github.com/1boombacks1/stat_dice/app"
	"github.com/1boombacks1/stat_dice/appctx"
	"github.com/1boombacks1/stat_dice/config"
	"github.com/1boombacks1/stat_dice/models"
)

type GameCMD struct {
	Create CreateGameCMD `cmd:"" default:"withargs"`
	List   ListGamesCMD  `cmd:""`
	Delete DeleteGameCMD `cmd:""`
}

type CreateGameCMD struct {
	Name string `short:"n" required:"" help:"game name"`
}

func (cmd *CreateGameCMD) Run(cfg *config.Config) error {
	return app.WithApp(cfg, func(ctx *appctx.AppCtx) error {
		if cmd.Name == "" {
			return errors.New("empty name")
		}

		game := models.Game{
			Name: cmd.Name,
		}

		if err := game.Create(ctx); err != nil {
			return fmt.Errorf("failed to create game: %s", err.Error())
		}

		ctx.Log().Str("gameID", game.GetID()).Msg("game created")
		return nil
	})
}

type ListGamesCMD struct{}

func (cmd *ListGamesCMD) Run(cfg *config.Config) error {
	return app.WithApp(cfg, func(ctx *appctx.AppCtx) error {
		games, err := models.GetGames(ctx)
		if err != nil {
			return fmt.Errorf("failed to get games: %s", err.Error())
		}

		for _, game := range games {
			ctx.Log().Str("gameID", game.GetID()).Str("name", game.Name).Send()
		}

		return nil
	})
}

type DeleteGameCMD struct {
	Name string `short:"n" required:"" help:"game name"`
}

func (cmd *DeleteGameCMD) Run(cfg *config.Config) error {
	return app.WithApp(cfg, func(ctx *appctx.AppCtx) error {
		if cmd.Name == "" {
			return errors.New("empty name")
		}

		game, err := models.GetGameByName(ctx, cmd.Name)
		if err != nil {
			return fmt.Errorf("failed to get game: %s", err.Error())
		}

		if err := game.Delete(ctx); err != nil {
			return fmt.Errorf("failed to delete game: %s", err.Error())
		}

		ctx.Log().Str("gameID", game.GetID()).Msg("game deleted")
		return nil
	})
}

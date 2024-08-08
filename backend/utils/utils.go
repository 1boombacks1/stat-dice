package utils

import (
	"fmt"

	"github.com/1boombacks1/stat_dice/appctx"
	"golang.org/x/crypto/bcrypt"
)

func GenerateBcryptHash(ctx *appctx.AppCtx, password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), ctx.Config().BcryptCost)
	if err != nil {
		return "", fmt.Errorf("generating hash: %w", err)
	}

	return string(hash), nil
}

func CompareBcryptHash(ctx *appctx.AppCtx, hash string, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

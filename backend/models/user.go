package models

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/1boombacks1/stat_dice/appctx"
	"github.com/1boombacks1/stat_dice/utils"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type User struct {
	Base

	Login          string `gorm:"unique;not null"`
	Password       string `gorm:"-:all"`
	HashedPassword string `gorm:"not null"`
	Name           string `gorm:"not null"`
}

type UserJWTClaims struct {
	Login string `json:"login"`
	jwt.RegisteredClaims
}

type userCtxKey struct{}

func GetUsers(ctx *appctx.AppCtx) ([]User, error) {
	var users []User
	if err := ctx.DB().Find(&users).Error; err != nil {
		return nil, fmt.Errorf("getting users: %w", err)
	}
	return users, nil
}

func GetUserByCredentials(ctx *appctx.AppCtx, login, password string) (*User, error) {
	user, err := getUserByLogin(ctx, login)
	if err != nil {
		return nil, err
	}

	if !utils.CompareBcryptHash(ctx, user.HashedPassword, password) {
		return nil, fmt.Errorf("invalid password for user %s", login)
	}

	return user, nil
}

func GetUserByJWT(ctx *appctx.AppCtx, token string) (*User, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("empty JWT")
	}

	parsedToken, err := jwt.ParseWithClaims(token, &UserJWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(ctx.Config().JWTKey), nil
	}, jwt.WithLeeway(time.Duration(20*time.Minute)), jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenMalformed):
			return nil, fmt.Errorf("malformed JWT token: %w", err)
		case errors.Is(err, jwt.ErrTokenSignatureInvalid):
			return nil, fmt.Errorf("invalid JWT token signature: %w", err)
		case errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet):
			return nil, fmt.Errorf("token expired or not yet valid : %w", err)
		default:
			return nil, fmt.Errorf("failed to parse JWT: %w", err)
		}
	}

	if !parsedToken.Valid {
		return nil, errors.New("invalid JWT")
	}

	claims, ok := parsedToken.Claims.(*UserJWTClaims)
	if !ok {
		return nil, errors.New("invalid JWT claims format")
	}

	return getUserByLogin(ctx, claims.Login)
}

func getUserByLogin(ctx *appctx.AppCtx, login string) (*User, error) {
	var user *User
	// Take() ищет без сортировки, в отличии от First() или Last()
	// Find() не возвращает ошибку ErrRecordNotFound. Принимает как одну струкутуру так и срез
	if err := ctx.DB().Where("login = ?", login).Take(user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting user: %w", err)
	}

	return user, nil
}

func GetUserFromContext(ctx *appctx.AppCtx) *User {
	u, ok := ctx.Value(userCtxKey{}).(*User)
	if !ok {
		return nil
	}
	return u
}

func (u *User) WithContext(ctx *appctx.AppCtx) *appctx.AppCtx {
	return ctx.WithValue(userCtxKey{}, u)
}

func (u *User) GenerateJWT(ctx *appctx.AppCtx) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &UserJWTClaims{
		Login: u.Login,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ctx.Config().JWTDuration)),
		},
	})

	res, err := token.SignedString([]byte(ctx.Config().JWTKey))
	if err != nil {
		return "", fmt.Errorf("generating JWT: %w", err)
	}

	return res, nil
}

func (u *User) Create(ctx *appctx.AppCtx) error {
	if u.Password == "" {
		return errors.New("empty password")
	}

	var err error
	u.HashedPassword, err = utils.GenerateBcryptHash(ctx, u.Password)
	if err != nil {
		return fmt.Errorf("generating bcrypt hash: %w", err)
	}

	if err := ctx.DB().Create(u).Error; err != nil {
		return fmt.Errorf("creating user: %w", err)
	}

	return nil
}

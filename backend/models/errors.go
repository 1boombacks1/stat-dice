package models

import "errors"

var (
	ErrPlayersNotFound = errors.New("no players found")
	ErrUserNotFound    = errors.New("user not found")
)

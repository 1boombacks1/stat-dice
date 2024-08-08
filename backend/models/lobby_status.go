package models

type LobbyStatus uint8

const (
	LOBBY_STATUS_OPEN LobbyStatus = iota
	LOBBY_STATUS_PROCESSING
	LOBBY_STATUS_CLOSED
)

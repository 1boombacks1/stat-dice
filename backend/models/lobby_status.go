package models

type LobbyStatus uint8

const (
	LOBBY_STATUS_OPEN LobbyStatus = iota
	LOBBY_STATUS_PROCESSING
	LOBBY_STATUS_RESULT
	LOBBY_STATUS_CLOSED
)

func (l LobbyStatus) String() string {
	switch l {
	case LOBBY_STATUS_OPEN:
		return "OPEN"
	case LOBBY_STATUS_PROCESSING:
		return "PROCESSING"
	case LOBBY_STATUS_RESULT:
		return "RESULT"
	case LOBBY_STATUS_CLOSED:
		return "CLOSED"
	default:
		return "UNKNOWN"
	}
}

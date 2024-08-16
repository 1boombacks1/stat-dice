package models

type ResultStatus uint8

const (
	RESULT_STATUS_WIN = iota
	RESULT_STATUS_LOSE
	RESULT_STATUS_LEAVE
	RESULT_STATUS_PLAYING
)

func (r ResultStatus) String() string {
	switch r {
	case RESULT_STATUS_WIN:
		return "Win"
	case RESULT_STATUS_LOSE:
		return "Lose"
	case RESULT_STATUS_LEAVE:
		return "Leave"
	case RESULT_STATUS_PLAYING:
		return "Playing"
	default:
		return "unknown"
	}
}

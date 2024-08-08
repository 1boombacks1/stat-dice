package models

type ResultStatus uint8

const (
	RESULT_STATUS_WIN = iota
	RESULT_STATUS_LOSE
	RESULT_STATUS_LEAVE
	RESULT_STATUS_PLAYING
)

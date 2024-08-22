package models

import (
	"fmt"

	"github.com/1boombacks1/stat_dice/appctx"
)

type PlayerStat struct {
	Name  string
	Total int
	Win   int
	Lose  int
}

func (s *PlayerStat) FormatNum(num int) string {
	return fmt.Sprintf("%02d", num)
}

func (s *PlayerStat) IsChampion(status ResultStatus, max int) bool {
	switch status {
	case RESULT_STATUS_WIN:
		return max == s.Win
	case RESULT_STATUS_LOSE:
		return max == s.Lose
	default:
		return max == s.Total
	}
}

func GetFilterStats(ctx *appctx.AppCtx, orderQuery string) ([]PlayerStat, error) {
	var stats []PlayerStat
	err := ctx.DB().Model(&User{}).
		Select(
			"users.name as name, "+
				"COUNT(matches.user_id) as total, "+
				"SUM(CASE WHEN matches.result = ? THEN 1 ELSE 0 END) as win, "+
				"SUM(CASE WHEN matches.result = ? THEN 1 ELSE 0 END) as lose",
			RESULT_STATUS_WIN, RESULT_STATUS_LOSE,
		).
		Joins("JOIN matches ON matches.user_id = users.id").
		Group("users.name").
		Order(orderQuery).
		Scan(&stats).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get filter stats with query='%s': %v", orderQuery, err)
	}

	return stats, nil
}

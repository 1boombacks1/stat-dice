package db

import (
	"fmt"
)

func (db *DB) GetVersion() (string, error) {
	var version string
	err := db.Raw("SELECT version();").Scan(&version).Error
	if err != nil {
		return "", fmt.Errorf("getting database version: %w", err)
	}

	return version, nil
}

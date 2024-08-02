package models

type User struct {
	Base

	Login          string `gorm:"unique;not null"`
	Password       string `gorm:"-:all"`
	HashedPassword string `gorm:"not null"`
	Name           string `gorm:"not null"`
}

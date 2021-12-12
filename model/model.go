package model

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	gorm.Model
	Username string `gorm:"unique"`
	Password string
	Email    string `gorm:"unique"`
	Nickname string
	Avatar   string
}

type Session struct {
	gorm.Model
	UserID       uint
	User         User
	SessionID    string `gorm:"unique"`
	UserAgent    string
	LastAccessed time.Time
}

func Bind(db *gorm.DB) {
	_ = db.AutoMigrate(&User{})
	_ = db.AutoMigrate(&Session{})
}

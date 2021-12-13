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
	IP           string
}

type Token struct {
	gorm.Model
	TokenID  string `gorm:"unique"`
	Name     string
	URL      string
	Token    string
	Priority int
	UserID   uint
	User     User
}

func Bind(db *gorm.DB) {
	_ = db.AutoMigrate(&User{})
	_ = db.AutoMigrate(&Session{})
	_ = db.AutoMigrate(&Token{})
}

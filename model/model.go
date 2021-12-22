package model

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	gorm.Model
	Password  string
	Email     string `gorm:"unique"`
	Nickname  string
	Avatar    string
	Enable2FA bool
	Secret    string
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

type Playlist struct {
	gorm.Model
	Name         string
	Description  string
	UserID       uint
	User         User
	IsPublic     bool
	CoverAlbumID string
	CoverDiscID  int
}

type PlaylistSong struct {
	gorm.Model
	PlaylistID  uint
	Playlist    Playlist
	AlbumID     string
	DiscID      int
	TrackID     int
	Description string
	Order       uint
}

type Share struct {
	gorm.Model
	UserID  uint
	User    User
	Data    string
	ShareID string `gorm:"unique"`
}

type FavoriteMusic struct {
	UserID  uint
	User    User
	AlbumID string
	DiscID  int
	TrackID int
}

type FavoritePlaylist struct {
	UserID     uint
	User       User
	PlaylistID uint
	Playlist   Playlist
}

func Bind(db *gorm.DB) {
	_ = db.AutoMigrate(&User{})
	_ = db.AutoMigrate(&Session{})
	_ = db.AutoMigrate(&Token{})
	_ = db.AutoMigrate(&Playlist{})
	_ = db.AutoMigrate(&PlaylistSong{})
	_ = db.AutoMigrate(&Share{})
	_ = db.AutoMigrate(&FavoriteMusic{})
	_ = db.AutoMigrate(&FavoritePlaylist{})
}

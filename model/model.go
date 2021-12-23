package model

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	gorm.Model
	Password  string
	Email     string `gorm:"uniqueIndex"`
	Nickname  string
	Avatar    string
	Enable2FA bool
	Secret    string
}

type Session struct {
	gorm.Model
	UserID       uint
	User         User
	SessionID    string `gorm:"uniqueIndex"`
	UserAgent    string
	LastAccessed time.Time
	IP           string
}

type Token struct {
	gorm.Model
	TokenID  string `gorm:"uniqueIndex"`
	Name     string
	URL      string
	Token    string
	Priority int
	UserID   uint `gorm:"index"`
	User     User
}

type Playlist struct {
	gorm.Model
	Name         string
	Description  string
	UserID       uint `gorm:"index"`
	User         User
	IsPublic     bool
	CoverAlbumID string
	CoverDiscID  int
}

type PlaylistSong struct {
	gorm.Model
	PlaylistID  uint `gorm:"uniqueIndex:playlist_song_index"`
	Playlist    Playlist
	AlbumID     string `gorm:"uniqueIndex:playlist_song_index"`
	DiscID      int    `gorm:"uniqueIndex:playlist_song_index"`
	TrackID     int    `gorm:"uniqueIndex:playlist_song_index"`
	Description string
	Order       uint
}

type Share struct {
	gorm.Model
	UserID  uint
	User    User
	Data    string
	ShareID string `gorm:"uniqueIndex"`
}

type FavoriteMusic struct {
	UserID  uint `gorm:"uniqueIndex:favorite_music_index"`
	User    User
	AlbumID string `gorm:"uniqueIndex:favorite_music_index"`
	DiscID  int    `gorm:"uniqueIndex:favorite_music_index"`
	TrackID int    `gorm:"uniqueIndex:favorite_music_index"`
}

type FavoritePlaylist struct {
	UserID     uint `gorm:"uniqueIndex:favorite_playlist_index"`
	User       User
	PlaylistID uint `gorm:"uniqueIndex:favorite_playlist_index"`
	Playlist   Playlist
}

type Lyric struct {
	gorm.Model
	AlbumID  string `gorm:"uniqueIndex:lyric_index"`
	DiscID   int    `gorm:"uniqueIndex:lyric_index"`
	TrackID  int    `gorm:"uniqueIndex:lyric_index"`
	Language string `gorm:"uniqueIndex:lyric_index"`
	Type     string `gorm:"check:type='text' OR type='lrc'"`
	Data     string
	UserID   uint
	User     User
	Source   bool // TODO missing constraint
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
	_ = db.AutoMigrate(&Lyric{})
}

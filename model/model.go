package model

import (
	"errors"
	"time"

	"github.com/ProjectAnni/anniv-go/meta"
	"gorm.io/gorm"
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
	TokenID    string `gorm:"uniqueIndex"`
	Name       string
	URL        string
	Token      string
	Priority   int
	UserID     uint `gorm:"index"`
	User       User
	Controlled bool
}

type Playlist struct {
	gorm.Model
	Name         string
	Description  string
	UserID       uint `gorm:"index"`
	User         User
	IsPublic     bool
	CoverAlbumID string
	CoverDiscID  uint
}

type PlaylistSong struct {
	gorm.Model
	PlaylistID  uint
	Playlist    Playlist
	AlbumID     string
	DiscID      uint
	TrackID     uint
	Description string
	Type        string         `gorm:"check:type='normal' OR type='dummy' OR type='album'"`
	TrackInfo   meta.TrackInfo `gorm:"embedded;embeddedPrefix:track_info_"`
	Order       uint
}

type Share struct {
	gorm.Model
	UserID     uint
	User       User
	Info       []byte
	TokenGrant []byte
	ShareID    string `gorm:"uniqueIndex"`
}

type FavoriteMusic struct {
	gorm.Model
	UserID  uint `gorm:"uniqueIndex:favorite_music_index"`
	User    User
	AlbumID string `gorm:"uniqueIndex:favorite_music_index"`
	DiscID  uint   `gorm:"uniqueIndex:favorite_music_index"`
	TrackID uint   `gorm:"uniqueIndex:favorite_music_index"`
}

type FavoritePlaylist struct {
	gorm.Model
	UserID     uint `gorm:"uniqueIndex:favorite_playlist_index"`
	User       User
	PlaylistID uint `gorm:"uniqueIndex:favorite_playlist_index"`
	Playlist   Playlist
}

type FavoriteAlbum struct {
	gorm.Model
	UserID  uint `gorm:"uniqueIndex:favorite_album_index"`
	User    User
	AlbumID string `gorm:"uniqueIndex:favorite_album_index"`
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
	// Is track original language
	Source bool
	// Lyric source
	LyricSource string
}

type PlayRecord struct {
	gorm.Model
	UserID uint
	User   User
	Track  meta.TrackIdentifier `gorm:"column:track;type:text"`
	At     time.Time
}

func (l *Lyric) BeforeCreate(tx *gorm.DB) error {
	var cnt int64
	err := tx.Model(&Lyric{}).
		Where("album_id = ? AND disc_id = ? AND track_id = ?", l.AlbumID, l.DiscID, l.TrackID).
		Count(&cnt).Error
	if err != nil {
		return err
	}
	if cnt == 0 && !l.Source {
		return errors.New("cannot create translation before source is defined")
	}
	if cnt != 0 && l.Source {
		return errors.New("duplicated source lang")
	}
	return nil
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&Session{},
		&Token{},
		&Playlist{},
		&PlaylistSong{},
		&Share{},
		&FavoriteMusic{},
		&FavoritePlaylist{},
		&FavoriteAlbum{},
		&Lyric{},
		&PlayRecord{},
	)
}

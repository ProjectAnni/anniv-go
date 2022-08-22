package services

import (
	"fmt"
	"github.com/ProjectAnni/anniv-go/model"
	"strconv"
)

type Response struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func resOk(data interface{}) Response {
	return Response{
		Status:  StatusOK,
		Message: "OK",
		Data:    data,
	}
}

func illegalParams(msg string) Response {
	return Response{
		Status:  IllegalParams,
		Message: msg,
		Data:    nil,
	}
}

func writeErr(err error) Response {
	return Response{
		Status:  WriteErr,
		Message: fmt.Sprintf("%v", err),
		Data:    nil,
	}
}

func readErr(err error) Response {
	return Response{
		Status:  ReadErr,
		Message: fmt.Sprintf("%v", err),
		Data:    nil,
	}
}

func resErr(status int, msg string) Response {
	return Response{
		Status:  status,
		Message: msg,
		Data:    nil,
	}
}

type UserInfo struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	Enable2FA bool   `json:"2fa_enabled"`
}

func userInfo(u model.User) UserInfo {
	return UserInfo{
		UserID:    strconv.Itoa(int(u.ID)),
		Email:     u.Email,
		Nickname:  u.Nickname,
		Avatar:    u.Avatar,
		Enable2FA: u.Enable2FA,
	}
}

type UserIntro struct {
	UserID   string `json:"user_id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

func userIntro(u model.User) UserIntro {
	return UserIntro{
		UserID:   strconv.Itoa(int(u.ID)),
		Nickname: u.Nickname,
		Avatar:   u.Avatar,
	}
}

type RegisterForm struct {
	Password   string `json:"password"`
	Email      string `json:"email"`
	Nickname   string `json:"nickname"`
	Avatar     string `json:"avatar"`
	Secret     string `json:"2fa_secret"`
	Code       string `json:"2fa_code"`
	InviteCode string `json:"invite_code"`
}

type LoginForm struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Code     string `json:"2fa_code"`
}

type ChangePasswordForm struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type UserIntroForm struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

type DeleteForm struct {
	ID string `json:"id"`
}

type Enable2FAForm struct {
	Secret string `json:"2fa_secret"`
	Code   string `json:"2fa_code"`
}

type Cover struct {
	AlbumID *string `json:"album_id" mapstructure:"album_id"`
	DiscID  *uint   `json:"disc_id" mapstructure:"disc_id"`
}

type PlaylistResult struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Owner       string `json:"owner"`
}

type FavoriteMusicEntry struct {
	AlbumID string `json:"album_id"`
	DiscID  int    `json:"disc_id"`
	TrackID int    `json:"track_id"`
}

type FavoritePlaylistEntry struct {
	PlaylistID string `json:"playlist_id"`
	Name       string `json:"name"`
	Owner      string `json:"owner"`
}

type FavoritePlaylistForm struct {
	PlaylistID string `json:"playlist_id"`
}

type LyricResponse struct {
	Source       LyricLanguage   `json:"source"`
	Translations []LyricLanguage `json:"translations"`
}

type LyricLanguage struct {
	Language    string    `json:"language"`
	Type        string    `json:"type"`
	Data        string    `json:"data"`
	Contributor UserIntro `json:"contributor"`
	Source      string    `json:"source"`
}

func lyricLanguage(lyric model.Lyric) LyricLanguage {
	return LyricLanguage{
		Language:    lyric.Language,
		Type:        lyric.Type,
		Data:        lyric.Data,
		Contributor: userIntro(lyric.User),
		Source:      lyric.LyricSource,
	}
}

func lyricLanguages(lyrics []model.Lyric) []LyricLanguage {
	res := make([]LyricLanguage, 0, len(lyrics))
	for _, v := range lyrics {
		res = append(res, LyricLanguage{
			Language:    v.Language,
			Type:        v.Type,
			Data:        v.Data,
			Contributor: userIntro(v.User),
		})
	}
	return res
}

type LyricPatchForm struct {
	AlbumID string `json:"album_id"`
	DiscID  int    `json:"disc_id"`
	TrackID int    `json:"track_id"`
	Type    string `json:"type"`
	Lang    string `json:"lang"`
	Data    string `json:"data"`
}

package services

import (
	"fmt"
	"github.com/ProjectAnni/anniv-go/config"
	"github.com/ProjectAnni/anniv-go/meta"
	"github.com/ProjectAnni/anniv-go/model"
	"github.com/pelletier/go-toml/v2"
	"strconv"
	time2 "time"
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

type SiteInfo struct {
	SiteName        string   `json:"site_name"`
	Description     string   `json:"description"`
	ProtocolVersion string   `json:"protocol_version"`
	Features        []string `json:"features"`
}

func siteInfo() SiteInfo {
	features := make([]string, 0)
	if config.Cfg.Enforce2FA {
		features = append(features, "2fa_enforced")
	} else {
		features = append(features, "2fa")
	}
	return SiteInfo{
		SiteName:        config.Cfg.SiteName,
		Description:     config.Cfg.Description,
		ProtocolVersion: "1",
		Features:        features,
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
	Password string `json:"password"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Secret   string `json:"2fa_secret"`
	Code     string `json:"2fa_code"`
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

type TokenInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	Token    string `json:"token"`
	Priority int    `json:"priority"`
}

func tokenInfo(token model.Token) TokenInfo {
	return TokenInfo{
		ID:       token.TokenID,
		Name:     token.Name,
		URL:      token.URL,
		Token:    token.Token,
		Priority: token.Priority,
	}
}

func tokenInfos(tokens []model.Token) []TokenInfo {
	ret := make([]TokenInfo, 0, len(tokens))
	for _, v := range tokens {
		ret = append(ret, tokenInfo(v))
	}
	return ret
}

type TokenForm struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Token    string `json:"token"`
	Priority int    `json:"priority"`
}

type TokenPatchForm struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	Token    string `json:"token"`
	Priority int    `json:"priority"`
}

type DeleteForm struct {
	ID string `json:"id"`
}

type Enable2FAForm struct {
	Secret string `json:"2fa_secret"`
	Code   string `json:"2fa_code"`
}

type PlaylistSong struct {
	ID          string `json:"id"`
	AlbumID     string `json:"album_id"`
	DiscID      int    `json:"disc_id"`
	TrackID     int    `json:"track_id"`
	Description string `json:"description"`
}

type Cover struct {
	AlbumID string `json:"album_id"`
	DiscID  *int   `json:"disc_id"`
}

type Playlist struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Owner       string         `json:"owner"`
	IsPublic    bool           `json:"is_public"`
	Cover       Cover          `json:"cover"`
	Songs       []PlaylistSong `json:"songs"`
}

type PlaylistSongForm struct {
	AlbumID     string `json:"album_id"`
	DiscID      int    `json:"disc_id"`
	TrackID     int    `json:"track_id"`
	Description string `json:"description"`
}

type PlaylistForm struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	IsPublic    bool               `json:"is_public"`
	Cover       Cover              `json:"cover"`
	Songs       []PlaylistSongForm `json:"songs"`
}

type PlaylistPatchForm struct {
	ID      string      `json:"id"`
	Command string      `json:"command"`
	Payload interface{} `json:"payload"`
}

type AlbumInfo struct {
	AlbumID string           `json:"album_id"`
	Title   string           `json:"title"`
	Edition *string          `json:"edition"`
	Catalog string           `json:"catalog"`
	Artist  string           `json:"artist"`
	Date    string           `json:"date"`
	Tags    []string         `json:"tags"`
	Type    string           `json:"type"`
	Discs   []*meta.DiscInfo `json:"discs"`
}

func albumInfo(album meta.AlbumInfo) *AlbumInfo {
	ret := &AlbumInfo{
		AlbumID: album.AlbumID,
		Title:   album.Title,
		Edition: album.Edition,
		Catalog: album.Catalog,
		Artist:  album.Artist,
		Date:    parseDate(album.Date),
		Tags:    album.Tags,
		Type:    album.Type,
		Discs:   album.Discs,
	}
	if ret.Tags == nil {
		ret.Tags = []string{}
	}
	return ret
}

func parseDate(v interface{}) string {
	str, ok := v.(string)
	if ok {
		return str
	}
	date, ok := v.(toml.LocalDate)
	if ok {
		return date.String()
	}
	time, ok := v.(time2.Time)
	if ok {
		return time.Format("2006-01-02")
	}
	m, ok := v.(map[string]int)
	if ok {
		year := m["year"]
		month := m["month"]
		day, containsDay := m["day"]
		ret := strconv.Itoa(year) + "-" + strconv.Itoa(month)
		if containsDay {
			ret += "-" + strconv.Itoa(day)
		}
		return ret
	}
	return ""
}

type PlaylistResult struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Owner       string `json:"owner"`
}

type SearchResult struct {
	Albums    []*AlbumInfo               `json:"albums"`
	Tracks    []*meta.TrackInfoWithAlbum `json:"tracks"`
	Playlists []*PlaylistResult          `json:"playlists"`
}

type ShareEntry struct {
	ID   string `json:"id"`
	Date int64  `json:"date"`
}

type CreateShareForm struct {
	Data string `json:"data"`
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
}

func lyricLanguage(lyric model.Lyric) LyricLanguage {
	return LyricLanguage{
		Language:    lyric.Language,
		Type:        lyric.Type,
		Data:        lyric.Data,
		Contributor: userIntro(lyric.User),
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

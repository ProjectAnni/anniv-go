package services

import (
	"fmt"
	"github.com/ProjectAnni/anniv-go/config"
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
	Username  string `json:"username"`
	Email     string `json:"email"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	Enable2FA bool   `json:"2fa_enabled"`
}

func userInfo(u model.User) UserInfo {
	return UserInfo{
		UserID:    strconv.Itoa(int(u.ID)),
		Username:  u.Username,
		Email:     u.Email,
		Nickname:  u.Nickname,
		Avatar:    u.Avatar,
		Enable2FA: u.Enable2FA,
	}
}

type RegisterForm struct {
	Username string `json:"username"`
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

type DeleteTokenForm struct {
	ID string `json:"id"`
}

type Enable2FAForm struct {
	Secret string `json:"2fa_secret"`
	Code   string `json:"2fa_code"`
}

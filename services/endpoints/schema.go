package endpoints

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
	return SiteInfo{
		SiteName:        config.Cfg.SiteName,
		Description:     config.Cfg.Description,
		ProtocolVersion: "1",
		Features:        []string{},
	}
}

type UserInfo struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

func userInfo(u model.User) UserInfo {
	return UserInfo{
		UserID:   strconv.Itoa(int(u.ID)),
		Username: u.Username,
		Email:    u.Email,
		Nickname: u.Nickname,
		Avatar:   u.Avatar,
	}
}

type RegisterForm struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

type LoginForm struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

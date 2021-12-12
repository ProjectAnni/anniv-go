package endpoints

import (
	"github.com/ProjectAnni/anniv-go/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"regexp"
	"strings"
)

var usernameReg = regexp.MustCompile("^[A-Za-z0-9-_]{5,11}$")

func User(ng *gin.Engine, db *gorm.DB) {
	g := ng.Group("/api/user")

	g.POST("/register", func(ctx *gin.Context) {
		form := RegisterForm{}
		if err := ctx.Bind(&form); err != nil {
			ctx.JSON(http.StatusOK, illegalParams("malformed register form"))
			return
		}
		if !usernameReg.MatchString(form.Username) || strings.Contains(form.Username, "\n") {
			ctx.JSON(http.StatusOK, illegalUsername("Username must match [A-Za-z0-9-_]{5,11}"))
			return
		}
		if db.Where("username = ?", form.Username).First(&model.User{}).RowsAffected != 0 {
			ctx.JSON(http.StatusOK, illegalParams("username already taken"))
			return
		}
		// TODO email nickname avatar check
		user := model.User{
			Username: form.Username,
			Password: form.Password,
			Email:    form.Email,
			Nickname: form.Nickname,
			Avatar:   form.Avatar,
		}
		err := db.Create(&user).Error
		if err != nil {
			ctx.JSON(http.StatusOK, writeErr(err))
			return
		}
		ctx.JSON(http.StatusOK, resOk(nil))
	})
}

func illegalUsername(msg string) Response {
	return Response{
		Status:  UsernameUnavailable,
		Message: msg,
		Data:    nil,
	}
}

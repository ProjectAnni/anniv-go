package endpoints

import (
	"encoding/base64"
	"github.com/ProjectAnni/anniv-go/model"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
	"regexp"
	"strings"
)

var usernameReg = regexp.MustCompile("^[A-Za-z0-9-_]{5,11}$")
var emailReg = regexp.MustCompile("^[A-Za-z0-9.]@((?!-))(xn--)?[a-z0-9][a-z0-9-_]{0,61}[a-z0-9]{0,1}\\.(xn--)?([a-z0-9\\-]{1,61}|[a-z0-9-]{1,30}\\.[a-z]{2,})$")

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
			ctx.JSON(http.StatusOK, illegalUsername("username already taken"))
			return
		}
		if !emailReg.MatchString(form.Email) || strings.Contains(form.Email, "\n") {
			ctx.JSON(http.StatusOK, illegalEmail("invalid email"))
			return
		}
		if db.Where("email = ?", form.Email).First(&model.User{}).RowsAffected != 0 {
			ctx.JSON(http.StatusOK, illegalEmail("email already taken"))
			return
		}
		// TODO nickname avatar check
		hash, err := bcrypt.GenerateFromPassword([]byte(form.Password), bcrypt.DefaultCost)
		if err != nil {
			ctx.JSON(http.StatusOK, resErr(InternalError, "error hashing password"))
			return
		}
		user := model.User{
			Username: form.Username,
			Password: base64.StdEncoding.EncodeToString(hash),
			Email:    form.Email,
			Nickname: form.Nickname,
			Avatar:   form.Avatar,
		}
		err = db.Create(&user).Error
		if err != nil {
			ctx.JSON(http.StatusOK, writeErr(err))
			return
		}
		ctx.JSON(http.StatusOK, resOk(nil))
	})

	g.POST("login", func(ctx *gin.Context) {
		// TODO
	})
}

func illegalUsername(msg string) Response {
	return Response{
		Status:  UsernameUnavailable,
		Message: msg,
		Data:    nil,
	}
}

func illegalEmail(msg string) Response {
	return Response{
		Status:  EmailUnavailable,
		Message: msg,
		Data:    nil,
	}
}

package services

import (
	"github.com/ProjectAnni/anniv-go/model"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"regexp"
	"time"
)

var usernameReg = regexp.MustCompile("^[A-Za-z0-9-_]{5,11}$")
var emailReg = regexp.MustCompile("^(?:[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*|\"(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21\\x23-\\x5b\\x5d-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])*\")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\\[(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9])\\.){3}(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9])|[a-z0-9-]*[a-z0-9]:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21-\\x5a\\x53-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])+)])$")

func EndpointUser(ng *gin.Engine) {
	g := ng.Group("/api/user")

	g.POST("/register", func(ctx *gin.Context) {
		form := RegisterForm{}
		if err := ctx.BindJSON(&form); err != nil {
			ctx.JSON(http.StatusOK, illegalParams("malformed register form"))
			return
		}
		if !usernameReg.MatchString(form.Username) {
			ctx.JSON(http.StatusOK, illegalUsername("Username must match [A-Za-z0-9-_]{5,11}"))
			return
		}
		if db.Where("username = ?", form.Username).First(&model.User{}).RowsAffected != 0 {
			ctx.JSON(http.StatusOK, illegalUsername("username already taken"))
			return
		}
		if !emailReg.MatchString(form.Email) {
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
			Password: string(hash),
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

	g.POST("/login", func(ctx *gin.Context) {
		form := LoginForm{}
		err := ctx.BindJSON(&form)
		if err != nil {
			ctx.JSON(http.StatusOK, illegalParams("malformed login form"))
			return
		}
		user := model.User{}
		if db.Where("email = ?", form.Email).First(&user).RowsAffected == 0 {
			ctx.JSON(http.StatusOK, userNotFound())
			return
		}
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(form.Password))
		if err != nil {
			ctx.JSON(http.StatusOK, wrongPassword())
			return
		}
		session := model.Session{
			UserID:       user.ID,
			User:         user,
			SessionID:    uuid.NewV4().String(),
			UserAgent:    ctx.Request.UserAgent(),
			LastAccessed: time.Now(),
			IP:           ctx.ClientIP(),
		}
		if err = db.Save(&session).Error; err != nil {
			ctx.JSON(http.StatusOK, writeErr(err))
			return
		}
		ctx.SetCookie("session", session.SessionID, 86400, "/", "", true, true)
		ctx.JSON(http.StatusOK, resOk(userInfo(user)))
	})

	g.POST("/logout", AuthRequired, func(ctx *gin.Context) {
		session := ctx.MustGet("session").(model.Session)
		if err := db.Delete(&session).Error; err != nil {
			ctx.JSON(http.StatusOK, writeErr(err))
			return
		}
		ctx.SetCookie("session", "", -1, "", "", true, true)
		ctx.JSON(http.StatusOK, resOk(nil))
	})

	g.POST("/revoke", AuthRequired, func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		if err := db.Delete(&user).Error; err != nil {
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

func illegalEmail(msg string) Response {
	return Response{
		Status:  EmailUnavailable,
		Message: msg,
		Data:    nil,
	}
}

func userNotFound() Response {
	return Response{
		Status:  UserNotExist,
		Message: "user not exist",
		Data:    nil,
	}
}

func wrongPassword() Response {
	return Response{
		Status:  InvalidPassword,
		Message: "email and password does not match",
		Data:    nil,
	}
}

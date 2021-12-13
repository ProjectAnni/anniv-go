package endpoints

import (
	"github.com/ProjectAnni/anniv-go/model"
	"github.com/ProjectAnni/anniv-go/services"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var usernameReg = regexp.MustCompile("^[A-Za-z0-9-_]{5,11}$")
var emailReg = regexp.MustCompile("^[A-Za-z0-9.]@((?!-))(xn--)?[a-z0-9][a-z0-9-_]{0,61}[a-z0-9]?\\.(xn--)?([a-z0-9\\-]{1,61}|[a-z0-9-]{1,30}\\.[a-z]{2,})$")

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
		err := ctx.Bind(&form)
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
		}
		if err = db.Save(&session).Error; err != nil {
			ctx.JSON(http.StatusOK, writeErr(err))
			return
		}
		ctx.SetCookie("session", session.SessionID, 86400, "/", "", true, true)
		ctx.JSON(http.StatusOK, resOk(userInfo(user)))
	})

	g.POST("/logout", services.AuthRequired, func(ctx *gin.Context) {
		ctx.SetCookie("session", "", -1, "", "", true, true)
		ctx.Status(http.StatusNoContent)
	})

	g.POST("/revoke", services.AuthRequired, func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		if err := db.Delete(&user).Error; err != nil {
			ctx.JSON(http.StatusOK, writeErr(err))
			return
		}
		ctx.Status(http.StatusNoContent)
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

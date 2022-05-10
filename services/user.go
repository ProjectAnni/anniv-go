package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/ProjectAnni/anniv-go/config"
	"github.com/ProjectAnni/anniv-go/model"
	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

var emailReg = regexp.MustCompile("^(?:[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*|\"(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21\\x23-\\x5b\\x5d-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])*\")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\\[(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9])\\.){3}(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9])|[a-z0-9-]*[a-z0-9]:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21-\\x5a\\x53-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])+)])$")

func EndpointUser(ng *gin.Engine) {
	g := ng.Group("/api/user")

	g.POST("/register", func(ctx *gin.Context) {
		form := RegisterForm{}
		if err := ctx.ShouldBind(&form); err != nil {
			ctx.JSON(http.StatusOK, illegalParams("malformed register form"))
			return
		}
		if config.Cfg.RequireInvite {
			if form.InviteCode != config.Cfg.InviteCode {
				ctx.JSON(http.StatusOK, resErr(InvalidInviteCode, "invalid invite code"))
				return
			}
		}
		if len(form.Nickname) > NicknameMaxLen {
			ctx.JSON(http.StatusOK, resErr(InvalidNickname, "nickname too long"))
			return
		}
		if !emailReg.MatchString(form.Email) {
			ctx.JSON(http.StatusOK, illegalEmail("invalid email"))
			return
		}
		if db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).
			Where("email = ?", form.Email).First(&model.User{}).RowsAffected != 0 {
			ctx.JSON(http.StatusOK, illegalEmail("email already taken"))
			return
		}
		if config.Cfg.Enforce2FA && form.Secret == "" {
			ctx.JSON(http.StatusOK, Response{
				Status:  Illegal2FASecret,
				Message: "2fa enforced",
				Data:    nil,
			})
			return
		}
		if form.Secret != "" && !totp.Validate(form.Code, form.Secret) {
			ctx.JSON(http.StatusOK, wrong2FACode())
			return
		}
		// TODO avatar check
		hash, err := bcrypt.GenerateFromPassword([]byte(form.Password), bcrypt.DefaultCost)
		if err != nil {
			ctx.JSON(http.StatusOK, resErr(InternalError, "error hashing password"))
			return
		}
		user := model.User{
			Password:  string(hash),
			Email:     form.Email,
			Nickname:  form.Nickname,
			Avatar:    form.Avatar,
			Enable2FA: form.Secret != "",
			Secret:    form.Secret,
		}
		tokens, err := signUserTokens(form.Email)
		if err != nil {
			// TODO: decide which error code to use
			ctx.JSON(http.StatusOK, resErr(0, "failed to sign default tokens"))
		}
		err = db.Transaction(func(tx *gorm.DB) error {
			err := tx.Create(&user).Error
			if err != nil {
				return err
			}
			for _, v := range tokens {
				t := model.Token{
					TokenID:    uuid.NewV4().String(),
					Name:       v.Name,
					URL:        v.URL,
					Token:      v.Token,
					Priority:   v.Priority,
					UserID:     user.ID,
					Controlled: true,
				}
				err = tx.Create(&t).Error
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			ctx.JSON(http.StatusOK, writeErr(err))
			return
		}
		ctx.JSON(http.StatusOK, resOk(nil))
	})

	g.POST("/login", func(ctx *gin.Context) {
		form := LoginForm{}
		err := ctx.ShouldBind(&form)
		if err != nil {
			ctx.JSON(http.StatusOK, illegalParams("malformed login form"))
			return
		}
		user := model.User{}
		if db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).
			Where("email = ?", form.Email).First(&user).RowsAffected == 0 {
			ctx.JSON(http.StatusOK, userNotFound())
			return
		}
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(form.Password))
		if err != nil {
			ctx.JSON(http.StatusOK, wrongPassword())
			return
		}
		if user.Enable2FA && !totp.Validate(form.Code, user.Secret) {
			ctx.JSON(http.StatusOK, wrong2FACode())
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

	g.POST("/revoke", AuthRequired, TFARequired, func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		if err := db.Delete(&user).Error; err != nil {
			ctx.JSON(http.StatusOK, writeErr(err))
			return
		}
		ctx.JSON(http.StatusOK, resOk(nil))
	})

	g.GET("/info", AuthRequired, func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		ctx.JSON(http.StatusOK, resOk(userInfo(user)))
	})

	g.GET("/intro", AuthRequired, func(ctx *gin.Context) {
		id := ctx.Query("user_id")
		uid, err := strconv.Atoi(id)
		if err != nil {
			ctx.JSON(http.StatusOK, userNotFound())
			return
		}
		var u model.User
		if err := db.Where("id = ?", uid).First(&u).Error; err != nil {
			ctx.JSON(http.StatusOK, userNotFound())
			return
		}
		ctx.JSON(http.StatusOK, resOk(userIntro(u)))
	})

	g.PATCH("/password", AuthRequired, TFARequired, func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		form := ChangePasswordForm{}
		if err := ctx.ShouldBind(&form); err != nil {
			ctx.JSON(http.StatusOK, illegalParams("malformed password form"))
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(form.OldPassword)); err != nil {
			ctx.JSON(http.StatusOK, wrongPassword())
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(form.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			ctx.JSON(http.StatusOK, illegalParams("failed to hash password"))
			return
		}
		user.Password = string(hash)
		err = db.Transaction(func(tx *gorm.DB) error {
			err := tx.Save(&user).Error
			if err != nil {
				return err
			}
			return tx.Where("user_id=?", user.ID).Unscoped().Delete(&model.Session{}).Error
		})
		if err != nil {
			ctx.JSON(http.StatusOK, writeErr(err))
			return
		}
		ctx.JSON(http.StatusOK, resOk(nil))
	})

	g.PATCH("/intro", AuthRequired, func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		form := UserIntroForm{}
		if err := ctx.ShouldBind(&form); err != nil {
			ctx.JSON(http.StatusOK, illegalParams("malformed intro form"))
			return
		}
		if len(form.Nickname) > NicknameMaxLen {
			ctx.JSON(http.StatusOK, resErr(InvalidNickname, "nickname too long"))
			return
		}
		user.Nickname = form.Nickname
		user.Avatar = form.Avatar
		if err := db.Save(&user).Error; err != nil {
			ctx.JSON(http.StatusOK, writeErr(err))
			return
		}
		ctx.JSON(http.StatusOK, resOk(nil))
	})
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

var client = &http.Client{Timeout: 10}

func signUserTokens(user string) ([]Token, error) {
	res := []Token(nil)
	for _, v := range config.Cfg.AnnilToken {
		if !v.Enabled {
			continue
		}
		payload := map[string]interface{}{
			"user_id": user,
			"share":   v.AllowShare,
		}
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequest(http.MethodPost, v.URL+"/admin/sign", bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != 200 {
			return nil, errors.New("invalid response code")
		}

		tokenBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		_ = resp.Body.Close()
		token := string(tokenBytes)
		res = append(res, Token{
			Name:     v.Name,
			URL:      v.URL,
			Token:    token,
			Priority: 0,
		})
	}
	return res, nil
}

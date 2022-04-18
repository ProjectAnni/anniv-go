package services

import (
	"github.com/ProjectAnni/anniv-go/config"
	"github.com/ProjectAnni/anniv-go/model"
	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"net/http"
	"time"
)

func AuthRequired(ctx *gin.Context) {
	sid, err := ctx.Cookie("session")
	if err != nil {
		ctx.JSON(http.StatusOK, unauthorized())
		ctx.Abort()
		return
	}
	session := model.Session{}
	if db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).
		Preload("User").
		Where("session_id=?", sid).
		First(&session).Error != nil {
		ctx.JSON(http.StatusOK, unauthorized())
		ctx.Abort()
		return
	}
	if time.Now().Sub(session.LastAccessed) > time.Minute*5 ||
		session.IP != ctx.ClientIP() ||
		session.UserAgent != ctx.Request.UserAgent() {
		session.LastAccessed = time.Now()
		session.UserAgent = ctx.Request.UserAgent()
		session.IP = ctx.ClientIP()
		db.Save(&session)
	}
	// Renew cookie
	ctx.SetCookie("session", session.SessionID, 86400*7, "/", "", true, true)
	ctx.Set("user", session.User)
	ctx.Set("session", session)
	if config.Cfg.Enforce2FA {
		enforce2FA(ctx)
	}
}

func unauthorized() Response {
	return Response{
		Status:  Unauthorized,
		Message: "unauthorized",
		Data:    nil,
	}
}

var tfaAttempts map[string]int

func initMiddleware() {
	go func() {
		t := time.NewTicker(time.Hour)
		for {
			<-t.C
			tfaAttempts = make(map[string]int)
		}
	}()
}

func TFARequired(ctx *gin.Context) {
	session := ctx.MustGet("session").(model.Session)
	if tfaAttempts[session.SessionID] >= 5 {
		ctx.JSON(http.StatusOK, Response{
			Status:  TFAAttemptLimited,
			Message: "attempt too frequent",
			Data:    nil,
		})
		ctx.Abort()
		return
	}
	user := ctx.MustGet("user").(model.User)
	if !user.Enable2FA {
		return
	}
	params := map[string]interface{}{}
	if err := ctx.ShouldBind(&params); err != nil {
		ctx.JSON(http.StatusOK, illegalParams("failed to read 2fa token"))
		ctx.Abort()
	}
	code, ok := params["2fa_code"].(string)
	if !ok {
		ctx.JSON(http.StatusOK, wrong2FACode())
		ctx.Abort()
	}
	if !totp.Validate(code, user.Secret) {
		tfaAttempts[session.SessionID]++
		ctx.JSON(http.StatusOK, wrong2FACode())
		ctx.Abort()
	}
}

func enforce2FA(ctx *gin.Context) {
	if whitelistEndpoint(ctx.Request.Method, ctx.Request.URL.Path) {
		return
	}
	user := ctx.MustGet("user").(model.User)
	if user.Enable2FA {
		return
	}
	ctx.JSON(http.StatusOK, Response{
		Status:  TFANotEnabled,
		Message: "2fa is enforced",
		Data:    nil,
	})
	ctx.Abort()
}

func whitelistEndpoint(method, uri string) bool {
	return (method == "POST") && (uri == "/api/user/logout" || uri == "/api/user/login" || uri == "/api/features/2fa")
}

func wrong2FACode() Response {
	return Response{
		Status:  Wrong2FACode,
		Message: "wrong 2fa code",
		Data:    nil,
	}
}

func CustomHeaders(ctx *gin.Context) {
	for k, v := range config.Cfg.Headers {
		ctx.Header(k, v)
	}
}

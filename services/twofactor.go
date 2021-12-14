package services

import (
	"github.com/ProjectAnni/anniv-go/model"
	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
	"net/http"
)

func Endpoint2FA(ng *gin.Engine) {
	g := ng.Group("/api/features/2fa", AuthRequired)

	g.POST("", func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		if user.Enable2FA {
			ctx.JSON(http.StatusOK, Response{
				Status:  TFAAlreadyEnabled,
				Message: "2fa is already enabled",
				Data:    nil,
			})
			return
		}
		form := Enable2FAForm{}
		if err := ctx.ShouldBind(&form); err != nil {
			ctx.JSON(http.StatusOK, illegalParams("malformed 2fa form"))
			return
		}
		if !totp.Validate(form.Code, form.Secret) {
			ctx.JSON(http.StatusOK, wrong2FACode())
			return
		}
		user.Enable2FA = true
		user.Secret = form.Secret
		if err := db.Save(&user).Error; err != nil {
			ctx.JSON(http.StatusOK, writeErr(err))
			return
		}
		ctx.JSON(http.StatusOK, resOk(nil))
	})

	g.DELETE("", TFARequired, func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		if !user.Enable2FA {
			ctx.JSON(http.StatusOK, Response{
				Status:  TFANotEnabled,
				Message: "2fa is not enabled",
				Data:    nil,
			})
			return
		}
		user.Enable2FA = false
		user.Secret = ""
		if err := db.Save(&user).Error; err != nil {
			ctx.JSON(http.StatusOK, writeErr(err))
			return
		}
		ctx.JSON(http.StatusOK, resOk(nil))
	})
}

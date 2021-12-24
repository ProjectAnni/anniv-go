package services

import (
	"github.com/ProjectAnni/anniv-go/model"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"net/http"
)

func EndpointToken(ng *gin.Engine) {
	g := ng.Group("/api/credential", AuthRequired)

	g.GET("", func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		var tokens []model.Token
		if err := db.Where("user_id = ?", user.ID).Find(&tokens).Error; err != nil {
			ctx.JSON(http.StatusOK, readErr(err))
			return
		}
		ctx.JSON(http.StatusOK, resOk(tokenInfos(tokens)))
	})

	g.POST("", func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		form := TokenForm{}
		if err := ctx.ShouldBind(&form); err != nil {
			ctx.JSON(http.StatusOK, illegalParams("malformed token form"))
			return
		}
		token := model.Token{
			TokenID:  uuid.NewV4().String(),
			Name:     form.Name,
			URL:      form.URL,
			Token:    form.Token,
			Priority: form.Priority,
			UserID:   user.ID,
			User:     user,
		}
		if err := db.Save(&token).Error; err != nil {
			ctx.JSON(http.StatusOK, writeErr(err))
			return
		}
		ctx.JSON(http.StatusOK, resOk(tokenInfo(token)))
	})

	g.PATCH("", func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		form := TokenPatchForm{}
		if err := ctx.ShouldBind(&form); err != nil {
			ctx.JSON(http.StatusOK, illegalParams("malformed token patch form"))
			return
		}
		token := model.Token{}
		if db.Where("token_id = ? AND user_id = ?", form.ID, user.ID).First(&token).RowsAffected == 0 {
			ctx.JSON(http.StatusOK, Response{
				Status:  NotFound,
				Message: "token not found",
				Data:    nil,
			})
			return
		}
		token.Token = form.Token
		token.Name = form.Name
		token.URL = form.URL
		token.Priority = form.Priority
		if err := db.Model(&token).Updates(&token).Error; err != nil {
			ctx.JSON(http.StatusOK, writeErr(err))
			return
		}
		ctx.JSON(http.StatusOK, resOk(nil))
	})

	g.DELETE("", func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		form := DeleteForm{}
		if err := ctx.ShouldBind(&form); err != nil {
			ctx.JSON(http.StatusOK, illegalParams("malformed delete form"))
			return
		}
		token := model.Token{}
		if db.Where("token_id = ? AND user_id = ?", form.ID, user.ID).First(&token).RowsAffected == 0 {
			ctx.JSON(http.StatusOK, Response{
				Status:  NotFound,
				Message: "token not exist",
				Data:    nil,
			})
			return
		}
		if err := db.Unscoped().Delete(&token).Error; err != nil {
			ctx.JSON(http.StatusOK, writeErr(err))
			return
		}
		ctx.JSON(http.StatusOK, resOk(nil))
	})
}

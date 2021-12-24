package services

import (
	"github.com/ProjectAnni/anniv-go/model"
	"github.com/gin-gonic/gin"
	"math/rand"
	"net/http"
	"time"
)

func EndpointShare(ng *gin.Engine) {
	g := ng.Group("/api/share")

	g.GET("", func(ctx *gin.Context) {
		id := ctx.Query("id")
		share := model.Share{}
		err := db.Where("share_id = ?", id).First(&share).Error
		if err != nil {
			ctx.JSON(http.StatusOK, shareNotFound())
			return
		}
		ctx.JSON(http.StatusOK, resOk(share.Data))
	})

	g.GET("/", AuthRequired, func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		var shares []model.Share
		if err := db.Where("user_id = ?", user.ID).Find(&shares).Error; err != nil {
			ctx.JSON(http.StatusOK, readErr(err))
			return
		}
		res := make([]ShareEntry, 0, len(shares))
		for _, v := range shares {
			res = append(res, ShareEntry{
				ID:   v.ShareID,
				Date: v.CreatedAt.Unix(),
			})
		}
		ctx.JSON(http.StatusOK, resOk(res))
	})

	g.DELETE("", AuthRequired, func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		form := DeleteForm{}
		if err := ctx.ShouldBind(&form); err != nil {
			ctx.JSON(http.StatusOK, illegalParams("malformed delete form"))
			return
		}
		share := model.Share{}
		if err := db.Where("share_id = ?", form.ID).First(&share).Error; err != nil {
			ctx.JSON(http.StatusOK, shareNotFound())
			return
		}
		if share.UserID != user.ID {
			ctx.JSON(http.StatusOK, Response{
				Status:  PermissionDenied,
				Message: "this share does not belong to you",
				Data:    nil,
			})
			return
		}
		if err := db.Unscoped().Delete(&share).Error; err != nil {
			ctx.JSON(http.StatusOK, writeErr(err))
		}
		ctx.JSON(http.StatusOK, resOk(nil))
	})

	g.POST("", AuthRequired, func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		form := CreateShareForm{}
		if err := ctx.ShouldBind(&form); err != nil {
			ctx.JSON(http.StatusOK, illegalParams("malformed create form"))
			return
		}
		share := model.Share{
			UserID:  user.ID,
			Data:    form.Data,
			ShareID: randSeq(8),
		}
		if err := db.Save(&share).Error; err != nil {
			ctx.JSON(http.StatusOK, writeErr(err))
			return
		}
		ctx.JSON(http.StatusOK, resOk(share.ShareID))
	})
}

func shareNotFound() Response {
	return Response{
		Status:  NotFound,
		Message: "share not found",
		Data:    nil,
	}
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

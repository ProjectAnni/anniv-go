package endpoints

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func Basics(g *gin.Engine) {
	g.GET("/api/info", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, resOk(siteInfo()))
	})
}

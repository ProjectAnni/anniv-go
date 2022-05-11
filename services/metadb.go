package services

import (
	"github.com/ProjectAnni/anniv-go/meta"
	"github.com/gin-gonic/gin"
	"net/http"
)

func EndpointAnniDB(g *gin.Engine) {
	g.GET("/api/features/metadata-db", AuthRequired, func(ctx *gin.Context) {
		if !meta.DBAvailable() {
			ctx.Status(http.StatusNotFound)
			return
		}
		ctx.File("./tmp/repo.db")
	})
}

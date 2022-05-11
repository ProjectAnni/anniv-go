package services

import (
	"github.com/ProjectAnni/anniv-go/meta"
	"github.com/gin-gonic/gin"
	"net/http"
)

func EndpointAnniDB(g *gin.Engine) {
	g.GET("/api/features/anni-db", AuthRequired, func(ctx *gin.Context) {
		if !meta.DBAvailable() {
			ctx.String(http.StatusNotFound, "404 page not found")
			return
		}
		ctx.File("./tmp/repo.db")
	})
}

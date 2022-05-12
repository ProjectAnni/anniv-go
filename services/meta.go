package services

import (
	"github.com/ProjectAnni/anniv-go/meta"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"net/http"
)

func EndpointMeta(ng *gin.Engine) {
	g := ng.Group("/api/meta", AuthRequired)

	g.GET("/tags", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, resOk(meta.GetTags()))
	})

	g.GET("/album", func(ctx *gin.Context) {
		ids := ctx.QueryArray("id[]")
		res := make(map[string]*meta.AlbumDetails)
		for _, id := range ids {
			album, ok := meta.GetAlbumDetails(id)
			if ok {
				res[id] = &album
			} else {
				res[id] = nil
			}
		}
		ctx.JSON(http.StatusOK, resOk(res))
	})

	g.GET("/albums/by-tag", func(ctx *gin.Context) {
		tag := ctx.Query("tag")
		_, recursive := ctx.GetQuery("recursive")
		albums, ok := meta.GetAlbumsByTag(tag, recursive)
		if !ok {
			ctx.JSON(http.StatusOK, Response{
				Status:  NotFound,
				Message: "tag not found",
				Data:    nil,
			})
			return
		}
		ctx.JSON(http.StatusOK, resOk(albums))
	})

	g.GET("/tag-graph", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, resOk(meta.GetTagGraph()))
	})

	g.GET("/db/*", static.ServeRoot("/api/meta/db", "./tmp/prebuilt"))
}

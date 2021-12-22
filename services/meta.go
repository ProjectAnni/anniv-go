package services

import (
	"github.com/ProjectAnni/anniv-go/meta"
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
		res := make(map[string]*AlbumInfo)
		for _, id := range ids {
			album, ok := meta.GetAlbumInfo(id)
			if ok {
				res[id] = albumInfo(album)
			} else {
				res[id] = nil
			}
		}
		ctx.JSON(http.StatusOK, resOk(res))
	})

	g.GET("/albums/by-tag", func(ctx *gin.Context) {
		tag := ctx.Query("tag")
		albums, ok := meta.GetAlbumsByTag(tag)
		if !ok {
			ctx.JSON(http.StatusOK, Response{
				Status:  NotFound,
				Message: "tag not found",
				Data:    nil,
			})
			return
		}
		res := make([]*AlbumInfo, 0, len(albums))
		for _, v := range albums {
			res = append(res, albumInfo(v))
		}
		ctx.JSON(http.StatusOK, resOk(res))
	})

	g.GET("/tag-graph", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, resOk(meta.GetTagGraph()))
	})
}

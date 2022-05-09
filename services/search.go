package services

import (
	"github.com/ProjectAnni/anniv-go/meta"
	"github.com/gin-gonic/gin"
	"net/http"
)

type SearchResult struct {
	Albums []meta.AlbumDetails       `json:"albums"`
	Tracks []meta.TrackInfoWithAlbum `json:"tracks"`
}

func EndpointSearch(ng *gin.Engine) {
	g := ng.Group("/api/search", AuthRequired)

	g.GET("", func(ctx *gin.Context) {
		//user := ctx.MustGet("user").(model.User)
		res := SearchResult{}
		keyword := ctx.Query("keyword")
		if _, f := ctx.GetQuery("search_albums"); f {
			res.Albums = meta.SearchAlbums(keyword)
		}
		if _, f := ctx.GetQuery("search_tracks"); f {
			res.Tracks = meta.SearchTracks(keyword)
		}
		ctx.JSON(http.StatusOK, resOk(res))
	})
}

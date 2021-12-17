package services

import (
	"github.com/ProjectAnni/anniv-go/meta"
	"github.com/ProjectAnni/anniv-go/model"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func EndpointSearch(ng *gin.Engine) {
	g := ng.Group("/api/search", AuthRequired)

	g.GET("", func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		res := SearchResult{}
		keyword := ctx.Query("keyword")
		if _, f := ctx.GetQuery("search_albums"); f {
			albums := meta.SearchAlbums(keyword)
			res.Albums = make([]*AlbumInfo, 0, len(albums))
			for _, v := range albums {
				res.Albums = append(res.Albums, albumInfo(v))
			}
		}
		if _, f := ctx.GetQuery("search_tracks"); f {
			res.Tracks = meta.SearchTracks(keyword)
		}
		if _, f := ctx.GetQuery("search_playlists"); f {
			var playlists []model.Playlist
			err := db.Where("is_public OR user_id = ?", user.ID).Where("name LIKE '%' + ? + '%'", keyword).Find(&playlists).Error
			if err != nil {
				ctx.JSON(http.StatusOK, readErr(err))
				return
			}
			res.Playlists = make([]*PlaylistResult, 0, len(playlists))
			for _, v := range playlists {
				res.Playlists = append(res.Playlists, &PlaylistResult{
					ID:          strconv.Itoa(int(v.ID)),
					Name:        v.Name,
					Description: v.Description,
					Owner:       strconv.Itoa(int(v.UserID)),
				})
			}
		}
		ctx.JSON(http.StatusOK, resOk(res))
	})
}

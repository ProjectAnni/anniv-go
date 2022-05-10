package services

import (
	"github.com/ProjectAnni/anniv-go/meta"
	"github.com/ProjectAnni/anniv-go/model"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type SearchResult struct {
	Albums    []meta.AlbumDetails       `json:"albums,omitempty"`
	Tracks    []meta.TrackInfoWithAlbum `json:"tracks,omitempty"`
	Playlists []PlaylistInfo            `json:"playlists,omitempty"`
}

func EndpointSearch(ng *gin.Engine) {
	g := ng.Group("/api/search", AuthRequired)

	g.GET("", func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		res := SearchResult{}
		keyword := ctx.Query("keyword")
		if _, f := ctx.GetQuery("search_albums"); f {
			res.Albums = meta.SearchAlbums(keyword)
		}
		if _, f := ctx.GetQuery("search_tracks"); f {
			res.Tracks = meta.SearchTracks(keyword)
		}
		if _, f := ctx.GetQuery("search_playlists"); f {
			var playlists []model.Playlist
			err := db.
				Where("LOWER(name) LIKE '%' || ? || '%'", strings.ToLower(keyword)).
				Where("is_public OR user_id=?", user.ID).
				Find(&playlists).
				Error
			if err != nil {
				ctx.JSON(http.StatusOK, readErr(err))
				return
			}
			res.Playlists = make([]PlaylistInfo, 0, len(playlists))
			for _, v := range playlists {
				res.Playlists = append(res.Playlists, playlistInfo(v))
			}
		}
		ctx.JSON(http.StatusOK, resOk(res))
	})
}

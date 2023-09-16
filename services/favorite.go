package services

import (
	"net/http"
	"strconv"

	"github.com/ProjectAnni/anniv-go/meta"
	"github.com/ProjectAnni/anniv-go/model"
	"github.com/gin-gonic/gin"
)

type AlbumFavForm struct {
	AlbumID string `json:"album_id"`
}

func EndpointFavorite(ng *gin.Engine) {
	g := ng.Group("/api/favorite", AuthRequired)

	g.GET("/music", func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		var music []model.FavoriteMusic
		if err := db.Order("created_at DESC").
			Where("user_id = ?", user.ID).Find(&music).Error; err != nil {
			ctx.JSON(http.StatusOK, readErr(err))
			return
		}
		res := make([]meta.TrackInfoWithAlbum, 0, len(music))
		for _, v := range music {
			res = append(res, meta.GetTrackInfo(meta.TrackIdentifier{
				DiscIdentifier: meta.DiscIdentifier{
					AlbumID: meta.AlbumIdentifier(v.AlbumID),
					DiscID:  v.DiscID,
				},
				TrackID: v.TrackID,
			}))
		}
		ctx.JSON(http.StatusOK, resOk(res))
	})

	g.PUT("/music", func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		form := meta.TrackIdentifier{}
		if err := ctx.ShouldBind(&form); err != nil {
			ctx.JSON(http.StatusOK, illegalParams("malformed music form"))
			return
		}
		music := model.FavoriteMusic{
			UserID:  user.ID,
			AlbumID: string(form.AlbumID),
			DiscID:  form.DiscID,
			TrackID: form.TrackID,
		}
		db.Save(&music)
		ctx.JSON(http.StatusOK, resOk(nil))
	})

	g.DELETE("/music", func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		form := FavoriteMusicEntry{}
		if err := ctx.ShouldBind(&form); err != nil {
			ctx.JSON(http.StatusOK, illegalParams("malformed music form"))
			return
		}
		db.Where("user_id = ? AND album_id = ? AND disc_id = ? AND track_id = ?",
			user.ID, form.AlbumID, form.DiscID, form.TrackID).Unscoped().Delete(&model.FavoriteMusic{})
		ctx.JSON(http.StatusOK, resOk(nil))
	})

	g.GET("/playlist", func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		var playlists []model.FavoritePlaylist
		if err := db.Preload("Playlist").Order("created_at DESC").
			Where("user_id = ?", user.ID).Find(&playlists).Error; err != nil {
			ctx.JSON(http.StatusOK, readErr(err))
			return
		}
		res := make([]FavoritePlaylistEntry, 0, len(playlists))
		for _, v := range playlists {
			if v.Playlist.UserID != user.ID && !v.Playlist.IsPublic {
				res = append(res, FavoritePlaylistEntry{PlaylistID: strconv.Itoa(int(v.PlaylistID))})
			} else {
				res = append(res, FavoritePlaylistEntry{
					PlaylistID: strconv.Itoa(int(v.PlaylistID)),
					Name:       v.Playlist.Name,
					Owner:      strconv.Itoa(int(v.Playlist.UserID)),
				})
			}
		}
		ctx.JSON(http.StatusOK, resOk(res))
	})

	g.PUT("/playlist", func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		form := FavoritePlaylistForm{}
		if err := ctx.ShouldBind(&form); err != nil {
			ctx.JSON(http.StatusOK, illegalParams("malformed playlist form"))
			return
		}
		pid, err := strconv.Atoi(form.PlaylistID)
		if err != nil {
			ctx.JSON(http.StatusOK, Response{
				Status:  NotFound,
				Message: "playlist not found",
				Data:    nil,
			})
			return
		}
		if db.Where("user_id = ? AND playlist_id = ?", user.ID, pid).
			Find(&model.FavoritePlaylist{}).RowsAffected != 0 {
			ctx.JSON(http.StatusOK, resOk(nil))
			return
		}
		playlist := model.FavoritePlaylist{
			UserID:     user.ID,
			PlaylistID: uint(pid),
		}
		if err := db.Save(&playlist).Error; err != nil {
			ctx.JSON(http.StatusOK, writeErr(err))
		}
		ctx.JSON(http.StatusOK, resOk(nil))
	})

	g.DELETE("/playlist", func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		form := FavoritePlaylistForm{}
		if err := ctx.ShouldBind(&form); err != nil {
			ctx.JSON(http.StatusOK, illegalParams("malformed playlist form"))
			return
		}
		pid, err := strconv.Atoi(form.PlaylistID)
		if err != nil {
			ctx.JSON(http.StatusOK, resOk(nil))
			return
		}
		db.Unscoped().Where("user_id = ? AND playlist_id = ?", user.ID, pid).Delete(&model.FavoritePlaylist{})
		ctx.JSON(http.StatusOK, resOk(nil))
	})

	g.GET("/album", func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		var albums []model.FavoriteAlbum
		err := db.Order("created_at DESC").
			Where("user_id=?", user.ID).
			Find(&albums).Error
		if err != nil {
			ctx.JSON(http.StatusOK, readErr(err))
			return
		}
		res := make([]string, 0, len(albums))
		for _, v := range albums {
			res = append(res, v.AlbumID)
		}
		ctx.JSON(http.StatusOK, resOk(res))
	})

	g.PUT("/album", func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		var form AlbumFavForm
		if err := ctx.ShouldBindJSON(&form); err != nil {
			ctx.JSON(http.StatusOK, illegalParams(err.Error()))
			return
		}
		entry := model.FavoriteAlbum{
			UserID:  user.ID,
			AlbumID: form.AlbumID,
		}
		err := db.Save(&entry).Error
		if err != nil {
			ctx.JSON(http.StatusOK, writeErr(err))
			return
		}
		ctx.JSON(http.StatusOK, resOk(nil))
	})

	g.DELETE("/album", func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		var form AlbumFavForm
		if err := ctx.ShouldBindJSON(&form); err != nil {
			ctx.JSON(http.StatusOK, illegalParams(err.Error()))
			return
		}
		db.Unscoped().Where("user_id=?", user.ID).Where("album_id=?", form.AlbumID).Delete(&model.FavoriteAlbum{})
		ctx.JSON(http.StatusOK, resOk(nil))
	})
}

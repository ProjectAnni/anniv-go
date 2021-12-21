package services

import (
	"github.com/ProjectAnni/anniv-go/model"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

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
		res := make([]FavoriteMusicEntry, 0, len(music))
		for _, v := range music {
			res = append(res, FavoriteMusicEntry{
				AlbumID: v.AlbumID,
				DiscID:  v.DiscID,
				TrackID: v.TrackID,
			})
		}
		ctx.JSON(http.StatusOK, resOk(res))
	})

	g.PUT("/music", func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		form := FavoriteMusicEntry{}
		if err := ctx.ShouldBind(&form); err != nil {
			ctx.JSON(http.StatusOK, illegalParams("malformed music form"))
			return
		}
		if db.Where("user_id = ? AND album_id = ? AND disc_id = ? AND track_id = ?",
			user.ID, form.AlbumID, form.DiscID, form.TrackID).First(&FavoriteMusicEntry{}).RowsAffected != 0 {
			ctx.JSON(http.StatusOK, Response{
				Status:  AlreadyExist,
				Message: "music already exist",
				Data:    nil,
			})
			return
		}
		music := model.FavoriteMusic{
			UserID:  user.ID,
			AlbumID: form.AlbumID,
			DiscID:  form.DiscID,
			TrackID: form.TrackID,
		}
		if err := db.Save(&music).Error; err != nil {
			ctx.JSON(http.StatusOK, writeErr(err))
			return
		}
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
			user.ID, form.AlbumID, form.DiscID, form.TrackID).Delete(&model.FavoriteMusic{})
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
			ctx.JSON(http.StatusOK, Response{
				Status:  AlreadyExist,
				Message: "playlist already exist",
				Data:    nil,
			})
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
		db.Where("user_id = ? AND playlist_id = ?", user.ID, pid).Delete(&model.FavoritePlaylist{})
		ctx.JSON(http.StatusOK, resOk(nil))
	})

}

package services

import (
	"errors"
	"github.com/ProjectAnni/anniv-go/model"
	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

func EndpointPlaylist(ng *gin.Engine) {
	g := ng.Group("/api/playlist", AuthRequired)

	g.PUT("", func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		form := PlaylistForm{}
		if err := ctx.ShouldBind(&form); err != nil {
			ctx.JSON(http.StatusOK, illegalParams("malformed playlist form"))
			return
		}
		if form.Cover.DiscID == nil {
			form.Cover.DiscID = new(int)
			*form.Cover.DiscID = 1
		}
		playlist := model.Playlist{
			Name:         form.Name,
			Description:  form.Description,
			UserID:       user.ID,
			User:         user,
			IsPublic:     form.IsPublic,
			CoverAlbumID: form.Cover.AlbumID,
			CoverDiscID:  *form.Cover.DiscID,
		}
		err := db.Transaction(func(db *gorm.DB) error {
			err := db.Save(&playlist).Error
			if err != nil {
				return err
			}
			for k, v := range form.Songs {
				song := model.PlaylistSong{
					PlaylistID:  playlist.ID,
					Playlist:    playlist,
					AlbumID:     v.AlbumID,
					DiscID:      v.DiscID,
					TrackID:     v.TrackID,
					Description: v.Description,
					Order:       uint(k),
				}
				err = db.Save(&song).Error
				if err != nil {
					return err
				}
			}
			return nil
		})

		if err != nil {
			ctx.JSON(http.StatusOK, writeErr(err))
			return
		}

		ctx.JSON(http.StatusOK, resOk(nil))
	})

	g.DELETE("", func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		form := DeleteForm{}
		if err := ctx.ShouldBind(&form); err != nil {
			ctx.JSON(http.StatusOK, illegalParams("malformed delete form"))
			return
		}
		id, err := strconv.Atoi(form.ID)
		if err != nil {
			ctx.JSON(http.StatusOK, Response{
				Status:  NotFound,
				Message: "playlist not found",
				Data:    nil,
			})
			return
		}
		playlist := model.Playlist{}
		if err := db.Where("id = ?", id).First(&playlist).Error; err != nil {
			ctx.JSON(http.StatusOK, Response{
				Status:  NotFound,
				Message: "playlist not found",
				Data:    nil,
			})
			return
		}
		if playlist.UserID != user.ID {
			ctx.JSON(http.StatusOK, Response{
				Status:  PermissionDenied,
				Message: "permission denied",
				Data:    nil,
			})
			return
		}
		err = db.Transaction(func(tx *gorm.DB) error {
			err := tx.Where("playlist_id = ?", playlist.ID).Unscoped().Delete(&model.PlaylistSong{}).Error
			if err != nil {
				return err
			}
			return db.Unscoped().Delete(&playlist).Error
		})
		if err != nil {
			ctx.JSON(http.StatusOK, writeErr(err))
			return
		}
		ctx.JSON(http.StatusOK, resOk(nil))
	})

	g.PATCH("", func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		form := PlaylistPatchForm{}
		if err := ctx.ShouldBind(&form); err != nil {
			ctx.JSON(http.StatusOK, illegalParams("malformed patch form"))
			return
		}
		id, err := strconv.Atoi(form.ID)
		if err != nil {
			ctx.JSON(http.StatusOK, Response{
				Status:  NotFound,
				Message: "playlist not found",
				Data:    nil,
			})
			return
		}
		playlist := model.Playlist{}
		if err := db.Where("id = ?", id).First(&playlist).Error; err != nil {
			ctx.JSON(http.StatusOK, Response{
				Status:  NotFound,
				Message: "playlist not found",
				Data:    nil,
			})
			return
		}
		if playlist.UserID != user.ID {
			ctx.JSON(http.StatusOK, Response{
				Status:  PermissionDenied,
				Message: "permission denied",
				Data:    nil,
			})
			return
		}

		if form.Command == "append" {
			var payload []PlaylistSongForm
			err := mapstructure.Decode(form.Payload, &payload)
			if err != nil {
				ctx.JSON(http.StatusOK, illegalParams("malformed payload"))
				return
			}
			var ord int
			if err := db.Where("playlist_id = ?", playlist.ID).Order("order desc").
				Select("order").First(&ord).Error; err != nil {
				ctx.JSON(http.StatusOK, readErr(err))
				return
			}
			ord++
			err = db.Transaction(func(tx *gorm.DB) error {
				for k, v := range payload {
					song := model.PlaylistSong{
						PlaylistID:  playlist.ID,
						AlbumID:     v.AlbumID,
						DiscID:      v.DiscID,
						TrackID:     v.TrackID,
						Description: v.Description,
						Order:       uint(ord + k),
					}
					if err := tx.Save(&song).Error; err != nil {
						return err
					}
				}
				return nil
			})
			if err != nil {
				ctx.JSON(http.StatusOK, writeErr(err))
				return
			}
		} else if form.Command == "remove" {
			var payload []string
			err = mapstructure.Decode(form.Payload, &payload)
			if err != nil {
				ctx.JSON(http.StatusOK, illegalParams("malformed payload"))
				return
			}
			err = db.Transaction(func(tx *gorm.DB) error {
				for _, v := range payload {
					id, err := strconv.Atoi(v)
					if err != nil {
						return err
					}
					res := db.Where("playlist_id = ? AND id = ?", playlist.ID, id).Unscoped().
						Delete(&model.PlaylistSong{})
					if res.Error != nil {
						return res.Error
					}
					if res.RowsAffected != 1 {
						return errors.New("song not found")
					}
				}
				return nil
			})
			if err != nil {
				ctx.JSON(http.StatusOK, writeErr(err))
				return
			}
		} else if form.Command == "reorder" {
			var payload []string
			err = mapstructure.Decode(form.Payload, &payload)
			if err != nil {
				ctx.JSON(http.StatusOK, illegalParams("malformed payload"))
				return
			}
			var cnt int64
			if err := db.Model(&model.PlaylistSong{}).Where("playlist_id = ?", playlist.ID).
				Count(&cnt).Error; err != nil {
				ctx.JSON(http.StatusOK, readErr(err))
				return
			}
			if int(cnt) != len(payload) {
				ctx.JSON(http.StatusOK, illegalParams("malformed payload"))
				return
			}
			err = db.Transaction(func(tx *gorm.DB) error {
				for k, v := range payload {
					id, err := strconv.Atoi(v)
					if err != nil {
						return err
					}
					res := db.Model(&model.PlaylistSong{}).Where("playlist_id = ? AND id = ?", playlist.ID, id).
						Update("order", k)
					if res.Error != nil {
						return res.Error
					}
					if res.RowsAffected != 1 {
						return errors.New("song not found")
					}
				}
				return nil
			})
			if err != nil {
				ctx.JSON(http.StatusOK, writeErr(err))
				return
			}
		} else if form.Command == "replace" {
			var payload []PlaylistSong
			err = mapstructure.Decode(form.Payload, &payload)
			if err != nil {
				ctx.JSON(http.StatusOK, illegalParams("malformed payload"))
				return
			}
			err = db.Transaction(func(tx *gorm.DB) error {
				for _, v := range payload {
					id, err := strconv.Atoi(v.ID)
					if err != nil {
						return err
					}
					t := db.Model(&model.PlaylistSong{}).Where("playlist_id = ? AND id = ?", playlist.ID, id)
					t = t.Updates(map[string]interface{}{
						"album_id":    v.AlbumID,
						"disc_id":     v.DiscID,
						"track_id":    v.TrackID,
						"description": v.Description,
					})
					if t.Error != nil {
						return t.Error
					}
					if t.RowsAffected != 1 {
						return errors.New("song not found")
					}
				}
				return nil
			})
			if err != nil {
				ctx.JSON(http.StatusOK, writeErr(err))
				return
			}
		} else {
			ctx.JSON(http.StatusOK, Response{
				Status:  InvalidPatchCommand,
				Message: "invalid patch command",
				Data:    nil,
			})
			return
		}
		res, err := queryPlaylist(playlist)
		if err != nil {
			ctx.JSON(http.StatusOK, readErr(err))
			return
		}
		ctx.JSON(http.StatusOK, resOk(res))
	})

	g.GET("", func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		id, err := strconv.Atoi(ctx.Query("id"))
		if err != nil {
			ctx.JSON(http.StatusOK, Response{
				Status:  NotFound,
				Message: "playlist not found",
				Data:    nil,
			})
			return
		}
		playlist := model.Playlist{}
		if db.Where("id = ?", id).First(&playlist).RowsAffected == 0 {
			ctx.JSON(http.StatusOK, Response{
				Status:  NotFound,
				Message: "playlist not found",
				Data:    nil,
			})
			return
		}
		if playlist.UserID != user.ID && !playlist.IsPublic {
			ctx.JSON(http.StatusOK, Response{
				Status:  PermissionDenied,
				Message: "permission denied",
				Data:    nil,
			})
			return
		}
		res, err := queryPlaylist(playlist)
		if err != nil {
			ctx.JSON(http.StatusOK, readErr(err))
			return
		}
		ctx.JSON(http.StatusOK, resOk(res))
	})

	ng.GET("/api/playlists", AuthRequired, func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		userId, err := strconv.Atoi(ctx.Query("user"))
		if err != nil {
			ctx.JSON(http.StatusOK, resErr(NotFound, "user not found"))
			return
		}
		var playlists []model.Playlist
		tx := db.Where("user_id = ?", userId)
		if int(user.ID) != userId {
			tx = tx.Where("is_public")
		}
		if err := tx.First(&playlists).Error; err != nil {
			ctx.JSON(http.StatusOK, resErr(NotFound, "user not found"))
			return
		}
		res := make([]PlaylistInfo, 0, len(playlists))
		for _, v := range playlists {
			res = append(res, PlaylistInfo{
				ID:          strconv.Itoa(int(v.ID)),
				Name:        v.Name,
				Description: v.Description,
				Owner:       strconv.Itoa(int(v.UserID)),
				IsPublic:    v.IsPublic,
				Cover: Cover{
					AlbumID: v.CoverAlbumID,
					DiscID:  &v.CoverDiscID,
				},
			})
		}
		ctx.JSON(http.StatusOK, resOk(res))
	})

}

func queryPlaylist(p model.Playlist) (*Playlist, error) {
	ret := Playlist{
		ID:          strconv.Itoa(int(p.ID)),
		Name:        p.Name,
		Description: p.Description,
		Owner:       strconv.Itoa(int(p.UserID)),
		IsPublic:    p.IsPublic,
		Songs:       []PlaylistSong{},
		Cover: Cover{
			AlbumID: p.CoverAlbumID,
			DiscID:  &p.CoverDiscID,
		},
	}

	songs := make([]model.PlaylistSong, 0)
	if err := db.Where("playlist_id = ?", p.ID).Order("order").Find(&songs).Error; err != nil {
		return nil, err
	}
	for _, v := range songs {
		ret.Songs = append(ret.Songs, PlaylistSong{
			ID:          strconv.Itoa(int(v.ID)),
			AlbumID:     v.AlbumID,
			DiscID:      v.DiscID,
			TrackID:     v.TrackID,
			Description: v.Description,
		})
	}

	return &ret, nil
}

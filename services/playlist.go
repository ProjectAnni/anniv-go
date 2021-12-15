package services

import (
	"github.com/ProjectAnni/anniv-go/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
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
		playlist := model.Playlist{
			Name:        form.Name,
			Description: form.Description,
			UserID:      user.ID,
			User:        user,
			IsPublic:    form.IsPublic,
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
		}

		ctx.JSON(http.StatusOK, resOk(nil))
	})
}

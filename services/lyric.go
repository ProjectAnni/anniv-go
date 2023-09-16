package services

import (
	"net/http"
	"strconv"

	"github.com/ProjectAnni/anniv-go/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func EndpointLyric(ng *gin.Engine) {
	g := ng.Group("/api/lyric", AuthRequired)

	g.GET("", func(ctx *gin.Context) {
		aid := ctx.Query("album_id")
		dids := ctx.Query("disc_id")
		tids := ctx.Query("track_id")
		did, err := strconv.Atoi(dids)
		if err != nil {
			ctx.JSON(http.StatusOK, illegalParams("invalid disc id"))
			return
		}
		tid, err := strconv.Atoi(tids)
		if err != nil {
			ctx.JSON(http.StatusOK, illegalParams("invalid track id"))
			return
		}
		source := model.Lyric{}
		if err := db.Preload("User").
			Where("album_id = ? AND disc_id = ? AND track_id = ?", aid, did, tid).
			Where("source").First(&source).Error; err != nil {
			ctx.JSON(http.StatusOK, resErr(NotFound, "lyric not found"))
			return
		}
		translations := make([]model.Lyric, 0)
		db.Preload("User").
			Where("album_id = ? AND disc_id = ? AND track_id = ?", aid, did, tid).
			Where("NOT source").Find(&translations)

		res := LyricResponse{
			Source:       lyricLanguage(source),
			Translations: lyricLanguages(translations),
		}
		ctx.JSON(http.StatusOK, resOk(res))
	})

	g.PATCH("", func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		form := LyricPatchForm{}
		if err := ctx.ShouldBind(&form); err != nil {
			ctx.JSON(http.StatusOK, illegalParams("malformed patch form"))
			return
		}
		err := db.Transaction(func(tx *gorm.DB) error {
			t := tx.Model(&model.Lyric{}).
				Where("album_id = ? AND disc_id = ? AND track_id = ?", form.AlbumID, form.DiscID, form.TrackID).
				Session(&gorm.Session{})
			cnt := int64(0)
			if err := t.Count(&cnt).Error; err != nil {
				return err
			}
			if cnt == 0 {
				lyric := model.Lyric{
					AlbumID:  form.AlbumID,
					DiscID:   form.DiscID,
					TrackID:  form.TrackID,
					Language: form.Lang,
					Type:     form.Type,
					Data:     form.Data,
					UserID:   user.ID,
					Source:   true,
				}
				return tx.Save(&lyric).Error
			}
			if t.Where("language = ?", form.Lang).Count(&cnt); cnt != 0 {
				return t.Where("language = ?", form.Lang).Updates(map[string]interface{}{
					"type": form.Type,
					"data": form.Data,
				}).Error
			}
			lyric := model.Lyric{
				AlbumID:  form.AlbumID,
				DiscID:   form.DiscID,
				TrackID:  form.TrackID,
				Language: form.Lang,
				Type:     form.Type,
				Data:     form.Data,
				UserID:   user.ID,
				Source:   false,
			}
			return tx.Save(&lyric).Error
		})
		if err != nil {
			ctx.JSON(http.StatusOK, writeErr(err))
			return
		}
		ctx.JSON(http.StatusOK, resOk(nil))
	})
}

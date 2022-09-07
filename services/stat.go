package services

import (
	"github.com/ProjectAnni/anniv-go/meta"
	"github.com/ProjectAnni/anniv-go/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
)

type SongPlayRecord struct {
	Track meta.TrackIdentifier `json:"track"`
	At    []int64              `json:"at"`
}

type PlayRecordRes struct {
	Track meta.TrackIdentifier `json:"track"`
	Total int64                `json:"count"`
}

type HistoryRecord struct {
	Track meta.TrackIdentifier `json:"track"`
	At    int64                `json:"at"`
}

func EndpointStat(ng *gin.Engine) {
	g := ng.Group("/api/stat", AuthRequired)

	g.POST("", func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		var form []SongPlayRecord
		if err := ctx.ShouldBindJSON(&form); err != nil {
			ctx.JSON(http.StatusOK, illegalParams("malformed play record form"))
			return
		}
		err := db.Transaction(func(tx *gorm.DB) error {
			for _, v := range form {
				for _, t := range v.At {
					record := model.PlayRecord{
						UserID: user.ID,
						Track:  v.Track,
						At:     time.Unix(t, 0),
					}
					if err := tx.Save(&record).Error; err != nil {
						return err
					}
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

	g.GET("/self", func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		from, err := strconv.ParseInt(ctx.Query("from"), 10, 64)
		if err != nil {
			ctx.JSON(http.StatusOK, illegalParams("from"))
			return
		}
		to, err := strconv.ParseInt(ctx.Query("to"), 10, 64)
		if err != nil {
			to = time.Now().Unix()
		}
		var res []PlayRecordRes
		err = db.Model(&model.PlayRecord{}).
			Select("track, COUNT(track) AS total").
			Where("user_id=?", user.ID).
			Where("at BETWEEN ? AND ?", time.Unix(from, 0), time.Unix(to, 0)).
			Group("track").
			Order("total DESC").
			Find(&res).Error
		if err != nil {
			ctx.JSON(http.StatusOK, readErr(err))
			return
		}
		ctx.JSON(http.StatusOK, resOk(res))
	})

	g.GET("/self/history", func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		limit := 10
		offset := 0
		if ctx.Query("limit") != "" {
			var err error
			limit, err = strconv.Atoi(ctx.Query("limit"))
			if err != nil {
				ctx.JSON(http.StatusOK, illegalParams("limit"))
				return
			}
		}
		if ctx.Query("offset") != "" {
			var err error
			offset, err = strconv.Atoi(ctx.Query("offset"))
			if err != nil {
				ctx.JSON(http.StatusOK, illegalParams("offset"))
				return
			}
		}
		var tracks []model.PlayRecord
		err := db.Where("user_id=?", user.ID).
			Order("at DESC").
			Limit(limit).
			Offset(offset).
			Find(&tracks).Error
		if err != nil {
			ctx.JSON(http.StatusOK, readErr(err))
			return
		}
		res := make([]HistoryRecord, 0, len(tracks))
		for _, v := range tracks {
			res = append(res, HistoryRecord{
				Track: v.Track,
				At:    v.At.Unix(),
			})
		}
		ctx.JSON(http.StatusOK, resOk(res))
	})

	g.GET("/song", func(ctx *gin.Context) {
		albumId := ctx.Query("album_id")
		discId, err := strconv.Atoi(ctx.Query("disc_id"))
		if err != nil {
			ctx.JSON(http.StatusOK, illegalParams("disc_id"))
			return
		}
		trackId, err := strconv.Atoi(ctx.Query("track_id"))
		if err != nil {
			ctx.JSON(http.StatusOK, illegalParams("track_id"))
			return
		}
		track := meta.TrackIdentifier{
			DiscIdentifier: meta.DiscIdentifier{
				AlbumID: meta.AlbumIdentifier(albumId),
				DiscID:  uint(discId),
			},
			TrackID: uint(trackId),
		}
		var cnt int64
		db.Model(&model.PlayRecord{}).
			Where("track=?", track).
			Count(&cnt)
		ctx.JSON(http.StatusOK, resOk(map[string]interface{}{
			"count": cnt,
		}))
	})

	g.GET("/self/song", func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		albumId := ctx.Query("album_id")
		discId, err := strconv.Atoi(ctx.Query("disc_id"))
		if err != nil {
			ctx.JSON(http.StatusOK, illegalParams("disc_id"))
			return
		}
		trackId, err := strconv.Atoi(ctx.Query("track_id"))
		if err != nil {
			ctx.JSON(http.StatusOK, illegalParams("track_id"))
			return
		}
		track := meta.TrackIdentifier{
			DiscIdentifier: meta.DiscIdentifier{
				AlbumID: meta.AlbumIdentifier(albumId),
				DiscID:  uint(discId),
			},
			TrackID: uint(trackId),
		}
		from := int64(0)
		to := time.Now().Unix()
		if ctx.Query("from") != "" {
			var err error
			from, err = strconv.ParseInt(ctx.Query("from"), 10, 64)
			if err != nil {
				ctx.JSON(http.StatusOK, illegalParams("from"))
				return
			}
		}
		if ctx.Query("to") != "" {
			var err error
			to, err = strconv.ParseInt(ctx.Query("to"), 10, 64)
			if err != nil {
				ctx.JSON(http.StatusOK, illegalParams("to"))
				return
			}
		}
		var cnt int64
		db.Model(&model.PlayRecord{}).
			Where("track=?", track).
			Where("user_id=?", user.ID).
			Where("at BETWEEN ? AND ?", time.Unix(from, 0), time.Unix(to, 0)).
			Count(&cnt)
		ctx.JSON(http.StatusOK, resOk(map[string]interface{}{
			"count": cnt,
		}))
	})
}

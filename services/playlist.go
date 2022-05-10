package services

import (
	"errors"
	"github.com/ProjectAnni/anniv-go/meta"
	"github.com/ProjectAnni/anniv-go/model"
	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"net/http"
	"strconv"
)

type PatchedPlaylistInfo struct {
	Name        *string `json:"name" mapstructure:"name"`
	Description *string `json:"description" mapstructure:"description"`
	IsPublic    *bool   `json:"is_public" mapstructure:"is_public"`
	Cover       *Cover  `json:"cover" mapstructure:"cover"`
}

type PlaylistInfo struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	IsPublic    bool                `json:"is_public"`
	Cover       meta.DiscIdentifier `json:"cover"`
	ID          string              `json:"id"`
	Owner       string              `json:"owner"`
}

type PlaylistDetails struct {
	PlaylistInfo
	Items []PlaylistItemWithId `json:"items"`
}

type PlaylistItem struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Info        interface{} `json:"info"`
}

type PlaylistForm struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	IsPublic    bool           `json:"is_public"`
	Cover       Cover          `json:"cover"`
	Items       []PlaylistItem `json:"items"`
}

type PlaylistPatchForm struct {
	ID      string      `json:"id"`
	Command string      `json:"command"`
	Payload interface{} `json:"payload"`
}

type PlaylistItemWithId struct {
	PlaylistItem
	ID string `json:"id"`
}

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
			form.Cover.DiscID = new(uint)
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
			for k, v := range form.Items {
				song := model.PlaylistSong{
					PlaylistID: playlist.ID,
					Playlist:   playlist,
					Order:      uint(k),
				}
				if err := parsePlaylistItem(&song, v); err != nil {
					return err
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

		info, err := queryPlaylist(playlist)
		if err != nil {
			ctx.JSON(http.StatusOK, readErr(err))
			return
		}

		ctx.JSON(http.StatusOK, resOk(*info))
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
			ctx.JSON(http.StatusOK, resErr(NotFound, "playlist not found"))
			return
		}
		playlist := model.Playlist{}
		if err := db.Where("id = ?", id).First(&playlist).Error; err != nil {
			ctx.JSON(http.StatusOK, resErr(NotFound, "playlist not found"))
			return
		}
		if playlist.UserID != user.ID {
			ctx.JSON(http.StatusOK, resErr(PermissionDenied, "you are not the owner of the play list"))
			return
		}
		err = db.Transaction(func(tx *gorm.DB) error {
			err := tx.Where("playlist_id = ?", playlist.ID).Unscoped().Delete(&model.PlaylistSong{}).Error
			if err != nil {
				return err
			}
			return tx.Unscoped().Delete(&playlist).Error
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
			var payload []PlaylistItem
			err := mapstructure.Decode(form.Payload, &payload)
			if err != nil {
				ctx.JSON(http.StatusOK, illegalParams("malformed payload"))
				return
			}
			var ord int
			err = db.Model(&model.PlaylistSong{}).Select("order").
				Where("playlist_id = ?", playlist.ID).Order(clause.OrderByColumn{
				Column: clause.Column{
					Table: clause.CurrentTable,
					Name:  "order",
				},
				Desc: true,
			}).Limit(1).Scan(&ord).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					ord = 0
				} else {
					ctx.JSON(http.StatusOK, readErr(err))
					return
				}
			}
			ord++
			err = db.Transaction(func(tx *gorm.DB) error {
				for k, v := range payload {
					song := model.PlaylistSong{
						PlaylistID: playlist.ID,
						Order:      uint(ord + k),
					}
					if err := parsePlaylistItem(&song, v); err != nil {
						return err
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
			var payload PlaylistItemWithId
			err = mapstructure.Decode(form.Payload, &payload)
			if err != nil {
				ctx.JSON(http.StatusOK, illegalParams("malformed payload"))
				return
			}
			err = db.Transaction(func(tx *gorm.DB) error {
				id, err := strconv.Atoi(payload.ID)
				if err != nil {
					return err
				}
				song := model.PlaylistSong{}
				err = db.Model(&model.PlaylistSong{}).Where("playlist_id = ? AND id = ?", playlist.ID, id).
					First(&song).Error
				if err != nil {
					return err
				}
				if err := parsePlaylistItem(&song, payload.PlaylistItem); err != nil {
					return err
				}
				return db.Save(&song).Error
			})
			if err != nil {
				ctx.JSON(http.StatusOK, writeErr(err))
				return
			}
		} else if form.Command == "info" {
			var payload PatchedPlaylistInfo
			err = mapstructure.Decode(form.Payload, &payload)
			if err != nil {
				ctx.JSON(http.StatusOK, illegalParams("malformed payload"))
				return
			}
			if payload.Name != nil {
				playlist.Name = *payload.Name
			}
			if payload.Description != nil {
				playlist.Description = *payload.Description
			}
			if payload.Cover != nil {
				if payload.Cover.DiscID == nil {
					var disc = uint(1)
					payload.Cover.DiscID = &disc
				}
				playlist.CoverAlbumID = payload.Cover.AlbumID
				playlist.CoverDiscID = *payload.Cover.DiscID
			}
			if payload.IsPublic != nil {
				playlist.IsPublic = *payload.IsPublic
			}
			err = db.Save(&playlist).Error
			if err != nil {
				ctx.JSON(http.StatusOK, writeErr(err))
				return
			}
		} else {
			ctx.JSON(http.StatusOK, resErr(InvalidPatchCommand, "invalid patch command"))
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
			ctx.JSON(http.StatusOK, resErr(NotFound, "playlist not found"))
			return
		}
		playlist := model.Playlist{}
		if db.Where("id = ?", id).First(&playlist).RowsAffected == 0 {
			ctx.JSON(http.StatusOK, resErr(NotFound, "playlist not found"))
			return
		}
		if playlist.UserID != user.ID && !playlist.IsPublic {
			ctx.JSON(http.StatusOK, resErr(PermissionDenied, "permission denied"))
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
		var userId int
		userIdStr, ok := ctx.GetQuery("user_id")
		if !ok {
			userId = int(user.ID)
		} else {
			var err error
			userId, err = strconv.Atoi(userIdStr)
			if err != nil {
				ctx.JSON(http.StatusOK, resErr(NotFound, "user not found"))
				return
			}
		}

		var playlists []model.Playlist
		tx := db.Where("user_id = ?", userId)
		if int(user.ID) != userId {
			tx = tx.Where("is_public")
		}
		if err := tx.Find(&playlists).Error; err != nil {
			ctx.JSON(http.StatusOK, resOk(make([]PlaylistInfo, 0, 0)))
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
				Cover: meta.DiscIdentifier{
					AlbumID: meta.AlbumIdentifier(v.CoverAlbumID),
					DiscID:  v.CoverDiscID,
				},
			})
		}
		ctx.JSON(http.StatusOK, resOk(res))
	})
}

func playlistInfo(p model.Playlist) PlaylistInfo {
	return PlaylistInfo{
		ID:          strconv.Itoa(int(p.ID)),
		Name:        p.Name,
		Description: p.Description,
		Owner:       strconv.Itoa(int(p.UserID)),
		IsPublic:    p.IsPublic,
		Cover: meta.DiscIdentifier{
			AlbumID: meta.AlbumIdentifier(p.CoverAlbumID),
			DiscID:  p.CoverDiscID,
		},
	}
}

func queryPlaylist(p model.Playlist) (*PlaylistDetails, error) {
	ret := PlaylistDetails{
		PlaylistInfo: playlistInfo(p),
	}

	var songs []model.PlaylistSong
	if err := db.Where("playlist_id = ?", p.ID).Order(clause.OrderByColumn{
		Column: clause.Column{
			Table: clause.CurrentTable,
			Name:  "order",
		},
	}).Find(&songs).Error; err != nil {
		return nil, err
	}
	for _, v := range songs {
		item := PlaylistItemWithId{
			PlaylistItem: PlaylistItem{
				Type:        v.Type,
				Description: v.Description,
			},
			ID: strconv.Itoa(int(v.ID)),
		}
		if v.Type == "normal" {
			id := meta.TrackIdentifier{
				DiscIdentifier: meta.DiscIdentifier{
					AlbumID: meta.AlbumIdentifier(v.AlbumID),
					DiscID:  v.DiscID,
				},
				TrackID: v.TrackID,
			}
			item.Info = meta.GetTrackInfo(id)
		} else if v.Type == "dummy" {
			item.Info = v.TrackInfo
		} else {
			// type = album
			album, ok := meta.GetAlbumDetails(v.AlbumID)
			info := meta.AlbumInfo{}
			if !ok {
				info.AlbumID = meta.AlbumIdentifier(v.AlbumID)
			} else {
				info = album.AlbumInfo
			}
			item.Info = info
		}
		ret.Items = append(ret.Items, item)
	}

	return &ret, nil
}

func parsePlaylistItem(song *model.PlaylistSong, item PlaylistItem) error {
	song.Description = item.Description
	song.Type = item.Type
	if item.Type == "normal" {
		info := meta.TrackIdentifier{}
		if err := mapstructure.Decode(item.Info, &info); err != nil {
			return err
		}
		song.AlbumID = string(info.AlbumID)
		song.DiscID = info.DiscID
		song.TrackID = info.TrackID
	} else if item.Type == "dummy" {
		if err := mapstructure.Decode(item.Info, &song.TrackInfo); err != nil {
			return err
		}
	} else if item.Type == "album" {
		albumId, ok := item.Info.(string)
		if !ok {
			return errors.New("invalid payload")
		}
		song.AlbumID = albumId
	} else {
		return errors.New("invalid item type")
	}
	return nil
}

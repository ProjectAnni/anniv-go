package services

import (
	"encoding/json"
	"errors"
	"github.com/ProjectAnni/anniv-go/meta"
	"github.com/ProjectAnni/anniv-go/model"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type ShareEntry struct {
	ID   string `json:"id"`
	Date int64  `json:"date"`
}

type CreateShareForm struct {
	Info     meta.ExportedPlaylistInfo                       `json:"info"`
	Metadata map[meta.AlbumIdentifier]meta.ExportedAlbumInfo `json:"metadata"`
	Albums   map[meta.AlbumIdentifier]string                 `json:"albums"`
}

type TokenGrant struct {
	Claims AnnilShareClaims `json:"claims"`
	Secret string           `json:"secret"`
	Server string           `json:"server"`
	Kid    string           `json:"kid"`
}

func (grant *TokenGrant) Grant() (*meta.ExportedToken, error) {
	claims := grant.Claims
	claims.ExpiresAt = &jwt.NumericDate{Time: time.Now().Add(time.Hour * 24)}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["kid"] = grant.Kid
	tokenString, err := token.SignedString([]byte(grant.Secret))
	if err != nil {
		return nil, err
	}
	return &meta.ExportedToken{
		Server: grant.Server,
		Token:  tokenString,
	}, nil
}

func EndpointShare(ng *gin.Engine) {
	g := ng.Group("/api/share")

	g.GET("", func(ctx *gin.Context) {
		id := ctx.Query("id")
		share := model.Share{}
		err := db.Where("share_id = ?", id).First(&share).Error
		if err != nil {
			ctx.JSON(http.StatusOK, shareNotFound())
			return
		}
		ctx.Header("Access-Control-Allow-Origin", "*")
		var resp meta.ExportedPlaylist
		err = json.Unmarshal(share.Info, &resp)
		if err != nil {
			ctx.JSON(http.StatusOK, resErr(InternalError, err.Error()))
			return
		}

		var grants []TokenGrant
		err = json.Unmarshal(share.TokenGrant, &grants)
		if err != nil {
			ctx.JSON(http.StatusOK, resErr(InternalError, err.Error()))
			return
		}

		for _, v := range grants {
			token, err := v.Grant()
			if err != nil {
				ctx.JSON(http.StatusOK, resErr(InternalError, err.Error()))
				return
			}
			resp.Tokens = append(resp.Tokens, *token)
		}

		ctx.JSON(http.StatusOK, resOk(resp))
	})

	g.GET("/", AuthRequired, func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		var shares []model.Share
		if err := db.Where("user_id = ?", user.ID).Find(&shares).Error; err != nil {
			ctx.JSON(http.StatusOK, readErr(err))
			return
		}
		res := make([]ShareEntry, 0, len(shares))
		for _, v := range shares {
			res = append(res, ShareEntry{
				ID:   v.ShareID,
				Date: v.CreatedAt.Unix(),
			})
		}
		ctx.JSON(http.StatusOK, resOk(res))
	})

	g.DELETE("", AuthRequired, func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		form := DeleteForm{}
		if err := ctx.ShouldBind(&form); err != nil {
			ctx.JSON(http.StatusOK, illegalParams("malformed delete form"))
			return
		}
		share := model.Share{}
		if err := db.Where("share_id = ?", form.ID).First(&share).Error; err != nil {
			ctx.JSON(http.StatusOK, shareNotFound())
			return
		}
		if share.UserID != user.ID {
			ctx.JSON(http.StatusOK, Response{
				Status:  PermissionDenied,
				Message: "this share does not belong to you",
				Data:    nil,
			})
			return
		}
		if err := db.Unscoped().Delete(&share).Error; err != nil {
			ctx.JSON(http.StatusOK, writeErr(err))
		}
		ctx.JSON(http.StatusOK, resOk(nil))
	})

	g.POST("", AuthRequired, func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		form := CreateShareForm{}
		if err := ctx.ShouldBind(&form); err != nil {
			ctx.JSON(http.StatusOK, illegalParams("malformed create form"))
			return
		}
		data := meta.ExportedPlaylist{
			ExportedPlaylistInfo: form.Info,
			Metadata:             form.Metadata,
		}
		completeMetadata(data.Metadata, data.Songs)
		if form.Albums == nil {
			ctx.JSON(http.StatusOK, resErr(IllegalParams, "albums must not be null if tokens are not signed yet"))
			return
		}
		grants, err := generateTokenGrants(user, data.Songs, form.Albums)
		if err != nil {
			ctx.JSON(http.StatusOK, resErr(InternalError, err.Error()))
		}
		d, err := json.Marshal(data)
		if err != nil {
			ctx.JSON(http.StatusOK, resErr(InternalError, err.Error()))
			return
		}
		g, err := json.Marshal(grants)
		if err != nil {
			ctx.JSON(http.StatusOK, resErr(InternalError, err.Error()))
			return
		}
		share := model.Share{
			UserID:     user.ID,
			Info:       d,
			TokenGrant: g,
			ShareID:    randSeq(8),
		}
		if err := db.Save(&share).Error; err != nil {
			ctx.JSON(http.StatusOK, writeErr(err))
			return
		}
		ctx.JSON(http.StatusOK, resOk(share.ShareID))
	})
}

func shareNotFound() Response {
	return Response{
		Status:  NotFound,
		Message: "share not found",
		Data:    nil,
	}
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func completeMetadata(metadata map[meta.AlbumIdentifier]meta.ExportedAlbumInfo, songs []meta.ExportedTrackList) {
	for _, song := range songs {
		albumId := song.AlbumID
		if _, ok := metadata[albumId]; !ok {
			details, exist := meta.GetAlbumDetails(string(albumId))
			if exist {
				metadata[albumId] = meta.ExportAlbumInfo(details)
			}
		}
	}
}

type albumShareEntries map[meta.AlbumIdentifier]map[uint]map[uint]bool

func generateTokenGrants(user model.User, tracks []meta.ExportedTrackList, m map[meta.AlbumIdentifier]string) ([]TokenGrant, error) {
	var tmp albumShareEntries
	for _, v := range tracks {
		for _, tid := range v.Tracks {
			tmp[v.AlbumID][v.DiscID][tid] = true
		}
	}

	var tokenAlbumsMap map[string]albumShareEntries

	for album, entries := range tmp {
		tokenId, ok := m[album]
		if !ok {
			return nil, errors.New("no token available for album " + string(album))
		}
		// merge entry sets
		for d, t := range entries {
			for tid := range t {
				tokenAlbumsMap[tokenId][album][d][tid] = true
			}
		}
	}

	res := make([]TokenGrant, 0, len(m))
	for tid, albums := range tokenAlbumsMap {
		var userToken model.Token
		if err := db.Where("token_id=?", tid).Where("user_id=?", user.ID).First(&userToken).Error; err != nil {
			return nil, err
		}
		grant, err := getTokenGrant(userToken.Token, userToken.URL, albums)
		if err != nil {
			return nil, err
		}
		res = append(res, *grant)
	}
	return res, nil
}

func getTokenGrant(userToken string, server string, albums albumShareEntries) (*TokenGrant, error) {
	userTokenParsed, _, err := new(jwt.Parser).ParseUnverified(userToken, AnnilUserClaims{})
	if err != nil {
		return nil, err
	}
	userTokenClaims := userTokenParsed.Claims.(AnnilUserClaims)
	if userTokenClaims.Share == nil {
		return nil, errors.New("token does not support share")
	}
	keyID := userTokenClaims.Share.KeyID
	secret := userTokenClaims.Share.Secret
	shareClaims := AnnilShareClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: "anniv",
		},
		Type:   "share",
		Audios: nil,
	}
	for album, entries := range albums {
		for discId, t := range entries {
			dId := strconv.Itoa(int(discId))
			shareClaims.Audios[string(album)][dId] = make([]uint, 0, len(t))
			for trackId := range t {
				shareClaims.Audios[string(album)][dId] = append(shareClaims.Audios[string(album)][dId], trackId)
			}
		}
	}
	return &TokenGrant{
		Claims: shareClaims,
		Secret: secret,
		Server: server,
		Kid:    keyID,
	}, nil
}

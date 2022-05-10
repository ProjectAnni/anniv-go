package services

import (
	"github.com/ProjectAnni/anniv-go/model"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"net/http"
)

type Token struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Token    string `json:"token"`
	Priority int    `json:"priority"`
}

type TokenResponse struct {
	ID         string `json:"id"`
	Controlled bool   `json:"controlled"`
	Token
}

type TokenPatch struct {
	ID       string  `json:"id"`
	Name     *string `json:"name"`
	URL      *string `json:"url"`
	Token    *string `json:"token"`
	Priority *int    `json:"priority"`
}

func tokenResponse(t model.Token) TokenResponse {
	return TokenResponse{
		ID:         t.TokenID,
		Controlled: t.Controlled,
		Token: Token{
			Name:     t.Name,
			URL:      t.URL,
			Token:    t.Token,
			Priority: t.Priority,
		},
	}
}

func EndpointToken(ng *gin.Engine) {
	g := ng.Group("/api/credential", AuthRequired)

	g.GET("", func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		var tokens []model.Token
		if err := db.Where("user_id = ?", user.ID).Find(&tokens).Error; err != nil {
			ctx.JSON(http.StatusOK, readErr(err))
			return
		}
		res := make([]TokenResponse, 0, len(tokens))
		for _, v := range tokens {
			res = append(res, tokenResponse(v))
		}
		ctx.JSON(http.StatusOK, resOk(res))
	})

	g.POST("", func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		form := Token{}
		if err := ctx.ShouldBind(&form); err != nil {
			ctx.JSON(http.StatusOK, illegalParams("malformed token form"))
			return
		}
		token := model.Token{
			TokenID:    uuid.NewV4().String(),
			Name:       form.Name,
			URL:        form.URL,
			Token:      form.Token,
			Priority:   form.Priority,
			UserID:     user.ID,
			User:       user,
			Controlled: false,
		}
		if err := db.Save(&token).Error; err != nil {
			ctx.JSON(http.StatusOK, writeErr(err))
			return
		}
		ctx.JSON(http.StatusOK, resOk(tokenResponse(token)))
	})

	g.PATCH("", func(ctx *gin.Context) {
		user := ctx.MustGet("user").(model.User)
		form := TokenPatch{}
		if err := ctx.ShouldBind(&form); err != nil {
			ctx.JSON(http.StatusOK, illegalParams("malformed token patch form"))
			return
		}
		token := model.Token{}
		if db.Where("token_id = ? AND user_id = ?", form.ID, user.ID).First(&token).RowsAffected == 0 {
			ctx.JSON(http.StatusOK, Response{
				Status:  NotFound,
				Message: "token not found",
				Data:    nil,
			})
			return
		}
		if token.Controlled && (form.Token != nil || form.Name != nil || form.URL != nil) {
			ctx.JSON(http.StatusOK, resErr(ControlledToken, "controlled token"))
			return
		}
		if form.Token != nil {
			token.Token = *form.Token
		}
		if form.Name != nil {
			token.Name = *form.Name
		}
		if form.URL != nil {
			token.URL = *form.URL
		}
		if form.Priority != nil {
			token.Priority = *form.Priority
		}
		if err := db.Model(&token).Updates(&token).Error; err != nil {
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
		token := model.Token{}
		if db.Where("token_id = ? AND user_id = ?", form.ID, user.ID).First(&token).RowsAffected == 0 {
			ctx.JSON(http.StatusOK, Response{
				Status:  NotFound,
				Message: "token not exist",
				Data:    nil,
			})
			return
		}
		if token.Controlled {
			ctx.JSON(http.StatusOK, resErr(ControlledToken, "controlled token"))
			return
		}
		if err := db.Unscoped().Delete(&token).Error; err != nil {
			ctx.JSON(http.StatusOK, writeErr(err))
			return
		}
		ctx.JSON(http.StatusOK, resOk(nil))
	})
}

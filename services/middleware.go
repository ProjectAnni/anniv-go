package services

import (
	"github.com/ProjectAnni/anniv-go/model"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func AuthRequired(ctx *gin.Context) {
	sid, err := ctx.Cookie("session")
	if err != nil {
		ctx.JSON(http.StatusOK, unauthorized())
		ctx.Abort()
		return
	}
	session := model.Session{}
	if db.Where("session_id = ?", sid).First(&session).RowsAffected == 0 {
		ctx.JSON(http.StatusOK, unauthorized())
		ctx.Abort()
		return
	}
	session.LastAccessed = time.Now()
	session.UserAgent = ctx.Request.UserAgent()
	session.IP = ctx.ClientIP()
	// Renew cookie
	ctx.SetCookie("session", session.SessionID, 86400*7, "/", "", true, true)
	ctx.Set("user", session.User)
	ctx.Set("session", session)
}

func unauthorized() Response {
	return Response{
		Status:  Unauthorized,
		Message: "unauthorized",
		Data:    nil,
	}
}

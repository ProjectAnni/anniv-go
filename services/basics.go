package services

import (
	"github.com/ProjectAnni/anniv-go/config"
	"github.com/ProjectAnni/anniv-go/meta"
	"github.com/gin-gonic/gin"
	"net/http"
)

type SiteInfo struct {
	SiteName        string   `json:"site_name"`
	Description     string   `json:"description"`
	ProtocolVersion string   `json:"protocol_version"`
	Features        []string `json:"features"`
}

func siteInfo() SiteInfo {
	features := make([]string, 0)
	if config.Cfg.Enforce2FA {
		features = append(features, "2fa_enforced")
	} else {
		features = append(features, "2fa")
	}
	if config.Cfg.RequireInvite {
		features = append(features, "invite")
	}
	if meta.DBAvailable() {
		features = append(features, "metadata-db")
	}
	return SiteInfo{
		SiteName:        config.Cfg.SiteName,
		Description:     config.Cfg.Description,
		ProtocolVersion: "1",
		Features:        features,
	}
}

func EndpointBasics(g *gin.Engine) {
	g.GET("/api/info", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, resOk(siteInfo()))
	})
}

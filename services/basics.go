package services

import (
	"net/http"
	"runtime"
	"runtime/debug"

	"github.com/ProjectAnni/anniv-go/config"
	"github.com/ProjectAnni/anniv-go/meta"
	"github.com/gin-gonic/gin"
)

type SiteInfo struct {
	SiteName        string   `json:"site_name"`
	Description     string   `json:"description"`
	ProtocolVersion string   `json:"protocol_version"`
	Features        []string `json:"features"`
	Custom          any      `json:"__anniv-go"`
}

var customInfo = func() any {
	info := make(map[string]string)

	info["goVersion"] = runtime.Version()

	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		for _, v := range buildInfo.Settings {
			if v.Key == "vcs.revision" || v.Key == "vcs.time" || v.Key == "vcs.modified" {
				info[v.Key] = v.Value
			}
		}
	}

	return info
}()

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
		Custom:          customInfo,
	}
}

func EndpointBasics(g *gin.Engine) {
	g.GET("/api/info", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, resOk(siteInfo()))
	})
}

package services

import (
	"flag"
	"github.com/ProjectAnni/anniv-go/config"
	"github.com/ProjectAnni/anniv-go/model"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB
var dbPath = flag.String("db", "./data/data.db", "")

func Start(listen string) error {
	var err error
	db, err = gorm.Open(sqlite.Open(*dbPath), &gorm.Config{})
	if err != nil {
		return err
	}
	model.Bind(db)

	initMiddleware()

	g := gin.Default()
	err = g.SetTrustedProxies(config.Cfg.TrustedProxies)
	if err != nil {
		return err
	}

	g.Use(CustomHeaders)

	EndpointBasics(g)
	EndpointUser(g)
	EndpointToken(g)
	Endpoint2FA(g)
	EndpointPlaylist(g)
	EndpointMeta(g)
	EndpointSearch(g)
	EndpointShare(g)
	EndpointFavorite(g)
	EndpointLyric(g)

	g.NoRoute(static.Serve("/", static.LocalFile("frontend", false)))

	return g.Run(listen)
}

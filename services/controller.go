package services

import (
	"errors"
	"github.com/ProjectAnni/anniv-go/config"
	"github.com/ProjectAnni/anniv-go/model"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
)

var db *gorm.DB

func Start(listen string) error {
	var err error
	err = initDb()
	if err != nil {
		return err
	}
	err = model.AutoMigrate(db)
	if err != nil {
		if !errors.Is(sqlite.ErrConstraintsNotImplemented, err) {
			return errors.New("failed to migrate db: " + err.Error())
		}
	}

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
	EndpointAnniDB(g)

	g.NoRoute(static.Serve("/", static.LocalFile("frontend", false)))

	return g.Run(listen)
}

func initDb() error {
	vendor := os.Getenv("DB_VENDOR")
	path := os.Getenv("DB_PATH")

	var err error
	if vendor == "sqlite" {
		db, err = gorm.Open(sqlite.Open(path), &gorm.Config{})
	} else if vendor == "postgres" {
		db, err = gorm.Open(postgres.Open(path))
	} else {
		err = errors.New("unknown db vendor: " + vendor)
	}

	return err
}

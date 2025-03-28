package services

import (
	"errors"
	"flag"
	"log"
	"os"

	"github.com/ProjectAnni/anniv-go/config"
	"github.com/ProjectAnni/anniv-go/model"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB
var migrateTokens = flag.Bool("migrate-tokens", false, "")

func Start(listen string) error {
	var err error
	err = initDb()
	if err != nil {
		return err
	}
	err = model.AutoMigrate(db)
	if err != nil {
		return errors.New("failed to migrate db: " + err.Error())
	}

	if *migrateTokens {
		log.Println("Start migrating tokens...")
		var users []model.User
		db.Find(&users)
		for _, user := range users {
			err := db.Transaction(func(tx *gorm.DB) error {
				err := tx.Unscoped().Where("user_id=?", user.ID).
					Where("controlled IS NULL").Delete(&model.Token{}).Error
				if err != nil {
					return err
				}
				err = tx.Unscoped().Where("user_id=?", user.ID).
					Where("controlled").Delete(&model.Token{}).Error
				if err != nil {
					return err
				}
				tokens, err := signUserTokens(user.Email)
				if err != nil {
					return err
				}
				for _, token := range tokens {
					err := tx.Create(&model.Token{
						TokenID:    uuid.NewString(),
						Name:       token.Name,
						URL:        token.URL,
						Token:      token.Token,
						Priority:   token.Priority,
						UserID:     user.ID,
						Controlled: true,
					}).Error
					if err != nil {
						return err
					}
				}
				return nil
			})
			if err != nil {
				log.Fatalf(err.Error())
			}
		}
		os.Exit(0)
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
	if config.Cfg.EnableMeta {
		EndpointMeta(g)
	}
	EndpointSearch(g)
	EndpointShare(g)
	EndpointFavorite(g)
	EndpointLyric(g)
	EndpointStat(g)

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

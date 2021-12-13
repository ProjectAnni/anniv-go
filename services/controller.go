package services

import (
	"github.com/ProjectAnni/anniv-go/config"
	"github.com/ProjectAnni/anniv-go/model"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"net/http"
)

var db *gorm.DB

func Start(listen string) error {
	var err error
	db, err = gorm.Open(sqlite.Open(config.Cfg.DBPath), &gorm.Config{})
	if err != nil {
		return err
	}
	model.Bind(db)

	initMiddleware()

	g := gin.Default()
	g.Use(CustomHeaders)

	EndpointBasics(g)
	EndpointUser(g)
	EndpointToken(g)
	Endpoint2FA(g)

	return http.ListenAndServe(listen, g)
}

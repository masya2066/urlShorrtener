package routes

import (
	"shortener/internal/models"

	"github.com/gin-gonic/gin"
	"shortener/internal/routes/middleware"
)

type App struct {
	Cfg models.Config
}

func New(config models.Config) error {

	app := App{
		Cfg: config,
	}
	r := gin.Default()
	r.Use(middleware.Logger, middleware.Compress)

	r.GET("/:id", app.getURL)
	r.POST("/", app.shortner)
	r.GET("/ping", app.pingDB)
	r.POST("/api/shorten/batch", app.shortenBatch)
	api := r.Group("/api")
	{
		api.POST("/shorten", app.shorten)
	}

	err := r.Run(app.Cfg.ServerAddress)
	if err != nil {
		return err
	}
	return nil
}

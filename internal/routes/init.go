package routes

import (
	"os"

	"github.com/gin-gonic/gin"
	"shortener/internal/routes/middleware"
)

func Init() error {

	r := gin.Default()
	r.Use(middleware.Logger, middleware.Compress)

	r.GET("/:id", getURL)
	r.POST("/", shortner)
	r.GET("/ping", pingDB)
	r.POST("/api/shorten/batch", shortenBatch)
	api := r.Group("/api")
	{
		api.POST("/shorten", shorten)
	}

	err := r.Run(os.Getenv("SERVER_ADDRESS"))
	if err != nil {
		return err
	}
	return nil
}

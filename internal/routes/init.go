package routes

import (
	"github.com/gin-gonic/gin"
	"os"
	"shortener/internal/routes/middleware"
)

func Init() error {

	r := gin.Default()
	r.Use(middleware.Logger, middleware.Compress)

	r.GET("/:id", middleware.Logger, getURL)
	r.POST("/", middleware.Logger, shortner)
	api := r.Group("/api")
	{
		api.POST("/shorten", middleware.Logger, shorten)
	}

	err := r.Run(os.Getenv("SERVER_ADDRESS"))
	if err != nil {
		return err
	}
	return nil
}

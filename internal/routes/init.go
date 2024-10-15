package routes

import (
	"github.com/gin-gonic/gin"
	"os"
)

func Init() error {

	r := gin.Default()

	r.GET("/:id", GetURL)
	r.POST("/", Shortner)

	err := r.Run(os.Getenv("SERVER_ADDRESS"))
	if err != nil {
		return err
	}
	return nil
}

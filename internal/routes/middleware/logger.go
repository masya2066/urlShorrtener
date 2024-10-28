package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

func Logger(c *gin.Context) {
	logger := logrus.Logger{
		Out:       os.Stdout,
		Formatter: &logrus.JSONFormatter{},
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.InfoLevel,
	}

	startTime := time.Now()

	c.Next()

	duration := time.Since(startTime)

	statusCode := c.Writer.Status()
	contentLength := c.Writer.Size()

	logger.WithFields(logrus.Fields{
		"method":         c.Request.Method,
		"path":           c.Request.URL.Path,
		"status_code":    statusCode,
		"duration":       duration.String(),
		"content_length": contentLength,
	}).Info("Handled request")
}

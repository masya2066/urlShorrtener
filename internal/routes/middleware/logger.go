package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w responseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func Logger(c *gin.Context) {
	logger := logrus.Logger{
		Out:       os.Stdout,
		Formatter: &logrus.JSONFormatter{},
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.InfoLevel,
	}

	var requestBody string
	if c.Request.Body != nil {
		bodyBytes, _ := io.ReadAll(c.Request.Body)
		if c.GetHeader("Content-Encoding") == "gzip" {
			gzReader, err := gzip.NewReader(bytes.NewBuffer(bodyBytes))
			if err == nil {
				defer gzReader.Close()
				decompressed, _ := io.ReadAll(gzReader)
				requestBody = string(decompressed)
			} else {
				requestBody = "[ERROR decompressing request body]"
			}
		} else {
			requestBody = string(bodyBytes)
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	w := &responseBodyWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = w

	startTime := time.Now()
	c.Next()
	duration := time.Since(startTime)

	var responseBody string
	if w.Header().Get("Content-Encoding") == "gzip" {
		gzReader, err := gzip.NewReader(w.body)
		if err == nil {
			defer gzReader.Close()
			decompressed, _ := io.ReadAll(gzReader)
			responseBody = string(decompressed)
		} else {
			responseBody = "[ERROR decompressing response body]"
		}
	} else {
		responseBody = w.body.String()
	}

	logger.WithFields(logrus.Fields{
		"method":         c.Request.Method,
		"path":           c.Request.URL.Path,
		"status_code":    c.Writer.Status(),
		"duration":       duration.String(),
		"content_length": c.Writer.Size(),
		"request_body":   requestBody,
		"response_body":  responseBody,
	}).Info("Handled request")
}

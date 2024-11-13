package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type gzipResponseWriter struct {
	gin.ResponseWriter
	Writer io.Writer
}

func Compress(c *gin.Context) {
	if c.GetHeader("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(c.Request.Body)
		if err != nil {
			c.String(http.StatusBadRequest, "Failed to decompress request body")
			return
		}
		defer gz.Close()
		c.Request.Body = io.NopCloser(gz)
	}

	if strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")

		gz := gzip.NewWriter(c.Writer)
		defer gz.Close()

		c.Writer = &gzipResponseWriter{
			ResponseWriter: c.Writer,
			Writer:         gz,
		}
	}

	c.Next()
}

func (w *gzipResponseWriter) Write(data []byte) (int, error) {
	return w.Writer.Write(data) // Write gzipped data to response
}

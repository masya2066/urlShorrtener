package middleware

import (
	"compress/gzip"
	"github.com/gin-gonic/gin"
	"io"
	"strings"
)

type gzipResponseWriter struct {
	gin.ResponseWriter
	Writer io.Writer
}

func Compress(c *gin.Context) {
	if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
		c.Next()
		return
	}

	c.Header("Content-Encoding", "gzip")
	c.Header("Vary", "Accept-Encoding")

	gz := gzip.NewWriter(c.Writer)
	defer gz.Close()

	c.Writer = &gzipResponseWriter{
		ResponseWriter: c.Writer,
		Writer:         gz,
	}

	c.Next()
}

func (w *gzipResponseWriter) Write(data []byte) (int, error) {
	return w.Writer.Write(data)
}

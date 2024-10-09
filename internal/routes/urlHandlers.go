package routes

import (
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
	"shortener/internal/db"
)

type CreateBody struct {
	string
}

func Shortner(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		c.Writer.WriteHeader(http.StatusMethodNotAllowed)
		_, err := c.Writer.Write([]byte("Method must be a POST request"))
		if err != nil {
			c.Writer.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer c.Request.Body.Close()
	strBody := string(body)

	result, err := db.CreateUrl(strBody)
	if err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		_, err := c.Writer.Write([]byte(err.Error()))
		if err != nil {
			c.Writer.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	c.Writer.WriteHeader(http.StatusCreated)
	c.Header("Content-Type", "text/plain")
	_, errWrite := c.Writer.Write([]byte(os.Getenv("BASE_URL") + "/" + result))
	if errWrite != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func GetURL(c *gin.Context) {
	if c.Request.Method != http.MethodGet {
		c.Writer.WriteHeader(http.StatusMethodNotAllowed)
		_, err := c.Writer.Write([]byte("Method must be a GET request"))
		if err != nil {
			c.Writer.WriteHeader(http.StatusTemporaryRedirect)
		}
		return
	}

	id := c.Request.URL.Path[1:]

	result, err := db.GetUrl(id)
	if err != nil {
		c.Writer.WriteHeader(http.StatusTemporaryRedirect)
		_, err := c.Writer.Write([]byte(err.Error()))
		if err != nil {
			c.Writer.WriteHeader(http.StatusTemporaryRedirect)
		}
		return
	}

	if result == "" {
		c.Writer.WriteHeader(http.StatusTemporaryRedirect)
		_, err := c.Writer.Write([]byte("URL not found"))
		if err != nil {
			c.Writer.WriteHeader(http.StatusTemporaryRedirect)
		}
		return
	}

	c.Header("Location", result)
	c.Redirect(http.StatusTemporaryRedirect, result)
}

package routes

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
	"shortener/internal/db"
	"shortener/internal/models/request"
	"shortener/internal/models/response"
)

type CreateBody struct {
	string
}

func shortner(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		c.Writer.WriteHeader(http.StatusMethodNotAllowed)
		_, err := c.Writer.Write([]byte("Method must be a POST request"))
		if err != nil {
			fmt.Println(err)
			c.Writer.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		fmt.Println(err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer c.Request.Body.Close()
	strBody := string(body)

	result, err := db.AppendUrl(strBody)
	if err != nil {
		fmt.Println(err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		_, err := c.Writer.Write([]byte(err.Error()))
		if err != nil {
			fmt.Println(err)
			c.Writer.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	c.Writer.WriteHeader(http.StatusCreated)
	c.Header("Content-Type", "text/plain")
	_, errWrite := c.Writer.Write([]byte(os.Getenv("BASE_URL") + "/" + result))
	if errWrite != nil {
		fmt.Println(errWrite)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func getURL(c *gin.Context) {
	if c.Request.Method != http.MethodGet {
		c.Writer.WriteHeader(http.StatusMethodNotAllowed)
		_, err := c.Writer.Write([]byte("Method must be a GET request"))
		if err != nil {
			c.Writer.WriteHeader(http.StatusTemporaryRedirect)
		}
		return
	}

	id := c.Request.URL.Path[1:]

	result, err := db.GetURLByCode(id)
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

func shorten(c *gin.Context) {
	var body request.Shortener

	if err := c.ShouldBindJSON(&body); err != nil {
		c.Writer.WriteHeader(http.StatusBadRequest)
		_, err := c.Writer.Write([]byte(err.Error()))
		if err != nil {
			c.Writer.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if body.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "URL is required",
		})
		return
	}

	result, err := db.AppendUrl(body.URL)
	if err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		_, err := c.Writer.Write([]byte(err.Error()))
		if err != nil {
			c.Writer.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	c.JSON(http.StatusCreated, response.Shortener{
		Result: os.Getenv("BASE_URL") + "/" + result,
	})
}

package routes

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"log/slog"
	"net/http"

	"shortener/internal/db"
	"shortener/internal/models/request"
	"shortener/internal/models/response"
)

type CreateBody struct {
	string
}

func (a *App) shortner(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		c.Writer.WriteHeader(http.StatusMethodNotAllowed)
		_, err := c.Writer.Write([]byte("Method must be a POST request"))
		if err != nil {
			slog.Default().Error("Error method", err)
			c.Writer.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		slog.Default().Error("Error read", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer c.Request.Body.Close()
	strBody := string(body)

	result, err := db.CreateURL(strBody, a.Cfg)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code != "23505" {
			c.Writer.WriteHeader(http.StatusInternalServerError)
			_, err := c.Writer.Write([]byte(err.Error()))
			if err != nil {
				c.Writer.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		code, err := db.GetShortURLByLongURL(strBody, a.Cfg)
		if err != nil {
			c.Writer.WriteHeader(http.StatusInternalServerError)
			_, err := c.Writer.Write([]byte(err.Error()))
			if err != nil {
				c.Writer.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		c.Writer.WriteHeader(http.StatusConflict)
		c.Header("Content-Type", "text/plain")
		_, errWrite := c.Writer.Write([]byte(a.Cfg.BaseURL + "/" + code))
		if errWrite != nil {
			slog.Default().Error("Error write", errWrite)
			c.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}

	c.Writer.WriteHeader(http.StatusCreated)
	c.Header("Content-Type", "text/plain")
	_, errWrite := c.Writer.Write([]byte(a.Cfg.BaseURL + "/" + result))
	if errWrite != nil {
		slog.Default().Error("Error write", errWrite)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (a *App) getURL(c *gin.Context) {
	if c.Request.Method != http.MethodGet {
		c.Writer.WriteHeader(http.StatusMethodNotAllowed)
		_, err := c.Writer.Write([]byte("Method must be a GET request"))
		if err != nil {
			c.Writer.WriteHeader(http.StatusTemporaryRedirect)
		}
		return
	}

	id := c.Request.URL.Path[1:]

	result, err := db.GetURL(id, a.Cfg)
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

func (a *App) shorten(c *gin.Context) {
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

	result, err := db.CreateURL(body.URL, a.Cfg)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code != "23505" {
			c.Writer.WriteHeader(http.StatusInternalServerError)
			_, err := c.Writer.Write([]byte(err.Error()))
			if err != nil {
				c.Writer.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		code, err := db.GetShortURLByLongURL(body.URL, a.Cfg)
		if err != nil {
			c.Writer.WriteHeader(http.StatusInternalServerError)
			_, err := c.Writer.Write([]byte(err.Error()))
			if err != nil {
				c.Writer.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		c.JSON(http.StatusConflict, response.Shortener{
			Result: a.Cfg.BaseURL + "/" + code,
		})
		return
	}

	c.JSON(http.StatusCreated, response.Shortener{
		Result: a.Cfg.BaseURL + "/" + result,
	})
}

func (a *App) shortenBatch(c *gin.Context) {
	var body []request.Batch

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if len(body) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "batch cannot be empty"})
		return
	}

	result, err := db.CreateBatchURL(body, a.Cfg)
	if err != nil {
		c.Writer.WriteHeader(http.StatusBadRequest)
		_, err := c.Writer.Write([]byte(err.Error()))
		if err != nil {
			c.Writer.WriteHeader(http.StatusBadRequest)
		}
		return
	}

	c.JSON(http.StatusCreated, result)

}

func (a *App) pingDB(c *gin.Context) {
	if err := db.DB.PingDB(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "OK",
	})
}

package routes

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"net/http"
	"shortener/internal/pkg/auth"
	"strings"

	"shortener/internal/db"
	"shortener/internal/models/request"
	"shortener/internal/models/response"
)

var (
	createURLFunc      = db.CreateURL
	getURLFunc         = db.GetURL
	getShortByLongFunc = db.GetShortURLByLongURL
	createBatchFunc    = db.CreateBatchURL
)

type CreateBody struct {
	string
}

func (a *App) readUserID(c *gin.Context) (string, bool) {
	return auth.UserIDFromRequest(c.Request, a.Cfg.CookieName, []byte(a.Cfg.AuthSecret))
}

func (a *App) ensureUserID(c *gin.Context) (string, error) {
	if uid, ok := a.readUserID(c); ok {
		return uid, nil
	}
	return auth.IssueUserCookie(c.Writer, a.Cfg.CookieName, []byte(a.Cfg.AuthSecret))
}

func (a *App) shortner(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		c.Writer.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = c.Writer.Write([]byte("Method must be a POST request"))
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer c.Request.Body.Close()

	orig := strings.TrimSpace(string(body))
	if orig == "" {
		c.Writer.WriteHeader(http.StatusBadRequest)
		_, _ = c.Writer.Write([]byte("empty body"))
		return
	}

	userID, err := a.ensureUserID(c)
	if err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		_, _ = c.Writer.Write([]byte("cannot issue cookie"))
		return
	}

	code, err := createURLFunc(userID, orig, a.Cfg)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code != "23505" {
			c.Writer.WriteHeader(http.StatusInternalServerError)
			_, _ = c.Writer.Write([]byte(err.Error()))
			return
		}

		existing, getErr := getShortByLongFunc(userID, orig, a.Cfg)
		if getErr != nil {
			c.Writer.WriteHeader(http.StatusInternalServerError)
			_, _ = c.Writer.Write([]byte(getErr.Error()))
			return
		}

		c.Header("Content-Type", "text/plain; charset=utf-8")
		c.Writer.WriteHeader(http.StatusConflict)
		_, _ = c.Writer.Write([]byte(a.Cfg.BaseURL + "/" + existing))
		return
	}

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Writer.WriteHeader(http.StatusCreated)
	_, _ = c.Writer.Write([]byte(a.Cfg.BaseURL + "/" + code))
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

	result, err := getURLFunc(id, a.Cfg)
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
		_, _ = c.Writer.Write([]byte(err.Error()))
		return
	}
	if body.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL is required"})
		return
	}

	userID, err := a.ensureUserID(c)
	if err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		_, _ = c.Writer.Write([]byte("cannot issue cookie"))
		return
	}

	result, err := createURLFunc(userID, body.URL, a.Cfg)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code != "23505" {
			c.Writer.WriteHeader(http.StatusInternalServerError)
			_, _ = c.Writer.Write([]byte(err.Error()))
			return
		}

		code, err := getShortByLongFunc(userID, body.URL, a.Cfg)
		if err != nil {
			c.Writer.WriteHeader(http.StatusInternalServerError)
			_, _ = c.Writer.Write([]byte(err.Error()))
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

	userID, err := a.ensureUserID(c)
	if err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		_, _ = c.Writer.Write([]byte("cannot issue cookie"))
		return
	}

	result, err := createBatchFunc(userID, body, a.Cfg)
	if err != nil {
		c.Writer.WriteHeader(http.StatusBadRequest)
		_, _ = c.Writer.Write([]byte(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, result)
}

func (a *App) getUserURLs(c *gin.Context) {
	userID, err := a.ensureUserID(c)
	if err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		_, _ = c.Writer.Write([]byte("cannot issue cookie"))
		return
	}

	list, err := db.GetAllUserURLsFunc(userID, a.Cfg)
	if err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		_, _ = c.Writer.Write([]byte(err.Error()))
		return
	}
	if len(list) == 0 {
		c.Status(http.StatusNoContent) // 204
		return
	}
	c.JSON(http.StatusOK, list)
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

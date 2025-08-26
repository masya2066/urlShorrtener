package routes

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"shortener/internal/db"
	"shortener/internal/models"
	"shortener/internal/models/request"
	"shortener/internal/models/response"
	"testing"
)

type MockDB struct {
	PingError         error
	CreateError       error
	GetError          error
	CreatePostgresErr error
	GetPostgresErr    error
	MockURL           string
	MockCode          string
}

func (m *MockDB) PingDB() error {
	return m.PingError
}

func (m *MockDB) CreateURL(url string) (string, error) {
	if m.CreateError != nil {
		return "", m.CreateError
	}
	return m.MockCode, nil // Return a mock code
}

func (m *MockDB) GetURL(id string) (string, error) {
	if m.GetError != nil {
		return "", m.GetError
	}
	return m.MockURL, nil // Return a mock URL
}

func (m *MockDB) CreateURLPostgres(code string, url string) (string, error) {
	if m.CreatePostgresErr != nil {
		return "", m.CreatePostgresErr
	}
	return code, nil
}

func (m *MockDB) GetURLPostgres(id string) (string, error) {
	if m.GetPostgresErr != nil {
		return "", m.GetPostgresErr
	}
	return m.MockURL, nil // Return a mock URL
}

func (m *MockDB) GetShortURLByLongURLPostgres(longURL string) (string, error) {
	return m.MockCode, nil
}

func (m *MockDB) CreateBatchURLPostgres(items []request.Batch) (resItems []response.Batch, err error) {

	for _, item := range items {
		resItems = append(resItems, response.Batch{
			CorrelationID: item.CorrelationID,
			ShortURL:      "http://" + os.Getenv("SERVER_ADDRESS") + "/" + item.OriginalURL,
		})
	}

	return resItems, nil
}

func TestPingDB(t *testing.T) {
	originalDB := db.DB
	defer func() { db.DB = originalDB }()

	tests := []struct {
		name       string
		mockError  error
		wantStatus int
		wantBody   string
	}{
		{"Ping success", nil, http.StatusOK, `{"status":"OK"}`},
		{"Ping failure", errors.New("mock ping error"), http.StatusInternalServerError, `{"error":"mock ping error"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.DB = &MockDB{PingError: tt.mockError}

			app := App{
				Cfg: models.Config{
					ServerAddress:   "localhost:8080",
					BaseURL:         "http://localhost:8080",
					FileStoragePath: "tmp/JADAF",
					DatabaseDSN:     "postgres://myuser:mypassword@localhost:5432/url",
				},
			}

			gin.SetMode(gin.TestMode)
			router := gin.Default()
			router.GET("/ping", app.pingDB)

			req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			assert.Equal(t, tt.wantStatus, resp.Code)
			assert.JSONEq(t, tt.wantBody, resp.Body.String())
		})
	}
}

func TestShortner_Success(t *testing.T) {
	// save originals
	origCreate := createURLFunc
	t.Cleanup(func() { createURLFunc = origCreate })

	// stub
	createURLFunc = func(url string, cfg models.Config) (string, error) {
		return "abc123", nil
	}

	app := App{Cfg: models.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080",
	}}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/", app.shortner)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("https://example.com"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "text/plain", w.Header().Get("Content-Type"))
	assert.Equal(t, "http://localhost:8080/abc123", w.Body.String())
}

func TestShortner_DuplicateConflict(t *testing.T) {
	origCreate := createURLFunc
	origGetByLong := getShortByLongFunc
	t.Cleanup(func() {
		createURLFunc = origCreate
		getShortByLongFunc = origGetByLong
	})

	createURLFunc = func(url string, cfg models.Config) (string, error) {
		return "", &pgconn.PgError{Code: "23505"}
	}
	getShortByLongFunc = func(longURL string, cfg models.Config) (string, error) {
		return "dup001", nil
	}

	app := App{Cfg: models.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080",
	}}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/", app.shortner)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("https://dup.example"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Equal(t, "text/plain", w.Header().Get("Content-Type"))
	assert.Equal(t, "http://localhost:8080/dup001", w.Body.String())
}

func TestShortner_InternalErrorOnCreate(t *testing.T) {
	origCreate := createURLFunc
	t.Cleanup(func() { createURLFunc = origCreate })

	createURLFunc = func(url string, cfg models.Config) (string, error) {
		return "", &pgconn.PgError{Code: "23514"}
	}

	app := App{Cfg: models.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080",
	}}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/", app.shortner)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("https://err.example"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "23514") // хендлер пишет err.Error()
}

func TestShortner_DuplicateButGetShortFails(t *testing.T) {
	origCreate := createURLFunc
	origGetByLong := getShortByLongFunc
	t.Cleanup(func() {
		createURLFunc = origCreate
		getShortByLongFunc = origGetByLong
	})

	createURLFunc = func(url string, cfg models.Config) (string, error) {
		return "", &pgconn.PgError{Code: "23505"}
	}
	getShortByLongFunc = func(longURL string, cfg models.Config) (string, error) {
		return "", errors.New("cannot fetch short by long")
	}

	app := App{Cfg: models.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080",
	}}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/", app.shortner)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("https://dup.example"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "cannot fetch short by long")
}

func TestShortner_WrongMethod(t *testing.T) {
	app := App{Cfg: models.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080",
	}}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	// ВАЖНО: Any, чтобы GET тоже попал в тот же хендлер
	r.Any("/", app.shortner)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code) // 405
	assert.Contains(t, w.Body.String(), "Method must be a POST request")
}

func TestGetURL_SuccessRedirect(t *testing.T) {
	origGet := getURLFunc
	t.Cleanup(func() { getURLFunc = origGet })

	getURLFunc = func(id string, cfg models.Config) (string, error) {
		return "https://dest.example/path", nil
	}

	app := App{}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/:id", app.getURL)

	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
	assert.Equal(t, "https://dest.example/path", w.Header().Get("Location"))
}

func TestGetURL_NotFound(t *testing.T) {
	origGet := getURLFunc
	t.Cleanup(func() { getURLFunc = origGet })

	getURLFunc = func(id string, cfg models.Config) (string, error) {
		return "", nil // not found
	}

	app := App{}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/:id", app.getURL)

	req := httptest.NewRequest(http.MethodGet, "/nope", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
	assert.Equal(t, "", w.Header().Get("Location"))
	assert.Equal(t, "URL not found", w.Body.String())
}

func TestGetURL_DBError(t *testing.T) {
	origGet := getURLFunc
	t.Cleanup(func() { getURLFunc = origGet })

	getURLFunc = func(id string, cfg models.Config) (string, error) {
		return "", errors.New("db get failed")
	}

	app := App{}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/:id", app.getURL)

	req := httptest.NewRequest(http.MethodGet, "/err", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
	assert.Contains(t, w.Body.String(), "db get failed")
}

func TestGetURL_WrongMethod(t *testing.T) {
	app := App{}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Any("/:id", app.getURL) // чтобы хендлер вызвался и вернул 405

	req := httptest.NewRequest(http.MethodPost, "/abc", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	assert.Contains(t, w.Body.String(), "Method must be a GET request")
}

func TestShorten_Success(t *testing.T) {
	origCreate := createURLFunc
	t.Cleanup(func() { createURLFunc = origCreate })

	createURLFunc = func(u string, cfg models.Config) (string, error) {
		return "xyz987", nil
	}

	app := App{Cfg: models.Config{BaseURL: "http://localhost:8080"}}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/shorten", app.shorten)

	body, _ := json.Marshal(request.Shortener{URL: "https://ok.example"})
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var out response.Shortener
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	assert.Equal(t, "http://localhost:8080/xyz987", out.Result)
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
}

func TestShorten_InvalidJSON(t *testing.T) {
	app := App{}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/shorten", app.shorten)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString("{invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.NotEmpty(t, w.Body.String())
}

func TestShorten_URLEmpty(t *testing.T) {
	app := App{}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/shorten", app.shorten)

	body, _ := json.Marshal(request.Shortener{URL: ""})
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"URL is required"}`, w.Body.String())
}

func TestShorten_DuplicateConflict(t *testing.T) {
	origCreate := createURLFunc
	origGetByLong := getShortByLongFunc
	t.Cleanup(func() {
		createURLFunc = origCreate
		getShortByLongFunc = origGetByLong
	})

	// emulate duplicate
	createURLFunc = func(u string, cfg models.Config) (string, error) {
		return "", &pgconn.PgError{Code: "23505"}
	}
	getShortByLongFunc = func(longURL string, cfg models.Config) (string, error) {
		return "dup002", nil
	}

	app := App{Cfg: models.Config{
		BaseURL: "http://localhost:8080",
	}}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/shorten", app.shorten)

	body, _ := json.Marshal(request.Shortener{URL: "https://dup.example"})
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	var out response.Shortener
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	assert.Equal(t, "http://localhost:8080/dup002", out.Result)
}

func TestShorten_InternalErrorOnCreate(t *testing.T) {
	origCreate := createURLFunc
	t.Cleanup(func() { createURLFunc = origCreate })

	createURLFunc = func(u string, cfg models.Config) (string, error) {
		return "", &pgconn.PgError{Code: "23514"}
	}

	app := App{Cfg: models.Config{
		BaseURL: "http://localhost:8080",
	}}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/shorten", app.shorten)

	body, _ := json.Marshal(request.Shortener{URL: "https://err.example"})
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "23514")
}

func TestShortenBatch_Success(t *testing.T) {
	origBatch := createBatchFunc
	t.Cleanup(func() { createBatchFunc = origBatch })

	createBatchFunc = func(items []request.Batch, cfg models.Config) ([]response.Batch, error) {
		out := make([]response.Batch, 0, len(items))
		for _, it := range items {
			out = append(out, response.Batch{
				CorrelationID: it.CorrelationID,
				ShortURL:      "http://localhost:8080/" + it.OriginalURL,
			})
		}
		return out, nil
	}

	app := App{}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/shorten/batch", app.shortenBatch)

	in := []request.Batch{
		{CorrelationID: "1", OriginalURL: "https://a.example"},
		{CorrelationID: "2", OriginalURL: "https://b.example"},
	}
	body, _ := json.Marshal(in)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var out []response.Batch
	err := json.Unmarshal(w.Body.Bytes(), &out)
	assert.NoError(t, err)
	assert.Len(t, out, 2)
	assert.Equal(t, "1", out[0].CorrelationID)
	assert.Equal(t, "2", out[1].CorrelationID)
}

func TestShortenBatch_InvalidJSON(t *testing.T) {
	app := App{}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/shorten/batch", app.shortenBatch)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewBufferString("{invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"invalid request"}`, w.Body.String())
}

func TestShortenBatch_EmptyArray(t *testing.T) {
	app := App{}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/shorten/batch", app.shortenBatch)

	body, _ := json.Marshal([]request.Batch{})
	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"batch cannot be empty"}`, w.Body.String())
}

func TestShortenBatch_DBError(t *testing.T) {
	origBatch := createBatchFunc
	t.Cleanup(func() { createBatchFunc = origBatch })

	createBatchFunc = func(items []request.Batch, cfg models.Config) ([]response.Batch, error) {
		return nil, errors.New("batch failed")
	}

	app := App{}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/shorten/batch", app.shortenBatch)

	in := []request.Batch{{CorrelationID: "1", OriginalURL: "https://a.example"}}
	body, _ := json.Marshal(in)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "batch failed")
}

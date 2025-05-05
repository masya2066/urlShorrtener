package routes

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
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

func TestShortner(t *testing.T) {
	r := gin.Default()

	app := App{
		Cfg: models.Config{
			ServerAddress:   "localhost:8080",
			BaseURL:         "http://localhost:8080",
			FileStoragePath: "tmp/JADAF",
			DatabaseDSN:     "postgres://myuser:mypassword@localhost:5432/url",
		},
	}

	r.POST("/", app.shortner)

	requestBody := []byte("https://playgate.store")
	req, err := http.NewRequest("POST", "/", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			rr.Code, http.StatusCreated)
	}
}

func TestGetURL(t *testing.T) {
	r := gin.Default()
	app := App{
		Cfg: models.Config{
			ServerAddress:   "localhost:8080",
			BaseURL:         "http://localhost:8080",
			FileStoragePath: "tmp/JADAF",
			DatabaseDSN:     "postgres://myuser:mypassword@localhost:5432/url",
		},
	}
	r.GET("/:id", app.getURL)

	req, err := http.NewRequest("GET", "/123", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("handler returned wrong status code: got %v want %v",
			rr.Code, http.StatusTemporaryRedirect)
	}
}

func TestShorten(t *testing.T) {
	r := gin.Default()
	app := App{
		Cfg: models.Config{
			ServerAddress:   "localhost:8080",
			BaseURL:         "http://localhost:8080",
			FileStoragePath: "tmp/JADAF",
			DatabaseDSN:     "postgres://myuser:mypassword@localhost:5432/url",
		},
	}
	r.POST("/api/shorten", app.shorten)

	requestBody, err := json.Marshal(request.Shortener{
		URL: "https://www.example.com",
	})
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/api/shorten", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			rr.Code, http.StatusCreated)
	}
}

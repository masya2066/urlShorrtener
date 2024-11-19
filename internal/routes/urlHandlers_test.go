package routes

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"shortener/internal/db"
	"shortener/internal/models/request"
	"testing"
)

// MockDB implements the Database interface
type MockDB struct {
	PingError error
}

func (m *MockDB) PingDB() error {
	return m.PingError
}

func TestPingDB(t *testing.T) {
	// Backup the original Database
	originalDB := db.DB
	defer func() { db.DB = originalDB }()

	// Test cases
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
			// Set the mock database
			db.DB = &MockDB{PingError: tt.mockError}

			// Set up the Gin router
			gin.SetMode(gin.TestMode)
			router := gin.Default()
			router.GET("/ping", pingDB)

			// Perform the request
			req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			// Assert response
			assert.Equal(t, tt.wantStatus, resp.Code)
			assert.JSONEq(t, tt.wantBody, resp.Body.String())
		})
	}
}

func TestShortner(t *testing.T) {
	r := gin.Default()

	r.POST("/", shortner)

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

	r.GET("/:id", getURL)

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

	r.POST("/api/shorten", shorten)

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

package routes

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"shortener/internal/models/request"
	"testing"
)

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

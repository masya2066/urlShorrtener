package routes

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestShortner(t *testing.T) {
	r := gin.Default()

	r.POST("/test-shorten", Shortner)

	requestBody := []byte("https://www.example.com")
	req, err := http.NewRequest("POST", "/test-shorten", bytes.NewBuffer(requestBody))
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

	r.GET("/:id", GetURL)

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

package routes

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestShortner(t *testing.T) {
	requestBody := []byte("https://www.example.com")
	req, err := http.NewRequest("POST", "/", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Shortner)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			rr.Code, http.StatusCreated)
	}
}

func TestGetURL(t *testing.T) {
	req, err := http.NewRequest("GET", "/123", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetURL)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("handler returned wrong status code: got %v want %v",
			rr.Code, http.StatusTemporaryRedirect)
	}
}

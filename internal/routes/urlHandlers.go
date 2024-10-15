package routes

import (
	"io"
	"net/http"
	"os"
	"shortener/internal/db"
)

func Shortner(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method must be a POST request", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	strBody := string(body)
	result, err := db.CreateURL(strBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	_, errWrite := w.Write([]byte(os.Getenv("BASE_URL") + "/" + result))
	if errWrite != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}

func GetURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method must be a GET request", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Path[1:]
	result, err := db.GetURL(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusTemporaryRedirect)
		return
	}

	if result == "" {
		http.Error(w, "URL not found", http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, result, http.StatusTemporaryRedirect)
}

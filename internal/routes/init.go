package routes

import (
	"net/http"
	"os"
)

func Init() error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			Shortner(w, r)
		} else if r.Method == http.MethodGet && r.URL.Path != "/" {
			GetURL(w, r)
		} else {
			http.NotFound(w, r)
		}
	})

	err := http.ListenAndServe(os.Getenv("SERVER_ADDRESS"), nil)
	if err != nil {
		return err
	}
	return nil
}

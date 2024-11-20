package db

import (
	"database/sql"
	"fmt"
	"shortener/internal/models/request"
	"shortener/internal/models/response"
)

func getURLSQLite(id string) (string, error) {
	db, err := sql.Open("sqlite3", "./urlShortener.db")
	if err != nil {
		return "", err
	}
	defer db.Close()

	var longURL string

	err = db.QueryRow("SELECT longURL FROM urlList WHERE url_id = ?", id).Scan(&longURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("no URL found with id: %s", id)
		}
		return "", err
	}
	return longURL, nil
}

func createURLSQLite(url string, code string) (string, error) {
	db, err := sql.Open("sqlite3", "./urlShortener.db")
	if err != nil {
		return "", err
	}

	defer db.Close()

	_, err = db.Exec("INSERT INTO urlList (url_id, longURL) VALUES (?, ?)", code, url)
	if err != nil {
		return "", err
	}

	return code, nil
}

func createBatchURLSQLite(items []request.Batch) (resItems []response.Batch, error error) {
	var res []response.Batch

	for _, req := range items {
		_, err := createURLSQLite(req.OriginalURL, req.CorrelationID)
		if err != nil {
			return nil, err
		}

		res = append(res, response.Batch{
			CorrelationID: req.CorrelationID,
			OriginalURL:   req.OriginalURL,
		})
	}

	return res, nil
}

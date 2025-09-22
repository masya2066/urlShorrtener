package db

import (
	"database/sql"
	"fmt"
	"os"
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
	err = db.QueryRow(`SELECT longURL FROM urllist WHERE url_id = ?`, id).Scan(&longURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("no URL found with id: %s", id)
		}
		return "", err
	}
	return longURL, nil
}

func getShortURLByLongURLSQLite(userID, longURL string) (string, error) {
	db, err := sql.Open("sqlite3", "./urlShortener.db")
	if err != nil {
		return "", err
	}
	defer db.Close()

	var shortID string
	err = db.QueryRow(`SELECT url_id FROM urllist WHERE longURL = ? AND userID = ? LIMIT 1`, longURL, userID).Scan(&shortID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("no URL found URL: %s", longURL)
		}
		return "", err
	}
	return shortID, nil
}

func createURLSQLite(userID, url, code string) (string, error) {
	db, err := sql.Open("sqlite3", "./urlShortener.db")
	if err != nil {
		return "", err
	}
	defer db.Close()

	_, err = db.Exec(`INSERT INTO urllist (url_id, longURL, userID) VALUES (?, ?, ?)`, code, url, userID)
	if err != nil {
		return "", err
	}
	return code, nil
}

func createBatchURLSQLite(userID string, items []request.Batch) (resItems []response.Batch, err error) {
	db, err := sql.Open("sqlite3", "./urlShortener.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.Prepare(`INSERT INTO urllist (url_id, longURL, userID) VALUES (?, ?, ?)`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var res []response.Batch
	base := os.Getenv("SERVER_ADDRESS") // например "localhost:8080"

	for _, req := range items {
		if _, err := stmt.Exec(req.CorrelationID, req.OriginalURL, userID); err != nil {
			return nil, err
		}
		res = append(res, response.Batch{
			CorrelationID: req.CorrelationID,
			ShortURL:      "http://" + base + "/" + req.CorrelationID,
		})
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return res, nil
}

type UserURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func getAllUserURLsSQLite(userID string, base string) ([]UserURL, error) {
	db, err := sql.Open("sqlite3", "./urlShortener.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`SELECT url_id, longURL FROM urllist WHERE userID = ? ORDER BY url_id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if base == "" {
		base = "http://" + os.Getenv("SERVER_ADDRESS")
	}

	var out []UserURL
	for rows.Next() {
		var id, longURL string
		if err := rows.Scan(&id, &longURL); err != nil {
			return nil, err
		}
		out = append(out, UserURL{
			ShortURL:    base + "/" + id,
			OriginalURL: longURL,
		})
	}
	return out, rows.Err()
}

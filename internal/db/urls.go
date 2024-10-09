package db

import (
	"database/sql"
	"fmt"
	"shortener/internal/pkg/generator"
)

func GetUrl(id string) (string, error) {
	db, err := sql.Open("sqlite3", "./urlShortener.db")
	if err != nil {
		return "", err
	}
	defer db.Close()

	var longUrl string

	err = db.QueryRow("SELECT longUrl FROM urlList WHERE url_id = ?", id).Scan(&longUrl)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("no URL found with id: %s", id)
		}
		return "", err
	}
	return longUrl, nil
}

func CreateUrl(url string) (string, error) {
	db, err := sql.Open("sqlite3", "./urlShortener.db")
	if err != nil {
		return "", err
	}

	defer db.Close()

	code := generator.GenerateRandomCode(12)

	res, err := db.Exec("INSERT INTO urlList (url_id, longUrl) VALUES (?, ?)", code, url)
	if err != nil {
		return "", err
	}

	fmt.Println(res)
	return code, nil
}

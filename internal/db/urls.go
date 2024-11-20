package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"

	"shortener/internal/pkg/generator"
)

func (r *RealDB) GetURL(id string) (string, error) {
	var longURL string
	err := r.conn.QueryRow(context.Background(), "SELECT longURL FROM urlList WHERE url_id = $1", id).Scan(&longURL)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("no URL found with id: %s", id)
		}
		return "", err
	}
	return longURL, nil
}

func (r *RealDB) CreateURL(url string) (string, error) {
	code := generator.GenerateRandomCode(12)
	_, err := r.conn.Exec(context.Background(), "INSERT INTO urlList (url_id, longURL) VALUES ($1, $2)", code, url)
	if err != nil {
		return "", err
	}
	return code, nil
}

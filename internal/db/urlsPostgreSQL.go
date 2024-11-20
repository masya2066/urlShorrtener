package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
)

func (r *RealDB) CreateURLPostgres(code string, url string) (string, error) {
	fmt.Println(code)
	_, err := r.conn.Exec(context.Background(), "INSERT INTO urlList (url_id, longURL) VALUES ($1, $2)", code, url)
	if err != nil {
		return "", err
	}
	return code, nil
}

func (r *RealDB) GetURLPostgres(id string) (string, error) {
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

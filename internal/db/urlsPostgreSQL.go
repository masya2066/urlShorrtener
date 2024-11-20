package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"shortener/internal/models/request"
	"shortener/internal/models/response"
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

func (r *RealDB) CreateBatchURLPostgres(items []request.Batch) (resItems []response.Batch, error error) {
	tx, err := r.conn.Begin(context.Background())
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	var res []response.Batch

	for _, req := range items {
		if _, err := tx.Exec(context.Background(),
			"INSERT INTO urlList (url_id, longURL) VALUES ($1, $2)",
			req.CorrelationID, req.OriginalURL); err != nil {
			return nil, err
		}

		res = append(res, response.Batch{
			CorrelationID: req.CorrelationID,
			OriginalURL:   req.OriginalURL,
		})
	}

	// Commit transaction
	if err := tx.Commit(context.Background()); err != nil {
		return nil, err
	}

	return res, nil
}

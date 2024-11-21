package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"os"
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

func (r *RealDB) CreateBatchURLPostgres(items []request.Batch) ([]response.Batch, error) {
	tx, err := r.conn.Begin(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(context.Background())

	for _, req := range items {
		var exists bool
		err := tx.QueryRow(context.Background(),
			"SELECT EXISTS(SELECT 1 FROM urlList WHERE url_id = $1)", req.CorrelationID).Scan(&exists)
		if err != nil {
			return nil, fmt.Errorf("error checking for existing ID %s: %w", req.CorrelationID, err)
		}
		if exists {
			return nil, fmt.Errorf("ID %s already exists", req.CorrelationID)
		}
	}

	var res []response.Batch
	for _, req := range items {
		_, err := tx.Exec(context.Background(),
			"INSERT INTO urlList (url_id, longURL) VALUES ($1, $2)",
			req.CorrelationID, req.OriginalURL)
		if err != nil {
			return nil, fmt.Errorf("failed to insert item with ID %s: %w", req.CorrelationID, err)
		}

		res = append(res, response.Batch{
			CorrelationID: req.CorrelationID,
			ShortURL:      "http://" + os.Getenv("SERVER_ADDRESS") + "/" + req.CorrelationID,
		})
	}

	if err := tx.Commit(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return res, nil
}

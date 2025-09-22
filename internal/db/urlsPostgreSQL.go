package db

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"shortener/internal/models/request"
	"shortener/internal/models/response"
)

func (r *RealDB) CreateURLPostgres(userID, code, url string) (string, error) {
	_, err := r.conn.Exec(context.Background(),
		`INSERT INTO "urllist" ("url_id","longURL","userID") VALUES ($1,$2,$3)`,
		code, url, userID)
	if err != nil {
		return "", err
	}
	return code, nil
}

func (r *RealDB) GetURLPostgres(id string) (string, error) {
	var longURL string
	err := r.conn.QueryRow(context.Background(),
		`SELECT "longURL" FROM "urllist" WHERE "url_id" = $1`, id).Scan(&longURL)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("no URL found with id: %s", id)
		}
		return "", err
	}
	return longURL, nil
}

func (r *RealDB) CreateBatchURLPostgres(userID string, items []request.Batch) ([]response.Batch, error) {
	ctx := context.Background()
	tx, err := r.conn.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, req := range items {
		var exists bool
		if err := tx.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM "urllist" WHERE "url_id" = $1)`,
			req.CorrelationID).Scan(&exists); err != nil {
			return nil, fmt.Errorf("error checking id %s: %w", req.CorrelationID, err)
		}
		if exists {
			return nil, fmt.Errorf("id %s already exists", req.CorrelationID)
		}
	}

	res := make([]response.Batch, 0, len(items))
	for _, req := range items {
		if _, err := tx.Exec(ctx,
			`INSERT INTO "urllist" ("url_id","longURL","userID") VALUES ($1,$2,$3)`,
			req.CorrelationID, req.OriginalURL, userID); err != nil {
			return nil, fmt.Errorf("failed to insert item %s: %w", req.CorrelationID, err)
		}
		res = append(res, response.Batch{
			CorrelationID: req.CorrelationID,
			ShortURL:      "http://" + os.Getenv("SERVER_ADDRESS") + "/" + req.CorrelationID,
		})
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return res, nil
}

func (r *RealDB) GetShortURLByLongURLPostgres(longURL string) (string, error) {
	var shortID string
	err := r.conn.QueryRow(context.Background(),
		`SELECT "url_id" FROM "urllist" WHERE "longURL" = $1 LIMIT 1`, longURL).Scan(&shortID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("no URL found URL: %s", longURL)
		}
		return "", err
	}
	return shortID, nil
}

func (r *RealDB) GetAllUserURLsPostgres(userID string, baseURL string) ([]UserURL, error) {
	rows, err := r.conn.Query(context.Background(),
		`SELECT "url_id","longURL" FROM "urllist" WHERE "userID" = $1 ORDER BY "url_id" ASC`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	base := strings.TrimRight(baseURL, "/")
	res := make([]UserURL, 0, 16)

	for rows.Next() {
		var id, longURL string
		if err := rows.Scan(&id, &longURL); err != nil {
			return nil, err
		}
		res = append(res, UserURL{
			ShortURL:    base + "/" + id,
			OriginalURL: longURL,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

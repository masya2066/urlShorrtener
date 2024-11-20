package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"sync"
)

type Storage interface {
	AppendItem(newItem Item) error
	DeleteItem(id string) error
	GetItem(id string) (*Item, error)
	GetItemByShortCode(code string) (*Item, error)
}

type Item struct {
	ID      string `json:"id"`
	URL     string `json:"url"`
	LongURL string `json:"long_url"`
}

type FileStorage struct {
	mu   sync.Mutex
	path string
}

type Database interface {
	PingDB() error
	GetURL(id string) (string, error)
	CreateURL(url string) (string, error)
}

type RealDB struct {
	conn *pgx.Conn
}

func (r *RealDB) PingDB() error {
	return r.conn.Ping(context.Background())
}

var DB Database

func Init() error {
	connString := os.Getenv("DATABASE_DSN")
	conn, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		return err
	}
	realDB := &RealDB{conn: conn}
	if err := realDB.MigrateDB(); err != nil {
		return err
	}

	DB = realDB
	return nil
}

func (r *RealDB) MigrateDB() error {
	query := `
	CREATE TABLE IF NOT EXISTS urlList (
		url_id TEXT PRIMARY KEY,
		longURL TEXT NOT NULL
	)`
	_, err := r.conn.Exec(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}
	return nil
}

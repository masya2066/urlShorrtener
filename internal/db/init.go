package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jackc/pgx/v5"
	_ "github.com/mattn/go-sqlite3"
	"log/slog"
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
	CreateURLPostgres(code string, url string) (string, error)
	GetURLPostgres(id string) (string, error)
}

type RealDB struct {
	conn *pgx.Conn
}

func (r *RealDB) PingDB() error {
	return r.conn.Ping(context.Background())
}

var DB Database

func InitPostgres() error {
	connString := os.Getenv("DATABASE_DSN")
	conn, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		return err
	}
	realDB := &RealDB{conn: conn}
	if err := realDB.migratePostgres(); err != nil {
		return err
	}

	DB = realDB
	return nil
}

func (r *RealDB) migratePostgres() error {
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

func InitSQLite() error {
	db, err := sql.Open("sqlite3", "./urlShortener.db")
	if err != nil {
		return err
	}
	defer db.Close()
	if err := migrateSQLite(db); err != nil {
		return err
	}

	slog.Default().Info("Connected to SQLite and Migrated")
	return db.Ping()
}

func migrateSQLite(db *sql.DB) error {
	createURLListTable := `CREATE TABLE IF NOT EXISTS urlList
	(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		url_id TEXT NOT NULL,
		longURL TEXT NOT NULL
	);`

	_, err := db.Exec(createURLListTable)
	if err != nil {
		return err
	}

	return nil
}

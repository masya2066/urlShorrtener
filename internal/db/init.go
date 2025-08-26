package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"shortener/internal/models"
	"sync"

	"shortener/internal/models/request"
	"shortener/internal/models/response"

	"github.com/jackc/pgx/v5"
	_ "github.com/mattn/go-sqlite3"
)

type Storage interface {
	AppendItem(newItem Item) error
	AppendBatch(newItems []Item) error
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
	CreateBatchURLPostgres(items []request.Batch) (resItems []response.Batch, err error)
	GetShortURLByLongURLPostgres(longURL string) (string, error)
}

type RealDB struct {
	conn *pgx.Conn
}

func (r *RealDB) PingDB() error {
	return r.conn.Ping(context.Background())
}

var DB Database

func InitPostgres(cfg models.Config) error {
	connString := cfg.DatabaseDSN
	fmt.Println(connString)
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
	ctx := context.Background()

	query := `
	CREATE TABLE IF NOT EXISTS urlList (
		url_id TEXT PRIMARY KEY,
		longURL TEXT NOT NULL
	);`
	_, err := r.conn.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	deleteDuplicates := `
	DELETE FROM urlList
	WHERE url_id NOT IN (
		SELECT MIN(url_id) FROM urlList GROUP BY longURL
	);`
	_, err = r.conn.Exec(ctx, deleteDuplicates)
	if err != nil {
		return fmt.Errorf("failed to delete duplicates: %w", err)
	}

	addUniqIndex := `CREATE UNIQUE INDEX IF NOT EXISTS unique_longURL ON urlList(longURL);`
	_, err = r.conn.Exec(ctx, addUniqIndex)
	if err != nil {
		return fmt.Errorf("failed to create unique index: %w", err)
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

	log.Println("Connected to SQLite and Migrated")
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

	deleteDuplicates := `
	DELETE FROM urlList
	WHERE id NOT IN (
		SELECT MIN(id) FROM urlList GROUP BY longURL
	);`

	_, err = db.Exec(deleteDuplicates)
	if err != nil {
		return err
	}

	addUniqIndex := `CREATE UNIQUE INDEX IF NOT EXISTS unique_longURL ON urlList(longURL);`

	_, err = db.Exec(addUniqIndex)
	if err != nil {
		return err
	}

	return nil
}

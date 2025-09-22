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
	UserID  string `json:"userID"`
}

type FileStorage struct {
	mu   sync.Mutex
	path string
}

type Database interface {
	PingDB() error
	CreateURLPostgres(userID string, code string, url string) (string, error) // ⬅ добавили userID
	GetURLPostgres(id string) (string, error)
	CreateBatchURLPostgres(userID string, items []request.Batch) ([]response.Batch, error)
	GetShortURLByLongURLPostgres(longURL string) (string, error) // ⬅ добавили userID
	GetAllUserURLsPostgres(userID string, baseURL string) ([]UserURL, error)
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

	createTable := `
	CREATE TABLE IF NOT EXISTS "urllist" (
		"url_id"  TEXT PRIMARY KEY,
		"longURL" TEXT NOT NULL,
		"userID"  TEXT
	);`
	if _, err := r.conn.Exec(ctx, createTable); err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	if _, err := r.conn.Exec(ctx, `ALTER TABLE "urllist" ADD COLUMN IF NOT EXISTS "userID" TEXT;`); err != nil {
		return fmt.Errorf("failed to add userID column: %w", err)
	}

	if _, err := r.conn.Exec(ctx, `UPDATE "urllist" SET "userID" = 'legacy' WHERE "userID" IS NULL OR "userID" = ''`); err != nil {
		return fmt.Errorf("failed to backfill userID: %w", err)
	}

	deleteDup := `
	DELETE FROM "urllist" t
	USING "urllist" d
	WHERE t.ctid < d.ctid
	  AND t."userID" = d."userID"
	  AND t."longURL" = d."longURL";`
	if _, err := r.conn.Exec(ctx, deleteDup); err != nil {
		return fmt.Errorf("failed to delete duplicates by (userID,longURL): %w", err)
	}

	if _, err := r.conn.Exec(ctx, `ALTER TABLE "urllist" ALTER COLUMN "userID" SET NOT NULL;`); err != nil {
		return fmt.Errorf("failed to set userID NOT NULL: %w", err)
	}

	if _, err := r.conn.Exec(ctx, `CREATE INDEX IF NOT EXISTS urlList_userID_idx ON "urllist" ("userID");`); err != nil {
		return fmt.Errorf("failed to create userID index: %w", err)
	}

	if _, err := r.conn.Exec(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS url_unique_per_user ON "urllist" ("userID","longURL");`); err != nil {
		return fmt.Errorf("failed to create unique (userID,longURL) index: %w", err)
	}

	if _, err := r.conn.Exec(ctx, `DROP INDEX IF EXISTS "unique_longURL";`); err != nil {
		return fmt.Errorf("failed to drop old unique_longURL index: %w", err)
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
	createURLListTable := `CREATE TABLE IF NOT EXISTS urllist
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
	DELETE FROM urllist
	WHERE id NOT IN (
		SELECT MIN(id) FROM urllist GROUP BY longURL
	);`

	_, err = db.Exec(deleteDuplicates)
	if err != nil {
		return err
	}

	addUniqIndex := `CREATE UNIQUE INDEX IF NOT EXISTS unique_longURL ON urllist(longURL);`

	_, err = db.Exec(addUniqIndex)
	if err != nil {
		return err
	}

	return nil
}

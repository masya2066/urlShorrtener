package db

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func Init() error {
	db, err := sql.Open("sqlite3", "./urlShortener.db")
	if err != nil {
		return err
	}

	defer db.Close()

	if err := migrate(db); err != nil {
		return err
	}

	return db.Ping()
}

func migrate(db *sql.DB) error {
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

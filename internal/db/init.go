package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"path/filepath"
	"sync"
)

type Item struct {
	ID      string `json:"id"`
	URL     string `json:"url"`
	LongURL string `json:"long_url"`
}

var mu sync.Mutex

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

func InitStorage() error {
	mu.Lock()
	defer mu.Unlock()

	path := os.Getenv("FILE_STORAGE_PATH")
	if path == "" {
		path = "tmp/JADAF\n"
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		emptyData := []Item{}
		file, err := os.Create(path)
		if err != nil {
			return err
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		return encoder.Encode(emptyData)
	}

	return nil
}

func AppendItem(newItem Item) error {
	mu.Lock()
	defer mu.Unlock()

	items, err := getAllItems()
	if err != nil {
		return err
	}

	items = append(items, newItem)

	return writeItemsToFile(items)
}

func DeleteItem(id string) error {
	mu.Lock()
	defer mu.Unlock()

	items, err := getAllItems()
	if err != nil {
		return err
	}

	newItems := []Item{}
	for _, item := range items {
		if item.ID != id {
			newItems = append(newItems, item)
		}
	}

	return writeItemsToFile(newItems)
}

func GetItem(id string) (*Item, error) {
	items, err := getAllItems()
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		if item.ID == id {
			return &item, nil
		}
	}
	return nil, errors.New("item not found")
}

func GetItemByShortCode(code string) (*Item, error) {

	items, err := getAllItems()
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		if item.URL == code {
			return &item, nil
		}
	}

	return nil, errors.New("item not found")
}

func getAllItems() ([]Item, error) {
	path := os.Getenv("FILE_STORAGE_PATH")
	if path == "" {
		path = "tmp/JADAF\n"
	}

	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var items []Item
	err = json.Unmarshal(file, &items)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func writeItemsToFile(items []Item) error {
	path := os.Getenv("FILE_STORAGE_PATH")
	if path == "" {
		path = "tmp/JADAF\n"
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(items)
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

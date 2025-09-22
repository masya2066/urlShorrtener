package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"shortener/internal/models/request"
	"shortener/internal/models/response"
	"shortener/internal/pkg/generator"
	"strconv"
	"strings"
)

func (fs *FileStorage) AppendBatchURL(userID string, items []request.Batch, baseURL string) (resItems []response.Batch, err error) {
	var res []response.Batch
	for _, req := range items {
		if _, err := fs.AppendURL(userID, req.OriginalURL, req.CorrelationID); err != nil {
			return nil, err
		}
		res = append(res, response.Batch{
			CorrelationID: req.CorrelationID,
			ShortURL:      strings.TrimRight(baseURL, "/") + "/" + req.CorrelationID,
		})
	}
	return res, nil
}

func (fs *FileStorage) AppendURL(userID, url, codeGen string) (string, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	items, err := fs.getAllItemsStorage()
	if err != nil {
		return "", err
	}

	for _, item := range items {
		if item.UserID == userID && item.LongURL == url {
			return item.URL, nil
		}
	}

	if codeGen == "" {
		codeGen = generator.GenerateRandomCode(12) // используй свой генератор
	}

	maxID := 0
	for _, item := range items {
		if idInt, _ := strconv.Atoi(item.ID); idInt > maxID {
			maxID = idInt
		}
	}
	nextID := strconv.Itoa(maxID + 1)

	newItem := Item{
		ID:      nextID,
		URL:     codeGen,
		LongURL: url,
		UserID:  userID,
	}
	items = append(items, newItem)

	if err := fs.writeItemsToFile(items); err != nil {
		return "", err
	}
	return codeGen, nil
}

func (fs *FileStorage) GetURLByCode(code string) (string, error) {
	item, err := fs.GetItemByShortCodeStorage(code)
	if err != nil {
		return "", err
	}
	return item.LongURL, nil
}

func (fs *FileStorage) GetShortURLByLongURL(userID, longURL string) (string, error) {
	items, err := fs.getAllItemsStorage()
	if err != nil {
		return "", err
	}
	for _, item := range items {
		if item.UserID == userID && item.LongURL == longURL {
			return item.URL, nil
		}
	}
	return "", errors.New("item not found")
}

func (fs *FileStorage) GetAllUserURLs(userID, base string) ([]UserURL, error) {
	items, err := fs.getAllItemsStorage()
	if err != nil {
		return nil, err
	}
	if base == "" {
		base = "http://" + os.Getenv("SERVER_ADDRESS")
	}
	base = strings.TrimRight(base, "/")

	out := make([]UserURL, 0, 16)
	for _, it := range items {
		if it.UserID == userID {
			out = append(out, UserURL{
				ShortURL:    base + "/" + it.URL,
				OriginalURL: it.LongURL,
			})
		}
	}
	return out, nil
}

func NewFileStorage(path string) *FileStorage { return &FileStorage{path: path} }

func (fs *FileStorage) InitStorage() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	dir := filepath.Dir(fs.path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	if _, err := os.Stat(fs.path); errors.Is(err, os.ErrNotExist) {
		file, err := os.Create(fs.path)
		if err != nil {
			return err
		}
		defer file.Close()
		return json.NewEncoder(file).Encode([]Item{})
	}
	return nil
}

func (fs *FileStorage) AppendItemStorage(newItem Item) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	items, err := fs.getAllItemsStorage()
	if err != nil {
		return err
	}
	items = append(items, newItem)
	return fs.writeItemsToFile(items)
}

func (fs *FileStorage) DeleteItemStorage(id string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	items, err := fs.getAllItemsStorage()
	if err != nil {
		return err
	}
	newItems := make([]Item, 0, len(items))
	for _, item := range items {
		if item.ID != id {
			newItems = append(newItems, item)
		}
	}
	return fs.writeItemsToFile(newItems)
}

func (fs *FileStorage) GetItemStorage(id string) (*Item, error) {
	items, err := fs.getAllItemsStorage()
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

func (fs *FileStorage) GetItemByShortCodeStorage(code string) (*Item, error) {
	items, err := fs.getAllItemsStorage()
	if err != nil {
		return nil, fmt.Errorf("error in getAllItemsStorage: %w", err)
	}
	for _, item := range items {
		if item.URL == code {
			return &item, nil
		}
	}
	return nil, errors.New("item not found")
}

func (fs *FileStorage) getAllItemsStorage() ([]Item, error) {
	file, err := os.ReadFile(fs.path)
	if err != nil {
		return nil, err
	}
	var items []Item
	if err := json.Unmarshal(file, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (fs *FileStorage) writeItemsToFile(items []Item) error {
	file, err := os.Create(fs.path)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewEncoder(file).Encode(items)
}

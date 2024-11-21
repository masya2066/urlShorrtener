package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"shortener/internal/models/request"
	"shortener/internal/models/response"
	"strconv"
)

func (fs *FileStorage) AppendBatchURL(items []request.Batch) (resItems []response.Batch, error error) {
	var res []response.Batch
	for _, req := range items {
		if _, err := fs.AppendURL(req.OriginalURL, req.CorrelationID); err != nil {
			return nil, err
		}

		res = append(res, response.Batch{
			CorrelationID: req.CorrelationID,
			ShortURL:      "http://" + os.Getenv("SERVER_ADDRESS") + "/" + req.CorrelationID,
		})
	}

	return res, nil
}

func (fs *FileStorage) AppendURL(url string, codeGen string) (code string, errCreate error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	items, err := fs.getAllItemsStorage()
	if err != nil {
		return "", err
	}

	maxID := 0
	for _, item := range items {
		idInt, _ := strconv.Atoi(item.ID)
		if idInt > maxID {
			maxID = idInt
		}
	}
	nextID := strconv.Itoa(maxID + 1)

	newItem := Item{ID: nextID, URL: codeGen, LongURL: url}
	items = append(items, newItem)

	err = fs.writeItemsToFile(items)
	if err != nil {
		return "", err
	}
	return codeGen, nil
}

func (fs *FileStorage) GetURLByCode(code string) (string, error) {
	var longURL string

	items, err := fs.GetItemByShortCodeStorage(code)
	if err != nil {
		fmt.Println(2)
		return "Error in GetItemByShortCodeStorage", err
	}

	longURL = items.LongURL
	return longURL, nil
}

func NewFileStorage(path string) *FileStorage {
	return &FileStorage{
		path: path,
	}
}

func (fs *FileStorage) InitStorage() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	dir := filepath.Dir(fs.path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	if _, err := os.Stat(fs.path); errors.Is(err, os.ErrNotExist) {
		emptyData := []Item{}
		file, err := os.Create(fs.path)
		if err != nil {
			return err
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		return encoder.Encode(emptyData)
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

	newItems := []Item{}
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
		fmt.Println(item.URL, code)
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
	err = json.Unmarshal(file, &items)
	if err != nil {
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

	encoder := json.NewEncoder(file)
	return encoder.Encode(items)
}

package db

import (
	"strconv"

	"shortener/internal/pkg/generator"
)

func (fs *FileStorage) AppendURL(url string) (code string, errCreate error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	codeGen := generator.GenerateRandomCode(12)

	items, err := getAllItems()
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

	err = writeItemsToFile(items)
	if err != nil {
		return "", err
	}
	return codeGen, nil
}

func GetURLByCode(code string) (string, error) {
	var longURL string

	items, err := GetItemByShortCode(code)
	if err != nil {
		return "", err
	}

	longURL = items.LongURL
	return longURL, nil
}

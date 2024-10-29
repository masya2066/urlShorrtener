package db

import (
	"shortener/internal/pkg/generator"
	"strconv"
)

func AppendUrl(url string) (code string, errCreate error) {
	mu.Lock()
	defer mu.Unlock()

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

	newItem := Item{ID: nextID, URL: codeGen, LongUrl: url}
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

	longURL = items.LongUrl
	return longURL, nil
}

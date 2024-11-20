package db

import (
	"fmt"
	"os"
	"shortener/internal/models/request"
	"shortener/internal/models/response"

	"shortener/internal/pkg/generator"
)

func GetURL(id string) (string, error) {
	if os.Getenv("DATABASE_DSN") != "" {
		res, err := DB.GetURLPostgres(id)
		if err != nil {
			return "", err
		}

		return res, nil
	} else if os.Getenv("FILE_STORAGE_PATH") != "" {
		storagePath := os.Getenv("FILE_STORAGE_PATH")
		fileStorage := NewFileStorage(storagePath)

		result, err := fileStorage.GetURLByCode(id)
		if err != nil {
			fmt.Println(1)
			return "", err
		}
		return result, nil
	} else {
		res, err := getURLSQLite(id)
		if err != nil {
			return "", err
		}

		return res, nil
	}
}

func CreateURL(url string) (string, error) {
	code := generator.GenerateRandomCode(12)

	if os.Getenv("DATABASE_DSN") != "" {
		res, err := DB.CreateURLPostgres(code, url)
		if err != nil {
			return "", err
		}

		return res, nil
	} else if os.Getenv("FILE_STORAGE_PATH") != "" {

		storagePath := os.Getenv("FILE_STORAGE_PATH")
		fileStorage := NewFileStorage(storagePath)

		_, err := fileStorage.AppendURL(url, code)
		if err != nil {
			return "", err
		}

		return code, nil
	} else {
		res, err := createURLSQLite(url, code)
		if err != nil {
			return "", err
		}

		return res, nil
	}
}

func CreateBatchURL(items []request.Batch) ([]response.Batch, error) {
	if os.Getenv("DATABASE_DSN") != "" {
		res, err := DB.CreateBatchURLPostgres(items)
		if err != nil {
			return nil, err
		}
		return res, nil
	} else if os.Getenv("FILE_STORAGE_PATH") != "" {
		storagePath := os.Getenv("FILE_STORAGE_PATH")
		fileStorage := NewFileStorage(storagePath)

		res, err := fileStorage.AppendBatchUrl(items)
		if err != nil {
			return nil, err
		}
		return res, nil
	} else {
		res, err := createBatchURLSQLite(items)
		if err != nil {
			return nil, err
		}
		return res, nil

	}
}

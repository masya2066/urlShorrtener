package db

import (
	"shortener/internal/models"
	"shortener/internal/models/request"
	"shortener/internal/models/response"

	"shortener/internal/pkg/generator"
)

func GetURL(id string, cfg models.Config) (string, error) {
	if cfg.DatabaseDSN != "" {
		res, err := DB.GetURLPostgres(id)
		if err != nil {
			return "", err
		}

		return res, nil
	} else if cfg.FileStoragePath != "" {
		storagePath := cfg.FileStoragePath
		fileStorage := NewFileStorage(storagePath)

		result, err := fileStorage.GetURLByCode(id)
		if err != nil {
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

func CreateURL(url string, cfg models.Config) (string, error) {
	code := generator.GenerateRandomCode(12)

	if cfg.DatabaseDSN != "" {
		res, err := DB.CreateURLPostgres(code, url)
		if err != nil {
			return "", err
		}

		return res, nil
	} else if cfg.FileStoragePath != "" {

		storagePath := cfg.FileStoragePath
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

func CreateBatchURL(items []request.Batch, cfg models.Config) ([]response.Batch, error) {
	if cfg.DatabaseDSN != "" {
		res, err := DB.CreateBatchURLPostgres(items)
		if err != nil {
			return nil, err
		}
		return res, nil
	} else if cfg.FileStoragePath != "" {
		storagePath := cfg.FileStoragePath
		fileStorage := NewFileStorage(storagePath)

		res, err := fileStorage.AppendBatchURL(items, "http://"+cfg.BaseURL)
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

func GetShortURLByLongURL(longURL string, cfg models.Config) (string, error) {
	if cfg.DatabaseDSN != "" {
		res, err := DB.GetShortURLByLongURLPostgres(longURL)
		if err != nil {
			return "", err
		}
		return res, nil
	} else if cfg.FileStoragePath != "" {
		storagePath := cfg.FileStoragePath
		fileStorage := NewFileStorage(storagePath)

		res, err := fileStorage.GetShortURLByLongURL(longURL)
		if err != nil {
			return "", err
		}
		return res, nil
	} else {
		res, err := getShortURLByLongURLSQLite(longURL)
		if err != nil {
			return "", err
		}
		return res, nil
	}
}

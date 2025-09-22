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

func CreateURL(userID, url string, cfg models.Config) (string, error) {
	code := generator.GenerateRandomCode(12)

	if cfg.DatabaseDSN != "" {
		res, err := DB.CreateURLPostgres(userID, code, url)
		if err != nil {
			return "", err
		}

		return res, nil
	} else if cfg.FileStoragePath != "" {

		storagePath := cfg.FileStoragePath
		fileStorage := NewFileStorage(storagePath)

		isCode, err := fileStorage.AppendURL(userID, url, code)
		if err != nil {
			return "", err
		}

		return isCode, nil
	} else {
		res, err := createURLSQLite(userID, url, code)
		if err != nil {
			return "", err
		}

		return res, nil
	}
}

func CreateBatchURL(userID string, items []request.Batch, cfg models.Config) ([]response.Batch, error) {
	if cfg.DatabaseDSN != "" {
		res, err := DB.CreateBatchURLPostgres(userID, items)
		if err != nil {
			return nil, err
		}
		return res, nil
	} else if cfg.FileStoragePath != "" {
		storagePath := cfg.FileStoragePath
		fileStorage := NewFileStorage(storagePath)

		res, err := fileStorage.AppendBatchURL(userID, items, "http://"+cfg.BaseURL)
		if err != nil {
			return nil, err
		}
		return res, nil
	} else {
		res, err := createBatchURLSQLite(userID, items)
		if err != nil {
			return nil, err
		}
		return res, nil

	}
}

func GetShortURLByLongURL(userID, longURL string, cfg models.Config) (string, error) {
	if cfg.DatabaseDSN != "" {
		res, err := DB.GetShortURLByLongURLPostgres(longURL)
		if err != nil {
			return "", err
		}
		return res, nil
	} else if cfg.FileStoragePath != "" {
		storagePath := cfg.FileStoragePath
		fileStorage := NewFileStorage(storagePath)

		res, err := fileStorage.GetShortURLByLongURL(userID, longURL)
		if err != nil {
			return "", err
		}
		return res, nil
	} else {
		res, err := getShortURLByLongURLSQLite(userID, longURL)
		if err != nil {
			return "", err
		}
		return res, nil
	}
}

func GetAllUserURLsFunc(userID string, cfg models.Config) ([]UserURL, error) {
	base := "http://" + cfg.BaseURL

	if cfg.DatabaseDSN != "" {
		return DB.GetAllUserURLsPostgres(userID, base)
	} else if cfg.FileStoragePath != "" {
		fs := NewFileStorage(cfg.FileStoragePath)
		return fs.GetAllUserURLs(userID, base)
	} else {
		return getAllUserURLsSQLite(userID, base)
	}
}

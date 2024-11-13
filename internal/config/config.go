package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"shortener/internal/models"
)

func checkEnv(config models.Config) {
	if os.Getenv("SERVER_ADDRESS") == "" {
		os.Setenv("SERVER_ADDRESS", config.ServerAddress)
	}
	if os.Getenv("BASE_URL") == "" {
		os.Setenv("BASE_URL", config.BaseURL)
	}
	if os.Getenv("FILE_STORAGE_PATH") == "" {
		os.Setenv("FILE_STORAGE_PATH", config.FileStoragePath)
	}
}

func LoadConfig(filename string) error {
	var config models.Config
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Println("Config file does not exist. Creating a new one...")

		config = models.Config{
			ServerAddress:   "localhost:8080",
			BaseURL:         "http://localhost:8080",
			FileStoragePath: "tmp/JADAF",
		}

		configBytes, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return err
		}

		err = os.WriteFile(filename, configBytes, 0644)
		if err != nil {
			return err
		}
		fmt.Println("Default config created:", filename)
	} else {
		file, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer file.Close()

		bytes, err := io.ReadAll(file)
		if err != nil {
			return err
		}

		err = json.Unmarshal(bytes, &config)
		if err != nil {
			return err
		}

		checkEnv(config)
		return nil
	}

	checkEnv(config)

	return nil
}

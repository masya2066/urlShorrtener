package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"shortener/internal/models"
)

func LoadConfig(filename string) (models.Config, error) {
	var config models.Config
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Println("Config file does not exist. Creating a new one...")

		config = models.Config{
			ServerAddress:   "localhost:8080",
			BaseURL:         "http://localhost:8080",
			FileStoragePath: "storage",
		}

		configBytes, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return config, err
		}

		err = os.WriteFile(filename, configBytes, 0644)
		if err != nil {
			return config, err
		}
		fmt.Println("Default config created:", filename)
	} else {
		file, err := os.Open(filename)
		if err != nil {
			return config, err
		}
		defer file.Close()

		bytes, err := io.ReadAll(file)
		if err != nil {
			return config, err
		}

		err = json.Unmarshal(bytes, &config)
		if err != nil {
			return config, err
		}

		return config, nil
	}

	return config, nil
}

package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"shortener/internal/models"
)

func LoadConfig(filename string) (models.Config, error) {
	aFlag := flag.String("a", "", "Value for the -a flag")
	bFlag := flag.String("b", "", "Value for the -b flag")
	fFlag := flag.String("f", "", "Value for the -f flag")
	dFlag := flag.String("d", "", "Value for the -d flag")

	flag.Parse()

	config := models.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "tmp/JADAF",
		DatabaseDSN:     "default_dsn",
	}

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Println("Config file does not exist. Creating a new one...")

		configBytes, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return models.Config{}, err
		}

		err = os.WriteFile(filename, configBytes, 0644)
		if err != nil {
			return models.Config{}, err
		}
		fmt.Println("Default config created:", filename)
	} else {
		file, err := os.Open(filename)
		if err != nil {
			return models.Config{}, err
		}
		defer file.Close()

		bytes, err := io.ReadAll(file)
		if err != nil {
			return models.Config{}, err
		}

		err = json.Unmarshal(bytes, &config)
		if err != nil {
			return models.Config{}, err
		}
	}

	if *aFlag != "" {
		config.ServerAddress = *aFlag
	}
	if *bFlag != "" {
		config.BaseURL = *bFlag
	}
	if *fFlag != "" {
		config.FileStoragePath = *fFlag
	}
	if *dFlag != "" {
		config.DatabaseDSN = *dFlag
	}

	fmt.Printf("Loaded Config:\nServerAddress: %s\nBaseURL: %s\nFileStoragePath: %s\nDatabaseDSN: %s\n",
		config.ServerAddress, config.BaseURL, config.FileStoragePath, config.DatabaseDSN)

	return config, nil
}

package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"shortener/internal/models"
)

func LoadConfig(filename string) (conf models.Config, error error) {
	aFlag := flag.String("a", "", "Value for the -a flag")
	bFlag := flag.String("b", "", "Value for the -b flag")
	fFlag := flag.String("f", "", "Value for the -f flag")
	dFlag := flag.String("d", "", "Value for the -d flag")

	flag.Parse()

	if *aFlag != "" {
		if err := os.Setenv("SERVER_ADDRESS", *aFlag); err != nil {
			fmt.Println("Error setting environment variable:", err)
			return models.Config{}, err
		}
		fmt.Println("Environment variable SERVER_ADDRESS set to:", *aFlag)
	} else {
		fmt.Println("No -a flag provided")
	}

	if *bFlag != "" {
		if err := os.Setenv("BASE_URL", *bFlag); err != nil {
			fmt.Println("Error setting environment variable:", err)
			return models.Config{}, err
		}
		fmt.Println("Environment variable BASE_URL set to:", *bFlag)
	} else {
		fmt.Println("No -b flag provided")
	}

	if *fFlag != "" {
		if err := os.Setenv("FILE_STORAGE_PATH", *fFlag); err != nil {
			fmt.Println("Error setting environment variable:", err)
			return models.Config{}, err
		}
		fmt.Println("Environment variable FILE_STORAGE_PATH set to:", *fFlag)
	} else {
		fmt.Println("No -f flag provided")
	}
	if *dFlag != "" {
		if err := os.Setenv("DATABASE_DSN", *dFlag); err != nil {
			fmt.Println("Error setting environment variable:", err)
			return models.Config{}, err
		}
		fmt.Println("Environment variable DATABASE_DSN set to:", *dFlag)
	} else {
		fmt.Println("No -d flag provided")
	}

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
			return models.Config{}, err
		}

		err = os.WriteFile(filename, configBytes, 0644)
		if err != nil {
			return models.Config{}, err
		}
		fmt.Println("Default config created:", filename)

		if err := setConfigEnv(config); err != nil {
			return models.Config{}, err
		}

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

		if err := setConfigEnv(config); err != nil {
			return models.Config{}, err
		}
		return config, nil
	}

	if err := setConfigEnv(config); err != nil {
		return models.Config{}, err
	}
	return config, nil
}

func setConfigEnv(config models.Config) error {
	if config.BaseURL == "" {
		if err := os.Setenv("BASE_URL", config.BaseURL); err != nil {
			fmt.Println("Error setting environment variable:", err)
			return err
		}
		fmt.Println("Environment variable BASE_URL from config set to:", config.BaseURL)
	}

	if config.ServerAddress == "" {
		if err := os.Setenv("SERVER_ADDRESS", config.ServerAddress); err != nil {
			fmt.Println("Error setting environment variable:", err)
			return err
		}
		fmt.Println("Environment variable SERVER_ADDRESS from config set to:", config.ServerAddress)
	}

	return nil
}

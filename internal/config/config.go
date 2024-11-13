package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"shortener/internal/models"
)

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

		return nil
	}

	if os.Getenv("SERVER_ADDRESS") == "" {
		os.Setenv("SERVER_ADDRESS", config.ServerAddress)
	}
	if os.Getenv("BASE_URL") == "" {
		os.Setenv("BASE_URL", config.BaseURL)
	}
	if os.Getenv("FILE_STORAGE_PATH") == "" {
		os.Setenv("FILE_STORAGE_PATH", config.FileStoragePath)
	}

	aFlag := flag.String("a", "", "Value for the -a flag")
	bFlag := flag.String("b", "", "Value for the -b flag")
	fFlag := flag.String("f", "", "Value for the -f flag")

	flag.Parse()

	if *aFlag != "" {
		err := os.Setenv("SERVER_ADDRESS", *aFlag)
		if err != nil {
			fmt.Println("Error setting environment variable:", err)
			return err
		}
		fmt.Println("Environment variable SERVER_ADDRESS set to:", *aFlag)
	} else {
		fmt.Println("No -a flag provided")
	}

	if *bFlag != "" {
		err := os.Setenv("BASE_URL", *bFlag)
		if err != nil {
			fmt.Println("Error setting environment variable:", err)
			return err
		}
		fmt.Println("Environment variable BASE_URL set to:", *bFlag)
	} else {
		fmt.Println("No -b flag provided")
	}

	if *fFlag != "" {
		err := os.Setenv("FILE_STORAGE_PATH", *fFlag)
		if err != nil {
			fmt.Println("Error setting environment variable:", err)
			return err
		}
		fmt.Println("Environment variable FILE_STORAGE_PATH set to:", *fFlag)
	} else {
		fmt.Println("No -f flag provided")
	}

	return nil
}

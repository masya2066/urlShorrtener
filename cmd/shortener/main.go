package main

import (
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"shortener/internal/config"
	"shortener/internal/db"
	"shortener/internal/routes"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	cfg, errLoad := config.LoadConfig("config.json")
	if errLoad != nil {
		panic(errLoad)
	}

	fmt.Println("STARTED PATH: " + os.Getenv("FILE_STORAGE_PATH"))

	if os.Getenv("SERVER_ADDRESS") == "" {
		os.Setenv("SERVER_ADDRESS", cfg.ServerAddress)
	}
	if os.Getenv("BASE_URL") == "" {
		os.Setenv("BASE_URL", cfg.BaseURL)
	}
	if os.Getenv("FILE_STORAGE_PATH") == "" {
		os.Setenv("FILE_STORAGE_PATH", cfg.FileStoragePath)
	}

	fmt.Println("After PATH: " + os.Getenv("FILE_STORAGE_PATH"))

	aFlag := flag.String("a", "", "Value for the -a flag")
	bFlag := flag.String("b", "", "Value for the -b flag")
	fFlag := flag.String("f", "", "Value for the -f flag")

	flag.Parse()

	if *aFlag != "" {
		err := os.Setenv("SERVER_ADDRESS", *aFlag)
		if err != nil {
			fmt.Println("Error setting environment variable:", err)
			return
		}
		fmt.Println("Environment variable SERVER_ADDRESS set to:", *aFlag)
	} else {
		fmt.Println("No -a flag provided")
	}

	if *bFlag != "" {
		err := os.Setenv("BASE_URL", *bFlag)
		if err != nil {
			fmt.Println("Error setting environment variable:", err)
			return
		}
		fmt.Println("Environment variable BASE_URL set to:", *bFlag)
	} else {
		fmt.Println("No -b flag provided")
	}

	if *fFlag != "" {
		err := os.Setenv("FILE_STORAGE_PATH", *fFlag)
		if err != nil {
			fmt.Println("Error setting environment variable:", err)
			return
		}
		fmt.Println("Environment variable FILE_STORAGE_PATH set to:", *fFlag)
	} else {
		fmt.Println("No -f flag provided")
	}

	fmt.Println("FINISH PATH: " + os.Getenv("`FILE_STORAGE_PATH`"))

	if err := db.InitStorage(); err != nil {
		panic(err)
	}
	if err := db.Init(); err != nil {
		panic(err)
	}

	if err := routes.Init(); err != nil {
		panic(err)
	}
}

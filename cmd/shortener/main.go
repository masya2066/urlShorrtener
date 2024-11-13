package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"shortener/internal/config"
	"shortener/internal/db"
	"shortener/internal/routes"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	errLoad := config.LoadConfig("config.json")

	if errLoad != nil {
		panic(errLoad)
	}

	storagePath := os.Getenv("FILE_STORAGE_PATH")
	fileStorage := db.NewFileStorage(storagePath)

	if err := fileStorage.InitStorage(); err != nil {
		panic(err)
	}
	if err := db.Init(); err != nil {
		panic(err)
	}

	if err := routes.Init(); err != nil {
		panic(err)
	}
}

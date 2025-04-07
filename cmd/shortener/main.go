package main

import (
	"log"
	"shortener/internal/config"
	"shortener/internal/db"
	"shortener/internal/routes"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	conf, errLoad := config.LoadConfig("config.json")

	if errLoad != nil {
		log.Fatalf("Error loading config: %v", errLoad)
	}

	fileStorage := db.NewFileStorage(conf.FileStoragePath)

	if err := fileStorage.InitStorage(); err != nil {
		log.Println("Error init storage", err)
	}
	if err := db.InitPostgres(conf); err != nil {
		log.Println("Error init postgres", err)
	}

	if err := db.InitSQLite(); err != nil {
		log.Println("Error init sqlite", err)
	}

	if err := routes.New(conf); err != nil {
		panic(err)
	}
}

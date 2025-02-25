package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"

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
		panic(errLoad)
	}

	fmt.Println(os.Getenv("FILE_STORAGE_PATH"))
	fmt.Println(conf)
	fileStorage := db.NewFileStorage(conf.FileStoragePath)

	if err := fileStorage.InitStorage(); err != nil {
		slog.Default().Error("Error init storage", err)
	}
	if err := db.InitPostgres(); err != nil {
		slog.Default().Error("Error init postgres", err)
	}

	if err := db.InitSQLite(); err != nil {
		slog.Default().Error("Error init sqlite", err)
	}

	if err := routes.Init(); err != nil {
		panic(err)
	}
}

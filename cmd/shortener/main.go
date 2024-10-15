package main

import (
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

	if os.Getenv("SERVER_ADDRESS") == "" {
		os.Setenv("SERVER_ADDRESS", cfg.ServerAddress)
	}
	if os.Getenv("BASE_URL") == "" {
		os.Setenv("BASE_URL", cfg.BaseURL)
	}
	
	if err := db.Init(); err != nil {
		panic(err)
	}

	if err := routes.Init(); err != nil {
		panic(err)
	}
}

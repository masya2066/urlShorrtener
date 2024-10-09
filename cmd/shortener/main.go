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

	if os.Getenv("SERVER_ADDRESS") == "" {
		os.Setenv("SERVER_ADDRESS", cfg.ServerAddress)
	}
	if os.Getenv("BASE_URL") == "" {
		os.Setenv("BASE_URL", cfg.BaseUrl)
	}

	aFlag := flag.String("a", "", "Value for the -a flag")
	bFlag := flag.String("b", "", "Value for the -b flag")

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

	if err := db.Init(); err != nil {
		panic(err)
	}

	if err := routes.Init(); err != nil {
		panic(err)
	}
}

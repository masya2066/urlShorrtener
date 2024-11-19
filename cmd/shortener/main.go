package main

import (
	"flag"
	"fmt"
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

	aFlag := flag.String("a", "", "Value for the -a flag")
	bFlag := flag.String("b", "", "Value for the -b flag")
	fFlag := flag.String("f", "", "Value for the -f flag")
	dFlag := flag.String("d", "", "Value for the -d flag")

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

	if *dFlag != "" {
		err := os.Setenv("DATABASE_DSN", *dFlag)
		if err != nil {
			fmt.Println("Error setting environment variable:", err)
			return
		}
		fmt.Println("Environment variable DATABASE_DSN set to:", *dFlag)
	} else {
		fmt.Println("No -d flag provided")
	}

	storagePath := os.Getenv("FILE_STORAGE_PATH")
	fileStorage := db.NewFileStorage(storagePath)

	if err := fileStorage.InitStorage(); err != nil {
		panic(err)
	}
	if err := db.Init(); err != nil {
		fmt.Println(err)
	}

	if err := routes.Init(); err != nil {
		panic(err)
	}
}

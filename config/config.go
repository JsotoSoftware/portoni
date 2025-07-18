package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func Get(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}

	return fallback
}

func Load() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, falling back to system env")
	}
}

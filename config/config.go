package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}
}

func GetEnv(key string) string {

	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return ""

}

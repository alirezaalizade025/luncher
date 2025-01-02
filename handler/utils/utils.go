package utils

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func Getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func LoadENV() {
	err := godotenv.Load()

	if err != nil {
		if os.IsNotExist(err) {

			log.Println(".env file does not exist, using default values")
		} else {
			log.Fatalln("Error loading .env file:", err)
		}
	}
}

func GetWeekNumber() int {
	t := time.Now()
	_, week := t.ISOWeek()

	return week
}
package configs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	REDIS_SERVER_URL       string
	REDIS_PASSWORD         string
	REDIS_DB               string
	REDIS_PROTOCOL_VERSION string
}

func NewConfig() *Config {
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) {
		log.Println("Error loading .env file:", err)
	}

	return &Config{
		REDIS_SERVER_URL: os.Getenv("REDIS_SERVER_URL"),
		REDIS_PASSWORD:   os.Getenv("REDIS_PASSWORD"),
	}
}

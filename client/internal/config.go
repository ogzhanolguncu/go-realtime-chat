package internal

import (
	"os"
)

type Config struct {
	Port string
}

func NewConfig() Config {
	return Config{
		Port: getEnvOrDefault("CHAT_PORT", "7007"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	port := os.Getenv(key)
	if port == "" {
		port = defaultValue
	}
	return port
}

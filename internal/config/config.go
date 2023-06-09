package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
)

// Get func to get env value
func Get(key string) string {
	// load .env file
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Print("Error loading .env file")
	}
	return os.Getenv(key)
}

package config

import (
	"github.com/jamesdube/ussd/internal/utils"
	"github.com/joho/godotenv"
	"os"
)

// Get func to get env value
func Get(key string) string {
	// load .env file
	err := godotenv.Load(".env")
	if err != nil {
		utils.Logger.Error("Error loading .env file")
	}
	return os.Getenv(key)
}

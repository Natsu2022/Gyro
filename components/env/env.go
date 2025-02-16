package env

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// TODO: Load environment variables
func LoadEnv() (bool, error) {
	env := os.Getenv("GO_ENV") // Set GO_ENV to "development" or "production"

	env = strings.ToLower(env)

	if env == "" {
		env = "dev" // Default to development
	} else if env != "dev" && env != "prod" {
		log.Fatalf("Invalid environment: %s", env)
		return false, nil
	}

	err := godotenv.Load(".env." + env)
	if err != nil {
		log.Fatalf("Error loading .env file for %s environment: %v", env, err)
		return false, err
	}
	return true, nil
}

// Get environment variable
func GetEnv(key string) string {
	return os.Getenv(key)
}

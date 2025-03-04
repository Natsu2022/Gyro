package env

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// LoadEnv loads the environment variables from the appropriate .env file
func LoadEnv() error {
	env := os.Getenv("GO_ENV")
	var envFile string

	switch env {
	case "Prod":
		envFile = ".env.prod"
	case "Dev":
		envFile = ".env.dev"
	default:
		envFile = ".env.dev"
	}

	err := godotenv.Load(envFile)
	if err != nil {
		log.Printf("Error loading .env file for %s environment: %v", env, err)
		return err
	}

	return nil
}

// GetEnv retrieves the value of the environment variable named by the key.
func GetEnv(key string) string {
	return os.Getenv(key)
}

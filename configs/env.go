package configs

import (
	"os"

	"github.com/joho/godotenv"
)

func RetrieveEnv(envName string) string {
	godotenv.Load()
	return os.Getenv(envName)
}

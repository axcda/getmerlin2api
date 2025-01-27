package utils

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// GetEnvOrDefault 从环境变量中读取值,如果不存在则返回默认值
func GetEnvOrDefault(key, defaultVal string) string {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		return defaultVal
	}
	return val
}

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: .env file not found, using system environment variables")
	}
}

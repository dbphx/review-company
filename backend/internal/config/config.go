package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	DBHost         string
	DBUser         string
	DBPassword     string
	DBName         string
	DBPort         string
	ESUrl          string
	RedisAddr      string
	RedisPassword  string
	RedisDB        string
	AdminJWTSecret string
	AdminEmails    string
	GoogleClient   string
	GoogleSecret   string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, reading from environment variables")
	}

	return &Config{
		Port:           getEnv("PORT", "3000"),
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBUser:         getEnv("DB_USER", "postgres"),
		DBPassword:     getEnv("DB_PASSWORD", "secret"),
		DBName:         getEnv("DB_NAME", "review_db"),
		DBPort:         getEnv("DB_PORT", "5432"),
		ESUrl:          getEnv("ES_URL", "http://localhost:9200"),
		RedisAddr:      getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:  getEnv("REDIS_PASSWORD", ""),
		RedisDB:        getEnv("REDIS_DB", "0"),
		AdminJWTSecret: getEnv("ADMIN_JWT_SECRET", "super-secret-admin-key"),
		AdminEmails:    getEnv("ADMIN_EMAILS", ""),
		GoogleClient:   getEnv("GOOGLE_OAUTH_CLIENT_ID", ""),
		GoogleSecret:   getEnv("GOOGLE_OAUTH_SECRET", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

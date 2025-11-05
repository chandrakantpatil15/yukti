package config

import "os"

type Config struct {
	DatabaseURL string
	AWSRegion   string
	Port        string
}

func Load() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://yukti:yukti123@localhost:5432/yukti_finops?sslmode=disable"),
		AWSRegion:   getEnv("AWS_REGION", "us-east-1"),
		Port:        getEnv("PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
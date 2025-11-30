package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL       string
	AWSRegion         string
	Port              string
	JWTSecret         string
	CORSAllowedOrigins []string
	Environment       string // development, staging, production
}

func Load() *Config {
	// Load .env file if exists (for local development)
	if err := godotenv.Load(); err != nil {
		log.Printf("[INFO] No .env file found, using environment variables")
	}
	
	env := getEnv("ENVIRONMENT", "development")
	
	cfg := &Config{
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://yukti:yukti123@localhost:5432/yukti_finops?sslmode=disable"),
		AWSRegion:         getEnv("AWS_REGION", "us-east-1"),
		Port:              os.Getenv("PORT"), // Must be set in .env.ports
		JWTSecret:         os.Getenv("JWT_SECRET"),
		CORSAllowedOrigins: parseCORSOrigins(os.Getenv("CORS_ALLOWED_ORIGINS")),
		Environment:       env,
	}
	
	// Validate critical configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("[FATAL] Configuration validation failed: %v", err)
	}
	
	return cfg
}

func (c *Config) Validate() error {
	// JWT_SECRET is required in production
	if c.Environment == "production" && c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET must be set in production environment")
	}
	
	// Warn if using default JWT secret in non-production
	if c.JWTSecret == "" {
		c.JWTSecret = "yukti-secret-key-change-in-production"
		log.Printf("[WARN] JWT_SECRET not set, using default (INSECURE - change in production!)")
	}
	
	// Default CORS origins if not set
	if len(c.CORSAllowedOrigins) == 0 {
		if c.Environment == "production" {
			return fmt.Errorf("CORS_ALLOWED_ORIGINS must be set in production environment")
		}
		return fmt.Errorf("CORS_ALLOWED_ORIGINS must be set (use .env.ports to configure)")
		// c.CORSAllowedOrigins = []string{"http://localhost:3000"}
	}
	
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseCORSOrigins(origins string) []string {
	if origins == "" {
		return nil
	}
	parts := strings.Split(origins, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
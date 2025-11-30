package config

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"os"
)

// Secrets holds all application secrets
type Secrets struct {
	JWTSecret string
}

var secrets *Secrets

// LoadSecrets loads secrets from environment variables
// Call this ONCE at application startup
func LoadSecrets() *Secrets {
	if secrets != nil {
		return secrets
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Println("[WARN] JWT_SECRET not set, generating random secret (NOT for production)")
		jwtSecret = generateRandomSecret()
	}

	secrets = &Secrets{
		JWTSecret: jwtSecret,
	}

	log.Printf("[INFO] Secrets loaded successfully (JWT secret length: %d)", len(secrets.JWTSecret))
	return secrets
}

// GetSecrets returns the loaded secrets
func GetSecrets() *Secrets {
	if secrets == nil {
		log.Fatal("[FATAL] Secrets not loaded. Call LoadSecrets() at startup")
	}
	return secrets
}

// generateRandomSecret generates a cryptographically secure random secret
func generateRandomSecret() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		log.Fatal("[FATAL] Failed to generate random secret:", err)
	}
	return base64.StdEncoding.EncodeToString(b)
}
